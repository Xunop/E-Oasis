package response // import "github.com/Xunop/e-oasis/internal/http/response"

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestResponseHasCommonHeaders(t *testing.T) {
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		New(w, r).Write()
	})

	handler.ServeHTTP(w, r)
	resp := w.Result()

	headers := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
	}

	for header, expected := range headers {
		actual := resp.Header.Get(header)
		if actual != expected {
			t.Fatalf(`Unexpected header value, got %q instead of %q`, actual, expected)
		}
	}
}
