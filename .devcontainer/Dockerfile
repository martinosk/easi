FROM mcr.microsoft.com/devcontainers/javascript-node:24-bookworm

ARG TZ=UTC
ENV TZ="$TZ"

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
      socat \
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
  && echo 'export PS1="[easi-dev] %~ %# "' >> /home/node/.zshrc \
  && echo '[ -f /workspace/.devcontainer/cs-env.sh ] && eval "$(bash /workspace/.devcontainer/cs-env.sh --export)"' >> /home/node/.zshrc

USER node
