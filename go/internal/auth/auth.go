package auth

import (
	"github.com/blagoySimandov/ampledata/go/internal/config"
	"github.com/workos/workos-go/v6/pkg/usermanagement"
)

func Configure() {
	cfg := config.GetConfig()
	usermanagement.SetAPIKey(cfg.WorkOSApiKey)
}
