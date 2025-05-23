package auth

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/coreos/go-oidc"
)

// NewVerifier initializes the OIDC provider once and returns a verifier.
func NewVerifier(ctx context.Context, issuerURL, clientID string) (*oidc.IDTokenVerifier, error) {
	provider, err := oidc.NewProvider(ctx, issuerURL)
	if err != nil {
		slog.Error("failed to create OIDC provider", "err", err)
		return nil, fmt.Errorf("failed to create OIDC provider: %w", err)
	}
	verifier := provider.Verifier(&oidc.Config{
		ClientID: clientID,
	})
	return verifier, nil
}
