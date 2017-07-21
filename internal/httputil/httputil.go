package httputil

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
)

func Error(w http.ResponseWriter, code int, err error) {
	var msg string
	if err != nil {
		msg = err.Error()
	} else {
		msg = http.StatusText(code)
	}

	// Log this so have record of what went wrong
	fmt.Fprintf(os.Stderr, "error response: %d %s\n", code, strconv.Quote(msg))
	http.Error(w, msg, code)
}
