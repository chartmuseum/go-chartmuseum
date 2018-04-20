package chartmuseum

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer ts.Close()

	_, err := NewClient(ts.URL, http.DefaultClient)
	if err != nil {
		t.Error("error creating test client", err)
	}
}
