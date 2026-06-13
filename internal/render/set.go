package render

// Set is replaced with the full implementation in the next task.
type Set struct {
	Path        string
	Value       string
	ForceString bool
}

func applySet(doc any, _ Set) (any, error) { return doc, nil }
