package gova

type View interface {
	viewNode() *viewNode
}

type viewNode struct {
	kind     viewKind
	key      any
	children []*viewNode
	modifier modifierSet

	text         *textData
	button       *buttonData
	layout       *layoutData
	textField    *textFieldData
	toggle       *toggleData
	slider       *sliderData
	picker       *pickerData
	progress     *progressData
	list         *listData
	grid         *gridData
	form         *formData
	scrollView   *scrollViewData
	scaffold     *scaffoldData
	tabView      *tabViewData
	image        *imageData
	nav          *navData
	componentRef *componentNode
	conditional  *conditional
}

func (n *viewNode) viewNode() *viewNode { return n }

type viewKind int

const (
	viewKindText viewKind = iota
	viewKindButton
	viewKindVStack
	viewKindHStack
	viewKindZStack
	viewKindGroup
	viewKindSpacer
	viewKindComponent
	viewKindTextField
	viewKindToggle
	viewKindSlider
	viewKindPicker
	viewKindProgress
	viewKindList
	viewKindGrid
	viewKindForm
	viewKindScrollView
	viewKindDivider
	viewKindScaffold
	viewKindTabView
	viewKindImage
	viewKindNavStack
)

type scaffoldData struct {
	top, bottom, leading, trailing View
}

type textData struct {
	content textContent
}

type buttonData struct {
	label   string
	handler func()
	style   ButtonStyle
}

type layoutData struct {
	spacing      float32
	alignment    Alignment
	hasAlignment bool
}

type textFieldData struct {
	state       any // *StateValue[string] stored as any to avoid generic field
	placeholder string
	onSubmit    func(string)
	multiline   bool
	password    bool
}

type toggleData struct {
	state    any // *StateValue[bool] when state-driven; nil for static
	label    string
	checked  bool
	onChange func(bool)
}

type sliderData struct {
	state          any // *StateValue[float64] when state-driven; nil for static
	min, max, step float64
	value          float64
	hasRange       bool
	onChange       func(float64)
}

type pickerData struct {
	options  []string
	selected int
	onChange func(string)
}

type progressData struct {
	value float64 // -1 = infinite/indeterminate
}

type listData struct {
	count    int
	renderFn func(i int) *viewNode
}

type gridData struct {
	columns int
}

type formData struct {
	items []formItem
}

type formItem struct {
	label string
	view  *viewNode
}

type scrollViewData struct {
	direction scrollDirection
}

type scrollDirection int

const (
	ScrollVertical scrollDirection = iota
	ScrollHorizontal
	ScrollBoth
)

func (n *viewNode) Key(k any) *viewNode {
	n.key = k
	return n
}
