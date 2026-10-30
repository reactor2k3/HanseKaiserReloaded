# HanseKaiserReloaded – ein Image mit gemeinsamem Spielserver, beiden Clients und Startseite.
#   docker build -t hansekaiserreloaded .
#   docker run -d -p 8080:8080 -v hk-data:/var/lib/hansekaiser hansekaiserreloaded
# Mehrstufig: Clients mit Node bauen, Server mit Go bauen, dann nur die Ergebnisse in ein
# minimales Laufzeit-Image (distroless, ohne Shell, läuft als Nicht-root).

# 1. Clients (TypeScript/Svelte) bauen; Vite schreibt nach /src/dist/kaiser bzw. /src/dist/hanse
FROM node:24-alpine AS client
WORKDIR /src/clients/kaiser
COPY clients/kaiser/package.json clients/kaiser/package-lock.json ./
RUN npm ci --no-audit --no-fund
WORKDIR /src/clients/hanse
COPY clients/hanse/package.json clients/hanse/package-lock.json ./
RUN npm ci --no-audit --no-fund
WORKDIR /src
COPY clients/ ./clients/
RUN cd clients/kaiser && npm run build && cd ../hanse && npm run build
# Startseite ist reines HTML/CSS/JS ohne Build-Schritt
RUN mkdir -p dist/lobby && cp clients/lobby/* dist/lobby/

# 2. Server (Go) bauen – reiner Go-SQLite-Treiber, daher ohne cgo statisch
FROM golang:1.27-alpine AS server
WORKDIR /src/server
COPY server/go.mod server/go.sum ./
RUN go mod download
COPY server/ ./
# timetzdata: Zeitzonendaten ins Programm, damit TZ (z. B. Europe/Berlin) im distroless-Image wirkt
RUN CGO_ENABLED=0 go build -trimpath -tags timetzdata -ldflags="-s -w" -o /out/hkd ./cmd/hkd \
 && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/healthcheck ./cmd/healthcheck
# Datenordner vorbereiten, damit das Volume beim ersten Start dem Nicht-root-Nutzer gehört
RUN mkdir -p /out/var/lib/hansekaiser/media

# 3. Laufzeit-Image
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=server /out/hkd /out/healthcheck /app/
COPY --from=client /src/dist /app/public
COPY data/ /app/data/
COPY --from=server --chown=65532:65532 /out/var/lib/hansekaiser /var/lib/hansekaiser
VOLUME ["/var/lib/hansekaiser"]
EXPOSE 8080
USER nonroot
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s CMD ["/app/healthcheck"]
# Clients und WebSockets kommen vom selben Server und Port, daher keine fremden Origins (-origins leer).
# hkd setzt Sicherheits-Header, gzip und Cache-Regeln selbst (internal/*/server/web.go).
ENTRYPOINT ["/app/hkd", "-addr", "0.0.0.0:8080", "-data", "/app/data", "-static", "/app/public", \
            "-state", "/var/lib/hansekaiser", "-origins", ""]
