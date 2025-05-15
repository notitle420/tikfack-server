package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/rs/cors"

	connecthandler "github.com/tikfack/server/internal/infrastructure/connect"
)

func main() {
	// 環境変数の読み込み
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env ファイルの読み込みに失敗しました: %v", err)
	}

	// サーバーポートの設定
	port := os.Getenv("PORT")
	if port == "" {
		port = "50051"
	}
	
	// ハンドラーの初期化
	videoHandler := connecthandler.NewVideoServiceHandler()
	
	// HTTPルーティングの設定
	mux := http.NewServeMux()
	pattern, handler := videoHandler.GetHandler()
	mux.Handle(pattern, handler)
	
	// CORSミドルウェアの適用
	handlerWithCORS := cors.AllowAll().Handler(mux)
	
	// サーバー起動
	log.Printf("Connect gRPC server is running on :%s", port)
	if err := http.ListenAndServe(":"+port, handlerWithCORS); err != nil {
		log.Fatalf("サーバー起動に失敗しました: %v", err)
	}
}