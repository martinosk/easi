FROM node:24-bookworm

ARG TZ=UTC
ENV TZ="$TZ"

ARG GO_VERSION=1.26.2
ARG GOLANGCI_LINT_VERSION=v2.1.6

RUN apt-get update && apt-get install -y --no-install-recommends \
  less \
  git \
  procps \
  sudo \
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
  postgresql-client \
  && apt-get clean && rm -rf /var/lib/apt/lists/*

RUN ARCH=$(dpkg --print-architecture) && \
  if [ "$ARCH" = "amd64" ]; then GOARCH=amd64; \
  elif [ "$ARCH" = "arm64" ]; then GOARCH=arm64; \
  else echo "Unsupported arch: $ARCH" && exit 1; fi && \
  wget -q "https://go.dev/dl/go${GO_VERSION}.linux-${GOARCH}.tar.gz" -O /tmp/go.tar.gz && \
  tar -C /usr/local -xzf /tmp/go.tar.gz && \
  rm /tmp/go.tar.gz

ENV PATH="/usr/local/go/bin:/home/node/go/bin:${PATH}"
ENV GOPATH="/home/node/go"

RUN ARCH=$(dpkg --print-architecture) && \
  if [ "$ARCH" = "amd64" ]; then GOARCH=amd64; \
  elif [ "$ARCH" = "arm64" ]; then GOARCH=arm64; \
  else echo "Unsupported arch: $ARCH" && exit 1; fi && \
  curl -sSfL "https://github.com/golangci/golangci-lint/releases/download/${GOLANGCI_LINT_VERSION}/golangci-lint-${GOLANGCI_LINT_VERSION#v}-linux-${GOARCH}.tar.gz" \
  | tar -xz -C /usr/local/bin --strip-components=1 "golangci-lint-${GOLANGCI_LINT_VERSION#v}-linux-${GOARCH}/golangci-lint"

RUN go install github.com/swaggo/swag/cmd/swag@v1.16.6

RUN npx playwright install --with-deps chromium

RUN mkdir -p /usr/local/share/npm-global && \
  chown -R node:node /usr/local/share/npm-global

ENV DEVCONTAINER=true

RUN mkdir -p /workspace /home/node/.claude /home/node/go && \
  chown -R node:node /workspace /home/node/.claude /home/node/go

RUN echo 'node ALL=(ALL) NOPASSWD:ALL' > /etc/sudoers.d/node && \
  chmod 440 /etc/sudoers.d/node

WORKDIR /workspace

USER node

ENV NPM_CONFIG_PREFIX=/usr/local/share/npm-global
ENV PATH="${PATH}:/usr/local/share/npm-global/bin"


ENV SHELL=/bin/zsh
ENV EDITOR=nano
ENV VISUAL=nano

RUN echo 'export PS1="[easi-dev] %~ %# "' > /home/node/.zshrc
