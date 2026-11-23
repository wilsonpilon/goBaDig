package gui

import (
	"os"
	"testing"

	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"

	"github.com/wilsonpilon/goBaDig/internal/options"
)

// TestMain sets up Fyne's in-memory test driver (test.NewApp, which
// calls fyne.SetCurrentApp) once for the whole package - every widget
// constructor used by buildOptionsFormItems (Entry.SetText in
// particular, which measures text for layout) panics without a
// current app to read theme/driver state from. theme_test.go, the
// only test file in this package before this one, never needed this
// since it only exercises the pure color/palette mapping, never
// actually constructs a widget.
func TestMain(m *testing.M) {
	test.NewApp()
	os.Exit(m.Run())
}

// These tests exercise buildOptionsFormItems directly - the pure
// wiring layer split out of showOptionsPanel precisely so it can be
// unit-tested without a running window/dialog (Fyne's own form/dialog
// rendering has no established test pattern in this package yet;
// theme_test.go, its closest precedent, likewise only tests the
// catalog/palette mapping function, not showThemeInstaller's dialog
// itself).

func TestBuildOptionsFormItemsMatchesCatalogFieldsInOrder(t *testing.T) {
	fields := options.ForCategory(options.CategoryCore)
	items, read := buildOptionsFormItems(options.CategoryCore, nil)

	if len(items) != len(fields) || len(read) != len(fields) {
		t.Fatalf("got %d items / %d getters, want %d (one per catalog field)", len(items), len(read), len(fields))
	}
	for i, field := range fields {
		if items[i].Text != field.Key {
			t.Fatalf("item %d label = %q, want %q", i, items[i].Text, field.Key)
		}
	}
}

func TestBuildOptionsFormItemsMapsEachTypeToItsOwnWidget(t *testing.T) {
	items, _ := buildOptionsFormItems(options.CategoryCore, nil)
	fields := options.ForCategory(options.CategoryCore)

	for i, field := range fields {
		switch field.Type {
		case options.TypeBool:
			if _, ok := items[i].Widget.(*widget.Check); !ok {
				t.Errorf("%s (TypeBool): widget = %T, want *widget.Check", field.Key, items[i].Widget)
			}
		case options.TypeEnum:
			if _, ok := items[i].Widget.(*widget.Select); !ok {
				t.Errorf("%s (TypeEnum): widget = %T, want *widget.Select", field.Key, items[i].Widget)
			}
		case options.TypeInt, options.TypeString:
			if _, ok := items[i].Widget.(*widget.Entry); !ok {
				t.Errorf("%s (%s): widget = %T, want *widget.Entry", field.Key, field.Type, items[i].Widget)
			}
		}
	}
}

func TestBuildOptionsFormItemsFallsBackToDefaultWhenNoStoredValue(t *testing.T) {
	items, read := buildOptionsFormItems(options.CategoryCore, map[string]string{})
	fields := options.ForCategory(options.CategoryCore)

	for i, field := range fields {
		if got := read[i](); got != field.Default {
			t.Errorf("%s with no stored value: read() = %q, want the catalog default %q", field.Key, got, field.Default)
		}
	}
	// Spot-check the widget itself reflects it too, not just the getter.
	remHeader := findItem(t, items, "REM_HEADER")
	if check, ok := remHeader.Widget.(*widget.Check); !ok || !check.Checked {
		t.Fatalf("REM_HEADER's Check.Checked = %v, want true (its own default)", check.Checked)
	}
}

func TestBuildOptionsFormItemsPrefersTheStoredValueOverDefault(t *testing.T) {
	items, read := buildOptionsFormItems(options.CategoryCore, map[string]string{
		"INITIAL_LINE": "1",
		"REM_HEADER":   "false",
	})

	entry := findItem(t, items, "INITIAL_LINE")
	if e, ok := entry.Widget.(*widget.Entry); !ok || e.Text != "1" {
		t.Fatalf("INITIAL_LINE entry text = %q, want %q (the stored value, not the default)", e.Text, "1")
	}
	if got := readFor(t, items, read, "INITIAL_LINE"); got != "1" {
		t.Fatalf("INITIAL_LINE read() = %q, want %q", got, "1")
	}
	if got := readFor(t, items, read, "REM_HEADER"); got != "false" {
		t.Fatalf("REM_HEADER read() = %q, want %q (the stored value, not the true default)", got, "false")
	}
}

func TestBuildOptionsFormItemsEnumSelectionIsCaseInsensitive(t *testing.T) {
	items, read := buildOptionsFormItems(options.CategoryMSX, map[string]string{
		"CONVERT_PRINT": "to_question_mark",
	})

	sel, ok := findItem(t, items, "CONVERT_PRINT").Widget.(*widget.Select)
	if !ok {
		t.Fatalf("CONVERT_PRINT widget is not a *widget.Select")
	}
	if sel.Selected != "TO_QUESTION_MARK" {
		t.Fatalf("CONVERT_PRINT Select.Selected = %q, want the catalog's own casing %q", sel.Selected, "TO_QUESTION_MARK")
	}
	if got := readFor(t, items, read, "CONVERT_PRINT"); got != "TO_QUESTION_MARK" {
		t.Fatalf("CONVERT_PRINT read() = %q, want %q", got, "TO_QUESTION_MARK")
	}
}

func TestBuildOptionsFormItemsIntEntryValidatorRejectsOutOfRange(t *testing.T) {
	items, _ := buildOptionsFormItems(options.CategoryCore, nil)
	entry, ok := findItem(t, items, "LINE_INCREMENT").Widget.(*widget.Entry)
	if !ok {
		t.Fatalf("LINE_INCREMENT widget is not a *widget.Entry")
	}
	if entry.Validator == nil {
		t.Fatalf("LINE_INCREMENT entry has no Validator wired")
	}
	if err := entry.Validator("1"); err != nil {
		t.Errorf("Validator(1): %v, want nil (1 is LINE_INCREMENT's own Min)", err)
	}
	if err := entry.Validator("0"); err == nil {
		t.Errorf("Validator(0): want an error (below LINE_INCREMENT's Min 1), got nil")
	}
}

func findItem(t *testing.T, items []*widget.FormItem, key string) *widget.FormItem {
	t.Helper()
	for _, item := range items {
		if item.Text == key {
			return item
		}
	}
	t.Fatalf("no form item for %q", key)
	return nil
}

func readFor(t *testing.T, items []*widget.FormItem, read []func() string, key string) string {
	t.Helper()
	for i, item := range items {
		if item.Text == key {
			return read[i]()
		}
	}
	t.Fatalf("no form item for %q", key)
	return ""
}
