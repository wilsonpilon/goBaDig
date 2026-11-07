package gui

import (
	"encoding/hex"
	"image/color"
	"testing"

	fynetheme "fyne.io/fyne/v2/theme"

	"github.com/wilsonpilon/goBaDig/internal/theme"
)

func TestCatalogThemesApplyTheirPalette(t *testing.T) {
	catalog := theme.Catalog()
	if len(catalog) < 5 || len(catalog) > 7 {
		t.Fatalf("catalog has %d themes; want 5 to 7", len(catalog))
	}

	for _, item := range catalog {
		t.Run(item.ID, func(t *testing.T) {
			fyneTheme, err := newPaletteTheme(item)
			if err != nil {
				t.Fatal(err)
			}
			wantBytes, err := hex.DecodeString(item.Palette.Primary[1:])
			if err != nil {
				t.Fatal(err)
			}
			got := color.NRGBAModel.Convert(
				fyneTheme.Color(fynetheme.ColorNamePrimary, fynetheme.VariantDark),
			).(color.NRGBA)
			want := color.NRGBA{R: wantBytes[0], G: wantBytes[1], B: wantBytes[2], A: 0xff}
			if got != want {
				t.Fatalf("primary color = %#v; want %#v", got, want)
			}
		})
	}
}
