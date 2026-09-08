FROM mcr.microsoft.com/devcontainers/javascript-node:24-bookworm

ARG TZ=UTC
ENV TZ="$TZ"

ARG GOLANGCI_LINT_VERSION=v2.1.6

# System packages plus the PostgreSQL 17 client (psql) from the PGDG apt repository,
# matching the postgres:17 service the dev container talks to.
RUN apt-get update && apt-get install -y --no-install-recommends \
      less \
      procps \
      fzf \
      zsh \
      man-db \
      unzip \
      gnupg2 \
      jq \
      nano \
      vim \
      make \
      curl \
      wget \
      ca-certificates \
      lsb-release \
  && install -d /usr/share/postgresql-common/pgdg \
  && curl -fsSL https://www.postgresql.org/media/keys/ACCC4CF8.asc \
       -o /usr/share/postgresql-common/pgdg/apt.postgresql.org.asc \
  && echo "deb [signed-by=/usr/share/postgresql-common/pgdg/apt.postgresql.org.asc] https://apt.postgresql.org/pub/repos/apt bookworm-pgdg main" \
       > /etc/apt/sources.list.d/pgdg.list \
  && apt-get update && apt-get install -y --no-install-recommends \
      postgresql-client-17 \
  && apt-get clean && rm -rf /var/lib/apt/lists/*

# golangci-lint (pinned to match the backend lint configuration).
RUN ARCH="$(dpkg --print-architecture)" \
  && case "$ARCH" in \
       amd64) GOARCH=amd64 ;; \
       arm64) GOARCH=arm64 ;; \
       *) echo "Unsupported arch: $ARCH" >&2; exit 1 ;; \
     esac \
  && curl -sSfL "https://github.com/golangci/golangci-lint/releases/download/${GOLANGCI_LINT_VERSION}/golangci-lint-${GOLANGCI_LINT_VERSION#v}-linux-${GOARCH}.tar.gz" \
     | tar -xz -C /usr/local/bin --strip-components=1 "golangci-lint-${GOLANGCI_LINT_VERSION#v}-linux-${GOARCH}/golangci-lint"

# Playwright OS dependencies. Browser binaries are installed per-project in post-create.sh.
RUN npx -y playwright install-deps chromium

# Global npm prefix that does not require root (Claude Code CLI, CodeHealth MCP, etc.).
ENV NPM_CONFIG_PREFIX=/usr/local/share/npm-global
ENV PATH="/usr/local/share/npm-global/bin:${PATH}"

ENV DEVCONTAINER=true
ENV SHELL=/bin/zsh
ENV EDITOR=nano
ENV VISUAL=nano

RUN mkdir -p /usr/local/share/npm-global /home/node/go /home/node/.claude \
  && chown -R node:node /usr/local/share/npm-global /home/node/go /home/node/.claude \
  && echo 'export PS1="[easi-dev] %~ %# "' >> /home/node/.zshrc

USER node
