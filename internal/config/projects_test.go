package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wilsonpilon/goBaDig/internal/options"
)

func TestCreateProjectThenListProjectsRoundTrips(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "settings.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	projectDir := filepath.Join(t.TempDir(), "MeuJogo")
	if err := store.CreateProject("MeuJogo", projectDir, options.CategoryZXSpectrum); err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	if info, err := os.Stat(projectDir); err != nil || !info.IsDir() {
		t.Fatalf("expected %s to have been created as a directory: %v", projectDir, err)
	}

	projects, err := store.ListProjects()
	if err != nil {
		t.Fatalf("ListProjects: %v", err)
	}
	if len(projects) != 1 {
		t.Fatalf("ListProjects() = %d entries, want 1", len(projects))
	}
	got := projects[0]
	if got.Name != "MeuJogo" || got.Path != projectDir || got.Dialect != options.CategoryZXSpectrum {
		t.Fatalf("ListProjects()[0] = %+v, want Name=MeuJogo Path=%s Dialect=%s", got, projectDir, options.CategoryZXSpectrum)
	}
	if got.CreatedAt.IsZero() {
		t.Fatal("ListProjects()[0].CreatedAt is zero, want a real timestamp")
	}
}

func TestCreateProjectAcceptsAnAlreadyExistingDirectory(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "settings.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	projectDir := t.TempDir() // already exists
	if err := store.CreateProject("Existing", projectDir, options.CategoryMSX); err != nil {
		t.Fatalf("CreateProject on an existing directory: %v", err)
	}
}

func TestCreateProjectRejectsAnUnknownDialect(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "settings.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	if err := store.CreateProject("Nope", t.TempDir(), "atari"); err == nil {
		t.Fatal("CreateProject with an unknown dialect: want error, got nil")
	}
}

func TestCreateProjectRejectsAnEmptyName(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "settings.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	if err := store.CreateProject("", t.TempDir(), options.CategoryMSX); err == nil {
		t.Fatal("CreateProject with an empty name: want error, got nil")
	}
}

func TestCreateProjectRejectsADuplicateName(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "settings.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	dir := filepath.Join(t.TempDir(), "proj")
	if err := store.CreateProject("Dup", dir, options.CategoryMSX); err != nil {
		t.Fatalf("first CreateProject: %v", err)
	}
	if err := store.CreateProject("Dup", dir, options.CategoryMSX); err == nil {
		t.Fatal("second CreateProject with the same name: want error, got nil")
	}
}

func TestListProjectsOrdersMostRecentFirst(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "settings.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	if err := store.CreateProject("First", t.TempDir(), options.CategoryMSX); err != nil {
		t.Fatal(err)
	}
	if err := store.CreateProject("Second", t.TempDir(), options.CategoryZX81); err != nil {
		t.Fatal(err)
	}

	projects, err := store.ListProjects()
	if err != nil {
		t.Fatal(err)
	}
	if len(projects) != 2 || projects[0].Name != "Second" || projects[1].Name != "First" {
		t.Fatalf("ListProjects() = %+v, want [Second, First]", projects)
	}
}
