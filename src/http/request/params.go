package request

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// RouteIntParam returns an URL route parameter as int.
func RouteIntParam(r *http.Request, param string) int {
	vars := mux.Vars(r)
	value, err := strconv.Atoi(vars[param])
	if err != nil {
		return 0
	}

	if value < 0 {
		return 0
	}

	return value
}
