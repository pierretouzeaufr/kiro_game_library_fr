#!/bin/bash

echo "🔍 Test de l'interface avec durée d'emprunt"
echo "=========================================="

# Démarrer le serveur
echo "📡 Démarrage du serveur..."
./board-game-library &
SERVER_PID=$!

# Attendre que le serveur démarre
sleep 3

echo "🌐 Serveur démarré sur http://localhost:8080"
echo ""

# Tester la page des emprunts
echo "🧪 Test de la page des emprunts..."
if curl -s http://localhost:8080/borrowings | grep -q "Durée d'emprunt"; then
    echo "✅ Le champ 'Durée d'emprunt' est présent dans le HTML"
else
    echo "❌ Le champ 'Durée d'emprunt' n'est PAS trouvé"
fi

# Vérifier les options de durée
echo ""
echo "🔍 Options de durée disponibles :"
curl -s http://localhost:8080/borrowings | grep -o 'value="[0-9]*">[0-9]* jours' | head -5

echo ""
echo "🌐 Ouvre ton navigateur sur : http://localhost:8080/borrowings"
echo "📝 Instructions :"
echo "   1. Vide le cache de ton navigateur (Ctrl+F5 ou Cmd+Shift+R)"
echo "   2. Va sur la page des emprunts"
echo "   3. Tu devrais voir un champ 'Durée d'emprunt' avec 5 options"
echo ""
echo "⏹️  Appuie sur Ctrl+C pour arrêter le serveur"

# Attendre l'arrêt manuel
wait $SERVER_PID