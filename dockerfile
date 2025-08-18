# Goの公式イメージを使用
FROM golang:1.23

# 作業ディレクトリを設定
WORKDIR /app

# Goアプリケーションのソースコードをコピー
COPY . .

# 依存関係を解決
RUN go mod tidy

# アプリケーションをビルド
RUN go build -o main .

# コンテナが起動時に実行するコマンド
CMD ["./main"]

# ポート8080を開放
EXPOSE 8080
