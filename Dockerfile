# syntax=docker/dockerfile:1

# Mattermost self-hosted image, built from source (server + web app).
#
# The heavy build work (webpack + Go cross-compile) is pinned to the build
# host architecture for speed. The final runtime stage is per-architecture,
# so the same Dockerfile produces working linux/amd64 and linux/arm64 images.

# Configurable base images
ARG NODE_IMAGE=node:24.11-bookworm
ARG GO_IMAGE=golang:1.26.7-bookworm
ARG RUNTIME_IMAGE=debian:bookworm-slim

# ---------------------------------------------------------------------------
# Stage 1: build the web app (architecture-independent static assets)
# ---------------------------------------------------------------------------
FROM --platform=${BUILDPLATFORM} ${NODE_IMAGE} AS webapp
WORKDIR /src/webapp

# Copy workspace package definitions first to cache `npm ci` layer
COPY webapp/package.json webapp/package-lock.json ./
COPY webapp/channels/package.json ./channels/
COPY webapp/platform/client/package.json ./platform/client/
COPY webapp/platform/components/package.json ./platform/components/
COPY webapp/platform/eslint-plugin/package.json ./platform/eslint-plugin/
COPY webapp/platform/mattermost-redux/package.json ./platform/mattermost-redux/
COPY webapp/platform/shared/package.json ./platform/shared/
COPY webapp/platform/types/package.json ./platform/types/
COPY webapp/patches/ ./patches/

# Cache npm download cache across builds
RUN --mount=type=cache,target=/root/.npm \
    npm ci

# Copy the rest of the webapp source and build
COPY webapp/ ./
RUN npm run build

# ---------------------------------------------------------------------------
# Stage 2: build the server binaries and assemble the /mattermost directory
# ---------------------------------------------------------------------------
FROM --platform=${BUILDPLATFORM} ${GO_IMAGE} AS builder
ARG TARGETOS
ARG TARGETARCH
ARG BUILD_NUMBER=dev
ARG BUILD_DATE=unknown
ARG BUILD_HASH=unknown

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates git \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /src

# Copy go.mod / go.sum files first to cache `go mod download` layer
COPY server/go.mod server/go.sum ./server/
COPY server/public/go.mod server/public/go.sum ./server/public/

RUN cd server \
    && go work init \
    && go work use . \
    && go work use ./public \
    && go mod download

# Copy the rest of server source and the pre-built web client
COPY server/ ./server/
COPY --from=webapp /src/webapp/channels/dist ./webapp/channels/dist

# Build the two server binaries for the target architecture with Go build cache
RUN --mount=type=cache,target=/root/.cache/go-build \
    cd server \
    && mkdir -p "bin/${TARGETOS}_${TARGETARCH}" \
    && echo "Building mattermost and mmctl for TARGETOS=${TARGETOS} TARGETARCH=${TARGETARCH}..." \
    && CGO_ENABLED=0 GOOS="${TARGETOS}" GOARCH="${TARGETARCH}" go build \
         -buildvcs=false -trimpath \
         -tags 'production' \
         -ldflags "-X github.com/mattermost/mattermost/server/public/model.BuildNumber=${BUILD_NUMBER} -X github.com/mattermost/mattermost/server/public/model.BuildDate=${BUILD_DATE} -X github.com/mattermost/mattermost/server/public/model.BuildHash=${BUILD_HASH} -X github.com/mattermost/mattermost/server/public/model.BuildHashEnterprise=none -X github.com/mattermost/mattermost/server/public/model.BuildEnterpriseReady=false" \
         -o "bin/${TARGETOS}_${TARGETARCH}" \
         ./cmd/mattermost ./cmd/mmctl

# Assemble the same directory layout the official image ships: config, fonts,
# templates, i18n, the built web client, and the two binaries.
RUN cd server \
    && rm -rf dist \
    && mkdir -p dist/mattermost/config \
    && cp config/README.md dist/mattermost/config/ \
    && OUTPUT_CONFIG="$PWD/dist/mattermost/config/config.json" go run ./scripts/config_generator \
    && cp -RL fonts dist/mattermost/ \
    && cp -RL templates dist/mattermost/ \
    && rm -rf dist/mattermost/templates/*.mjml dist/mattermost/templates/partials/ \
    && cp -RL i18n dist/mattermost/ \
    && sed -i 's|"ConsoleLevel": "DEBUG"|"ConsoleLevel": "INFO"|g' dist/mattermost/config/config.json \
    && sed -i 's|"SiteURL": "http://localhost:8065"|"SiteURL": ""|g' dist/mattermost/config/config.json \
    && sed -i 's|"SendEmailNotifications": true,|"SendEmailNotifications": false,|g' dist/mattermost/config/config.json \
    && sed -i 's|"FeedbackEmail": "test@example.com",|"FeedbackEmail": "",|g' dist/mattermost/config/config.json \
    && sed -i 's|"ReplyToAddress": "test@example.com",|"ReplyToAddress": "",|g' dist/mattermost/config/config.json \
    && sed -i 's|"SMTPServer": "localhost",|"SMTPServer": "",|g' dist/mattermost/config/config.json \
    && sed -i 's|"SMTPPort": "10025",|"SMTPPort": "",|g' dist/mattermost/config/config.json \
    && chmod 600 dist/mattermost/config/config.json \
    && mkdir -p dist/mattermost/client dist/mattermost/bin dist/mattermost/logs dist/mattermost/plugins dist/mattermost/data dist/mattermost/client/plugins \
    && cp -RL ../webapp/channels/dist/* dist/mattermost/client/ \
    && cp "bin/${TARGETOS}_${TARGETARCH}/mattermost" "bin/${TARGETOS}_${TARGETARCH}/mmctl" dist/mattermost/bin/

# ---------------------------------------------------------------------------
# Stage 3: runtime
# ---------------------------------------------------------------------------
FROM ${RUNTIME_IMAGE} AS runtime

# Document preview helpers, certs, mime types and tzdata used by the server.
RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        ca-certificates \
        file \
        media-types \
        poppler-utils \
        tidy \
        tzdata \
        unrtf \
        wv \
    && rm -rf /var/lib/apt/lists/* \
    && groupadd --gid 2000 mattermost \
    && useradd --uid 2000 --gid 2000 --home-dir /mattermost --shell /usr/sbin/nologin mattermost

ENV PATH="/mattermost/bin:${PATH}" \
    MM_SERVICESETTINGS_ENABLELOCALMODE="true" \
    MM_INSTALL_TYPE="docker"

COPY --from=builder /src/server/dist/mattermost /mattermost
# Verify binary architecture matches target platform during build
RUN file /mattermost/bin/mattermost && chown -R mattermost:mattermost /mattermost

USER mattermost
WORKDIR /mattermost

HEALTHCHECK --interval=30s --timeout=10s \
    CMD ["/mattermost/bin/mmctl", "system", "status", "--local"]

CMD ["/mattermost/bin/mattermost"]

EXPOSE 8065 8067 8074

VOLUME ["/mattermost/data", "/mattermost/logs", "/mattermost/config", "/mattermost/plugins", "/mattermost/client/plugins"]
