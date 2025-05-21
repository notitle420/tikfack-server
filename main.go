package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/bufbuild/connect-go"
	"github.com/joho/godotenv"
	"github.com/rs/cors"

	"github.com/tikfack/server/internal/infrastructure/auth"
	connecthandler "github.com/tikfack/server/internal/infrastructure/connect"
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

	ctx := context.Background()

	verifier, err := auth.NewVerifier(ctx, "http://localhost:8080/realms/myrealm", "backend-service")
	if err != nil {
		slog.Error("OIDC verifier init failed", "error", err)
		os.Exit(1)
	}

	oidcInterceptor := auth.OIDCInterceptor(verifier)

	videoHandler := connecthandler.NewVideoServiceHandler(connect.WithInterceptors(oidcInterceptor))
	mux := http.NewServeMux()
	pattern, handler := videoHandler.GetHandler()
	mux.Handle(pattern, handler)

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

	logger.Info("ロガーを設定しました", "level", logLevel.String())
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
