package gova

import "image/color"

type ThemeVariant int

const (
	ThemeLight ThemeVariant = iota
	ThemeDark
	ThemeSystem
)

type ThemeColorName int

const (
	ColorPrimary ThemeColorName = iota
	ColorBackground
	ColorForeground
	ColorButton
	ColorError
	ColorSuccess
	ColorWarning
	ColorInputBackground
	ColorInputBorder
	ColorDisabled
	ColorFocus
	ColorSelection
	ColorHover
	ColorSecondary
	ColorSurface
	ColorAccent
	ColorBorder
	ColorOnPrimary
	ColorOverlay
	ColorMenu
	ColorPlaceholder
	ColorPressed
	ColorHeader
	ColorSeparator
	ColorHyperlink
	ColorDisabledBG
	ColorOnError
	ColorOnSuccess
	ColorOnWarning
)

type ThemeSizeName int

const (
	SizePadding ThemeSizeName = iota
	SizeInnerPadding
	SizeText
	SizeHeadingText
	SizeSubHeadingText
	SizeCaptionText
	SizeSeparatorThickness
	SizeInputBorder
	SizeInputRadius
	SizeSelectionRadius
	SizeScrollBar
	SizeScrollBarRadius
	SizeIconInline
	SizeLineSpacing
)

// themeColor is a sentinel value for a theme-resolved color. It implements
// color.Color (via a zero RGBA) so it passes through the `any` modifier
// field; the renderer detects it and resolves against the active theme.
type themeColor struct{ name ThemeColorName }

func (themeColor) RGBA() (r, g, b, a uint32) { return 0, 0, 0, 0 }

// Semantic colors: theme-resolved at render time.
var (
	Primary     color.Color = themeColor{ColorForeground}
	Secondary   color.Color = themeColor{ColorSecondary}
	Background  color.Color = themeColor{ColorBackground}
	Surface     color.Color = themeColor{ColorSurface}
	Accent      color.Color = themeColor{ColorAccent}
	Destructive color.Color = themeColor{ColorError}
	Success     color.Color = themeColor{ColorSuccess}
	Warning     color.Color = themeColor{ColorWarning}
	BorderColor color.Color = themeColor{ColorBorder}
)

// resolveColor converts either a concrete color.Color or a themeColor
// sentinel into a concrete color.Color. When a theme is provided and a
// semantic sentinel is seen, the theme's mapping is consulted; falls back
// to sensible defaults so the renderer never returns transparent black.
func resolveColor(c any, theme *Theme) color.Color {
	if c == nil {
		return nil
	}
	if tc, ok := c.(themeColor); ok {
		if theme != nil {
			if theme.Colors != nil {
				if cc, ok := theme.Colors[tc.name]; ok {
					return cc
				}
			}
			if tc.name == ColorAccent && theme.Accent != nil {
				return theme.Accent
			}
		}
		return defaultSemantic(tc.name, theme)
	}
	if col, ok := c.(color.Color); ok {
		return col
	}
	return nil
}

func defaultSemantic(name ThemeColorName, theme *Theme) color.Color {
	dark := theme == nil || theme.Variant != ThemeLight
	if dark {
		return govaDarkColor(name)
	}
	return govaLightColor(name)
}

