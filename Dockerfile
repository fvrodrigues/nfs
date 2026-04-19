# syntax=docker/dockerfile:1.7

# ---------- build stage ----------
FROM golang:1.25.4-bookworm AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/nfse ./cmd/main

# ---------- chromium stage ----------
# Rod uses a pinned Chromium snapshot. Download it once here so the
# runtime image has a deterministic binary at the hardcoded path expected
# by pkg/rod/rod.go (overridable via CHROMIUM_BIN).
FROM debian:bookworm-slim AS chromium
ARG CHROMIUM_REV=1321438
RUN apt-get update \
 && apt-get install -y --no-install-recommends ca-certificates curl unzip \
 && rm -rf /var/lib/apt/lists/* \
 && mkdir -p /opt/chromium \
 && curl -fsSL -o /tmp/chrome.zip \
      "https://storage.googleapis.com/chromium-browser-snapshots/Linux_x64/${CHROMIUM_REV}/chrome-linux.zip" \
 && unzip -q /tmp/chrome.zip -d /opt/chromium/ \
 && rm /tmp/chrome.zip \
 && chmod -R a+rx /opt/chromium/chrome-linux

# ---------- runtime stage ----------
FROM debian:bookworm-slim AS runtime

ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update && apt-get install -y --no-install-recommends \
      ca-certificates \
      xvfb \
      fonts-liberation \
      libasound2 libatk-bridge2.0-0 libatk1.0-0 libatspi2.0-0 \
      libcairo2 libcups2 libdbus-1-3 libdrm2 libexpat1 libgbm1 \
      libglib2.0-0 libnspr4 libnss3 libpango-1.0-0 libx11-6 \
      libx11-xcb1 libxcb1 libxcomposite1 libxdamage1 libxext6 \
      libxfixes3 libxkbcommon0 libxrandr2 libxshmfence1 libu2f-udev \
      libvulkan1 libgtk-3-0 \
 && rm -rf /var/lib/apt/lists/* \
 && groupadd --system nfse \
 && useradd  --system --gid nfse --home-dir /opt/nfse --create-home nfse

COPY --from=chromium /opt/chromium /opt/chromium

WORKDIR /opt/nfse
COPY --from=build /out/nfse /opt/nfse/nfse

# Default port so `docker run -p 8080:8080` works. Overridable with PORT env.
ENV PORT=8080
# Default to headful Chromium driven by Xvfb. The SP Login Único portal
# (pmspauth.prefeitura.sp.gov.br) resets connections for true-headless
# Chromium, so any container running with NFSE_HEADLESS=1 fails to reach
# the login form at all (ERR_CONNECTION_RESET). Headful + Xvfb looks like
# a regular browser from the network's perspective and gets through.
ENV NFSE_HEADLESS=0
ENV CHROMIUM_BIN=/opt/chromium/chrome-linux/chrome
EXPOSE 8080

RUN chown -R nfse:nfse /opt/nfse
USER nfse

# Xvfb is present so users can flip NFSE_HEADLESS=0 without rebuilding.
ENTRYPOINT ["/bin/sh","-c","if [ \"${NFSE_HEADLESS}\" = \"0\" ] || [ \"${NFSE_HEADLESS}\" = \"false\" ]; then Xvfb :99 -screen 0 1920x1080x24 -nolisten tcp & export DISPLAY=:99; fi; exec /opt/nfse/nfse"]
