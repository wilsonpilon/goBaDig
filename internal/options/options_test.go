package options

import "testing"

func TestCatalogHasNoDuplicateCategoryKeyPairs(t *testing.T) {
	seen := map[string]bool{}
	for _, f := range Catalog() {
		id := f.Category + "." + f.Key
		if seen[id] {
			t.Fatalf("duplicate option %s in Catalog()", id)
		}
		seen[id] = true
	}
}

func TestCatalogEveryFieldValidatesItsOwnDefault(t *testing.T) {
	for _, f := range Catalog() {
		if err := Validate(f.Category, f.Key, f.Default); err != nil {
			t.Errorf("option %s.%s's own Default %q fails Validate: %v", f.Category, f.Key, f.Default, err)
		}
	}
}

func TestFindIsCaseInsensitiveOnKey(t *testing.T) {
	if _, ok := Find(CategoryCore, "initial_line"); !ok {
		t.Fatalf("Find(core, initial_line) = not found, want INITIAL_LINE")
	}
	if _, ok := Find(CategoryCore, "Initial_Line"); !ok {
		t.Fatalf("Find(core, Initial_Line) = not found, want INITIAL_LINE")
	}
	if _, ok := Find(CategoryCore, "NOT_A_REAL_OPTION"); ok {
		t.Fatalf("Find(core, NOT_A_REAL_OPTION) = found, want not found")
	}
	if _, ok := Find("not-a-category", "INITIAL_LINE"); ok {
		t.Fatalf("Find(not-a-category, INITIAL_LINE) = found, want not found (category is exact-match)")
	}
}

func TestValidateUnknownOption(t *testing.T) {
	if err := Validate(CategoryCore, "NOT_A_REAL_OPTION", "1"); err == nil {
		t.Fatalf("Validate: want error for unknown option, got nil")
	}
}

func TestValidateBool(t *testing.T) {
	for _, raw := range []string{"true", "false"} {
		if err := Validate(CategoryCore, "REM_HEADER", raw); err != nil {
			t.Errorf("Validate(REM_HEADER, %q): %v, want nil", raw, err)
		}
	}
	for _, raw := range []string{"True", "1", "yes", ""} {
		if err := Validate(CategoryCore, "REM_HEADER", raw); err == nil {
			t.Errorf("Validate(REM_HEADER, %q): want error, got nil", raw)
		}
	}
}

func TestValidateIntRespectsMinAndMax(t *testing.T) {
	if err := Validate(CategoryCore, "VERBOSE_LEVEL", "2"); err != nil {
		t.Errorf("Validate(VERBOSE_LEVEL, 2): %v, want nil", err)
	}
	if err := Validate(CategoryCore, "VERBOSE_LEVEL", "-1"); err == nil {
		t.Errorf("Validate(VERBOSE_LEVEL, -1): want error (below Min 0), got nil")
	}
	if err := Validate(CategoryCore, "VERBOSE_LEVEL", "5"); err == nil {
		t.Errorf("Validate(VERBOSE_LEVEL, 5): want error (above Max 4), got nil")
	}
	if err := Validate(CategoryCore, "VERBOSE_LEVEL", "not-a-number"); err == nil {
		t.Errorf("Validate(VERBOSE_LEVEL, not-a-number): want error, got nil")
	}
}

func TestValidateIntAllowsNegativeWhenFieldHasNoMin(t *testing.T) {
	if err := Validate(CategoryZXSpectrum, "AUTORUN_LINE", "-1"); err != nil {
		t.Errorf("Validate(AUTORUN_LINE, -1): %v, want nil (-1 is its own documented off-value)", err)
	}
	if err := Validate(CategoryZXSpectrum, "AUTORUN_LINE", "-2"); err == nil {
		t.Errorf("Validate(AUTORUN_LINE, -2): want error (below Min -1), got nil")
	}
}

func TestValidateEnumIsCaseInsensitive(t *testing.T) {
	for _, raw := range []string{"48", "128"} {
		if err := Validate(CategoryZXSpectrum, "SPECTRUM_MODEL", raw); err != nil {
			t.Errorf("Validate(SPECTRUM_MODEL, %q): %v, want nil", raw, err)
		}
	}
	if err := Validate(CategoryZXSpectrum, "SPECTRUM_MODEL", "64"); err == nil {
		t.Errorf("Validate(SPECTRUM_MODEL, 64): want error, got nil")
	}
	if err := Validate(CategoryMSX, "CONVERT_PRINT", "to_question_mark"); err != nil {
		t.Errorf("Validate(CONVERT_PRINT, to_question_mark): %v, want nil (case-insensitive)", err)
	}
}

func TestValidateStringAcceptsAnything(t *testing.T) {
	for _, raw := range []string{"", "anything at all", `C:\some\path.exe`} {
		if err := Validate(CategoryMSX, "EMULATOR_PATH", raw); err != nil {
			t.Errorf("Validate(EMULATOR_PATH, %q): %v, want nil", raw, err)
		}
	}
}

func TestForCategoryOnlyReturnsThatCategory(t *testing.T) {
	fields := ForCategory(CategoryZX81)
	if len(fields) == 0 {
		t.Fatalf("ForCategory(zx81) = empty, want at least one field")
	}
	for _, f := range fields {
		if f.Category != CategoryZX81 {
			t.Fatalf("ForCategory(zx81) returned a %s field: %+v", f.Category, f)
		}
	}
}

// TestZX81AndMSXTokenizeDefaultsDiffer confirms a real, deliberate
// divergence from the original engine carried over on purpose: MSX's
// own TOKENIZE defaults to off (false) but ZX81's defaults to on
// (true) - badig.py itself has this same asymmetry (msx/
// tokenizer_interface.py's tokenize=False vs zx81/
// tokenizer_interface.py's tokenize=True).
func TestZX81AndMSXTokenizeDefaultsDiffer(t *testing.T) {
	msx, ok := Find(CategoryMSX, "TOKENIZE")
	if !ok || msx.Default != "false" {
		t.Fatalf("msx.TOKENIZE default = %+v, want false", msx)
	}
	zx81, ok := Find(CategoryZX81, "TOKENIZE")
	if !ok || zx81.Default != "true" {
		t.Fatalf("zx81.TOKENIZE default = %+v, want true", zx81)
	}
}
