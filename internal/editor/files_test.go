package edit

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestReadEditorFilePreservesRawBytes(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "source.PRG")
	want := []byte{0xef, 0xbb, 0xbf, 'a', 0x00, 0xff, '\r', '\n', 0x1a}
	if err := os.WriteFile(path, want, 0600); err != nil {
		t.Fatal(err)
	}

	gotPath, text, err := readEditorFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != path || text != string(want) {
		t.Fatalf("readEditorFile() = (%q, %v bytes), want path %q and exact raw bytes", gotPath, len(text), path)
	}
	if _, _, err := readEditorFile(dir); err == nil {
		t.Fatal("reading a directory must fail")
	}
	if _, _, err := readEditorFile(" \t "); err == nil {
		t.Fatal("an empty file name must fail")
	}
}

func TestSaveEditorFileBacksUpAndDetectsExternalChanges(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "source.TXT")
	original := []byte{0xef, 0xbb, 0xbf, 'a', 0x00, 0xff, '\r', '\n', 0x1a}
	updated := []byte{'b', 0x00, 0xfe, '\n'}
	if err := os.WriteFile(path, original, 0600); err != nil {
		t.Fatal(err)
	}

	expected := string(original)
	if err := saveEditorFile(path, string(updated), '{', &expected, false); err != nil {
		t.Fatal(err)
	}
	assertFileBytes(t, path, updated)
	assertFileBytes(t, filepath.Join(dir, "source.{TX"), original)

	external := []byte("external edit")
	if err := os.WriteFile(path, external, 0600); err != nil {
		t.Fatal(err)
	}
	err := saveEditorFile(path, "must not replace external data", '{', &expected, false)
	if !errors.Is(err, errFileChanged) {
		t.Fatalf("saveEditorFile() error = %v, want errFileChanged", err)
	}
	assertFileBytes(t, path, external)
}

func TestSaveEditorFileRequiresOverwriteAndCreatesNewFiles(t *testing.T) {
	dir := t.TempDir()
	existing := filepath.Join(dir, "existing.prg")
	if err := os.WriteFile(existing, []byte("old"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := saveEditorFile(existing, "new", '{', nil, false); !errors.Is(err, errFileExists) {
		t.Fatalf("saveEditorFile() error = %v, want errFileExists", err)
	}
	assertFileBytes(t, existing, []byte("old"))

	if err := saveEditorFile(existing, "new", '{', nil, true); err != nil {
		t.Fatal(err)
	}
	assertFileBytes(t, existing, []byte("new"))
	assertFileBytes(t, filepath.Join(dir, "existing.{PR"), []byte("old"))

	created := filepath.Join(dir, "created.prg")
	if err := saveEditorFile(created, "fresh", '{', nil, false); err != nil {
		t.Fatal(err)
	}
	assertFileBytes(t, created, []byte("fresh"))
	if _, err := os.Stat(filepath.Join(dir, "created.{PR")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("new file unexpectedly has a backup: %v", err)
	}
}

func TestAtomicWriteFailureKeepsOriginalAndRemovesTemporaryFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "source.txt")
	if err := os.WriteFile(path, []byte("original"), 0600); err != nil {
		t.Fatal(err)
	}
	failure := errors.New("stop before commit")
	err := atomicWriteFile(filepath.Join(dir, "temporary.txt"), "partial", 0600, func() error {
		return failure
	})
	if !errors.Is(err, failure) {
		t.Fatalf("atomicWriteFile() error = %v, want injected failure", err)
	}
	assertFileBytes(t, path, []byte("original"))
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if len(entry.Name()) >= len(".wibedit-") && entry.Name()[:len(".wibedit-")] == ".wibedit-" {
			t.Errorf("temporary file remains after failed commit: %s", entry.Name())
		}
	}
}

func assertFileBytes(t *testing.T, path string, want []byte) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(want) {
		t.Fatalf("%s bytes = %v, want %v", path, got, want)
	}
}
