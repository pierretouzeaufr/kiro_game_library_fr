# 🐳 Déploiement Docker

## 📋 **Prérequis**
- Docker installé sur le système
- Docker Compose (optionnel, mais recommandé)

## 🚀 **Démarrage avec Docker Compose (Recommandé)**

```bash
# Construire et démarrer l'application
docker-compose up -d

# Voir les logs
docker-compose logs -f

# Arrêter l'application
docker-compose down
```

## 🔧 **Démarrage avec Docker uniquement**

```bash
# Construire l'image
docker build -t board-game-library .

# Démarrer le conteneur
docker run -d \
  --name board-game-library \
  -p 8080:8080 \
  -v $(pwd)/data:/app/data \
  board-game-library

# Voir les logs
docker logs -f board-game-library

# Arrêter le conteneur
docker stop board-game-library
docker rm board-game-library
```

## 🌐 **Accès**

L'application sera accessible à : http://localhost:8080

## 💾 **Persistance des Données**

Les données sont automatiquement sauvegardées dans le dossier `./data` sur votre machine hôte.