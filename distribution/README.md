# 🎲 Bibliothèque de Jeux de Société

## 📋 Installation et Utilisation

### ✅ **Prérequis**
- **Aucun !** Cette application est un binaire autonome
- Fonctionne sur macOS, Linux et Windows
- Aucune installation de Go requise

### 🚀 **Démarrage Rapide**

#### **Sur macOS/Linux :**
```bash
# Rendre le script exécutable
chmod +x start.sh

# Démarrer l'application
./start.sh
```

#### **Sur Windows :**
```cmd
# Double-cliquer sur start.bat
# Ou depuis l'invite de commande :
start.bat
```

#### **Démarrage Manuel :**
```bash
# Directement avec le binaire
./board-game-library        # macOS/Linux
board-game-library.exe      # Windows
```

### 🌐 **Accès à l'Application**

Une fois démarrée, l'application est accessible à :
- **Page d'accueil :** http://localhost:8080/
- **Guide utilisateur :** http://localhost:8080/guide

### 📚 **Interfaces Disponibles**

| Interface | URL | Description |
|-----------|-----|-------------|
| 🏠 **Accueil** | http://localhost:8080/ | Page principale avec navigation |
| 🎲 **Jeux** | http://localhost:8080/games | Gestion de la collection de jeux |
| 👥 **Utilisateurs** | http://localhost:8080/users | Gestion des membres |
| 📚 **Emprunts** | http://localhost:8080/borrowings | Suivi des prêts de jeux |
| 🚨 **Alertes** | http://localhost:8080/alerts | Notifications et rappels |
| 📖 **Guide** | http://localhost:8080/guide | Guide d'utilisation complet |
| 🔧 **API Swagger** | http://localhost:8080/swagger/index.html | Documentation API |

### 💾 **Données**

- Les données sont stockées dans le dossier `data/library.db`
- Base de données SQLite (aucune configuration requise)
- Sauvegarde automatique de toutes les données

### 🛑 **Arrêt de l'Application**

- Appuyez sur `Ctrl+C` dans le terminal
- Ou fermez simplement la fenêtre du terminal

### 🔧 **Dépannage**

#### **Port déjà utilisé**
Si le port 8080 est occupé, l'application affichera une erreur. Dans ce cas :
1. Fermez l'autre application utilisant le port 8080
2. Ou modifiez le port dans la configuration

#### **Problème de permissions**
Sur macOS/Linux, si vous avez une erreur de permission :
```bash
chmod +x board-game-library
chmod +x start.sh
```

#### **Base de données corrompue**
En cas de problème avec la base de données :
1. Arrêtez l'application
2. Supprimez le fichier `data/library.db`
3. Redémarrez l'application (une nouvelle base sera créée)

### 📞 **Support**

- Consultez le guide intégré : http://localhost:8080/guide
- Documentation API : http://localhost:8080/swagger/index.html

### 🎯 **Fonctionnalités Principales**

- ✅ **Gestion complète** de la collection de jeux
- ✅ **Suivi des emprunts** avec échéances automatiques
- ✅ **Système d'alertes** intelligent
- ✅ **Interface web** intuitive en français
- ✅ **API REST** complète avec documentation Swagger
- ✅ **Aucune installation** requise
- ✅ **Base de données** intégrée (SQLite)

---

**Version :** 1.0  
**Plateforme :** Multi-plateforme (macOS, Linux, Windows)  
**Licence :** MIT