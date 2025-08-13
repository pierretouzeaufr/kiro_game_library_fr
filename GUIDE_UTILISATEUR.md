# 🎲 Guide Utilisateur - Bibliothèque de Jeux de Société

## 🏠 Page d'Accueil

**URL :** `http://localhost:8081/`

La page d'accueil vous donne accès à toutes les fonctionnalités :

```
┌─────────────────────────────────────────┐
│  🎲 Bibliothèque de Jeux de Société     │
├─────────────────────────────────────────┤
│  [🎲 Jeux]  [👥 Utilisateurs]          │
│  [📚 Emprunts]  [🚨 Alertes]           │
└─────────────────────────────────────────┘
```

---

## 🎲 Gestion des Jeux (`/games`)

### **Vue d'ensemble**
- **Titre :** "Collection de Jeux" (bleu)
- **Compteur :** Affiche le nombre total de jeux
- **Couleur thématique :** Bleu

### **Sections de la page :**

#### **1. Formulaire d'Ajout**
```
📚 Ajouter un Nouveau Jeu
┌─────────────────────────────────────────┐
│ Nom du Jeu * : [___________________]    │
│ Catégorie *  : [___________________]    │
│ Description  : [___________________]    │
│ État *       : [Excellent ▼]           │
│              [➕ Ajouter le Jeu]       │
└─────────────────────────────────────────┘
```

#### **2. Liste des Jeux**
Chaque jeu s'affiche dans une carte :
```
┌─────────────────────────────────────────┐
│ Monopoly                    [🗑️ Supprimer] │
│ Classic property trading board game     │
│ Catégorie : Strategy    ✅ Disponible   │
│ État : excellent | Ajouté : 2025-08-12 │
└─────────────────────────────────────────┘
```

**Statuts possibles :**
- ✅ **Disponible** (vert) : Peut être emprunté
- 🔴 **Emprunté** (rouge) : Actuellement prêté

---

## 👥 Gestion des Utilisateurs (`/users`)

### **Vue d'ensemble**
- **Titre :** "Membres de la Bibliothèque" (vert)
- **Compteur :** Affiche le nombre total d'utilisateurs
- **Couleur thématique :** Vert

### **Sections de la page :**

#### **1. Formulaire d'Inscription**
```
👤 Inscrire un Nouveau Membre
┌─────────────────────────────────────────┐
│ Nom Complet * : [___________________]   │
│ Adresse Email*: [___________________]   │
│              [➕ Inscrire le Membre]    │
└─────────────────────────────────────────┘
```

#### **2. Liste des Membres**
Chaque utilisateur s'affiche dans une carte :
```
┌─────────────────────────────────────────┐
│ Pierre Touzeau          [🗑️ Supprimer]  │
│ 📧 pierre@example.com                   │
│ ID : 1                  ✅ Actif        │
│ Inscrit : 2025-08-12                    │
└─────────────────────────────────────────┘
```

**Statuts possibles :**
- ✅ **Actif** (vert) : Peut emprunter des jeux
- 🔴 **Inactif** (rouge) : Compte désactivé

---

## 📚 Gestion des Emprunts (`/borrowings`)

### **Vue d'ensemble**
- **Titre :** "Gestion des Emprunts" (jaune)
- **Compteur :** Affiche le nombre total d'emprunts
- **Couleur thématique :** Jaune

### **Sections de la page :**

#### **1. Formulaire de Création**
```
📚 Créer un Nouvel Emprunt
┌─────────────────────────────────────────┐
│ Utilisateur * : [Sélectionner ▼]       │
│ Jeu Disponible*:[Sélectionner ▼]       │
│              [➕ Créer l'Emprunt]       │
└─────────────────────────────────────────┘
```

#### **2. Liste des Emprunts**
Chaque emprunt s'affiche dans une carte détaillée :
```
┌─────────────────────────────────────────┐
│ Emprunt #5    🟡 En cours  [✅ Retourner] │
│ Utilisateur : Marie Dupont (marie@...)  │
│ Jeu : Monopoly (Strategy)               │
│ Emprunté : 2025-08-13                   │
│ Échéance : 2025-08-27                   │
└─────────────────────────────────────────┘
```

**Statuts possibles :**
- 🟡 **En cours** (jaune) : Emprunt actif avec bouton "Retourner"
- ✅ **Retourné** (vert) : Jeu rendu, avec date de retour
- 🔴 **En retard** (rouge) : Dépassement de l'échéance

#### **3. Processus d'Emprunt**
1. **Sélectionner un utilisateur** actif
2. **Choisir un jeu** disponible
3. **Cliquer "Créer l'Emprunt"**
4. **Échéance automatique** : 14 jours
5. **Le jeu devient indisponible**

#### **4. Processus de Retour**
1. **Cliquer "Retourner"** sur un emprunt actif
2. **Confirmation automatique**
3. **Le jeu redevient disponible**
4. **L'emprunt passe en "Retourné"**

---

## 🚨 Gestion des Alertes (`/alerts`)

### **Vue d'ensemble**
- **Titre :** "Gestion des Alertes" (rouge)
- **Compteur :** Affiche le nombre d'alertes actives
- **Couleur thématique :** Rouge

### **Sections de la page :**