func govaDarkColor(name ThemeColorName) color.Color {
	switch name {
	case ColorBackground:
		return rgb(0x0F, 0x0F, 0x0F)
	case ColorSurface:
		return rgb(0x17, 0x17, 0x17)
	case ColorInputBackground:
		return rgb(0x17, 0x17, 0x17)
	case ColorMenu:
		return rgb(0x1A, 0x1A, 0x1A)
	case ColorOverlay:
		return rgba(0x0F, 0x0F, 0x0F, 0xF2)
	case ColorHeader:
		return rgb(0x0F, 0x0F, 0x0F)
	case ColorButton:
		return rgb(0x27, 0x27, 0x27)
	case ColorDisabledBG:
		return rgb(0x1F, 0x1F, 0x1F)
	case ColorBorder, ColorSeparator, ColorInputBorder:
		return rgb(0x27, 0x27, 0x27)
	case ColorForeground, ColorPrimary, ColorAccent, ColorHyperlink:
		return rgb(0xF1, 0xF1, 0xF1)
	case ColorOnPrimary:
		return rgb(0x0F, 0x0F, 0x0F)
	case ColorSecondary:
		return rgb(0xA3, 0xA3, 0xA3)
	case ColorPlaceholder:
		return rgb(0x73, 0x73, 0x73)
	case ColorDisabled:
		return rgb(0x5C, 0x5C, 0x5C)
	case ColorHover:
		return rgba(0xFF, 0xFF, 0xFF, 0x14)
	case ColorPressed:
		return rgba(0xFF, 0xFF, 0xFF, 0x1F)
	case ColorFocus:
		return rgba(0xF1, 0xF1, 0xF1, 0x59)
	case ColorSelection:
		return rgba(0xF1, 0xF1, 0xF1, 0x29)
	case ColorError:
		return rgb(0xE5, 0x48, 0x4D)
	case ColorSuccess:
		return rgb(0x46, 0xA7, 0x58)
	case ColorWarning:
		return rgb(0xE5, 0xA2, 0x3B)
	case ColorOnError, ColorOnSuccess, ColorOnWarning:
		return rgb(0xF1, 0xF1, 0xF1)
	}
	return color.Black
}

func govaLightColor(name ThemeColorName) color.Color {
	switch name {
	case ColorBackground:
		return rgb(0xF1, 0xF1, 0xF1)
	case ColorSurface:
		return rgb(0xFF, 0xFF, 0xFF)
	case ColorInputBackground:
		return rgb(0xFF, 0xFF, 0xFF)
	case ColorMenu:
		return rgb(0xFF, 0xFF, 0xFF)
	case ColorOverlay:
		return rgba(0xFF, 0xFF, 0xFF, 0xF2)
	case ColorHeader:
		return rgb(0xF1, 0xF1, 0xF1)
	case ColorButton:
		return rgb(0xE5, 0xE5, 0xE5)
	case ColorDisabledBG:
		return rgb(0xEB, 0xEB, 0xEB)
	case ColorBorder, ColorSeparator, ColorInputBorder:
		return rgb(0xD4, 0xD4, 0xD4)
	case ColorForeground, ColorPrimary, ColorAccent, ColorHyperlink:
		return rgb(0x0F, 0x0F, 0x0F)
	case ColorOnPrimary:
		return rgb(0xF1, 0xF1, 0xF1)
	case ColorSecondary:
		return rgb(0x73, 0x73, 0x73)
	case ColorPlaceholder:
		return rgb(0x73, 0x73, 0x73)
	case ColorDisabled:
		return rgb(0xA3, 0xA3, 0xA3)
	case ColorHover:
		return rgba(0x00, 0x00, 0x00, 0x0A)
	case ColorPressed:
		return rgba(0x00, 0x00, 0x00, 0x14)
	case ColorFocus:
		return rgba(0x0F, 0x0F, 0x0F, 0x59)
	case ColorSelection:
		return rgba(0x0F, 0x0F, 0x0F, 0x29)
	case ColorError:
		return rgb(0xCD, 0x2B, 0x31)
	case ColorSuccess:
		return rgb(0x2A, 0x7E, 0x3B)
	case ColorWarning:
		return rgb(0xAD, 0x57, 0x00)
	case ColorOnError, ColorOnSuccess, ColorOnWarning:
		return rgb(0xF1, 0xF1, 0xF1)
	}
	return color.Black
}

func rgb(r, g, b uint8) color.Color     { return color.NRGBA{R: r, G: g, B: b, A: 0xFF} }
func rgba(r, g, b, a uint8) color.Color { return color.NRGBA{R: r, G: g, B: b, A: a} }

