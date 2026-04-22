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
	dark := theme != nil && theme.Variant == ThemeDark
	switch name {
	case ColorForeground, ColorPrimary:
		if dark {
			return color.NRGBA{R: 240, G: 240, B: 240, A: 255}
		}
		return color.NRGBA{R: 20, G: 20, B: 20, A: 255}
	case ColorSecondary:
		if dark {
			return color.NRGBA{R: 170, G: 170, B: 170, A: 255}
		}
		return color.NRGBA{R: 110, G: 110, B: 110, A: 255}
	case ColorBackground:
		if dark {
			return color.NRGBA{R: 20, G: 20, B: 22, A: 255}
		}
		return color.NRGBA{R: 250, G: 250, B: 250, A: 255}
	case ColorSurface:
		if dark {
			return color.NRGBA{R: 40, G: 40, B: 44, A: 255}
		}
		return color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	case ColorAccent:
		return color.NRGBA{R: 0, G: 122, B: 255, A: 255}
	case ColorError:
		return color.NRGBA{R: 255, G: 59, B: 48, A: 255}
	case ColorSuccess:
		return color.NRGBA{R: 52, G: 199, B: 89, A: 255}
	case ColorWarning:
		return color.NRGBA{R: 255, G: 149, B: 0, A: 255}
	case ColorBorder:
		if dark {
			return color.NRGBA{R: 70, G: 70, B: 75, A: 255}
		}
		return color.NRGBA{R: 200, G: 200, B: 205, A: 255}
	}
	return color.Black
}

type Theme struct {
	Variant ThemeVariant
	Accent  color.Color
	Colors  map[ThemeColorName]color.Color
}

func LightTheme() *Theme {
	return &Theme{Variant: ThemeLight, Colors: make(map[ThemeColorName]color.Color)}
}

func DarkTheme() *Theme {
	return &Theme{Variant: ThemeDark, Colors: make(map[ThemeColorName]color.Color)}
}

func SystemTheme() *Theme {
	return &Theme{Variant: ThemeSystem, Colors: make(map[ThemeColorName]color.Color)}
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
