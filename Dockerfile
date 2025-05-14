# ======================
# 1. ビルド用ステージ
# ======================
FROM golang:1.22-alpine AS builder

# 作業ディレクトリを設定
WORKDIR /app

# 依存ファイルをコピーしてgo mod download
COPY go.mod go.sum ./
RUN go mod download

# 残りのソースをコピー
COPY . .

# ビルド (静的リンク)
RUN go build -o server main.go

# ======================
# 2. 実行用ステージ
# ======================
FROM alpine:3.15

WORKDIR /app

# builderステージでビルドしたバイナリをコピー
COPY --from=builder /app/server .

# コンテナがListenするポート (REST API と gRPC)
EXPOSE 8080 50051

# 実行コマンド
CMD ["./server"]
