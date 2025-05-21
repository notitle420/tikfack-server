package auth

import (
	"context"
	"fmt"

	"github.com/coreos/go-oidc"
	"golang.org/x/oauth2"
)

// NewVerifier initializes the OIDC provider once and returns a verifier.
func NewVerifier(ctx context.Context, issuerURL, clientID string) (*oidc.IDTokenVerifier, error) {
	provider, err := oidc.NewProvider(ctx, issuerURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create OIDC provider: %w", err)
	}
	// oauth2.Config can be used if needed
	_ = &oauth2.Config{ClientID: clientID, Endpoint: provider.Endpoint()}
	verifier := provider.Verifier(&oidc.Config{ClientID: clientID})
	return verifier, nil
}
