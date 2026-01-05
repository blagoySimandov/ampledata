package api_test

import (
	"testing"

	"github.com/blagoySimandov/ampledata/go/internal/api"
	"github.com/gorilla/mux"
)

func TestSetupRoutes(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		enrHandler *api.EnrichHandler
		want       *mux.Router
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := api.SetupRoutes(tt.enrHandler)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("SetupRoutes() = %v, want %v", got, tt.want)
			}
		})
	}
}
