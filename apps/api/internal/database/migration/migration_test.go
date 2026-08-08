package migration

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverSortsAndRequiresPairs(t *testing.T) {
	directory := t.TempDir()
	writeMigration(t, directory, "0002_second")
	writeMigration(t, directory, "0001_first")

	files, err := Discover(directory)
	if err != nil {
		t.Fatalf("Discover() error = %v", err)
	}
	if len(files) != 2 || files[0].Version != 1 || files[1].Version != 2 {
		t.Fatalf("files = %#v, want ordered versions 1 and 2", files)
	}

	if err := os.Remove(files[1].DownPath); err != nil {
		t.Fatal(err)
	}
	if _, err := Discover(directory); err == nil {
		t.Fatal("Discover() accepted a migration without a down file")
	}
}

func TestCreateProducesNonCollidingPair(t *testing.T) {
	directory := t.TempDir()
	created, err := Create(directory, "add_demo_table")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.Version != 1 || filepath.Base(created.UpPath) != "0001_add_demo_table.up.sql" {
		t.Fatalf("created = %#v", created)
	}
	second, err := Create(directory, "add_demo_table")
	if err != nil {
		t.Fatalf("Create() second error = %v", err)
	}
	if second.Version != 2 {
		t.Fatalf("second version = %d, want 2", second.Version)
	}
}

func writeMigration(t *testing.T, directory, prefix string) {
	t.Helper()
	content := "BEGIN;\nCOMMIT;\n"
	for _, direction := range []string{"up", "down"} {
		path := filepath.Join(directory, prefix+"."+direction+".sql")
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}
