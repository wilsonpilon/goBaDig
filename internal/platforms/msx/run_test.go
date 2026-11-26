package msx

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/wilsonpilon/goBaDig/internal/platforms/msx/dsk"
)

func TestBuildTestDiskContainsTheCompiledProgramAsAutoexec(t *testing.T) {
	dir := t.TempDir()
	dmxPath := filepath.Join(dir, "hello.dmx")
	if err := os.WriteFile(dmxPath, []byte("print 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	diskPath, err := BuildTestDisk(dmxPath)
	if err != nil {
		t.Fatalf("BuildTestDisk: %v", err)
	}
	if want := filepath.Join(dir, "hello.dsk"); diskPath != want {
		t.Fatalf("diskPath = %q, want %q", diskPath, want)
	}

	img, err := dsk.Open(diskPath)
	if err != nil {
		t.Fatalf("Open the built disk: %v", err)
	}
	defer img.Close()

	files, err := img.ListFiles()
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 || files[0].Name != "autoexec.bas" {
		t.Fatalf("disk contents = %+v, want exactly one file named autoexec.bas", files)
	}

	extracted := filepath.Join(dir, "extracted.bas")
	if err := img.ExtractFile("autoexec.bas", extracted); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(extracted)
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile(filepath.Join(dir, "hello.bmx"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("AUTOEXEC.BAS on the disk (%d bytes) differs from the compiled hello.bmx (%d bytes)", len(got), len(want))
	}
}

func TestBuildTestDiskReturnsAnErrorForAMissingSourceFile(t *testing.T) {
	if _, err := BuildTestDisk(filepath.Join(t.TempDir(), "nope.dmx")); err == nil {
		t.Fatal("BuildTestDisk(missing file): want error, got nil")
	}
}

func TestRunInEmulatorReturnsAnErrorWhenOpenmsxIsNotOnPath(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	if err := RunInEmulator("whatever.dsk"); err == nil {
		t.Fatal("RunInEmulator with an empty PATH: want error, got nil")
	}
}

func TestRunInFMSXReturnsAnErrorWhenFmsxIsNotOnPath(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	if err := RunInFMSX("whatever.dsk"); err == nil {
		t.Fatal("RunInFMSX with an empty PATH: want error, got nil")
	}
}

func TestRunInBlueMSXReturnsAnErrorWhenBlueMSXIsNotOnPath(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	if err := RunInBlueMSX("whatever.dsk"); err == nil {
		t.Fatal("RunInBlueMSX with an empty PATH: want error, got nil")
	}
}
