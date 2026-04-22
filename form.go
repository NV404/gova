package gova

func Form(items ...FormEntry) *viewNode {
	fi := make([]formItem, len(items))
	for i, entry := range items {
		fi[i] = formItem{label: entry.Label, view: entry.View.viewNode()}
	}
	return &viewNode{
		kind: viewKindForm,
		form: &formData{items: fi},
	}
}

type FormEntry struct {
	Label string
	View  View
}

func Divider() *viewNode {
	return &viewNode{kind: viewKindDivider}
}
