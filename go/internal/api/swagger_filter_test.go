package api

import (
	"strings"
	"testing"
)

func TestPublicSpecHidesInternal(t *testing.T) {
	spec, err := publicSpec()
	if err != nil {
		t.Fatal(err)
	}
	s := string(spec)
	hiddenPaths := []string{`"/subscription"`, `"/subscribe"`, `"/subscription/upgrade"`, `"/subscription/cancel"`, `"/subscription/portal"`, `"/select-key"`, `"/webhooks/stripe"`}
	for _, p := range hiddenPaths {
		if strings.Contains(s, p) {
			t.Errorf("public spec leaks internal path: %s", p)
		}
	}
	hiddenOps := []string{`"operationId":"GetSubscriptionStatus"`, `"operationId":"SelectKey"`, `"operationId":"HandleStripeWebhook"`, `"operationId":"UpgradeSubscription"`}
	for _, op := range hiddenOps {
		if strings.Contains(s, op) {
			t.Errorf("public spec leaks internal op: %s", op)
		}
	}
	shownOps := []string{`"operationId":"EnrichSource"`, `"operationId":"ListSources"`, `"operationId":"GetJobResults"`, `"operationId":"ListTiers"`, `"operationId":"GetMe"`}
	for _, op := range shownOps {
		if !strings.Contains(s, op) {
			t.Errorf("public spec missing public op: %s", op)
		}
	}
}
