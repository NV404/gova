package gova

// Slider creates a continuous value slider. Accepts either a
// *StateValue[float64] for state-driven binding or a plain float64 for
// static use. Default range is 0..1; override with .Range(min, max).
//
//	g.Slider(volume).Range(0, 100).Step(5)
func Slider(arg any) *viewNode {
	data := &sliderData{min: 0, max: 1}
	switch v := arg.(type) {
	case *StateValue[float64]:
		data.state = v
		data.value = v.Get()
	case float64:
		data.value = v
	case int:
		data.value = float64(v)
	default:
		panic("gova: Slider requires *StateValue[float64] or float64")
	}
	return &viewNode{kind: viewKindSlider, slider: data}
}

// Range overrides the default 0..1 range.
func (n *viewNode) Range(min, max float64) *viewNode {
	if n.slider != nil {
		n.slider.min = min
		n.slider.max = max
		n.slider.hasRange = true
	}
	return n
}

func (n *viewNode) Step(s float64) *viewNode {
	if n.slider != nil {
		n.slider.step = s
	}
	return n
}

func (n *viewNode) OnSliderChange(fn func(float64)) *viewNode {
	if n.slider != nil {
		n.slider.onChange = fn
	}
	return n
}
