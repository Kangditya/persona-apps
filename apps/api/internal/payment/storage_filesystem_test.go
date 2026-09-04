package payment

import (
    "bytes"
    "context"
    "crypto/sha256"
    "errors"
    "io"
    "os"
    "path/filepath"
    "strings"
    "testing"
    "time"
)

type referenceChecker func(context.Context, string) (bool, error)

func (check referenceChecker) ReferenceExists(ctx context.Context, reference string) (bool, error) {
    return check(ctx, reference)
}

func TestFilesystemStorePrivateImmutableLifecycle(t *testing.T) {
    root := t.TempDir()
    if err := os.Chmod(root, 0o700); err != nil {
        t.Fatal(err)
    }
    store, err := NewFilesystemStore(root)
    if err != nil {
        t.Fatal(err)
    }
    defer store.Close()
    for _, directory := range []string{"staging", "objects"} {
        info, statErr := os.Stat(filepath.Join(store.root.Name(), directory))
        if statErr != nil {
            t.Fatalf("%s stat = %v", directory, statErr)
        }
        if info.Mode().Perm() != 0o700 {
            t.Fatalf("%s mode = %o", directory, info.Mode().Perm())
        }
    }
    data := []byte("%PDF-1.4\nprivate")
    digest := sha256.Sum256(data)
    reference, err := NewEvidenceReference()
    if err != nil {
        t.Fatal(err)
    }
    stored, err := store.Put(context.Background(), EvidenceObject{Reference: reference, MediaType: "application/pdf", SizeBytes: int64(len(data)), SHA256: digest[:], Body: bytes.NewReader(data)})
    if err != nil || stored.Reference != reference || stored.SizeBytes != int64(len(data)) {
        t.Fatalf("stored/error = %#v/%v", stored, err)
    }
    info, err := os.Stat(filepath.Join(store.root.Name(), "objects", reference))
    if err != nil {
        t.Fatal(err)
    }
    if info.Mode().Perm() != 0o600 {
        t.Fatalf("file mode = %o", info.Mode().Perm())
    }
    file, err := store.Open(context.Background(), reference)
    if err != nil {
        t.Fatal(err)
    }
    got, _ := io.ReadAll(file)
    _ = file.Close()
    if !bytes.Equal(got, data) {
        t.Fatalf("open = %q", got)
    }
    if _, err := store.Put(context.Background(), EvidenceObject{Reference: reference, MediaType: "application/pdf", SizeBytes: int64(len(data)), SHA256: digest[:], Body: bytes.NewReader(data)}); !errors.Is(err, ErrStorageCollision) {
        t.Fatalf("collision = %v", err)
    }
    if err := store.Delete(context.Background(), reference); err != nil {
        t.Fatal(err)
    }
    if _, err := store.Open(context.Background(), reference); !errors.Is(err, ErrEvidenceNotFound) {
        t.Fatalf("open after delete = %v", err)
    }
}

func TestFilesystemStoreLeaseAndReconciliation(t *testing.T) {
    root := t.TempDir()
    if err := os.Chmod(root, 0o700); err != nil {
        t.Fatal(err)
    }
    store, err := NewFilesystemStore(root)
    if err != nil {
        t.Fatal(err)
    }
    defer store.Close()
    if _, err := NewFilesystemStore(root); err == nil {
        t.Fatal("second lease acquired")
    }
    lease, err := os.Stat(filepath.Join(root, ".lease"))
    if err != nil || lease.Mode().Perm() != 0o600 {
        t.Fatalf("lease=%v/%v", lease, err)
    }
    data := []byte("%PDF-1.4\norphan")
    digest := sha256.Sum256(data)
    stale, _ := NewEvidenceReference()
    referenced, _ := NewEvidenceReference()
    young, _ := NewEvidenceReference()
    for _, reference := range []string{stale, referenced, young} {
        if _, err := store.Put(context.Background(), EvidenceObject{Reference: reference, MediaType: "application/pdf", SizeBytes: int64(len(data)), SHA256: digest[:], Body: bytes.NewReader(data)}); err != nil {
            t.Fatal(err)
        }
    }
    old := time.Now().Add(-evidenceOrphanGrace - time.Hour)
    for _, reference := range []string{stale, referenced} {
        if err := os.Chtimes(filepath.Join(store.root.Name(), "objects", reference), old, old); err != nil {
            t.Fatal(err)
        }
    }
    if err := store.Reconcile(context.Background(), referenceChecker(func(_ context.Context, reference string) (bool, error) { return reference == referenced, nil })); err != nil {
        t.Fatal(err)
    }
    if _, err := store.Open(context.Background(), stale); !errors.Is(err, ErrEvidenceNotFound) {
        t.Fatalf("stale open = %v", err)
    }
    if file, err := store.Open(context.Background(), referenced); err != nil {
        t.Fatalf("referenced was removed: %v", err)
    } else {
        _ = file.Close()
    }
    if file, err := store.Open(context.Background(), young); err != nil {
        t.Fatalf("young was removed: %v", err)
    } else {
        _ = file.Close()
    }
    if err := store.Reconcile(context.Background(), referenceChecker(func(context.Context, string) (bool, error) { return false, errors.New("database unavailable") })); err == nil {
        t.Fatal("query failure accepted")
    }
}

