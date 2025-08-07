# ======================
# 1. ビルド用ステージ
# ======================
FROM golang:1.23-alpine AS builder

# 作業ディレクトリを設定
WORKDIR /app

# 依存ファイルをコピーしてgo mod download
COPY go.mod go.sum ./
RUN go mod download

# 残りのソースをコピー
COPY . .

# ビルド (静的リンク)
RUN go build -o server ./cmd/app

# ======================
# 2. 実行用ステージ
# ======================
FROM alpine:3.15

WORKDIR /app

# builderステージでビルドしたバイナリをコピー
COPY --from=builder /app/server .

# コンテナがListenするポート (REST API と gRPC)
EXPOSE 50051

# 実行コマンド
CMD ["./server"]
