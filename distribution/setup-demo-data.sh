#!/bin/bash

echo "🎲 Configuration des données de démonstration"
echo "=============================================="

# Vérifier que le serveur fonctionne
if ! curl -s http://localhost:8080/ > /dev/null; then
    echo "❌ Le serveur n'est pas démarré. Lancez d'abord ./start.sh"
    exit 1
fi

echo "📡 Serveur détecté sur http://localhost:8080"
echo ""

# Créer des utilisateurs de test
echo "👥 Création d'utilisateurs de test..."
curl -X POST http://localhost:8080/users/create \
  -d "name=Pierre Touzeau&email=pierre@example.com" \
  -s > /dev/null

curl -X POST http://localhost:8080/users/create \
  -d "name=Marie Dupont&email=marie@example.com" \
  -s > /dev/null

echo "✅ Utilisateurs créés"

# Créer des jeux de test
echo "🎮 Création de jeux de test..."
curl -X POST http://localhost:8080/games/create \
  -d "name=Monopoly&category=Stratégie&description=Jeu de société classique&condition=excellent" \
  -s > /dev/null

curl -X POST http://localhost:8080/games/create \
  -d "name=Scrabble&category=Mots&description=Jeu de formation de mots&condition=bon" \
  -s > /dev/null

curl -X POST http://localhost:8080/games/create \
  -d "name=Catan&category=Stratégie&description=Les Colons de Catane&condition=excellent" \
  -s > /dev/null

echo "✅ Jeux créés"

# Attendre un peu pour que les données soient bien enregistrées
sleep 1

# Créer des emprunts de test
echo "📚 Création d'emprunts de test..."
curl -X POST http://localhost:8080/borrowings/create \
  -d "user_id=1&game_id=1&duration_days=14" \
  -s > /dev/null

curl -X POST http://localhost:8080/borrowings/create \
  -d "user_id=2&game_id=2&duration_days=7" \
  -s > /dev/null

echo "✅ Emprunts créés"

# Créer des alertes de test
echo "🚨 Création d'alertes de test..."

# Alerte personnalisée
curl -X POST http://localhost:8080/api/v1/alerts \
  -H "Content-Type: application/json" \
  -d '{"user_id": 1, "game_id": 1, "type": "custom", "message": "Alerte de test : Veuillez vérifier l état du jeu avant le prochain emprunt."}' \
  -s > /dev/null

# Alerte de rappel
curl -X POST http://localhost:8080/api/v1/alerts \
  -H "Content-Type: application/json" \
  -d '{"user_id": 1, "game_id": 1, "type": "reminder", "message": "Rappel : Le jeu Monopoly doit être rendu dans 2 jours."}' \
  -s > /dev/null

# Alerte de retard
curl -X POST http://localhost:8080/api/v1/alerts \
  -H "Content-Type: application/json" \
  -d '{"user_id": 2, "game_id": 2, "type": "overdue", "message": "Game Scrabble is overdue by 3 day(s). Please return it as soon as possible."}' \
  -s > /dev/null

echo "✅ Alertes créées"

echo ""
echo "🎉 Configuration terminée !"
echo ""

# Afficher un résumé
USER_COUNT=$(curl -s http://localhost:8080/api/v1/users | grep -o '"count":[0-9]*' | cut -d':' -f2)
GAME_COUNT=$(curl -s http://localhost:8080/api/v1/games | grep -o '"count":[0-9]*' | cut -d':' -f2)
ALERT_COUNT=$(curl -s http://localhost:8080/api/v1/alerts | grep -o '"count":[0-9]*' | cut -d':' -f2)

echo "📊 Résumé des données créées :"
echo "   👥 Utilisateurs : $USER_COUNT"
echo "   🎮 Jeux : $GAME_COUNT"
echo "   🚨 Alertes : $ALERT_COUNT"
echo ""
echo "🌐 Vous pouvez maintenant explorer :"
echo "   - Accueil : http://localhost:8080/"
echo "   - Utilisateurs : http://localhost:8080/users"
echo "   - Jeux : http://localhost:8080/games"
echo "   - Emprunts : http://localhost:8080/borrowings"
echo "   - Alertes : http://localhost:8080/alerts"