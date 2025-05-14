package main

import (
	"log"
	"net/http"
	"os"

	"github.com/tikfack/server/internal/infrastructure/connect"

	"github.com/joho/godotenv"
	"github.com/rs/cors"
	// Connect ハンドラー実装
)

func main() {
	// .env の読み込み
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env ファイルの読み込みに失敗しました: %v", err)
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "50051"
	}

	// Connect ハンドラーの初期化
	handler := connect.NewVideoServiceHandler()

	mux := http.NewServeMux()
	// Connect サービスのパターンとハンドラーを登録
	pattern, h := handler.GetHandler()
	mux.Handle(pattern, h)

	// CORS 設定を適用
	handlerWithCORS := cors.AllowAll().Handler(mux)
	log.Printf("Connect gRPC server is running on :%s", port)
	if err := http.ListenAndServe(":"+port, handlerWithCORS); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
