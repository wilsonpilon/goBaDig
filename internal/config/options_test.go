package config

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/wilsonpilon/goBaDig/internal/options"
)

func TestEnsureOptionDefaultsSeedsOnlyOnce(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "settings.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	if err := store.EnsureOptionDefaults(options.CategoryCore, options.Catalog()); err != nil {
		t.Fatal(err)
	}
	value, err := store.GetOption(options.CategoryCore, "INITIAL_LINE")
	if err != nil {
		t.Fatal(err)
	}
	if value != "10" {
		t.Fatalf("INITIAL_LINE = %q, want the catalog default %q", value, "10")
	}

	if err := store.SetOption(options.CategoryCore, "INITIAL_LINE", "1"); err != nil {
		t.Fatal(err)
	}

	// Running EnsureOptionDefaults again must NOT revert the value the
	// "user" just set - this is the whole point of ON CONFLICT DO
	// NOTHING here, unlike InstallThemes's always-upsert.
	if err := store.EnsureOptionDefaults(options.CategoryCore, options.Catalog()); err != nil {
		t.Fatal(err)
	}
	value, err = store.GetOption(options.CategoryCore, "INITIAL_LINE")
	if err != nil {
		t.Fatal(err)
	}
	if value != "1" {
		t.Fatalf("INITIAL_LINE after re-running EnsureOptionDefaults = %q, want it to stay %q", value, "1")
	}
}

func TestSetOptionRejectsInvalidValueAndLeavesStoreUnchanged(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "settings.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	if err := store.EnsureOptionDefaults(options.CategoryCore, options.Catalog()); err != nil {
		t.Fatal(err)
	}
	if err := store.SetOption(options.CategoryCore, "REM_HEADER", "maybe"); err == nil {
		t.Fatal("SetOption(REM_HEADER, maybe): want error, got nil")
	}
	value, err := store.GetOption(options.CategoryCore, "REM_HEADER")
	if err != nil {
		t.Fatal(err)
	}
	if value != "true" {
		t.Fatalf("REM_HEADER after a rejected SetOption = %q, want the untouched default %q", value, "true")
	}
}

func TestSetOptionRejectsUnknownOption(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "settings.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	if err := store.SetOption(options.CategoryCore, "NOT_A_REAL_OPTION", "1"); err == nil {
		t.Fatal("SetOption(NOT_A_REAL_OPTION, 1): want error, got nil")
	}
}

func TestSetOptionNormalizesKeyCaseToTheCatalogSpelling(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "settings.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	if err := store.SetOption(options.CategoryCore, "initial_line", "7"); err != nil {
		t.Fatal(err)
	}
	value, err := store.GetOption(options.CategoryCore, "INITIAL_LINE")
	if err != nil {
		t.Fatal(err)
	}
	if value != "7" {
		t.Fatalf("GetOption(INITIAL_LINE) = %q, want %q (SetOption with a lowercase key must still land under the canonical key)", value, "7")
	}
}

func TestGetOptionBeforeEnsureDefaultsReturnsErrOptionNotSet(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "settings.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	if _, err := store.GetOption(options.CategoryCore, "INITIAL_LINE"); !errors.Is(err, ErrOptionNotSet) {
		t.Fatalf("GetOption before EnsureOptionDefaults: err = %v, want ErrOptionNotSet", err)
	}
}

func TestListOptionsReflectsWhatWasStored(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "settings.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	if err := store.EnsureOptionDefaults(options.CategoryMSX, options.Catalog()); err != nil {
		t.Fatal(err)
	}
	if err := store.SetOption(options.CategoryMSX, "EMULATOR_MACHINE", "Philips_NMS_8250"); err != nil {
		t.Fatal(err)
	}

	all, err := store.ListOptions(options.CategoryMSX)
	if err != nil {
		t.Fatal(err)
	}
	if all["EMULATOR_MACHINE"] != "Philips_NMS_8250" {
		t.Fatalf("ListOptions()[EMULATOR_MACHINE] = %q, want %q", all["EMULATOR_MACHINE"], "Philips_NMS_8250")
	}
	if all["TOKENIZE"] != "false" {
		t.Fatalf("ListOptions()[TOKENIZE] = %q, want the seeded default %q", all["TOKENIZE"], "false")
	}
	// core/zx81/zxspectrum options were never seeded for this Store -
	// ListOptions(msx) must not leak them in.
	if _, leaked := all["SPECTRUM_MODEL"]; leaked {
		t.Fatalf("ListOptions(msx) leaked a zxspectrum-only key: %v", all)
	}
}

func TestEnsureOptionDefaultsOnlySeedsItsOwnCategory(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "settings.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	if err := store.EnsureOptionDefaults(options.CategoryCore, options.Catalog()); err != nil {
		t.Fatal(err)
	}
	if _, err := store.GetOption(options.CategoryMSX, "TOKENIZE"); !errors.Is(err, ErrOptionNotSet) {
		t.Fatalf("GetOption(msx, TOKENIZE) after seeding only core: err = %v, want ErrOptionNotSet", err)
	}
}
