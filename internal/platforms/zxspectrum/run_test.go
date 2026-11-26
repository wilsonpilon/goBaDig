package zxspectrum

import "testing"

func TestRunInEmulatorReturnsAnErrorWhenEightyoneIsNotOnPath(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	if err := RunInEmulator("whatever.tap"); err == nil {
		t.Fatal("RunInEmulator with an empty PATH: want error, got nil")
	}
}
