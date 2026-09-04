package payment

import (
    "context"
    "crypto/rand"
    "crypto/sha256"
    "encoding/base64"
    "errors"
    "fmt"
    "io"
    "net/http"
    "os"
    "path/filepath"
    "sort"
    "strings"
    "sync"
    "syscall"
    "time"
)

const evidenceOrphanGrace = 24 * time.Hour

// FilesystemStore is the single-instance private EvidenceStore selected by ADR-057.
type FilesystemStore struct {
    root    *os.Root
    lease   *os.File
    mu      sync.RWMutex
    now     func() time.Time
    staging string
    objects string
}

type EvidenceReferenceChecker interface {
    ReferenceExists(context.Context, string) (bool, error)
}

func NewFilesystemStore(root string) (*FilesystemStore, error) {
    if err := validateEvidenceRoot(root); err != nil {
        return nil, err
    }
    opened, err := os.OpenRoot(root)
    if err != nil {
        return nil, ErrStorageUnavailable
    }
    closeRoot := true
    defer func() {
        if closeRoot {
            _ = opened.Close()
        }
    }()
    for _, directory := range []string{"staging", "objects"} {
        if err := opened.MkdirAll(directory, 0o700); err != nil {
            return nil, ErrStorageUnavailable
        }
        if err := opened.Chmod(directory, 0o700); err != nil {
            return nil, ErrStorageUnavailable
        }
        info, statErr := opened.Lstat(directory)
        if statErr != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 || info.Mode().Perm() != 0o700 {
            return nil, ErrStorageUnavailable
        }
    }
    lease, err := opened.OpenFile(".lease", os.O_CREATE|os.O_RDWR, 0o600)
    if err != nil {
        return nil, ErrStorageUnavailable
    }
    leaseInfo, statErr := lease.Stat()
    if statErr != nil || leaseInfo.Mode().Perm() != 0o600 {
        _ = lease.Close()
        return nil, ErrStorageUnavailable
    }
    if err := syncRootDirectory(opened, "."); err != nil {
        _ = lease.Close()
        return nil, ErrStorageUnavailable
    }
    if err := syscall.Flock(int(lease.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
        _ = lease.Close()
        return nil, errors.New("evidence storage is already in use")
    }
    store := &FilesystemStore{root: opened, lease: lease, now: time.Now, staging: "staging", objects: "objects"}
    if err := store.probe(); err != nil {
        _ = store.Close()
        return nil, ErrStorageUnavailable
    }
    closeRoot = false
    return store, nil
}

func validateEvidenceRoot(value string) error {
    if !filepath.IsAbs(value) || filepath.Clean(value) == string(filepath.Separator) {
        return errors.New("evidence storage root must be an absolute non-root directory")
    }
    info, err := os.Lstat(value)
    if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 || info.Mode().Perm()&0o077 != 0 {
        return errors.New("evidence storage root is unsafe")
    }
    resolved, err := filepath.EvalSymlinks(value)
    if err != nil || filepath.Clean(resolved) != filepath.Clean(value) {
        return errors.New("evidence storage root must not contain symlinks")
    }
    for directory := filepath.Clean(value); ; directory = filepath.Dir(directory) {
        if _, err := os.Lstat(filepath.Join(directory, ".git")); err == nil {
            return errors.New("evidence storage root must not be in a repository")
        }
        parent := filepath.Dir(directory)
        if parent == directory {
            break
        }
    }
    return nil
}

func (store *FilesystemStore) probe() error {
    name, err := store.stagingName()
    if err != nil {
        return err
    }
    file, err := store.root.OpenFile(store.staging+"/"+name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
    if err != nil {
        return err
    }
    if _, err = file.Write([]byte{0}); err == nil {
        err = file.Sync()
    }
    if closeErr := file.Close(); err == nil {
        err = closeErr
    }
    if err != nil {
        _ = store.root.Remove(store.staging + "/" + name)
        return err
    }
    if err = store.root.Link(store.staging+"/"+name, store.objects+"/"+name); err != nil {
        _ = store.root.Remove(store.staging + "/" + name)
        return err
    }
    if err = syncRootDirectory(store.root, store.objects); err == nil {
        err = store.root.Remove(store.staging + "/" + name)
    }
    if removeErr := store.root.Remove(store.objects + "/" + name); err == nil {
        err = removeErr
    }
    return err
}

func (store *FilesystemStore) Put(ctx context.Context, object EvidenceObject) (StoredEvidence, error) {
    store.mu.Lock()
    defer store.mu.Unlock()
    if err := validEvidenceKey(object.Reference); err != nil || object.Body == nil || object.SizeBytes < 1 || object.SizeBytes > MaxEvidenceBytes || len(object.SHA256) != sha256.Size || !validMediaType(object.MediaType) {
        return StoredEvidence{}, ErrInvalidInput
    }
    if err := ctx.Err(); err != nil {
        return StoredEvidence{}, err
    }
    stagingName, err := store.stagingName()
    if err != nil {
        return StoredEvidence{}, ErrStorageUnavailable
    }
    stagingPath := store.staging + "/" + stagingName
    file, err := store.root.OpenFile(stagingPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
    if err != nil {
        return StoredEvidence{}, ErrStorageUnavailable
    }
    removeStaging := true
    defer func() {
        if removeStaging {
            _ = store.root.Remove(stagingPath)
        }
    }()
    count, digest, detected, copyErr := copyEvidence(ctx, file, object.Body)
    if syncErr := file.Sync(); copyErr == nil {
        copyErr = syncErr
    }
    if closeErr := file.Close(); copyErr == nil {
        copyErr = closeErr
    }
    if copyErr != nil || count != object.SizeBytes || digest != [sha256.Size]byte(object.SHA256) || detected != object.MediaType {
        if copyErr != nil {
            return StoredEvidence{}, copyErr
        }
        return StoredEvidence{}, ErrInvalidInput
    }
    finalPath := store.objects + "/" + object.Reference
    if err := store.root.Link(stagingPath, finalPath); err != nil {
        if errors.Is(err, os.ErrExist) {
            return StoredEvidence{}, ErrStorageCollision
        }
        return StoredEvidence{}, ErrStorageUnavailable
    }
    if err := syncRootDirectory(store.root, store.objects); err != nil {
        return StoredEvidence{}, ErrStorageUnavailable
    }
    if err := store.root.Remove(stagingPath); err != nil {
        return StoredEvidence{}, ErrStorageUnavailable
    }
    removeStaging = false
    return StoredEvidence{Reference: object.Reference, SizeBytes: count}, nil
}

func copyEvidence(ctx context.Context, destination *os.File, source io.Reader) (int64, [sha256.Size]byte, string, error) {
    hash := sha256.New()
    buffer := make([]byte, 32<<10)
    sample := make([]byte, 0, 512)
    source = io.LimitReader(source, MaxEvidenceBytes+1)
    var count int64
    for {
        if err := ctx.Err(); err != nil {
            return 0, [sha256.Size]byte{}, "", err
        }
        read, readErr := source.Read(buffer)
        if read > 0 {
            count += int64(read)
            if count > MaxEvidenceBytes {
                return 0, [sha256.Size]byte{}, "", ErrEvidenceTooLarge
            }
            if len(sample) < 512 {
                take := min(512-len(sample), read)
                sample = append(sample, buffer[:take]...)
            }
            if _, err := hash.Write(buffer[:read]); err != nil {
                return 0, [sha256.Size]byte{}, "", err
            }
            if _, err := destination.Write(buffer[:read]); err != nil {
                return 0, [sha256.Size]byte{}, "", err
            }
        }
        if errors.Is(readErr, io.EOF) {
            var result [sha256.Size]byte
            copy(result[:], hash.Sum(nil))
            return count, result, http.DetectContentType(sample), nil
        }
        if readErr != nil {
            return 0, [sha256.Size]byte{}, "", readErr
        }
    }
}

func (store *FilesystemStore) Open(ctx context.Context, reference string) (io.ReadCloser, error) {
    store.mu.RLock()
    unlock := store.mu.RUnlock
    if err := validEvidenceKey(reference); err != nil {
        unlock()
        return nil, ErrEvidenceNotFound
    }
    if err := ctx.Err(); err != nil {
        unlock()
        return nil, err
    }
    path := store.objects + "/" + reference
    info, err := store.root.Lstat(path)
    if errors.Is(err, os.ErrNotExist) {
        unlock()
        return nil, ErrEvidenceNotFound
    }
    if err != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
        unlock()
        return nil, ErrEvidenceNotFound
    }
    file, err := store.root.Open(path)
    if errors.Is(err, os.ErrNotExist) {
        unlock()
        return nil, ErrEvidenceNotFound
    }
    if err != nil {
        unlock()
        return nil, ErrStorageUnavailable
    }
    return &unlockReadCloser{ReadCloser: file, unlock: unlock}, nil
}

type unlockReadCloser struct {
    io.ReadCloser
    unlock func()
}

func (reader *unlockReadCloser) Close() error {
    err := reader.ReadCloser.Close()
    if reader.unlock != nil {
        reader.unlock()
        reader.unlock = nil
    }
    return err
}

func (store *FilesystemStore) Delete(ctx context.Context, reference string) error {
    store.mu.Lock()
    defer store.mu.Unlock()
    if err := validEvidenceKey(reference); err != nil {
        return ErrEvidenceNotFound
    }
    if err := ctx.Err(); err != nil {
        return err
    }
    path := store.objects + "/" + reference
    info, err := store.root.Lstat(path)
    if errors.Is(err, os.ErrNotExist) {
        return nil
    }
    if err != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
        return ErrEvidenceNotFound
    }
    if err := store.root.Remove(path); err != nil {
        return ErrStorageUnavailable
    }
    return syncRootDirectory(store.root, store.objects)
}

func (store *FilesystemStore) Reconcile(ctx context.Context, checker EvidenceReferenceChecker) error {
    store.mu.Lock()
    defer store.mu.Unlock()
    if checker == nil {
        return ErrInvalidInput
    }
    cutoff := store.clock().Add(-evidenceOrphanGrace)
    for _, directory := range []string{store.staging, store.objects} {
        opened, err := store.root.Open(directory)
        if err != nil {
            return ErrStorageUnavailable
        }
        entries, err := opened.ReadDir(-1)
        _ = opened.Close()
        if err != nil {
            return ErrStorageUnavailable
        }
        sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
        for _, entry := range entries {
            if err := ctx.Err(); err != nil {
                return err
            }
            if !entry.Type().IsRegular() {
                continue
            }
            info, err := entry.Info()
            if err != nil || !info.Mode().IsRegular() || !info.ModTime().Before(cutoff) {
                continue
            }
            if directory == store.objects {
                if validEvidenceKey(entry.Name()) != nil {
                    continue
                }
                referenced, queryErr := checker.ReferenceExists(ctx, entry.Name())
                if queryErr != nil {
                    if errors.Is(queryErr, context.Canceled) || errors.Is(queryErr, context.DeadlineExceeded) {
                        return queryErr
                    }
                    return ErrStorageUnavailable
                }
                if referenced {
                    continue
                }
            }
            if err := store.root.Remove(directory + "/" + entry.Name()); err != nil && !errors.Is(err, os.ErrNotExist) {
                return ErrStorageUnavailable
            }
        }
        if err := syncRootDirectory(store.root, directory); err != nil {
            return ErrStorageUnavailable
        }
    }
    return nil
}

func (store *FilesystemStore) Close() error {
    store.mu.Lock()
    defer store.mu.Unlock()
    var err error
    if store.lease != nil {
        _ = syscall.Flock(int(store.lease.Fd()), syscall.LOCK_UN)
        err = store.lease.Close()
        store.lease = nil
    }
    if store.root != nil {
        if closeErr := store.root.Close(); err == nil {
            err = closeErr
        }
        store.root = nil
    }
    return err
}

func (store *FilesystemStore) stagingName() (string, error) { return randomFlatName() }
func (store *FilesystemStore) clock() time.Time {
    if store.now == nil {
        return time.Now()
    }
    return store.now()
}
func randomFlatName() (string, error) {
    bytes := make([]byte, 16)
    if _, err := io.ReadFull(rand.Reader, bytes); err != nil {
        return "", err
    }
    return fmt.Sprintf("%x", bytes), nil
}
func validEvidenceKey(value string) error {
    if len(value) != 43 || strings.ContainsAny(value, "/\\") {
        return ErrInvalidInput
    }
    for _, character := range value {
        if !(character >= 'A' && character <= 'Z' || character >= 'a' && character <= 'z' || character >= '0' && character <= '9' || character == '-' || character == '_') {
            return ErrInvalidInput
        }
    }
    decoded, err := base64.RawURLEncoding.DecodeString(value)
    if err != nil || len(decoded) != evidenceKeyEntropyBytes {
        return ErrInvalidInput
    }
    return nil
}
func syncRootDirectory(root *os.Root, name string) error {
    directory, err := root.Open(name)
    if err != nil {
        return err
    }
    defer directory.Close()
    return directory.Sync()
}
