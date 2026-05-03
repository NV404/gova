package fyneBridge

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"

	"github.com/nv404/gova/internal/fyne/fonts"
)

type ThemeConfig struct {
	Variant int // 0=light, 1=dark, 2=system
	Accent  color.Color
	Colors  map[int]color.Color
	Sizes   map[int]float32
}

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
		t.variant = 255 // sentinel: pass through whatever Fyne reports
	}
	return t
}

const (
	govaColorPrimary = iota
	govaColorBackground
	govaColorForeground
	govaColorButton
	govaColorError
	govaColorSuccess
	govaColorWarning
	govaColorInputBackground
	govaColorInputBorder
	govaColorDisabled
	govaColorFocus
	govaColorSelection
	govaColorHover
	govaColorSecondary
	govaColorSurface
	govaColorAccent
	govaColorBorder
	govaColorOnPrimary
	govaColorOverlay
	govaColorMenu
	govaColorPlaceholder
	govaColorPressed
	govaColorHeader
	govaColorSeparator
	govaColorHyperlink
	govaColorDisabledBG
	govaColorOnError
	govaColorOnSuccess
	govaColorOnWarning
)

const (
	govaSizePadding = iota
	govaSizeInnerPadding
	govaSizeText
	govaSizeHeadingText
	govaSizeSubHeadingText
	govaSizeCaptionText
	govaSizeSeparatorThickness
	govaSizeInputBorder
	govaSizeInputRadius
	govaSizeSelectionRadius
	govaSizeScrollBar
	govaSizeScrollBarRadius
	govaSizeIconInline
	govaSizeLineSpacing
)

var fyneToGovaColor = map[fyne.ThemeColorName]int{
	theme.ColorNamePrimary:             govaColorPrimary,
	theme.ColorNameBackground:          govaColorBackground,
	theme.ColorNameForeground:          govaColorForeground,
	theme.ColorNameButton:              govaColorButton,
	theme.ColorNameError:               govaColorError,
	theme.ColorNameSuccess:             govaColorSuccess,
	theme.ColorNameWarning:             govaColorWarning,
	theme.ColorNameInputBackground:     govaColorInputBackground,
	theme.ColorNameInputBorder:         govaColorInputBorder,
	theme.ColorNameDisabled:            govaColorDisabled,
	theme.ColorNameFocus:               govaColorFocus,
	theme.ColorNameSelection:           govaColorSelection,
	theme.ColorNameHover:               govaColorHover,
	theme.ColorNameMenuBackground:      govaColorMenu,
	theme.ColorNameOverlayBackground:   govaColorOverlay,
	theme.ColorNamePlaceHolder:         govaColorPlaceholder,
	theme.ColorNamePressed:             govaColorPressed,
	theme.ColorNameHeaderBackground:    govaColorHeader,
	theme.ColorNameSeparator:           govaColorSeparator,
	theme.ColorNameHyperlink:           govaColorHyperlink,
	theme.ColorNameDisabledButton:      govaColorDisabledBG,
	theme.ColorNameForegroundOnPrimary: govaColorOnPrimary,
	theme.ColorNameForegroundOnError:   govaColorOnError,
	theme.ColorNameForegroundOnSuccess: govaColorOnSuccess,
	theme.ColorNameForegroundOnWarning: govaColorOnWarning,
}

var fyneToGovaSize = map[fyne.ThemeSizeName]int{
	theme.SizeNamePadding:            govaSizePadding,
	theme.SizeNameInnerPadding:       govaSizeInnerPadding,
	theme.SizeNameText:               govaSizeText,
	theme.SizeNameHeadingText:        govaSizeHeadingText,
	theme.SizeNameSubHeadingText:     govaSizeSubHeadingText,
	theme.SizeNameCaptionText:        govaSizeCaptionText,
	theme.SizeNameSeparatorThickness: govaSizeSeparatorThickness,
	theme.SizeNameInputBorder:        govaSizeInputBorder,
	theme.SizeNameInputRadius:        govaSizeInputRadius,
	theme.SizeNameSelectionRadius:    govaSizeSelectionRadius,
	theme.SizeNameScrollBar:          govaSizeScrollBar,
	theme.SizeNameScrollBarRadius:    govaSizeScrollBarRadius,
	theme.SizeNameInlineIcon:         govaSizeIconInline,
	theme.SizeNameLineSpacing:        govaSizeLineSpacing,
}

func (t *govaTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	if t.variant != 255 {
		variant = t.variant
	}

	govaName, mapped := fyneToGovaColor[name]
	if mapped {
		if c, ok := t.config.Colors[govaName]; ok {
			return c
		}
		// Accent overrides Primary specifically — buttons, focus, etc.
		if name == theme.ColorNamePrimary && t.config.Accent != nil {
			return t.config.Accent
		}
		if c := govaColor(govaName, variant); c != nil {
			return c
		}
	}

	// Fall through: tokens we don't have a direct opinion on (scrollbar
	// fill, shadow, etc.) borrow from Fyne's default.
	return t.base.Color(name, variant)
}

func (t *govaTheme) Size(name fyne.ThemeSizeName) float32 {
	govaName, mapped := fyneToGovaSize[name]
	if mapped {
		if s, ok := t.config.Sizes[govaName]; ok {
			return s
		}
		if s, ok := govaSize(govaName); ok {
			return s
		}
	}
	return t.base.Size(name)
}

func (t *govaTheme) Font(style fyne.TextStyle) fyne.Resource {
	switch {
	case style.Bold:
		return fonts.SansSemiBold
	case style.Italic:
		return fonts.SansRegular
	}
	return fonts.SansRegular
}