func TestFilesystemStoreReaderAndQueryAbortSafety(t *testing.T) {
    root := t.TempDir()
    _ = os.Chmod(root, 0o700)
    store, err := NewFilesystemStore(root)
    if err != nil {
        t.Fatal(err)
    }
    defer store.Close()
    data := []byte("%PDF-1.4\nx")
    d := sha256.Sum256(data)
    put := func() string {
        k, _ := NewEvidenceReference()
        if _, e := store.Put(context.Background(), EvidenceObject{Reference: k, MediaType: "application/pdf", SizeBytes: int64(len(data)), SHA256: d[:], Body: bytes.NewReader(data)}); e != nil {
            t.Fatal(e)
        }
        _ = os.Chtimes(filepath.Join(root, "objects", k), time.Now().Add(-25*time.Hour), time.Now().Add(-25*time.Hour))
        return k
    }
    first := put()
    later := put()
    if first > later {
        first, later = later, first
    }
    reader, err := store.Open(context.Background(), first)
    if err != nil {
        t.Fatal(err)
    }
    done := make(chan error, 1)
    go func() {
        done <- store.Reconcile(context.Background(), referenceChecker(func(_ context.Context, key string) (bool, error) {
            if key == first {
                return false, errors.New("opaque sentinel")
            }
            return false, nil
        }))
    }()
    select {
    case <-done:
        t.Fatal("reconcile completed while reader open")
    case <-time.After(20 * time.Millisecond):
    }
    _ = reader.Close()
    err = <-done
    if err == nil || strings.Contains(err.Error(), first) || strings.Contains(err.Error(), root) {
        t.Fatalf("query err=%v", err)
    }
    if laterReader, err := store.Open(context.Background(), later); err != nil {
        t.Fatalf("later removed=%v", err)
    } else {
        _ = laterReader.Close()
    }
    symlink, _ := NewEvidenceReference()
    if err := store.root.Symlink("../staging/missing", "objects/"+symlink); err == nil {
        if _, e := store.Open(context.Background(), symlink); !errors.Is(e, ErrEvidenceNotFound) {
            t.Fatalf("symlink open=%v", e)
        }
        if e := store.Delete(context.Background(), symlink); !errors.Is(e, ErrEvidenceNotFound) {
            t.Fatalf("symlink delete=%v", e)
        }
    }
}

func TestFilesystemStoreRejectsUnsafeRootsAndKeys(t *testing.T) {
    if _, err := NewFilesystemStore("relative"); err == nil {
        t.Fatal("relative root accepted")
    }
    if _, err := NewFilesystemStore("/"); err == nil {
        t.Fatal("root accepted")
    }
    if _, err := NewFilesystemStore(filepath.Join(t.TempDir(), "missing")); err == nil {
        t.Fatal("missing root accepted")
    }
    world := t.TempDir()
    if err := os.Chmod(world, 0o755); err != nil {
        t.Fatal(err)
    }
    if _, err := NewFilesystemStore(world); err == nil {
        t.Fatal("world-readable root accepted")
    }
    git := t.TempDir()
    if err := os.Chmod(git, 0o700); err != nil {
        t.Fatal(err)
    }
    if err := os.Mkdir(filepath.Join(git, ".git"), 0o700); err != nil {
        t.Fatal(err)
    }
    if _, err := NewFilesystemStore(git); err == nil {
        t.Fatal("repository root accepted")
    }
    target := t.TempDir()
    link := filepath.Join(t.TempDir(), "link")
    if err := os.Symlink(target, link); err == nil {
        if _, openErr := NewFilesystemStore(link); openErr == nil {
            t.Fatal("symlink root accepted")
        }
    }
    root := t.TempDir()
    if err := os.Chmod(root, 0o700); err != nil {
        t.Fatal(err)
    }
    store, err := NewFilesystemStore(root)
    if err != nil {
        t.Fatal(err)
    }
    defer store.Close()
    for _, key := range []string{"../x", "objects/x", "", strings.Repeat("a", 43)} {
        if _, err := store.Open(context.Background(), key); !errors.Is(err, ErrEvidenceNotFound) {
            t.Fatalf("key %q error = %v", key, err)
        }
    }
}

