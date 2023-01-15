package templates

type templateNotFoundError struct {
	name string
}

func (e *templateNotFoundError) Error() string {
	return "template '" + e.name + "' not found"
}
