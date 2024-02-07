package main

import (
	"net/http"
	"testing"

	"snippetbox.volcanoeyes.net/internal/assert"
)

func TestPing(t *testing.T) {
	app := newTestApplication(t)

	// huuuuuuuuge testing upside for isolating the application routing in app.routes()
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	status, _, body := ts.get(t, "/ping")
	assert.Equal(t, status, http.StatusOK)
	assert.Equal(t, body, "OK")
}
