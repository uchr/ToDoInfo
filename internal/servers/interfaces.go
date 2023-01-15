package servers

import "io"

type Templates interface {
	Render(w io.Writer, name string, data interface{}) error
}
