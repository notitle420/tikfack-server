package auth

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/bufbuild/connect-go"
	"github.com/coreos/go-oidc"
)

type ContextKey string

const SubKey ContextKey = "sub"

// OIDCInterceptor validates Authorization header on every RPC and stores sub in context.
func OIDCInterceptor(verifier *oidc.IDTokenVerifier) connect.Interceptor {
	return connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			header := req.Header().Get("Authorization")
			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				slog.Error("invalid auth header", "parts", parts)
				return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("invalid auth header"))
			}
			idt, err := verifier.Verify(ctx, parts[1])
			slog.Debug("idt", "idt", idt)
			if err != nil {
				slog.Error("failed to verify idt", "err", err)
				return nil, connect.NewError(connect.CodeUnauthenticated, err)
			}
			var claims struct {
				Sub string `json:"sub"`
			}
			if err := idt.Claims(&claims); err != nil {
				return nil, connect.NewError(connect.CodeInternal, err)
			}
			slog.Debug("claims", "claims", claims)
			ctx = context.WithValue(ctx, SubKey, claims.Sub)
			return next(ctx, req)
		}
	})
}