func (t *govaTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return t.base.Icon(name)
}

func govaColor(token int, variant fyne.ThemeVariant) color.Color {
	if variant == theme.VariantLight {
		return govaLight(token)
	}
	return govaDark(token)
}

func govaDark(token int) color.Color {
	switch token {
	case govaColorBackground:
		return rgb(0x0F, 0x0F, 0x0F)
	case govaColorSurface, govaColorInputBackground:
		return rgb(0x17, 0x17, 0x17)
	case govaColorMenu:
		return rgb(0x1A, 0x1A, 0x1A)
	case govaColorOverlay:
		return rgba(0x0F, 0x0F, 0x0F, 0xF2)
	case govaColorHeader:
		return rgb(0x0F, 0x0F, 0x0F)
	case govaColorButton:
		return rgb(0x27, 0x27, 0x27)
	case govaColorDisabledBG:
		return rgb(0x1F, 0x1F, 0x1F)
	case govaColorBorder, govaColorSeparator, govaColorInputBorder:
		return rgb(0x27, 0x27, 0x27)
	case govaColorForeground, govaColorPrimary, govaColorAccent, govaColorHyperlink:
		return rgb(0xF1, 0xF1, 0xF1)
	case govaColorOnPrimary:
		return rgb(0x0F, 0x0F, 0x0F)
	case govaColorSecondary:
		return rgb(0xA3, 0xA3, 0xA3)
	case govaColorPlaceholder:
		return rgb(0x73, 0x73, 0x73)
	case govaColorDisabled:
		return rgb(0x5C, 0x5C, 0x5C)
	case govaColorHover:
		return rgba(0xFF, 0xFF, 0xFF, 0x14)
	case govaColorPressed:
		return rgba(0xFF, 0xFF, 0xFF, 0x1F)
	case govaColorFocus:
		return rgba(0xF1, 0xF1, 0xF1, 0x59)
	case govaColorSelection:
		return rgba(0xF1, 0xF1, 0xF1, 0x29)
	case govaColorError:
		return rgb(0xE5, 0x48, 0x4D)
	case govaColorSuccess:
		return rgb(0x46, 0xA7, 0x58)
	case govaColorWarning:
		return rgb(0xE5, 0xA2, 0x3B)
	case govaColorOnError, govaColorOnSuccess, govaColorOnWarning:
		return rgb(0xF1, 0xF1, 0xF1)
	}
	return nil
}

func govaLight(token int) color.Color {
	switch token {
	case govaColorBackground:
		return rgb(0xF1, 0xF1, 0xF1)
	case govaColorSurface, govaColorInputBackground, govaColorMenu:
		return rgb(0xFF, 0xFF, 0xFF)
	case govaColorOverlay:
		return rgba(0xFF, 0xFF, 0xFF, 0xF2)
	case govaColorHeader:
		return rgb(0xF1, 0xF1, 0xF1)
	case govaColorButton:
		return rgb(0xE5, 0xE5, 0xE5)
	case govaColorDisabledBG:
		return rgb(0xEB, 0xEB, 0xEB)
	case govaColorBorder, govaColorSeparator, govaColorInputBorder:
		return rgb(0xD4, 0xD4, 0xD4)
	case govaColorForeground, govaColorPrimary, govaColorAccent, govaColorHyperlink:
		return rgb(0x0F, 0x0F, 0x0F)
	case govaColorOnPrimary:
		return rgb(0xF1, 0xF1, 0xF1)
	case govaColorSecondary, govaColorPlaceholder:
		return rgb(0x73, 0x73, 0x73)
	case govaColorDisabled:
		return rgb(0xA3, 0xA3, 0xA3)
	case govaColorHover:
		return rgba(0x00, 0x00, 0x00, 0x0A)
	case govaColorPressed:
		return rgba(0x00, 0x00, 0x00, 0x14)
	case govaColorFocus:
		return rgba(0x0F, 0x0F, 0x0F, 0x59)
	case govaColorSelection:
		return rgba(0x0F, 0x0F, 0x0F, 0x29)
	case govaColorError:
		return rgb(0xCD, 0x2B, 0x31)
	case govaColorSuccess:
		return rgb(0x2A, 0x7E, 0x3B)
	case govaColorWarning:
		return rgb(0xAD, 0x57, 0x00)
	case govaColorOnError, govaColorOnSuccess, govaColorOnWarning:
		return rgb(0xF1, 0xF1, 0xF1)
	}
	return nil
}

func govaSize(token int) (float32, bool) {
	switch token {
	case govaSizePadding:
		return 6, true
	case govaSizeInnerPadding:
		return 7, true
	case govaSizeText:
		return 13, true
	case govaSizeHeadingText:
		return 19, true
	case govaSizeSubHeadingText:
		return 15, true
	case govaSizeCaptionText:
		return 11, true
	case govaSizeSeparatorThickness:
		return 1, true
	case govaSizeInputBorder:
		return 1, true
	case govaSizeInputRadius:
		return 6, true
	case govaSizeSelectionRadius:
		return 4, true
	case govaSizeScrollBar:
		return 8, true
	case govaSizeScrollBarRadius:
		return 4, true
	case govaSizeIconInline:
		return 14, true
	case govaSizeLineSpacing:
		return 3, true
	}
	return 0, false
}

func rgb(r, g, b uint8) color.Color     { return color.NRGBA{R: r, G: g, B: b, A: 0xFF} }
func rgba(r, g, b, a uint8) color.Color { return color.NRGBA{R: r, G: g, B: b, A: a} }
