package workos

import (
	"context"
	"errors"
	"log"

	"github.com/workos/workos-go/v4/pkg/usermanagement"
	"github.com/workos/workos-go/v4/pkg/workos_errors"
)

type MembershipClient struct {
	client       *usermanagement.Client
	defaultOrgID string
}

func NewMembershipClient(apiKey, defaultOrgID string) *MembershipClient {
	return &MembershipClient{
		client:       usermanagement.NewClient(apiKey),
		defaultOrgID: defaultOrgID,
	}
}

func (c *MembershipClient) EnsureMembership(ctx context.Context, userID string) error {
	_, err := c.client.CreateOrganizationMembership(ctx, usermanagement.CreateOrganizationMembershipOpts{
		UserID:         userID,
		OrganizationID: c.defaultOrgID,
	})
	if isAlreadyMember(err) {
		return nil
	}
	return err
}

func isAlreadyMember(err error) bool {
	var httpErr workos_errors.HTTPError
	if !errors.As(err, &httpErr) {
		return false
	}
	if httpErr.Code == 409 || httpErr.Code == 422 {
		log.Printf("WorkOS membership already exists (status %d), treating as success", httpErr.Code)
		return true
	}
	return false
}
