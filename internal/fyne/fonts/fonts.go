package fonts

import (
	_ "embed"

	"fyne.io/fyne/v2"
)

//go:embed Inter-Regular.ttf
var interRegular []byte

//go:embed Inter-Medium.ttf
var interMedium []byte

//go:embed Inter-SemiBold.ttf
var interSemiBold []byte

//go:embed Inter-Bold.ttf
var interBold []byte

// Cached resources: Fyne re-reads these on every Font() call so we hand
// out the same *StaticResource each time rather than allocating per call.
var (
	SansRegular  = fyne.NewStaticResource("Inter-Regular.ttf", interRegular)
	SansMedium   = fyne.NewStaticResource("Inter-Medium.ttf", interMedium)
	SansSemiBold = fyne.NewStaticResource("Inter-SemiBold.ttf", interSemiBold)
	SansBold     = fyne.NewStaticResource("Inter-Bold.ttf", interBold)
)
