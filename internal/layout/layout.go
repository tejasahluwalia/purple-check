package layout

import (
	"net/http"

	"github.com/a-h/templ"
)

func Handler(content templ.Component, head Head) http.Handler {
	return templ.Handler(Page(content, head))
}
