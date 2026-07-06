# ============================================================================
# Section 9.1 : Dockerfile multi-stage, images distroless / scratch
# Description : Le Dockerfile multi-stage de la section — étape builder
#               (toute la chaîne Go), étape finale distroless (le seul
#               binaire). Cache de couches : go.mod copié AVANT le code.
#               Résultat constaté : ~8 Mo, non-root, sans shell.
# Fichier source : 01-docker.md
# ============================================================================

# syntax=docker/dockerfile:1

# --- Étape 1 : construction ---
FROM golang:1.26 AS builder
WORKDIR /src

# Dépendances d'abord : cette couche reste en cache tant que go.mod ne change pas.
COPY go.mod ./
RUN go mod download

# Puis le code source.
COPY . .

# Binaire statique (CGO_ENABLED=0, § 6.4), aminci et sans chemins locaux.
ARG VERSION=dev
RUN CGO_ENABLED=0 go build \
      -trimpath \
      -ldflags="-s -w -X main.version=${VERSION}" \
      -o /out/server ./cmd/server

# --- Étape 2 : image finale minimale ---
FROM gcr.io/distroless/static-debian13:nonroot
COPY --from=builder /out/server /server
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/server"]
