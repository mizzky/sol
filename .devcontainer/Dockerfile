FROM mcr.microsoft.com/devcontainers/go:1.25-trixie

ARG NODE_VERSION="22"
RUN su vscode -c "source /usr/local/share/nvm/nvm.sh && nvm install ${NODE_VERSION}"

RUN apt-get update && apt-get install -y postgresql-client

# Starshipをバイナリで直接インストール
RUN curl -sS https://starship.rs/install.sh | sh -s -- -y

# bash使用時にStarshipを起動するための設定
RUN echo 'eval "$(starship init bash)"' >> /home/vscode/.bashrc

RUN echo 'alias ll="ls -laF --color=auto"' >> /home/vscode/.bashrc