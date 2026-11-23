package dsk

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestCreateProducesACorrectlySizedFormattedImage(t *testing.T) {
	path := filepath.Join(t.TempDir(), "new.dsk")
	img, err := Create(path, nil)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := img.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() != 720*1024 {
		t.Fatalf("image size = %d, want %d (720KB)", info.Size(), 720*1024)
	}

	reopened, err := Open(path)
	if err != nil {
		t.Fatalf("Open a freshly created image: %v", err)
	}
	defer reopened.Close()
	files, err := reopened.ListFiles()
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 0 {
		t.Fatalf("ListFiles() on a freshly formatted image = %v, want empty", files)
	}
}

func TestCreateRejectsAWrongSizedCustomBootSector(t *testing.T) {
	path := filepath.Join(t.TempDir(), "new.dsk")
	if _, err := Create(path, []byte{1, 2, 3}); err == nil {
		t.Fatalf("Create with a 3-byte boot sector: want error, got nil")
	}
}

func TestAddFileThenListFilesShowsIt(t *testing.T) {
	dir := t.TempDir()
	diskPath := filepath.Join(dir, "disk.dsk")
	img, err := Create(diskPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer img.Close()

	localPath := filepath.Join(dir, "HELLO.BAS")
	content := []byte("10 PRINT \"HELLO\"\n")
	if err := os.WriteFile(localPath, content, 0o644); err != nil {
		t.Fatal(err)
	}

	if err := img.AddFile(localPath, ""); err != nil {
		t.Fatalf("AddFile: %v", err)
	}

	files, err := img.ListFiles()
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 {
		t.Fatalf("ListFiles() = %v, want exactly 1 file", files)
	}
	if files[0].Name != "hello.bas" {
		t.Fatalf("file name = %q, want %q (lower-cased)", files[0].Name, "hello.bas")
	}
	if files[0].Size != uint32(len(content)) {
		t.Fatalf("file size = %d, want %d", files[0].Size, len(content))
	}
	if files[0].FirstCluster < 2 {
		t.Fatalf("first cluster = %d, want >= 2 (a real allocated cluster)", files[0].FirstCluster)
	}
}

func TestAddFileWithTheSameNameOverwritesRatherThanDuplicates(t *testing.T) {
	dir := t.TempDir()
	diskPath := filepath.Join(dir, "disk.dsk")
	img, err := Create(diskPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer img.Close()

	localPath := filepath.Join(dir, "a.txt")
	if err := os.WriteFile(localPath, []byte("first"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := img.AddFile(localPath, "a.txt"); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(localPath, []byte("second version, longer"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := img.AddFile(localPath, "a.txt"); err != nil {
		t.Fatal(err)
	}

	files, err := img.ListFiles()
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 {
		t.Fatalf("ListFiles() after overwriting = %v, want exactly 1 entry, not a duplicate", files)
	}
	if files[0].Size != uint32(len("second version, longer")) {
		t.Fatalf("size after overwrite = %d, want the SECOND version's size", files[0].Size)
	}
}

func TestExtractFileRoundTripsTheExactBytes(t *testing.T) {
	dir := t.TempDir()
	diskPath := filepath.Join(dir, "disk.dsk")
	img, err := Create(diskPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer img.Close()

	// Larger than one cluster (1024 bytes for a 720KB disk) to also
	// exercise the multi-cluster chain path, not just a single sector.
	original := bytes.Repeat([]byte("MSX disk image round-trip test. "), 200)
	srcPath := filepath.Join(dir, "DATA.BIN")
	if err := os.WriteFile(srcPath, original, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := img.AddFile(srcPath, ""); err != nil {
		t.Fatal(err)
	}

	destPath := filepath.Join(dir, "extracted.bin")
	if err := img.ExtractFile("data.bin", destPath); err != nil {
		t.Fatalf("ExtractFile: %v", err)
	}
	got, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, original) {
		t.Fatalf("extracted %d bytes, want %d identical bytes (round-trip mismatch)", len(got), len(original))
	}
}

func TestExtractFileUnknownNameIsAnError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "disk.dsk")
	img, err := Create(path, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer img.Close()
	if err := img.ExtractFile("nope.txt", filepath.Join(t.TempDir(), "out.txt")); err == nil {
		t.Fatal("ExtractFile(unknown name): want error, got nil")
	}
}

func TestDeleteFileFreesClustersForReuse(t *testing.T) {
	dir := t.TempDir()
	diskPath := filepath.Join(dir, "disk.dsk")
	img, err := Create(diskPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer img.Close()

	content := bytes.Repeat([]byte("x"), 3000) // several clusters
	srcPath := filepath.Join(dir, "a.bin")
	if err := os.WriteFile(srcPath, content, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := img.AddFile(srcPath, ""); err != nil {
		t.Fatal(err)
	}
	files, err := img.ListFiles()
	if err != nil {
		t.Fatal(err)
	}
	firstCluster := files[0].FirstCluster

	if err := img.DeleteFile("a.bin"); err != nil {
		t.Fatalf("DeleteFile: %v", err)
	}
	files, err = img.ListFiles()
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 0 {
		t.Fatalf("ListFiles() after delete = %v, want empty", files)
	}

	// Adding a new file must be able to reuse the clusters DeleteFile
	// just freed - confirms the FAT chain was actually zeroed, not
	// just the directory entry marked deleted.
	src2 := filepath.Join(dir, "b.bin")
	if err := os.WriteFile(src2, content, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := img.AddFile(src2, ""); err != nil {
		t.Fatal(err)
	}
	files, err = img.ListFiles()
	if err != nil {
		t.Fatal(err)
	}
	if files[0].FirstCluster != firstCluster {
		t.Fatalf("new file's first cluster = %d, want the reused %d (freed clusters should be reusable)", files[0].FirstCluster, firstCluster)
	}
}

func TestDeleteFileUnknownNameIsNotAnError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "disk.dsk")
	img, err := Create(path, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer img.Close()
	if err := img.DeleteFile("nope.txt"); err != nil {
		t.Fatalf("DeleteFile(unknown name): %v, want nil (matches the original's own \"nothing to delete, that's fine\")", err)
	}
}

func TestAddFileWithZeroBytesHasNoAllocatedCluster(t *testing.T) {
	dir := t.TempDir()
	diskPath := filepath.Join(dir, "disk.dsk")
	img, err := Create(diskPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer img.Close()

	emptyPath := filepath.Join(dir, "empty.txt")
	if err := os.WriteFile(emptyPath, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := img.AddFile(emptyPath, ""); err != nil {
		t.Fatal(err)
	}
	files, err := img.ListFiles()
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 || files[0].Size != 0 || files[0].FirstCluster != 0 {
		t.Fatalf("zero-byte file entry = %+v, want Size=0 and FirstCluster=0", files)
	}
}

func TestAddFileReturnsErrorWhenDiskIsFull(t *testing.T) {
	dir := t.TempDir()
	diskPath := filepath.Join(dir, "disk.dsk")
	img, err := Create(diskPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer img.Close()

	// A 720KB disk has far fewer usable bytes than this - guaranteed
	// to exhaust every free cluster.
	big := make([]byte, 800*1024)
	bigPath := filepath.Join(dir, "big.bin")
	if err := os.WriteFile(bigPath, big, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := img.AddFile(bigPath, ""); err == nil {
		t.Fatal("AddFile with a file bigger than the whole disk: want error, got nil")
	}

	// The failed attempt must not leave a half-written directory
	// entry or dangling allocated clusters behind.
	files, err := img.ListFiles()
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 0 {
		t.Fatalf("ListFiles() after a failed AddFile = %v, want empty (no partial entry left behind)", files)
	}
}

func TestConvertToFAT11PadsUppercasesAndSplitsExtension(t *testing.T) {
	cases := []struct{ in, want string }{
		{"file.txt", "FILE    TXT"},
		{"a.b", "A       B  "},
		{"noext", "NOEXT      "},
		{"*.bas", "????????BAS"},
		{"auto*.*", "AUTO???????"},
	}
	for _, c := range cases {
		if got := ConvertToFAT11(c.in); got != c.want {
			t.Errorf("ConvertToFAT11(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestMatchesFAT11Wildcards(t *testing.T) {
	entry := ConvertToFAT11("hello.bas")
	if !MatchesFAT11(entry, ConvertToFAT11("*.bas")) {
		t.Errorf("MatchesFAT11(hello.bas, *.bas) = false, want true")
	}
	if !MatchesFAT11(entry, ConvertToFAT11("*.*")) {
		t.Errorf("MatchesFAT11(hello.bas, *.*) = false, want true")
	}
	if MatchesFAT11(entry, ConvertToFAT11("*.txt")) {
		t.Errorf("MatchesFAT11(hello.bas, *.txt) = true, want false")
	}
}

func TestOpenRejectsAWrongSizedFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "notadisk.dsk")
	if err := os.WriteFile(path, make([]byte, 1000), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Open(path); err == nil {
		t.Fatal("Open a file too small to be a real boot sector: want error, got nil")
	}
}

func TestMethodsAfterCloseReturnErrClosed(t *testing.T) {
	path := filepath.Join(t.TempDir(), "disk.dsk")
	img, err := Create(path, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := img.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := img.ListFiles(); err != ErrClosed {
		t.Errorf("ListFiles after Close: err = %v, want ErrClosed", err)
	}
	if err := img.DeleteFile("x"); err != ErrClosed {
		t.Errorf("DeleteFile after Close: err = %v, want ErrClosed", err)
	}
}
