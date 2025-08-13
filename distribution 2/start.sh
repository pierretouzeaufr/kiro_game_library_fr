#!/bin/bash

# Script de démarrage pour la Bibliothèque de Jeux de Société
# Board Game Library Startup Script

echo "🎲 Démarrage de la Bibliothèque de Jeux de Société..."
echo "   Starting Board Game Library..."
echo ""

# Créer le dossier data s'il n'existe pas
mkdir -p data

# Démarrer l'application
echo "📚 Serveur démarré sur : http://localhost:8080"
echo "   Server started at: http://localhost:8080"
echo ""
echo "🌐 Interfaces disponibles :"
echo "   - Page d'accueil : http://localhost:8080/"
echo "   - Jeux : http://localhost:8080/games"
echo "   - Utilisateurs : http://localhost:8080/users"
echo "   - Emprunts : http://localhost:8080/borrowings (avec durée personnalisable)"
echo "   - Alertes : http://localhost:8080/alerts"
echo "   - Guide : http://localhost:8080/guide"
echo "   - API Swagger : http://localhost:8080/swagger/index.html"
echo ""
echo "⏹️  Pour arrêter : Ctrl+C"
echo ""

./board-game-library