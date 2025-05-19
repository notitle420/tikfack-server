# TikFack Server

TikFack サーバーは、動画配信のバックエンドを担当する Go で書かれたアプリケーションです。このサーバーは Connect プロトコルを使用して API を提供し、DMM API を利用して動画データを取得します。

テスト

## 機能

- DMM API を使用した動画データの取得
- Connect プロトコル (gRPC と REST) を使用した API の提供
- 動画情報の取得と提供

## 必要条件

- Go 1.24.1 以上
- DMM API アカウント (API ID と アフィリエイト ID が必要)
- Protocol Buffers コンパイラ (protoc)
- buf ツール (Protocol Buffers 生成用)

## 環境変数

プロジェクトのルートディレクトリに `.env` ファイルを作成し、以下の環境変数を設定してください：

```
DMM_API_ID=あなたのDMM_API_ID
DMM_API_AFFILIATE_ID=あなたのDMM_AFFILIATE_ID
PORT=8080 # オプション: デフォルトは 8080
BASE_URL=http://localhost:8080 # オプション: API のベース URL
HITS=10 # オプション: DMM API から取得する動画の数
```

## Protocol Buffers のセットアップ

サーバーを実行する前に、Protocol Buffers からコードを生成する必要があります：

```bash
# buf ツールのインストール（必要な場合）
go install github.com/bufbuild/buf/cmd/buf@latest

# Protocol Buffers コードの生成
buf generate
```

これにより、`gen` ディレクトリに必要なコードが生成されます。

## インストールと実行

### ローカル開発環境

```bash
# 依存関係のインストール
go mod download

# Protocol Buffers コードの生成（初回または proto ファイル変更時）
buf generate

# サーバーの起動
go run main.go
```

サーバーは `http://localhost:8080` で動作します。

### Docker を使った実行

```bash
# Docker イメージのビルド
docker build -t tikfack-server .

# コンテナの実行
docker run -p 8080:8080 -p 50051:50051 --env-file .env tikfack-server
```

## トラブルシューティング

### パッケージのインポートエラー

以下のようなエラーが表示される場合：

```
package server/generated is not in std
package server/generated/protoconnect is not in std
```

Protocol Buffers のコードが正しく生成されていない可能性があります。以下の手順を試してください：

1. `buf generate` コマンドを実行して Protocol Buffers コードを生成する
2. `go mod tidy` を実行して依存関係を更新する
3. もし問題が解決しない場合は、`go.mod` ファイルの `module` 行が正しいパスになっているか確認する

## API エンドポイント

サーバーは以下の API エンドポイントを提供します：

### GetVideos

動画のリストを取得します。

- URL: `/proto.video.v1.VideoService/GetVideos`
- メソッド: POST
- リクエスト形式: Connect プロトコル
- レスポンス: 動画のリスト

### GetVideoById

指定した ID の動画情報を取得します。

- URL: `/proto.video.v1.VideoService/GetVideoById`
- メソッド: POST
- リクエスト形式: Connect プロトコル
- レスポンス: 単一の動画情報

## プロジェクト構造

```
server/
├── cmd/            # アプリケーションエントリポイント
├── gen/            # 生成されたコード (Protocol Buffers)
├── internal/       # 内部パッケージ
├── migrations/     # データベースマイグレーション
├── pkg/            # 公開パッケージ
├── proto/          # プロトコル定義 (.proto ファイル)
├── tests/          # テスト
├── Dockerfile      # Docker ビルド設定
├── go.mod          # Go モジュール定義
├── go.sum          # 依存関係のハッシュ
├── main.go         # メインエントリポイント
└── README.md       # このファイル
```

## ライセンス

Copyright © 2024 TikFack 