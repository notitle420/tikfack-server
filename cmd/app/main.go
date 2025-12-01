package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/bufbuild/connect-go"
	"github.com/joho/godotenv"
	gocloak "github.com/mviniciusgc/gocloak/v13"
	"github.com/rs/cors"
	"github.com/tikfack/server/internal/di"
	auth "github.com/tikfack/server/internal/middleware/auth"
	"github.com/tikfack/server/internal/middleware/logger"
)

func main() {
	// 環境変数の読み込み
	if err := godotenv.Load(); err != nil {
		// .envファイルがなくてもエラーではない（本番環境では環境変数で設定する場合がある）
		slog.Info("環境変数を.envから読み込めませんでした", "error", err)
	}

	// ログレベルの設定
	logLevel := os.Getenv("LOG_LEVEL")
	setupLogger(logLevel)

	port := os.Getenv("PORT")
	if port == "" {
		port = "50051"
	}

	keycloakBaseURL := os.Getenv("KEYCLOAK_BASE_URL")
	issuerURL := os.Getenv("ISSUER_URL")
	clientID := os.Getenv("CLIENT_ID")
	realm := os.Getenv("KEYCLOAK_REALM")
	clientSecret := os.Getenv("KEYCLOAK_BACKEND_CLIENT_SECRET")
	//keycloakTokenEndpoint := os.Getenv("KEYCLOAK_TOKEN_ENDPOINT")
	slog.Info("issuerURL", "issuerURL", issuerURL)
	slog.Info("clientID", "clientID", clientID)

	gocloakClient := gocloak.NewClient(keycloakBaseURL)

	//ユーザーIDを取得するためにOIDCを使用
	ctx := context.Background()
	verifier, err := auth.NewVerifier(ctx, issuerURL, clientID)
	if err != nil {
		slog.Error("OIDC verifier init failed", "error", err)
		os.Exit(1)
	}
	slog.Info("OIDC verifier initialized", "verifier", verifier)

	introspectionInterceptor := auth.IntrospectionInterceptor(
		verifier,
		gocloakClient,
		realm,
		clientID,
		clientSecret,
	)
	permInterceptor := auth.PermissionInterceptor(
		gocloakClient,
		realm,
		clientID,
		auth.CheckPermissionFunc,
	)

	videoHandler, err := di.InitializeVideoHandler([]connect.HandlerOption{
		connect.WithInterceptors(
			introspectionInterceptor,
			logger.LoggingInterceptor(),
			permInterceptor,
		),
	})
	if err != nil {
		slog.Error("failed to initialize video handler", "error", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()
	pattern, handler := videoHandler.GetHandler()
	mux.Handle(pattern, handler)

	eventHandler, err := di.InitializeEventLogHandler([]connect.HandlerOption{
		connect.WithInterceptors(
			logger.LoggingInterceptor(),
		),
	})
	if err != nil {
		slog.Error("failed to initialize event log handler", "error", err)
		os.Exit(1)
	}
	epattern, ehandler := eventHandler.GetHandler()
	mux.Handle(epattern, ehandler)

	// ミドルウェアチェイン
	loggedHandler := loggingMiddleware(mux)
	handlerWithCORS := cors.AllowAll().Handler(loggedHandler)

	slog.Info("サーバーを起動しています", "port", port)
	if err := http.ListenAndServe(":"+port, handlerWithCORS); err != nil {
		slog.Error("サーバー起動に失敗しました", "error", err)
		os.Exit(1)
	}
}

// setupLogger configures the global slog logger based on the environment
func setupLogger(level string) {
	var logLevel slog.Level
	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo // デフォルトはInfo
	}

	// JSONハンドラーを使用
	opts := &slog.HandlerOptions{
		Level: logLevel,
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, opts))
	slog.SetDefault(logger)

	slog.Info("ロガーを設定しました", "level", logLevel.String())
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.Info("新しいリクエスト",
			"method", r.Method,
			"path", r.URL.Path,
			"remote_addr", r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}
