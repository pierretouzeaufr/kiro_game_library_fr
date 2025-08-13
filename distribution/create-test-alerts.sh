#!/bin/bash

echo "🚨 Création d'alertes de test"
echo "============================="

# Vérifier que le serveur fonctionne
if ! curl -s http://localhost:8080/ > /dev/null; then
    echo "❌ Le serveur n'est pas démarré. Lancez d'abord ./start.sh"
    exit 1
fi

echo "📡 Serveur détecté sur http://localhost:8080"
echo ""

# Créer une alerte personnalisée
echo "📝 Création d'une alerte personnalisée..."
curl -X POST http://localhost:8080/api/v1/alerts \
  -H "Content-Type: application/json" \
  -d '{"user_id": 1, "game_id": 1, "type": "custom", "message": "Alerte de test : Veuillez vérifier l état du jeu avant le prochain emprunt."}' \
  -s > /dev/null

# Créer une alerte de rappel
echo "⏰ Création d'une alerte de rappel..."
curl -X POST http://localhost:8080/api/v1/alerts \
  -H "Content-Type: application/json" \
  -d '{"user_id": 1, "game_id": 1, "type": "reminder", "message": "Rappel : Pensez à nettoyer les pièces du jeu avant de le rendre."}' \
  -s > /dev/null

# Créer une alerte de retard
echo "🔴 Création d'une alerte de retard..."
curl -X POST http://localhost:8080/api/v1/alerts \
  -H "Content-Type: application/json" \
  -d '{"user_id": 1, "game_id": 1, "type": "overdue", "message": "Game Monopoly is overdue by 3 day(s). Please return it as soon as possible."}' \
  -s > /dev/null

echo ""
echo "✅ Alertes de test créées avec succès !"
echo ""
echo "🌐 Vous pouvez maintenant voir les alertes sur :"
echo "   http://localhost:8080/alerts"
echo ""
echo "📊 Types d'alertes créées :"
echo "   🔴 overdue   - Jeu en retard"
echo "   ⏰ reminder  - Rappel"
echo "   📝 custom    - Alerte personnalisée"
echo ""

# Afficher le nombre d'alertes
ALERT_COUNT=$(curl -s http://localhost:8080/api/v1/alerts | grep -o '"count":[0-9]*' | cut -d':' -f2)
echo "📈 Total des alertes : $ALERT_COUNT"