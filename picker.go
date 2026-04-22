package gova

// Picker is the simple string-option picker.
//
//	g.Picker([]string{"Low", "Medium", "High"}, selectedIdx).OnPickerChange(...)
func Picker(options []string, selected int) *viewNode {
	return &viewNode{
		kind: viewKindPicker,
		picker: &pickerData{
			options:  options,
			selected: selected,
		},
	}
}

func (n *viewNode) OnPickerChange(fn func(string)) *viewNode {
	if n.picker != nil {
		n.picker.onChange = fn
	}
	return n
}

// PickerOf is a typed picker for any option type. The label function maps
// each value to the display string; the state holds the currently selected
// value. Selection changes write through to state.
//
//	g.PickerOf(priority, []Priority{Low, Med, High},
//	    func(p Priority) string { return p.String() })
func PickerOf[T comparable](state *StateValue[T], options []T, label func(T) string) *viewNode {
	strs := make([]string, len(options))
	selected := -1
	cur := state.Get()
	for i, o := range options {
		strs[i] = label(o)
		if o == cur {
			selected = i
		}
	}
	if selected < 0 {
		selected = 0
	}
	return &viewNode{
		kind: viewKindPicker,
		picker: &pickerData{
			options:  strs,
			selected: selected,
			onChange: func(s string) {
				for i, o := range options {
					if strs[i] == s {
						state.Set(o)
						return
					}
				}
			},
		},
	}
}