type Theme struct {
	Variant ThemeVariant
	Accent  color.Color
	Colors  map[ThemeColorName]color.Color
	Sizes   map[ThemeSizeName]float32
}

func LightTheme() *Theme {
	return &Theme{Variant: ThemeLight}
}

func DarkTheme() *Theme {
	return &Theme{Variant: ThemeDark}
}

func SystemTheme() *Theme {
	return &Theme{Variant: ThemeSystem}
}

func GovaTheme() *Theme {
	return &Theme{Variant: ThemeDark}
}

func GovaLightTheme() *Theme {
	return &Theme{Variant: ThemeLight}
}

func (t *Theme) WithAccent(c color.Color) *Theme {
	t.Accent = c
	return t
}

func (t *Theme) SetColor(name ThemeColorName, c color.Color) *Theme {
	if t.Colors == nil {
		t.Colors = make(map[ThemeColorName]color.Color)
	}
	t.Colors[name] = c
	return t
}

func (t *Theme) SetSize(name ThemeSizeName, v float32) *Theme {
	if t.Sizes == nil {
		t.Sizes = make(map[ThemeSizeName]float32)
	}
	t.Sizes[name] = v
	return t
}

// FadeColor returns c with its alpha channel multiplied by alpha (0..1).
// Useful for hand-rolled fade animations: pick a base color, drive a
// float32 from UseAnimation, and pass FadeColor(base, progress) as the
// reactive color for text or background.
func FadeColor(c color.Color, alpha float64) color.Color {
	if c == nil {
		return nil
	}
	if alpha <= 0 {
		return color.NRGBA{}
	}
	if alpha > 1 {
		alpha = 1
	}
	r, g, b, a := c.RGBA()
	// RGBA() returns premultiplied 16-bit; convert to 8-bit NRGBA.
	return color.NRGBA{
		R: uint8(r >> 8),
		G: uint8(g >> 8),
		B: uint8(b >> 8),
		A: uint8(float64(a>>8) * alpha),
	}
}

// MixColor linearly interpolates between two colors. t=0 returns a, t=1 returns b.
// Useful for animated color transitions (e.g. fading text from Primary to Secondary).
func MixColor(a, b color.Color, t float64) color.Color {
	if t <= 0 {
		return a
	}
	if t >= 1 {
		return b
	}
	ar, ag, ab, aa := a.RGBA()
	br, bg, bb, ba := b.RGBA()
	return color.NRGBA{
		R: uint8(lerpU32(ar, br, t) >> 8),
		G: uint8(lerpU32(ag, bg, t) >> 8),
		B: uint8(lerpU32(ab, bb, t) >> 8),
		A: uint8(lerpU32(aa, ba, t) >> 8),
	}
}

func lerpU32(a, b uint32, t float64) uint32 {
	return uint32(float64(a)*(1-t) + float64(b)*t)
}

func Hex(hex string) color.Color {
	if len(hex) > 0 && hex[0] == '#' {
		hex = hex[1:]
	}
	if len(hex) == 6 {
		hex = hex + "FF"
	}
	if len(hex) != 8 {
		return color.Black
	}
	r := hexByte(hex[0], hex[1])
	g := hexByte(hex[2], hex[3])
	b := hexByte(hex[4], hex[5])
	a := hexByte(hex[6], hex[7])
	return color.NRGBA{R: r, G: g, B: b, A: a}
}

func hexByte(hi, lo byte) uint8 {
	return hexDigit(hi)<<4 | hexDigit(lo)
}

func hexDigit(b byte) uint8 {
	switch {
	case b >= '0' && b <= '9':
		return b - '0'
	case b >= 'a' && b <= 'f':
		return b - 'a' + 10
	case b >= 'A' && b <= 'F':
		return b - 'A' + 10
	default:
		return 0
	}
}
