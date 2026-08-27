FROM mcr.microsoft.com/devcontainers/go:1.25-trixie

ARG NODE_VERSION="22"
RUN su vscode -c "source /usr/local/share/nvm/nvm.sh && nvm install ${NODE_VERSION}"

# Codex CLIをvscodeユーザー環境にインストール
RUN su vscode -c "source /usr/local/share/nvm/nvm.sh && \
    nvm use ${NODE_VERSION} && \
    npm install -g @openai/codex"

RUN apt-get update && apt-get install -y postgresql-client

# Starshipをバイナリで直接インストール
RUN curl -sS https://starship.rs/install.sh | sh -s -- -y

#goツールインストール
RUN go install github.com/air-verse/air@v1.64.5 && \
    go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.30.0 && \
    go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest &&\
    go install golang.org/x/tools/gopls@v0.21.0

RUN chmod -R 755 /go/bin

# bash使用時にStarshipを起動するための設定
RUN echo 'eval "$(starship init bash)"' >> /home/vscode/.bashrc

RUN echo 'alias ll="ls -laF --color=auto"' >> /home/vscode/.bashrc

RUN echo 'alias go-cover-html="go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out -o coverage.html && echo \"Coverage report generated: coverage.html\""' >> /home/vscode/.bashrc