func TestFilesystemStoreRejectsStreamMismatchAndCancellation(t *testing.T) {
    root := t.TempDir()
    if err := os.Chmod(root, 0o700); err != nil {
        t.Fatal(err)
    }
    store, err := NewFilesystemStore(root)
    if err != nil {
        t.Fatal(err)
    }
    defer store.Close()
    data := []byte("%PDF-1.4\nproof")
    digest := sha256.Sum256(data)
    newObject := func(body io.Reader) EvidenceObject {
        key, _ := NewEvidenceReference()
        return EvidenceObject{Reference: key, MediaType: "application/pdf", SizeBytes: int64(len(data)), SHA256: digest[:], Body: body}
    }
    if _, err := store.Put(context.Background(), newObject(bytes.NewReader(data[:3]))); !errors.Is(err, ErrInvalidInput) {
        t.Fatalf("short read = %v", err)
    }
    if _, err := store.Put(context.Background(), newObject(bytes.NewReader([]byte("not a pdf")))); !errors.Is(err, ErrInvalidInput) {
        t.Fatalf("mime/digest mismatch = %v", err)
    }
    wrong := newObject(bytes.NewReader(data))
    wrong.SHA256 = make([]byte, sha256.Size)
    if _, err := store.Put(context.Background(), wrong); !errors.Is(err, ErrInvalidInput) {
        t.Fatalf("digest mismatch = %v", err)
    }
    tooLarge := append([]byte("%PDF-1.4\n"), make([]byte, MaxEvidenceBytes)...)
    largeDigest := sha256.Sum256(tooLarge)
    key, _ := NewEvidenceReference()
    if _, err := store.Put(context.Background(), EvidenceObject{Reference: key, MediaType: "application/pdf", SizeBytes: MaxEvidenceBytes, SHA256: largeDigest[:], Body: bytes.NewReader(tooLarge)}); !errors.Is(err, ErrEvidenceTooLarge) {
        t.Fatalf("limit+1 = %v", err)
    }
    ctx, cancel := context.WithCancel(context.Background())
    cancel()
    if _, err := store.Put(ctx, newObject(bytes.NewReader(data))); !errors.Is(err, context.Canceled) {
        t.Fatalf("cancel = %v", err)
    }
}

type failingEvidenceReader struct{}

func (failingEvidenceReader) Read([]byte) (int, error) { return 0, errors.New("source failed") }

func TestFilesystemStoreExactLimitStagingAndSafeErrors(t *testing.T) {
    root := t.TempDir()
    if err := os.Chmod(root, 0o700); err != nil {
        t.Fatal(err)
    }
    store, err := NewFilesystemStore(root)
    if err != nil {
        t.Fatal(err)
    }
    defer store.Close()
    data := append([]byte("%PDF-1.4\n"), make([]byte, MaxEvidenceBytes-int64(len("%PDF-1.4\n")))...)
    digest := sha256.Sum256(data)
    key, _ := NewEvidenceReference()
    if _, err := store.Put(context.Background(), EvidenceObject{Reference: key, MediaType: "application/pdf", SizeBytes: MaxEvidenceBytes, SHA256: digest[:], Body: bytes.NewReader(data)}); err != nil {
        t.Fatalf("exact limit=%v", err)
    }
    failKey, _ := NewEvidenceReference()
    if _, err := store.Put(context.Background(), EvidenceObject{Reference: failKey, MediaType: "application/pdf", SizeBytes: 1, SHA256: make([]byte, 32), Body: failingEvidenceReader{}}); err == nil || strings.Contains(err.Error(), root) || strings.Contains(err.Error(), failKey) {
        t.Fatalf("unsafe source error=%v", err)
    }
    staging := "staging/stale"
    file, err := store.root.OpenFile(staging, os.O_CREATE|os.O_WRONLY, 0o600)
    if err != nil {
        t.Fatal(err)
    }
    _ = file.Close()
    old := time.Now().Add(-25 * time.Hour)
    if err := store.root.Chtimes(staging, old, old); err != nil {
        t.Fatal(err)
    }
    if err := store.Reconcile(context.Background(), referenceChecker(func(context.Context, string) (bool, error) { return false, nil })); err != nil {
        t.Fatal(err)
    }
    if _, err := store.root.Lstat(staging); !errors.Is(err, os.ErrNotExist) {
        t.Fatalf("staging=%v", err)
    }
    if err := store.Delete(context.Background(), "../"+key); err == nil || strings.Contains(err.Error(), key) || strings.Contains(err.Error(), root) {
        t.Fatalf("unsafe delete=%v", err)
    }
}
