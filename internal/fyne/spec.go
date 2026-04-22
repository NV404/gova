package fyneBridge

type ViewKind int

const (
	KindText ViewKind = iota
	KindButton
	KindVStack
	KindHStack
	KindZStack
	KindGroup
	KindSpacer
	KindTextField
	KindToggle
	KindSlider
	KindPicker
	KindProgress
	KindList
	KindGrid
	KindForm
	KindScrollView
	KindDivider
	KindScaffold
	KindTabView
	KindImage
	KindNavStack
)

type ScaffoldSpec struct {
	Top, Bottom, Leading, Trailing *ViewSpec
}

type NavBarSpec struct {
	Title    string
	CanPop   bool
	OnBack   func()
	Leading  []ViewSpec
	Trailing []ViewSpec
}

type ViewSpec struct {
	Kind     ViewKind
	Key      any
	Children []ViewSpec

	TextValue     string
	TextSubscribe func(func(string)) func()

	ButtonLabel   string
	ButtonHandler func()

	TextFieldGet         func() string
	TextFieldSet         func(string)
	TextFieldSetSilent   func(string)
	TextFieldPlaceholder string
	TextFieldOnSubmit    func(string)
	TextFieldMultiline   bool
	TextFieldPassword    bool

	ToggleLabel    string
	ToggleChecked  bool
	ToggleOnChange func(bool)

	SliderMin, SliderMax, SliderStep, SliderValue float64
	SliderOnChange                                func(float64)

	PickerOptions  []string
	PickerSelected int
	PickerOnChange func(string)

	ProgressValue float64

	GridColumns int

	FormItems []FormItemSpec

	ScrollDirection int

	Scaffold *ScaffoldSpec

	TabLabels    []string
	TabIcons     []string
	TabPlacement int

	ImageFilePath string
	ImageURL      string
	ImageFillMode int

	NavBar *NavBarSpec

	Bold      bool
	Italic    bool
	FontStyle int
	TextColor [4]uint8
	HasColor  bool

	PaddingTop, PaddingRight, PaddingBottom, PaddingLeft float32
	HasPadding                                           bool

	Spacing      float32
	HasSpacing   bool
	Alignment    int
	HasAlignment bool

	OnTap    func()
	HasOnTap bool

	CornerRadius    float32
	HasCornerRadius bool

	ShadowColor                 [4]uint8
	ShadowRadius                float32
	ShadowOffsetX, ShadowOffsetY float32
	HasShadow                   bool

	StrokeColor [4]uint8
	StrokeWidth float32
	HasStroke   bool

	BackgroundColor [4]uint8
	HasBackground   bool

	Opacity    float64
	HasOpacity bool

	A11yLabel string
	A11yHint  string
	A11yRole  int
	HasA11y   bool

	MinWidth, MinHeight float32
	HasMinSize          bool
	Grow                bool

	unsubs []func()
}

type FormItemSpec struct {
	Label string
	View  ViewSpec
}

func (s *ViewSpec) AddUnsub(fn func()) {
	s.unsubs = append(s.unsubs, fn)
}

func (s *ViewSpec) Cleanup() {
	for _, fn := range s.unsubs {
		fn()
	}
	s.unsubs = nil
	for i := range s.Children {
		s.Children[i].Cleanup()
	}
	for i := range s.FormItems {
		s.FormItems[i].View.Cleanup()
	}
	if s.Scaffold != nil {
		if s.Scaffold.Top != nil {
			s.Scaffold.Top.Cleanup()
		}
		if s.Scaffold.Bottom != nil {
			s.Scaffold.Bottom.Cleanup()
		}
		if s.Scaffold.Leading != nil {
			s.Scaffold.Leading.Cleanup()
		}
		if s.Scaffold.Trailing != nil {
			s.Scaffold.Trailing.Cleanup()
		}
	}
}
