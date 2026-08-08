package migration

import (
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
)

const DefaultDirectory = "migrations"

var filenamePattern = regexp.MustCompile(`^([0-9]+)_([a-z0-9][a-z0-9_-]*)\.(up|down)\.sql$`)
var migrationNamePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]*$`)

type File struct {
	Version  uint
	Name     string
	UpPath   string
	DownPath string
}

type State struct {
	Version    uint
	HasVersion bool
	Dirty      bool
}

func ResolveDirectory() string {
	if directory := os.Getenv("MIGRATIONS_DIR"); directory != "" {
		return directory
	}
	return DefaultDirectory
}

func Discover(directory string) ([]File, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, fmt.Errorf("read migration directory %q: %w", directory, err)
	}

	type discovered struct {
		version uint
		name    string
		up      string
		down    string
	}
	byVersion := make(map[uint]*discovered)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		match := filenamePattern.FindStringSubmatch(entry.Name())
		if match == nil {
			return nil, fmt.Errorf("invalid migration filename %q", entry.Name())
		}
		version, err := strconv.ParseUint(match[1], 10, 64)
		if err != nil || version == 0 || uint64(uint(version)) != version {
			return nil, fmt.Errorf("invalid migration version in %q", entry.Name())
		}

		migration := byVersion[uint(version)]
		if migration == nil {
			migration = &discovered{version: uint(version), name: match[2]}
			byVersion[uint(version)] = migration
		} else if migration.name != match[2] {
			return nil, fmt.Errorf("migration version %d has mismatched names %q and %q", version, migration.name, match[2])
		}

		path := filepath.Join(directory, entry.Name())
		contents, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read migration %q: %w", entry.Name(), err)
		}
		if strings.TrimSpace(string(contents)) == "" {
			return nil, fmt.Errorf("migration %q is empty", entry.Name())
		}

		switch match[3] {
		case "up":
			if migration.up != "" {
				return nil, fmt.Errorf("duplicate up migration version %d", version)
			}
			migration.up = path
		case "down":
			if migration.down != "" {
				return nil, fmt.Errorf("duplicate down migration version %d", version)
			}
			migration.down = path
		}
	}

	files := make([]File, 0, len(byVersion))
	for _, migration := range byVersion {
		if migration.up == "" || migration.down == "" {
			return nil, fmt.Errorf("migration %04d_%s must have both up and down files", migration.version, migration.name)
		}
		files = append(files, File{
			Version:  migration.version,
			Name:     migration.name,
			UpPath:   migration.up,
			DownPath: migration.down,
		})
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Version < files[j].Version })
	return files, nil
}

func New(db *sql.DB, directory string) (*migrate.Migrate, error) {
	absoluteDirectory, err := filepath.Abs(directory)
	if err != nil {
		return nil, fmt.Errorf("resolve migration directory: %w", err)
	}
	if _, err := Discover(absoluteDirectory); err != nil {
		return nil, err
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{MultiStatementEnabled: true})
	if err != nil {
		return nil, fmt.Errorf("configure PostgreSQL migration driver: %w", err)
	}
	source := (&url.URL{Scheme: "file", Path: absoluteDirectory}).String()
	migrator, err := migrate.NewWithDatabaseInstance(source, "persona_apps", driver)
	if err != nil {
		return nil, fmt.Errorf("open migration source: %w", err)
	}
	return migrator, nil
}

func Current(migrator *migrate.Migrate) (State, error) {
	version, dirty, err := migrator.Version()
	if errors.Is(err, migrate.ErrNilVersion) {
		return State{}, nil
	}
	if err != nil {
		return State{}, fmt.Errorf("read migration version: %w", err)
	}
	return State{Version: version, HasVersion: true, Dirty: dirty}, nil
}

func Create(directory, name string) (File, error) {
	name = strings.TrimSpace(name)
	if !migrationNamePattern.MatchString(name) {
		return File{}, errors.New("migration name must contain lowercase letters, numbers, hyphens, or underscores")
	}
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return File{}, fmt.Errorf("create migration directory: %w", err)
	}

	files, err := Discover(directory)
	if err != nil {
		return File{}, err
	}
	var next uint = 1
	if len(files) > 0 {
		next = files[len(files)-1].Version + 1
		if next == 0 {
			return File{}, errors.New("migration version overflow")
		}
	}

	prefix := fmt.Sprintf("%04d_%s", next, name)
	upPath := filepath.Join(directory, prefix+".up.sql")
	downPath := filepath.Join(directory, prefix+".down.sql")
	content := "BEGIN;\n\n-- Add migration SQL here.\n\nCOMMIT;\n"
	if err := createFile(upPath, content); err != nil {
		return File{}, fmt.Errorf("create %s: %w", filepath.Base(upPath), err)
	}
	if err := createFile(downPath, content); err != nil {
		_ = os.Remove(upPath)
		return File{}, fmt.Errorf("create %s: %w", filepath.Base(downPath), err)
	}
	return File{Version: next, Name: name, UpPath: upPath, DownPath: downPath}, nil
}

func createFile(path, content string) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	if _, err := file.WriteString(content); err != nil {
		_ = file.Close()
		_ = os.Remove(path)
		return err
	}
	return file.Close()
}
