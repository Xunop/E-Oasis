package request

import "net/http"

type ContextKey int

const (
	ClientIPContextKey ContextKey = iota
)

func getContextStringValue(r *http.Request, key ContextKey) string {
	if v := r.Context().Value(key); v != nil {
		if value, valid := v.(string); valid {
			return value
		}
	}
	return ""
}

// ClientIP returns the client IP address stored in the context.
func ClientIP(r *http.Request) string {
	return getContextStringValue(r, ClientIPContextKey)
}
