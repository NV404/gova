package fyneBridge

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type ThemeConfig struct {
	Variant int // 0=light, 1=dark, 2=system
	Accent  color.Color
	Colors  map[int]color.Color // keyed by gova ThemeColorName
}

// govaTheme implements fyne.Theme, delegating to user overrides and
// falling back to Fyne's default theme.
type govaTheme struct {
	config  ThemeConfig
	base    fyne.Theme
	variant fyne.ThemeVariant
}

func NewTheme(config ThemeConfig) fyne.Theme {
	t := &govaTheme{
		config: config,
		base:   theme.DefaultTheme(),
	}
	switch config.Variant {
	case 0:
		t.variant = theme.VariantLight
	case 1:
		t.variant = theme.VariantDark
	default:
		t.variant = 255 // sentinel: use system
	}
	return t
}

var colorNameMap = map[int]fyne.ThemeColorName{
	0:  theme.ColorNamePrimary,
	1:  theme.ColorNameBackground,
	2:  theme.ColorNameForeground,
	3:  theme.ColorNameButton,
	4:  theme.ColorNameError,
	5:  theme.ColorNameSuccess,
	6:  theme.ColorNameWarning,
	7:  theme.ColorNameInputBackground,
	8:  theme.ColorNameInputBorder,
	9:  theme.ColorNameDisabled,
	10: theme.ColorNameFocus,
	11: theme.ColorNameSelection,
	12: theme.ColorNameHover,
}

// reverse map: fyne color name → gova int (built once)
var fyneToGovaColor map[fyne.ThemeColorName]int

func init() {
	fyneToGovaColor = make(map[fyne.ThemeColorName]int, len(colorNameMap))
	for k, v := range colorNameMap {
		fyneToGovaColor[v] = k
	}
}

func (t *govaTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	if t.variant != 255 {
		variant = t.variant
	}

	if govaName, ok := fyneToGovaColor[name]; ok {
		if c, ok := t.config.Colors[govaName]; ok {
			return c
		}
	}

	if name == theme.ColorNamePrimary && t.config.Accent != nil {
		return t.config.Accent
	}

	return t.base.Color(name, variant)
}

func (t *govaTheme) Font(style fyne.TextStyle) fyne.Resource {
	return t.base.Font(style)
}

func (t *govaTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return t.base.Icon(name)
}

func (t *govaTheme) Size(name fyne.ThemeSizeName) float32 {
	return t.base.Size(name)
}