#### **1. Actions Rapides**
```
⚡ Actions Rapides
┌─────────────────────────────────────────┐
│ [⚠️ Générer Alertes Retard]            │
│ [⏰ Générer Rappels]                    │
│ [🧹 Nettoyer Alertes Résolues]         │
└─────────────────────────────────────────┘
```

**Fonctions :**
- **Générer Alertes Retard** : Crée des alertes pour tous les emprunts en retard
- **Générer Rappels** : Crée des rappels pour les emprunts dus dans 2 jours
- **Nettoyer Alertes Résolues** : Supprime les alertes pour les emprunts retournés

#### **2. Formulaire d'Alerte Personnalisée**
```
📝 Créer une Alerte Personnalisée
┌─────────────────────────────────────────┐
│ Utilisateur * : [Sélectionner ▼]       │
│ Jeu *        : [Sélectionner ▼]        │
│ Message *    : [___________________]    │
│              [___________________]      │
│              [➕ Créer l'Alerte]        │
└─────────────────────────────────────────┘
```

#### **3. Liste des Alertes Actives**
Chaque alerte s'affiche avec son type et ses actions :
```
┌─────────────────────────────────────────┐
│ 📝 Alerte #4  🟣 custom  [✅ Marquer comme lue] │
│                          [🗑️ Supprimer]        │
│ Rappel: Veuillez retourner le jeu...    │
│ Utilisateur : Pierre Touzeau (pierre@...) │
│ Jeu : Monopoly (Strategy)               │
│ Créée : 2025-08-13 09:37                │
└─────────────────────────────────────────┘
```

**Types d'alertes :**
- ⚠️ **overdue** (rouge) : Jeu en retard
- ⏰ **reminder** (jaune) : Rappel d'échéance proche
- 📝 **custom** (violet) : Alerte personnalisée

#### **4. Actions sur les Alertes**
- **Marquer comme lue** : L'alerte disparaît de la liste active
- **Supprimer** : Suppression définitive (avec confirmation)

---

## 🎯 Flux de Travail Typique

### **Scénario 1 : Nouvel Emprunt**
1. **Aller sur `/borrowings`**
2. **Remplir le formulaire** "Créer un Nouvel Emprunt"
3. **Sélectionner utilisateur et jeu**
4. **Cliquer "Créer l'Emprunt"**
5. **Voir la confirmation** avec détails
6. **L'emprunt apparaît** dans la liste avec statut "En cours"

### **Scénario 2 : Retour de Jeu**
1. **Aller sur `/borrowings`**
2. **Trouver l'emprunt** avec statut "En cours"
3. **Cliquer "Retourner"**
4. **Voir la confirmation** de retour
5. **L'emprunt passe** en statut "Retourné"
6. **Le jeu redevient** disponible

### **Scénario 3 : Gestion des Alertes**
1. **Aller sur `/alerts`**
2. **Cliquer "Générer Alertes Retard"** pour créer des alertes automatiques
3. **Voir les nouvelles alertes** dans la liste
4. **Traiter chaque alerte** : marquer comme lue ou supprimer
5. **Utiliser "Nettoyer Alertes Résolues"** pour la maintenance

---

## 🎨 Codes Couleurs

### **Statuts des Emprunts**
- 🟡 **Jaune** : En cours (normal)
- ✅ **Vert** : Retourné (terminé)
- 🔴 **Rouge** : En retard (attention requise)

### **Types d'Alertes**
- 🔴 **Rouge** : Alertes de retard (urgent)
- 🟡 **Jaune** : Rappels (bientôt dû)
- 🟣 **Violet** : Alertes personnalisées

### **Disponibilité des Jeux**
- ✅ **Vert** : Disponible pour emprunt
- 🔴 **Rouge** : Actuellement emprunté

---

## 💡 Conseils d'Utilisation

### **Bonnes Pratiques**
1. **Vérifiez régulièrement** la page des alertes
2. **Générez les rappels** avant les échéances
3. **Nettoyez les alertes** résolues périodiquement
4. **Utilisez les alertes personnalisées** pour communiquer avec les utilisateurs

### **Navigation**
- **Bouton "Retour à l'accueil"** sur chaque page
- **Messages de confirmation** avec redirection automatique
- **Validation des formulaires** avec messages d'erreur clairs

### **Maintenance**
- **Actions rapides** pour la gestion automatisée
- **Historique complet** des emprunts conservé
- **Intégrité des données** protégée par les contraintes

---

## 🔧 Dépannage

### **Problèmes Courants**

#### **"Impossible de supprimer ce jeu"**
- **Cause :** Le jeu a un historique d'emprunts
- **Solution :** C'est normal pour préserver l'intégrité des données

#### **"Jeu non disponible pour emprunt"**
- **Cause :** Le jeu est actuellement emprunté
- **Solution :** Attendre le retour ou vérifier les emprunts actifs

#### **"Aucune alerte active"**
- **Cause :** Toutes les alertes ont été traitées
- **Solution :** Utiliser les boutons "Actions Rapides" pour générer de nouvelles alertes

Cette interface vous permet de gérer efficacement votre bibliothèque de jeux de société avec un suivi complet des emprunts et un système d'alertes intelligent ! 🎲