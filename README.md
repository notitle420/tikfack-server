# TikFack Server

TikFack サーバーは、動画配信のバックエンドを担当する Go 製アプリケーションです。Connect プロトコルを用いて gRPC / gRPC-Web / REST を同時に公開し、DMM API から動画データを取得します。クリーンアーキテクチャの考え方で構成しているため、ユースケースの追加や外部サービスの差し替えが行いやすい構造になっています。

## 主な機能

### コア機能
- **DMM API 統合**: DMM の動画データベースからリアルタイムで動画情報を取得
- **Connect プロトコル対応**: gRPC と REST の両方をサポートする統一 API
- **動画検索**: 日付、キーワード、ID による柔軟な動画検索機能
- **認証・認可**: OIDC + Keycloak によるセキュアな認可フロー
- **イベントログ収集**: 視聴イベントを Kafka に書き込む EventLogService

### アーキテクチャ特徴
- **クリーンアーキテクチャ**: ドメイン駆動設計に基づく明確な責務分離
- **依存性注入**: モック置き換え可能な実装でテストが容易
- **包括的テスト**: テーブル駆動テストと gomock による堅牢なユニットテスト
- **ミドルウェア**: ログ、認証、CORS などの横断的関心ごとを疎結合で実装

## 必要条件

### システム要件
- **Go**: 1.23.0 以上（`toolchain go1.24.3` を使用）
- **Protocol Buffers**: `protoc` コンパイラ
- **buf**: Protocol Buffers 生成ツール
- **Docker / docker compose**: コンテナで動かす場合に必要

### 外部サービス
- **DMM API**: API ID とアフィリエイト ID
- **Keycloak**: OIDC イシューア
- **Kafka**: EventLogService の書き込み先（デフォルトは `localhost:9094`）

## セットアップ

### 1. リポジトリと依存関係

```bash
git clone https://github.com/tikfack/server.git
cd tikfack-server
go mod download
```

### 2. 環境変数

`godotenv` が自動的に `.env` を読み込むため、ルートにファイルを作成します。Docker では `docker.env` を参照するため、両方を同じ値で管理するのが簡単です。

```bash
cp docker.env .env  # 既存の値をベースにローカル用を作成（必要に応じて編集）
```

主要な変数は以下のとおりです。

| 変数 | 必須 | 説明 | デフォルト |
| --- | --- | --- | --- |
| `DMM_API_ID` | ✅ | DMM API ID | - |
| `DMM_API_AFFILIATE_ID` | ✅ | DMM アフィリエイト ID | - |
| `BASE_URL` | ⭕ | DMM API ベース URL | `https://api.dmm.com/affiliate/` |
| `HITS` | ⭕ | DMM API から取得する件数 | `10` |
| `PORT` | ⭕ | HTTP リッスンポート | `50051` |
| `LOG_LEVEL` | ⭕ | `debug/info/warn/error` | `info` |
| `ISSUER_URL` | ✅ | Keycloak Realm の Issuer URL | - |
| `CLIENT_ID` | ✅ | バックエンド用クライアント ID | - |
| `KEYCLOAK_REALM` | ✅ | Realm 名 | - |
| `KEYCLOAK_BACKEND_CLIENT_SECRET` | ✅ | クライアントシークレット | - |
| `KEYCLOAK_BASE_URL` | ✅ | gocloak が利用する Keycloak ベース URL | - |
| `KAFKA_BROKER_ADDRESSES` | ⭕ | Kafka ブローカー (`host:port` をカンマ区切り) | `localhost:9094` |

### 3. Protocol Buffers コード生成

#### Docker 利用（推奨）

Buf や protoc プラグインをローカルに入れなくても、Docker イメージ経由で生成できます。

```bash
./scripts/buf-generate.sh
```

必要に応じて追加の buf フラグを渡せます（例: `./scripts/buf-generate.sh --path proto/event_log`）。


## アプリケーションの起動

### ローカル起動

```bash
go run cmd/app/main.go
```

gRPC / Connect / REST を `http://localhost:50051` で公開します。

### Docker 起動

docker.envを用意し上記の内容を記述してください

```bash
docker compose up --build
```

`docker.env` が `env_file` として読み込まれます。Keycloak や Kafka などの外部コンテナは `backend` ネットワーク上に存在する必要があります。

## トラブルシューティング

### Protocol Buffers の生成エラー
- `buf generate` が実行済みか確認
- `go mod tidy` で依存関係を整理
- `go env GOPATH` に生成物が正しく配置されているか確認

## API エンドポイント

Connect は HTTP/1.1・HTTP/2 どちらでも使用でき、gRPC クライアントからもアクセス可能です。デフォルトでは JSON over HTTP/1.1 も有効なため、`curl` でも動作します。

### VideoService (`video.VideoService`)

| RPC | HTTP パス | 説明 |
| --- | --- | --- |
| `GetVideosByDate` | `/video.VideoService/GetVideosByDate` | 日付範囲で検索し、`SearchMetadata` を返す |
| `GetVideoById` | `/video.VideoService/GetVideoById` | DMM ID から単一動画を取得 |
| `SearchVideos` | `/video.VideoService/SearchVideos` | v3 互換の検索パラメータによる総合検索 |
| `GetVideosByID` | `/video.VideoService/GetVideosByID` | 女優/ジャンル/メーカーなどの ID 条件で絞り込み |
| `GetVideosByKeyword` | `/video.VideoService/GetVideosByKeyword` | キーワード + 期間 + ソートで検索 |

### EventLogService (`eventlog.EventLogService`)

| RPC | HTTP パス | 説明 |
| --- | --- | --- |
| `Record` | `/eventlog.EventLogService/Record` | 単一イベントを Kafka に送信 |
| `RecordBatch` | `/eventlog.EventLogService/RecordBatch` | 複数イベントをまとめて送信 |

## プロジェクト構造

```
tikfack-server/
├── cmd/
│   └── app/              # エントリポイント
├── internal/
│   ├── application/      # ユースケース層
│   ├── domain/           # エンティティ・リポジトリインターフェース
│   ├── infrastructure/   # DMM API クライアント、リポジトリ実装
│   └── middleware/       # 認証・ロギング・コンテキストキー
├── proto/                # Protocol Buffers 定義
├── gen/                  # buf generate による生成コード
├── docs/                 # 設計資料
├── docker-compose.yml    # Docker 起動設定
├── docker.env            # Docker 用環境変数サンプル
└── README.md
```

## 開発とテスト

### テスト実行

```bash
# すべてのテスト
go test ./...

# カバレッジ（HTML 出力例）
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# 特定パッケージのテスト
go test ./internal/application/usecase/video/...
```

### コード品質
- `gomock` を使用したモック生成（`go generate` タグが付いたファイルを参照）
- slog による構造化ログ
- CORS とミドルウェアをチェインして Connect ハンドラーに適用

## 参考資料
- `docs/clean_architecture.mmd`: レイヤー構成
- `docs/sequences/video/*.mmd`: 各ユースケースのシーケンス図
- `docs/entity_diagram/entity_diagram.mmd`: エンティティ関係図
