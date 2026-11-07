package gui

import (
	"encoding/hex"
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	fynetheme "fyne.io/fyne/v2/theme"

	apptheme "github.com/wilsonpilon/goBaDig/internal/theme"
)

type paletteTheme struct {
	base   fyne.Theme
	colors map[fyne.ThemeColorName]color.Color
}

func newPaletteTheme(item apptheme.Theme) (fyne.Theme, error) {
	if err := item.Validate(); err != nil {
		return nil, err
	}
	colors := map[fyne.ThemeColorName]string{
		fynetheme.ColorNameBackground:                item.Palette.Background,
		fynetheme.ColorNameForeground:                item.Palette.Foreground,
		fynetheme.ColorNameButton:                    item.Palette.Surface,
		fynetheme.ColorNameDisabledButton:            item.Palette.Surface,
		fynetheme.ColorNameInputBackground:           item.Palette.Surface,
		fynetheme.ColorNamePrimary:                   item.Palette.Primary,
		fynetheme.ColorNameHyperlink:                 item.Palette.Primary,
		fynetheme.ColorNameSuccess:                   item.Palette.Primary,
		fynetheme.ColorNameWarning:                   item.Palette.Primary,
		fynetheme.ColorNameHover:                     item.Palette.Hover,
		fynetheme.ColorNameSelection:                 item.Palette.Selection,
		fynetheme.ColorNameFocus:                     item.Palette.Primary,
		fynetheme.ColorNameDisabled:                  item.Palette.Muted,
		fynetheme.ColorNamePlaceHolder:               item.Palette.Muted,
		fynetheme.ColorNameSeparator:                 item.Palette.Border,
		fynetheme.ColorNameInputBorder:               item.Palette.Border,
		fynetheme.ColorNameScrollBar:                 item.Palette.Border,
		fynetheme.ColorNameError:                     item.Palette.Error,
		fynetheme.ColorNameForegroundOnError:         item.Palette.Foreground,
		fynetheme.ColorNameForegroundOnPrimary:       item.Palette.Foreground,
		fynetheme.ColorNamePressed:                   item.Palette.Selection,
		fynetheme.ColorNameMenuBackground:            item.Palette.Surface,
		fynetheme.ColorNameOverlayBackground:         item.Palette.Background,
		fynetheme.ColorNameHeaderBackground:          item.Palette.Surface,
		fynetheme.ColorNameInnerWindowBorder:         item.Palette.Border,
		fynetheme.ColorNameInnerWindowBorderInactive: item.Palette.Border,
		fynetheme.ColorNameScrollBarBackground:       item.Palette.Background,
	}
	parsed := make(map[fyne.ThemeColorName]color.Color, len(colors))
	for name, value := range colors {
		if len(value) != 7 || value[0] != '#' {
			return nil, fmt.Errorf("invalid color %q for theme %q", value, item.ID)
		}
		raw, err := hex.DecodeString(value[1:])
		if err != nil || len(raw) != 3 {
			return nil, fmt.Errorf("invalid color %q for theme %q", value, item.ID)
		}
		parsed[name] = color.NRGBA{R: raw[0], G: raw[1], B: raw[2], A: 0xff}
	}
	return paletteTheme{base: fynetheme.DefaultTheme(), colors: parsed}, nil
}

func (t paletteTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	if value, ok := t.colors[name]; ok {
		return value
	}
	return t.base.Color(name, variant)
}

func (t paletteTheme) Font(style fyne.TextStyle) fyne.Resource {
	return t.base.Font(style)
}

func (t paletteTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return t.base.Icon(name)
}

func (t paletteTheme) Size(name fyne.ThemeSizeName) float32 {
	return t.base.Size(name)
}
