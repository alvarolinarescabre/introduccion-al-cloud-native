package main

import (
	"strings"
	"testing"

	"github.com/danielgtaylor/huma/v2/humatest"
)

func TestGetHealthCheck(t *testing.T) {
	_, api := humatest.New(t)

	addRoutes(api)

	resp := api.Get("/")
	if !strings.Contains(resp.Body.String(), "ok") {
		t.Fatalf("Unexpected response: %s", resp.Body.String())
	}
}

func TestGetLink(t *testing.T) {
	_, api := humatest.New(t)

	addRoutes(api)

	resp := api.Get("/v1/link/0", map[string]any{
			"id": 0,
			"url": "https://www.holachamo.com",
			"links": 18,
        })

	if resp.Code != 200 {
		t.Fatalf("Unexpected status code: %d", resp.Code)
	}
}

func TestGetLinkError(t *testing.T) {
	_, api := humatest.New(t)

	addRoutes(api)

	resp := api.Get("/v1/link/10", map[string]any{
		"message": "id must be between 0 and 9",
	})

	if resp.Code != 500 {
		t.Fatalf("Unexpected status code: %d", resp.Code)
	}
}