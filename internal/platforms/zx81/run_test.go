package zx81

import "testing"

func TestRunInEmulatorReturnsAnErrorWhenEightyoneIsNotOnPath(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	if err := RunInEmulator("whatever.p"); err == nil {
		t.Fatal("RunInEmulator with an empty PATH: want error, got nil")
	}
}

func TestRunInZEsarUXReturnsAnErrorWhenZEsarUXIsNotOnPath(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	if err := RunInZEsarUX("whatever.p"); err == nil {
		t.Fatal("RunInZEsarUX with an empty PATH: want error, got nil")
	}
}
