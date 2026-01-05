package auth

import (
	"github.com/blagoySimandov/ampledata/go/internal/config"
	"github.com/workos/workos-go/v6/pkg/organizations"
	"github.com/workos/workos-go/v6/pkg/sso"
	"github.com/workos/workos-go/v6/pkg/usermanagement"
)

func Configure() {
	cfg := config.GetConfig()
	sso.Configure(cfg.WorkOSApiKey, cfg.WorkOSClientID)
	organizations.SetAPIKey(cfg.WorkOSApiKey)
	usermanagement.SetAPIKey(cfg.WorkOSApiKey)
}
