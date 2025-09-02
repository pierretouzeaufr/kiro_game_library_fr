# Dockerfile pour la Bibliothèque de Jeux de Société
FROM golang:1.24-alpine AS builder

# Installer les dépendances de build (SQLite nécessite CGO)
RUN apk add --no-cache git gcc musl-dev sqlite-dev

# Définir le répertoire de travail
WORKDIR /app

# Copier les fichiers de dépendances
COPY go.mod go.sum ./
RUN go mod download

# Copier le code source
COPY . .

# Construire l'application avec CGO pour SQLite
RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags="-s -w" -o board-game-library ./cmd/server

# Image finale
FROM alpine:latest

# Installer SQLite
RUN apk --no-cache add ca-certificates sqlite

# Créer un utilisateur non-root
RUN adduser -D -s /bin/sh appuser

# Définir le répertoire de travail
WORKDIR /app

# Copier le binaire depuis l'étape de build
COPY --from=builder /app/board-game-library .
COPY --from=builder /app/web ./web
COPY --from=builder /app/docs ./docs

# Créer le dossier data
RUN mkdir -p data && chown -R appuser:appuser /app

# Changer vers l'utilisateur non-root
USER appuser

# Exposer le port
EXPOSE 8080

# Commande de démarrage
CMD ["./board-game-library"]