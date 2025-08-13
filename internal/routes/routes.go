package routes

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	ginSwagger "github.com/swaggo/gin-swagger"
	swaggerFiles "github.com/swaggo/files"

	"board-game-library/internal/handlers"
	"board-game-library/internal/repositories"
	"board-game-library/internal/services"
	"board-game-library/pkg/database"
)

// SetupRoutes configures all application routes
func SetupRoutes(router *gin.Engine, db *database.DB) error {
	// Setup template functions
	setupTemplateFunctions(router)
	
	// Serve static files
	router.Static("/static", "./web/static")

	// Initialize repositories
	gameRepo := repositories.NewSQLiteGameRepository(db)
	userRepo := repositories.NewSQLiteUserRepository(db)
	borrowingRepo := repositories.NewSQLiteBorrowingRepository(db)
	alertRepo := repositories.NewSQLiteAlertRepository(db)

	// Initialize services
	gameService := services.NewGameService(gameRepo, borrowingRepo)
	userService := services.NewUserService(userRepo, borrowingRepo)
	borrowingService := services.NewBorrowingService(borrowingRepo, userRepo, gameRepo)
	alertService := services.NewAlertService(alertRepo, borrowingRepo, userRepo, gameRepo)

	// Initialize API handlers
	gameHandler := handlers.NewGameHandler(gameService)
	userHandler := handlers.NewUserHandler(userService)
	borrowingHandler := handlers.NewBorrowingHandler(borrowingService)
	alertHandler := handlers.NewAlertHandler(alertService)

	// Guide utilisateur route
	router.GET("/guide", func(c *gin.Context) {
		c.Header("Content-Type", "text/html")
		c.String(http.StatusOK, `
<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Guide Utilisateur - Bibliothèque de Jeux de Société</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <style>
        .section-card { transition: all 0.3s ease; }
        .section-card:hover { transform: translateY(-2px); box-shadow: 0 10px 25px rgba(0,0,0,0.1); }
        .status-badge { display: inline-block; padding: 4px 8px; border-radius: 12px; font-size: 12px; font-weight: 600; }
        .status-available { background: #dcfce7; color: #166534; }
        .status-borrowed { background: #fecaca; color: #991b1b; }
        .status-active { background: #fef3c7; color: #92400e; }
        .nav-sticky { position: sticky; top: 20px; max-height: calc(100vh - 40px); overflow-y: auto; }
    </style>
</head>
<body class="bg-gray-50">
    <!-- Header -->
    <header class="bg-gradient-to-r from-blue-600 to-purple-600 text-white py-8">
        <div class="container mx-auto px-4">
            <div class="text-center">
                <h1 class="text-4xl font-bold mb-2">🎲 Guide Utilisateur</h1>
                <p class="text-xl opacity-90">Bibliothèque de Jeux de Société</p>
                <p class="text-sm opacity-75 mt-2">Apprenez à utiliser toutes les fonctionnalités de l'interface</p>
                <a href="/" class="inline-block mt-4 bg-white text-blue-600 px-4 py-2 rounded-lg font-semibold hover:bg-gray-100 transition-colors">← Retour à l'accueil</a>
            </div>
        </div>
    </header>

    <div class="container mx-auto px-4 py-8">
        <div class="max-w-6xl mx-auto">
            <!-- Navigation rapide -->
            <div class="bg-white rounded-lg shadow-lg p-6 mb-8">
                <h2 class="text-2xl font-bold text-gray-800 mb-4">🚀 Accès Rapide aux Interfaces</h2>
                <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
                    <a href="/games" class="bg-blue-500 hover:bg-blue-600 text-white p-4 rounded-lg text-center transition-colors">
                        <div class="text-2xl mb-2">🎲</div>
                        <div class="font-semibold">Jeux</div>
                        <div class="text-xs opacity-75">Gérer la collection</div>
                    </a>
                    <a href="/users" class="bg-green-500 hover:bg-green-600 text-white p-4 rounded-lg text-center transition-colors">
                        <div class="text-2xl mb-2">👥</div>
                        <div class="font-semibold">Utilisateurs</div>
                        <div class="text-xs opacity-75">Gérer les membres</div>
                    </a>
                    <a href="/borrowings" class="bg-yellow-500 hover:bg-yellow-600 text-white p-4 rounded-lg text-center transition-colors">
                        <div class="text-2xl mb-2">📚</div>
                        <div class="font-semibold">Emprunts</div>
                        <div class="text-xs opacity-75">Suivre les prêts</div>
                    </a>
                    <a href="/alerts" class="bg-red-500 hover:bg-red-600 text-white p-4 rounded-lg text-center transition-colors">
                        <div class="text-2xl mb-2">🚨</div>
                        <div class="font-semibold">Alertes</div>
                        <div class="text-xs opacity-75">Voir les notifications</div>
                    </a>
                </div>
            </div>

            <!-- Guide des Jeux -->
            <div class="section-card bg-white rounded-lg shadow-lg p-8 mb-8">
                <h2 class="text-3xl font-bold text-blue-600 mb-6">🎲 Gestion des Jeux</h2>
                <p class="text-gray-600 mb-4"><strong>URL :</strong> <code class="bg-gray-100 px-2 py-1 rounded">http://localhost:8081/games</code></p>
                
                <div class="grid md:grid-cols-2 gap-6">
                    <div>
                        <h3 class="text-xl font-semibold text-blue-600 mb-4">📚 Comment ajouter un jeu</h3>
                        <ol class="list-decimal list-inside space-y-2 text-sm text-gray-700">
                            <li>Remplir le nom du jeu (obligatoire)</li>
                            <li>Sélectionner la catégorie (Stratégie, Famille, etc.)</li>
                            <li>Ajouter une description (optionnel)</li>
                            <li>Choisir l'état du jeu (Excellent, Bon, etc.)</li>
                            <li>Cliquer "Ajouter le Jeu"</li>
                        </ol>
                    </div>
                    <div>
                        <h3 class="text-xl font-semibold text-blue-600 mb-4">📊 Statuts des jeux</h3>
                        <div class="space-y-2">
                            <div><span class="status-badge status-available">✅ Disponible</span> - Peut être emprunté</div>
                            <div><span class="status-badge status-borrowed">🔴 Emprunté</span> - Actuellement prêté</div>
                        </div>
                    </div>
                </div>
            </div>

            <!-- Guide des Utilisateurs -->
            <div class="section-card bg-white rounded-lg shadow-lg p-8 mb-8">
                <h2 class="text-3xl font-bold text-green-600 mb-6">👥 Gestion des Utilisateurs</h2>
                <p class="text-gray-600 mb-4"><strong>URL :</strong> <code class="bg-gray-100 px-2 py-1 rounded">http://localhost:8081/users</code></p>
                
                <div class="grid md:grid-cols-2 gap-6">
                    <div>
                        <h3 class="text-xl font-semibold text-green-600 mb-4">👤 Comment inscrire un membre</h3>
                        <ol class="list-decimal list-inside space-y-2 text-sm text-gray-700">
                            <li>Saisir le nom complet (obligatoire)</li>
                            <li>Ajouter l'adresse email (obligatoire)</li>
                            <li>Cliquer "Inscrire le Membre"</li>
                            <li>Le membre devient automatiquement actif</li>
                        </ol>
                    </div>
                    <div>
                        <h3 class="text-xl font-semibold text-green-600 mb-4">📊 Statuts des utilisateurs</h3>
                        <div class="space-y-2">
                            <div><span class="status-badge status-available">✅ Actif</span> - Peut emprunter des jeux</div>
                            <div><span class="status-badge status-borrowed">🔴 Inactif</span> - Compte désactivé</div>
                        </div>
                    </div>
                </div>
            </div>

            <!-- Guide des Emprunts -->
            <div class="section-card bg-white rounded-lg shadow-lg p-8 mb-8">
                <h2 class="text-3xl font-bold text-yellow-600 mb-6">📚 Gestion des Emprunts</h2>
                <p class="text-gray-600 mb-4"><strong>URL :</strong> <code class="bg-gray-100 px-2 py-1 rounded">http://localhost:8081/borrowings</code></p>
                
                <div class="grid md:grid-cols-2 gap-6">
                    <div>
                        <h3 class="text-xl font-semibold text-yellow-600 mb-4">📚 Comment créer un emprunt</h3>
                        <ol class="list-decimal list-inside space-y-2 text-sm text-gray-700">
                            <li>Sélectionner un utilisateur actif</li>
                            <li>Choisir un jeu disponible</li>
                            <li>Sélectionner la durée d'emprunt (7, 14, 21, 30 ou 60 jours)</li>
                            <li>Cliquer "Créer l'Emprunt"</li>
                            <li>Le jeu devient indisponible</li>
                        </ol>
                    </div>
                    <div>
                        <h3 class="text-xl font-semibold text-yellow-600 mb-4">🔄 Comment retourner un jeu</h3>
                        <ol class="list-decimal list-inside space-y-2 text-sm text-gray-700">
                            <li>Trouver l'emprunt "En cours"</li>
                            <li>Cliquer le bouton "Retourner"</li>
                            <li>Confirmer le retour</li>
                            <li>Le jeu redevient disponible</li>
                        </ol>
                    </div>
                </div>
                
                <div class="mt-6">
                    <h3 class="text-lg font-semibold text-gray-800 mb-3">📊 Statuts des emprunts</h3>
                    <div class="flex flex-wrap gap-3">
                        <span class="status-badge status-active">🟡 En cours</span>
                        <span class="status-badge status-available">✅ Retourné</span>
                        <span class="status-badge status-borrowed">🔴 En retard</span>
                    </div>
                </div>
            </div>

            <!-- Guide des Alertes -->
            <div class="section-card bg-white rounded-lg shadow-lg p-8 mb-8">
                <h2 class="text-3xl font-bold text-red-600 mb-6">🚨 Gestion des Alertes</h2>
                <p class="text-gray-600 mb-4"><strong>URL :</strong> <code class="bg-gray-100 px-2 py-1 rounded">http://localhost:8081/alerts</code></p>
                
                <div class="grid md:grid-cols-3 gap-4 mb-6">
                    <div class="bg-red-50 p-4 rounded-lg">
                        <h4 class="font-semibold text-red-600 mb-2">⚠️ Alertes de Retard</h4>
                        <p class="text-sm text-red-700">Cliquez "Générer Alertes Retard" pour créer des alertes pour tous les emprunts en retard.</p>
                    </div>
                    <div class="bg-yellow-50 p-4 rounded-lg">
                        <h4 class="font-semibold text-yellow-600 mb-2">⏰ Rappels</h4>
                        <p class="text-sm text-yellow-700">Cliquez "Générer Rappels" pour créer des rappels pour les emprunts dus dans 2 jours.</p>
                    </div>
                    <div class="bg-blue-50 p-4 rounded-lg">
                        <h4 class="font-semibold text-blue-600 mb-2">🧹 Nettoyage</h4>
                        <p class="text-sm text-blue-700">Cliquez "Nettoyer Alertes Résolues" pour supprimer les alertes obsolètes.</p>
                    </div>
                </div>

                <div>
                    <h3 class="text-xl font-semibold text-red-600 mb-4">📝 Créer une alerte personnalisée</h3>
                    <ol class="list-decimal list-inside space-y-2 text-sm text-gray-700">
                        <li>Sélectionner un utilisateur</li>
                        <li>Choisir un jeu</li>
                        <li>Écrire un message personnalisé</li>
                        <li>Cliquer "Créer l'Alerte"</li>
                    </ol>
                </div>
            </div>

            <!-- Conseils -->
            <div class="section-card bg-white rounded-lg shadow-lg p-8 mb-8">
                <h2 class="text-3xl font-bold text-indigo-600 mb-6">💡 Conseils d'Utilisation</h2>
                
                <div class="grid md:grid-cols-2 gap-6">
                    <div class="bg-indigo-50 p-6 rounded-lg">
                        <h3 class="text-xl font-semibold text-indigo-600 mb-4">✅ Bonnes Pratiques</h3>
                        <ul class="space-y-2 text-sm text-indigo-700">
                            <li>• Vérifiez régulièrement la page des alertes</li>
                            <li>• Générez les rappels avant les échéances</li>
                            <li>• Nettoyez les alertes résolues périodiquement</li>
                            <li>• Utilisez les alertes personnalisées pour communiquer</li>
                        </ul>
                    </div>

                    <div class="bg-yellow-50 p-6 rounded-lg">
                        <h3 class="text-xl font-semibold text-yellow-600 mb-4">🔧 Problèmes Courants</h3>
                        <div class="space-y-3 text-sm">
                            <div>
                                <p class="font-semibold text-yellow-800">"Impossible de supprimer ce jeu"</p>
                                <p class="text-yellow-700">Le jeu a un historique d'emprunts. C'est normal pour préserver l'intégrité des données.</p>
                            </div>
                            <div>
                                <p class="font-semibold text-yellow-800">"Jeu non disponible"</p>
                                <p class="text-yellow-700">Le jeu est actuellement emprunté. Vérifiez les emprunts actifs.</p>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>

    <!-- Footer -->
    <footer class="bg-gray-800 text-white py-8">
        <div class="container mx-auto px-4 text-center">
            <h3 class="text-xl font-semibold mb-4">🎲 Bibliothèque de Jeux de Société</h3>
            <p class="text-gray-300 mb-4">Interface complète pour la gestion de votre collection de jeux</p>
            <a href="/" class="text-blue-400 hover:text-blue-300 transition-colors">← Retour à l'accueil</a>
        </div>
    </footer>
</body>
</html>`)
	})

	// Root route - page d'accueil en français
	router.GET("/", func(c *gin.Context) {
		c.Header("Content-Type", "text/html")
		c.String(http.StatusOK, `
<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Bibliothèque de Jeux de Société</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-100">
    <div class="container mx-auto px-4 py-8">
        <h1 class="text-4xl font-bold text-center text-blue-600 mb-8">Bibliothèque de Jeux de Société</h1>
        <div class="max-w-4xl mx-auto bg-white rounded-lg shadow-lg p-6">
            <h2 class="text-2xl font-semibold mb-4">Bienvenue dans le Système de Gestion de la Bibliothèque de Jeux</h2>
            <p class="text-gray-600 mb-6">Gérez votre collection de jeux de société, suivez les emprunts et restez organisé.</p>
            
            <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                <a href="/games" class="bg-blue-500 hover:bg-blue-600 text-white p-4 rounded-lg text-center transition-colors">
                    <h3 class="font-semibold">Jeux</h3>
                    <p class="text-sm">Gérer la collection</p>
                </a>
                <a href="/users" class="bg-green-500 hover:bg-green-600 text-white p-4 rounded-lg text-center transition-colors">
                    <h3 class="font-semibold">Utilisateurs</h3>
                    <p class="text-sm">Gérer les membres</p>
                </a>
                <a href="/borrowings" class="bg-yellow-500 hover:bg-yellow-600 text-white p-4 rounded-lg text-center transition-colors">
                    <h3 class="font-semibold">Emprunts</h3>
                    <p class="text-sm">Suivre les prêts</p>
                </a>
                <a href="/alerts" class="bg-red-500 hover:bg-red-600 text-white p-4 rounded-lg text-center transition-colors">
                    <h3 class="font-semibold">Alertes</h3>
                    <p class="text-sm">Voir les notifications</p>
                </a>
            </div>
            
            <!-- Section Guide Utilisateur -->
            <div class="mt-8 bg-gradient-to-r from-purple-50 to-blue-50 rounded-lg p-6 border border-purple-200">
                <div class="text-center">
                    <h3 class="text-xl font-semibold text-purple-800 mb-3">📖 Guide Utilisateur</h3>
                    <p class="text-gray-700 mb-4">Découvrez comment utiliser toutes les fonctionnalités de l'interface</p>
                    <a href="/guide" class="inline-block bg-purple-600 hover:bg-purple-700 text-white px-6 py-3 rounded-lg font-semibold transition-colors shadow-lg hover:shadow-xl transform hover:-translate-y-1">
                        🚀 Consulter le Guide Complet
                    </a>
                </div>
                <div class="mt-6 grid grid-cols-2 md:grid-cols-4 gap-3 text-center">
                    <div class="bg-white p-3 rounded-lg shadow-sm">
                        <div class="text-lg mb-1">🎲</div>
                        <div class="text-xs text-gray-600">Gestion des Jeux</div>
                    </div>
                    <div class="bg-white p-3 rounded-lg shadow-sm">
                        <div class="text-lg mb-1">👥</div>
                        <div class="text-xs text-gray-600">Gestion des Utilisateurs</div>
                    </div>
                    <div class="bg-white p-3 rounded-lg shadow-sm">
                        <div class="text-lg mb-1">📚</div>
                        <div class="text-xs text-gray-600">Gestion des Emprunts</div>
                    </div>
                    <div class="bg-white p-3 rounded-lg shadow-sm">
                        <div class="text-lg mb-1">🚨</div>
                        <div class="text-xs text-gray-600">Gestion des Alertes</div>
                    </div>
                </div>
            </div>
            
            <div class="mt-8 text-center">
                <h3 class="text-lg font-semibold mb-2">Points d'accès API</h3>
                <div class="space-y-2 text-sm">
                    <a href="/swagger/index.html" class="text-green-600 hover:underline block font-semibold">📚 Documentation Swagger API</a>
                    <a href="/api/v1/status" class="text-blue-600 hover:underline block">Statut API</a>
                    <a href="/health" class="text-blue-600 hover:underline block">Vérification Santé</a>
                </div>
            </div>
        </div>
    </div>
</body>
</html>`)
	})

	// Simple web routes
	setupSimpleWebRoutes(router, gameService, userService, borrowingService, alertService)
	
	// Form submission routes
	setupFormRoutes(router, gameService, userService, borrowingService, alertService)

	// Swagger documentation route
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API routes
	setupAPIRoutes(router, gameHandler, userHandler, borrowingHandler, alertHandler)

	return nil
}

// setupSimpleWebRoutes configures simple web routes for testing
func setupSimpleWebRoutes(router *gin.Engine, gameService *services.GameService, userService *services.UserService, borrowingService *services.BorrowingService, alertService *services.AlertService) {
	// Games route
	router.GET("/games", func(c *gin.Context) {
		games, err := gameService.GetAllGames()
		if err != nil {
			c.Header("Content-Type", "text/html")
			c.String(http.StatusInternalServerError, `
<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Jeux - Bibliothèque de Jeux de Société</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-100">
    <div class="container mx-auto px-4 py-8">
        <div class="max-w-4xl mx-auto bg-white rounded-lg shadow-lg p-6">
            <h1 class="text-3xl font-bold text-red-600 mb-4">Erreur</h1>
            <p class="text-gray-600">Échec du chargement des jeux : %s</p>
            <a href="/" class="mt-4 inline-block bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded">Retour à l'accueil</a>
        </div>
    </div>
</body>
</html>`, err.Error())
			return
		}
		
		gamesHTML := ""
		if len(games) == 0 {
			gamesHTML = `<p class="text-gray-500 text-center py-8">Aucun jeu trouvé. Utilisez le formulaire ci-dessous pour ajouter votre premier jeu !</p>`
		} else {
			gamesHTML = `<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">`
			for _, game := range games {
				status := "Disponible"
				statusColor := "text-green-600"
				if !game.IsAvailable {
					status = "Emprunté"
					statusColor = "text-red-600"
				}
				gamesHTML += fmt.Sprintf(`
					<div class="bg-gray-50 p-4 rounded-lg border">
						<div class="flex justify-between items-start">
							<div class="flex-1">
								<h3 class="font-semibold text-lg mb-2">%s</h3>
								<p class="text-gray-600 text-sm mb-2">%s</p>
								<div class="flex justify-between items-center mb-2">
									<span class="text-sm text-gray-500">Catégorie : %s</span>
									<span class="text-sm %s font-medium">%s</span>
								</div>
								<div class="text-xs text-gray-400">
									État : %s | Ajouté : %s
								</div>
							</div>
							<div class="ml-4">
								<form action="/games/%d/delete" method="POST" style="display: inline;">
									<button type="submit" class="text-xs px-2 py-1 bg-red-500 hover:bg-red-600 text-white rounded" 
											onclick="return confirm('Êtes-vous sûr de vouloir supprimer ce jeu ?')">
										🗑️ Supprimer
									</button>
								</form>
							</div>
						</div>
					</div>`, game.Name, game.Description, game.Category, statusColor, status, game.Condition, game.EntryDate.Format("2006-01-02"), game.ID)
			}
			gamesHTML += `</div>`
		}
		
		c.Header("Content-Type", "text/html")
		c.String(http.StatusOK, `
<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Jeux - Bibliothèque de Jeux de Société</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-100">
    <div class="container mx-auto px-4 py-8">
        <div class="max-w-6xl mx-auto space-y-6">
            <!-- En-tête -->
            <div class="bg-white rounded-lg shadow-lg p-6">
                <div class="flex justify-between items-center">
                    <h1 class="text-3xl font-bold text-blue-600">Collection de Jeux</h1>
                    <a href="/" class="bg-gray-500 hover:bg-gray-600 text-white px-4 py-2 rounded">Retour à l'accueil</a>
                </div>
                <p class="text-gray-600 mt-2">Total des jeux : %d</p>
            </div>

            <!-- Formulaire d'ajout de nouveau jeu -->
            <div class="bg-white rounded-lg shadow-lg p-6">
                <h2 class="text-xl font-semibold text-blue-600 mb-4">📚 Ajouter un Nouveau Jeu</h2>
                <form action="/games/create" method="POST" class="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div>
                        <label for="name" class="block text-sm font-medium text-gray-700 mb-1">Nom du Jeu *</label>
                        <input type="text" id="name" name="name" required 
                               class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                               placeholder="Saisir le nom du jeu">
                    </div>
                    <div>
                        <label for="category" class="block text-sm font-medium text-gray-700 mb-1">Catégorie *</label>
                        <input type="text" id="category" name="category" required 
                               class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                               placeholder="ex: Stratégie, Famille, Cartes">
                    </div>
                    <div class="md:col-span-2">
                        <label for="description" class="block text-sm font-medium text-gray-700 mb-1">Description</label>
                        <textarea id="description" name="description" rows="3"
                                  class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                                  placeholder="Brève description du jeu"></textarea>
                    </div>
                    <div>
                        <label for="condition" class="block text-sm font-medium text-gray-700 mb-1">État *</label>
                        <select id="condition" name="condition" required 
                                class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500">
                            <option value="">Sélectionner l'état</option>
                            <option value="excellent">Excellent</option>
                            <option value="good">Bon</option>
                            <option value="fair">Correct</option>
                            <option value="poor">Mauvais</option>
                        </select>
                    </div>
                    <div class="flex items-end">
                        <button type="submit" 
                                class="w-full bg-blue-500 hover:bg-blue-600 text-white font-medium py-2 px-4 rounded-md transition-colors">
                            ➕ Ajouter le Jeu
                        </button>
                    </div>
                </form>
            </div>

            <!-- Liste des jeux -->
            <div class="bg-white rounded-lg shadow-lg p-6">
                <h2 class="text-xl font-semibold text-blue-600 mb-4">🎲 Jeux Actuels</h2>
                %s
            </div>

            <!-- Informations API -->
            <div class="bg-gray-50 rounded-lg p-4 text-center">
                <h3 class="text-lg font-semibold mb-2">Accès API</h3>
                <div class="space-x-4 text-sm">
                    <a href="/api/v1/games" class="text-blue-600 hover:underline">Voir JSON</a>
                    <span class="text-gray-400">|</span>
                    <span class="text-gray-500">API REST disponible pour les opérations avancées</span>
                </div>
            </div>
        </div>
    </div>
</body>
</html>`, len(games), gamesHTML)
	})
	
	// Users route
	router.GET("/users", func(c *gin.Context) {
		users, err := userService.GetAllUsers()
		if err != nil {
			c.Header("Content-Type", "text/html")
			c.String(http.StatusInternalServerError, `
<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Utilisateurs - Bibliothèque de Jeux de Société</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-100">
    <div class="container mx-auto px-4 py-8">
        <div class="max-w-4xl mx-auto bg-white rounded-lg shadow-lg p-6">
            <h1 class="text-3xl font-bold text-red-600 mb-4">Erreur</h1>
            <p class="text-gray-600">Échec du chargement des utilisateurs : %s</p>
            <a href="/" class="mt-4 inline-block bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded">Retour à l'accueil</a>
        </div>
    </div>
</body>
</html>`, err.Error())
			return
		}
		
		usersHTML := ""
		if len(users) == 0 {
			usersHTML = `<p class="text-gray-500 text-center py-8">Aucun utilisateur trouvé. Utilisez le formulaire ci-dessous pour inscrire votre premier membre !</p>`
		} else {
			usersHTML = `<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">`
			for _, user := range users {
				statusColor := "text-green-600"
				status := "Actif"
				if !user.IsActive {
					statusColor = "text-red-600"
					status = "Inactif"
				}
				usersHTML += fmt.Sprintf(`
					<div class="bg-gray-50 p-4 rounded-lg border">
						<div class="flex justify-between items-start">
							<div class="flex-1">
								<h3 class="font-semibold text-lg mb-2">%s</h3>
								<p class="text-gray-600 text-sm mb-2">📧 %s</p>
								<div class="flex justify-between items-center mb-2">
									<span class="text-sm text-gray-500">ID : %d</span>
									<span class="text-sm %s font-medium">%s</span>
								</div>
								<div class="text-xs text-gray-400">
									Inscrit : %s
								</div>
							</div>
							<div class="ml-4">
								<form action="/users/%d/delete" method="POST" style="display: inline;">
									<button type="submit" class="text-xs px-2 py-1 bg-red-500 hover:bg-red-600 text-white rounded" 
											onclick="return confirm('Êtes-vous sûr de vouloir supprimer cet utilisateur ?')">
										🗑️ Supprimer
									</button>
								</form>
							</div>
						</div>
					</div>`, user.Name, user.Email, user.ID, statusColor, status, user.RegisteredAt.Format("2006-01-02"), user.ID)
			}
			usersHTML += `</div>`
		}
		
		c.Header("Content-Type", "text/html")
		c.String(http.StatusOK, `
<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Utilisateurs - Bibliothèque de Jeux de Société</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-100">
    <div class="container mx-auto px-4 py-8">
        <div class="max-w-6xl mx-auto space-y-6">
            <!-- En-tête -->
            <div class="bg-white rounded-lg shadow-lg p-6">
                <div class="flex justify-between items-center">
                    <h1 class="text-3xl font-bold text-green-600">Membres de la Bibliothèque</h1>
                    <a href="/" class="bg-gray-500 hover:bg-gray-600 text-white px-4 py-2 rounded">Retour à l'accueil</a>
                </div>
                <p class="text-gray-600 mt-2">Total des utilisateurs : %d</p>
            </div>

            <!-- Formulaire d'ajout de nouvel utilisateur -->
            <div class="bg-white rounded-lg shadow-lg p-6">
                <h2 class="text-xl font-semibold text-green-600 mb-4">👤 Inscrire un Nouveau Membre</h2>
                <form action="/users/create" method="POST" class="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div>
                        <label for="name" class="block text-sm font-medium text-gray-700 mb-1">Nom Complet *</label>
                        <input type="text" id="name" name="name" required 
                               class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-green-500"
                               placeholder="Saisir le nom complet">
                    </div>
                    <div>
                        <label for="email" class="block text-sm font-medium text-gray-700 mb-1">Adresse Email *</label>
                        <input type="email" id="email" name="email" required 
                               class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-green-500"
                               placeholder="Saisir l'adresse email">
                    </div>
                    <div class="md:col-span-2">
                        <button type="submit" 
                                class="w-full md:w-auto bg-green-500 hover:bg-green-600 text-white font-medium py-2 px-6 rounded-md transition-colors">
                            ➕ Inscrire le Membre
                        </button>
                    </div>
                </form>
            </div>

            <!-- Liste des utilisateurs -->
            <div class="bg-white rounded-lg shadow-lg p-6">
                <h2 class="text-xl font-semibold text-green-600 mb-4">👥 Membres Actuels</h2>
                %s
            </div>

            <!-- Informations API -->
            <div class="bg-gray-50 rounded-lg p-4 text-center">
                <h3 class="text-lg font-semibold mb-2">Accès API</h3>
                <div class="space-x-4 text-sm">
                    <a href="/api/v1/users" class="text-blue-600 hover:underline">Voir JSON</a>
                    <span class="text-gray-400">|</span>
                    <span class="text-gray-500">API REST disponible pour les opérations avancées</span>
                </div>
            </div>
        </div>
    </div>
</body>
</html>`, len(users), usersHTML)
	})
	
	// Borrowings route
	router.GET("/borrowings", func(c *gin.Context) {
		borrowings, err := borrowingService.GetAllBorrowings()
		if err != nil {
			c.Header("Content-Type", "text/html")
			c.String(http.StatusInternalServerError, `
<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Emprunts - Bibliothèque de Jeux de Société</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-100">
    <div class="container mx-auto px-4 py-8">
        <div class="max-w-4xl mx-auto bg-white rounded-lg shadow-lg p-6">
            <h1 class="text-3xl font-bold text-red-600 mb-4">Erreur</h1>
            <p class="text-gray-600">Échec du chargement des emprunts : %s</p>
            <a href="/" class="mt-4 inline-block bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded">Retour à l'accueil</a>
        </div>
    </div>
</body>
</html>`, err.Error())
			return
		}

		// Get all users and games for the form
		users, _ := userService.GetAllUsers()
		games, _ := gameService.GetAvailableGames()

		borrowingsHTML := ""
		if len(borrowings) == 0 {
			borrowingsHTML = `<p class="text-gray-500 text-center py-8">Aucun emprunt trouvé. Utilisez le formulaire ci-dessous pour créer votre premier emprunt !</p>`
		} else {
			borrowingsHTML = `<div class="space-y-4">`
			for _, borrowing := range borrowings {
				status := "En cours"
				statusColor := "text-yellow-600"
				statusBg := "bg-yellow-100"
				actionButton := ""
				
				if borrowing.ReturnedAt != nil {
					status = "Retourné"
					statusColor = "text-green-600"
					statusBg = "bg-green-100"
				} else if borrowing.IsOverdue {
					status = "En retard"
					statusColor = "text-red-600"
					statusBg = "bg-red-100"
				} else {
					actionButton = fmt.Sprintf(`
						<form action="/borrowings/%d/return" method="POST" style="display: inline;">
							<button type="submit" class="text-xs px-2 py-1 bg-green-500 hover:bg-green-600 text-white rounded">
								✅ Retourner
							</button>
						</form>`, borrowing.ID)
				}

				returnedText := ""
				if borrowing.ReturnedAt != nil {
					returnedText = fmt.Sprintf("| Retourné : %s", borrowing.ReturnedAt.Format("2006-01-02"))
				}

				// Get user and game details
				userName := fmt.Sprintf("ID %d", borrowing.UserID)
				gameName := fmt.Sprintf("ID %d", borrowing.GameID)
				
				if user, err := userService.GetUser(borrowing.UserID); err == nil {
					userName = fmt.Sprintf("%s (%s)", user.Name, user.Email)
				}
				
				if game, err := gameService.GetGame(borrowing.GameID); err == nil {
					gameName = fmt.Sprintf("%s (%s)", game.Name, game.Category)
				}

				borrowingsHTML += fmt.Sprintf(`
					<div class="bg-gray-50 p-4 rounded-lg border">
						<div class="flex justify-between items-start">
							<div class="flex-1">
								<div class="flex items-center mb-2">
									<h3 class="font-semibold text-lg mr-3">Emprunt #%d</h3>
									<span class="px-2 py-1 text-xs font-medium rounded-full %s %s">%s</span>
								</div>
								<div class="grid grid-cols-1 md:grid-cols-2 gap-2 text-sm text-gray-600 mb-2">
									<p><strong>Utilisateur :</strong> %s</p>
									<p><strong>Jeu :</strong> %s</p>
									<p><strong>Emprunté :</strong> %s</p>
									<p><strong>Échéance :</strong> %s</p>
								</div>
								<div class="text-xs text-gray-400">
									%s
								</div>
							</div>
							<div class="ml-4">
								%s
							</div>
						</div>
					</div>`, borrowing.ID, statusBg, statusColor, status, userName, gameName, 
					borrowing.BorrowedAt.Format("2006-01-02"), borrowing.DueDate.Format("2006-01-02"), returnedText, actionButton)
			}
			borrowingsHTML += `</div>`
		}

		// Build users options
		usersOptions := ""
		for _, user := range users {
			usersOptions += fmt.Sprintf(`<option value="%d">%s (%s)</option>`, user.ID, user.Name, user.Email)
		}

		// Build games options
		gamesOptions := ""
		for _, game := range games {
			gamesOptions += fmt.Sprintf(`<option value="%d">%s (%s)</option>`, game.ID, game.Name, game.Category)
		}

		c.Header("Content-Type", "text/html")
		c.String(http.StatusOK, `
<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Emprunts - Bibliothèque de Jeux de Société</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-100">
    <div class="container mx-auto px-4 py-8">
        <div class="max-w-6xl mx-auto space-y-6">
            <!-- En-tête -->
            <div class="bg-white rounded-lg shadow-lg p-6">
                <div class="flex justify-between items-center">
                    <h1 class="text-3xl font-bold text-yellow-600">Gestion des Emprunts</h1>
                    <a href="/" class="bg-gray-500 hover:bg-gray-600 text-white px-4 py-2 rounded">Retour à l'accueil</a>
                </div>
                <p class="text-gray-600 mt-2">Total des emprunts : %d</p>
            </div>

            <!-- Formulaire de nouvel emprunt -->
            <div class="bg-white rounded-lg shadow-lg p-6">
                <h2 class="text-xl font-semibold text-yellow-600 mb-4">📚 Créer un Nouvel Emprunt</h2>
                <form action="/borrowings/create" method="POST" class="grid grid-cols-1 md:grid-cols-3 gap-4">
                    <div>
                        <label for="user_id" class="block text-sm font-medium text-gray-700 mb-1">Utilisateur *</label>
                        <select id="user_id" name="user_id" required 
                                class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-yellow-500">
                            <option value="">Sélectionner un utilisateur</option>
                            %s
                        </select>
                    </div>
                    <div>
                        <label for="game_id" class="block text-sm font-medium text-gray-700 mb-1">Jeu Disponible *</label>
                        <select id="game_id" name="game_id" required 
                                class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-yellow-500">
                            <option value="">Sélectionner un jeu</option>
                            %s
                        </select>
                    </div>
                    <div>
                        <label for="duration_days" class="block text-sm font-medium text-gray-700 mb-1">Durée d'emprunt *</label>
                        <select id="duration_days" name="duration_days" required 
                                class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-yellow-500">
                            <option value="7">7 jours (1 semaine)</option>
                            <option value="14" selected>14 jours (2 semaines) - Défaut</option>
                            <option value="21">21 jours (3 semaines)</option>
                            <option value="30">30 jours (1 mois)</option>
                            <option value="60">60 jours (2 mois)</option>
                        </select>
                    </div>
                    <div class="flex items-end">
                        <button type="submit" 
                                class="w-full bg-yellow-500 hover:bg-yellow-600 text-white font-medium py-2 px-4 rounded-md transition-colors">
                            ➕ Créer l'Emprunt
                        </button>
                    </div>
                </form>
            </div>

            <!-- Liste des emprunts -->
            <div class="bg-white rounded-lg shadow-lg p-6">
                <h2 class="text-xl font-semibold text-yellow-600 mb-4">📋 Emprunts Actuels</h2>
                %s
            </div>

            <!-- Informations API -->
            <div class="bg-gray-50 rounded-lg p-4 text-center">
                <h3 class="text-lg font-semibold mb-2">Accès API</h3>
                <div class="space-x-4 text-sm">
                    <span class="text-gray-500">API REST disponible pour les opérations avancées</span>
                </div>
            </div>
        </div>
    </div>
</body>
</html>`, len(borrowings), usersOptions, gamesOptions, borrowingsHTML)
	})
	
	// Alerts route
	router.GET("/alerts", func(c *gin.Context) {
		alerts, err := alertService.GetActiveAlerts()
		if err != nil {
			c.Header("Content-Type", "text/html")
			c.String(http.StatusInternalServerError, `
<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Erreur - Bibliothèque de Jeux de Société</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-100">
    <div class="container mx-auto px-4 py-8">
        <div class="max-w-4xl mx-auto bg-white rounded-lg shadow-lg p-6">
            <h1 class="text-3xl font-bold text-red-600 mb-4">Erreur</h1>
            <p class="text-gray-600">Échec du chargement des alertes : %s</p>
            <a href="/" class="mt-4 inline-block bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded">Retour à l'accueil</a>
        </div>
    </div>
</body>
</html>`, err.Error())
			return
		}

		// Get all users and games for the custom alert form
		users, _ := userService.GetAllUsers()
		games, _ := gameService.GetAllGames()

		alertsHTML := ""
		if len(alerts) == 0 {
			alertsHTML = `<p class="text-gray-500 text-center py-8">Aucune alerte active. Toutes les notifications ont été traitées !</p>`
		} else {
			alertsHTML = `<div class="space-y-4">`
			for _, alert := range alerts {
				alertTypeColor := "text-blue-600"
				alertTypeBg := "bg-blue-100"
				alertIcon := "ℹ️"
				
				switch alert.Type {
				case "overdue":
					alertTypeColor = "text-red-600"
					alertTypeBg = "bg-red-100"
					alertIcon = "⚠️"
				case "reminder":
					alertTypeColor = "text-yellow-600"
					alertTypeBg = "bg-yellow-100"
					alertIcon = "⏰"
				case "custom":
					alertTypeColor = "text-purple-600"
					alertTypeBg = "bg-purple-100"
					alertIcon = "📝"
				}

				// Get user and game details
				userName := fmt.Sprintf("ID %d", alert.UserID)
				gameName := fmt.Sprintf("ID %d", alert.GameID)
				
				if user, err := userService.GetUser(alert.UserID); err == nil {
					userName = fmt.Sprintf("%s (%s)", user.Name, user.Email)
				}
				
				if game, err := gameService.GetGame(alert.GameID); err == nil {
					gameName = fmt.Sprintf("%s (%s)", game.Name, game.Category)
				}

				alertsHTML += fmt.Sprintf(`
					<div class="bg-gray-50 p-4 rounded-lg border border-l-4 border-l-red-500">
						<div class="flex justify-between items-start">
							<div class="flex-1">
								<div class="flex items-center mb-2">
									<span class="text-2xl mr-2">%s</span>
									<h3 class="font-semibold text-lg mr-3">Alerte #%d</h3>
									<span class="px-2 py-1 text-xs font-medium rounded-full %s %s">%s</span>
								</div>
								<div class="mb-3">
									<p class="text-gray-800 font-medium">%s</p>
								</div>
								<div class="grid grid-cols-1 md:grid-cols-2 gap-2 text-sm text-gray-600 mb-2">
									<p><strong>Utilisateur :</strong> %s</p>
									<p><strong>Jeu :</strong> %s</p>
									<p><strong>Créée :</strong> %s</p>
									<p><strong>Type :</strong> %s</p>
								</div>
							</div>
							<div class="ml-4 space-x-2">
								<form action="/alerts/%d/read" method="POST" style="display: inline;">
									<button type="submit" class="text-xs px-3 py-1 bg-green-500 hover:bg-green-600 text-white rounded">
										✅ Marquer comme lue
									</button>
								</form>
								<form action="/alerts/%d/delete" method="POST" style="display: inline;">
									<button type="submit" class="text-xs px-3 py-1 bg-red-500 hover:bg-red-600 text-white rounded" 
											onclick="return confirm('Êtes-vous sûr de vouloir supprimer cette alerte ?')">
										🗑️ Supprimer
									</button>
								</form>
							</div>
						</div>
					</div>`, alertIcon, alert.ID, alertTypeBg, alertTypeColor, alert.Type, alert.Message, userName, gameName, 
					alert.CreatedAt.Format("2006-01-02 15:04"), alert.Type, alert.ID, alert.ID)
			}
			alertsHTML += `</div>`
		}

		// Build users options
		usersOptions := ""
		for _, user := range users {
			usersOptions += fmt.Sprintf(`<option value="%d">%s (%s)</option>`, user.ID, user.Name, user.Email)
		}

		// Build games options
		gamesOptions := ""
		for _, game := range games {
			gamesOptions += fmt.Sprintf(`<option value="%d">%s (%s)</option>`, game.ID, game.Name, game.Category)
		}

		c.Header("Content-Type", "text/html")
		c.String(http.StatusOK, `
<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Alertes - Bibliothèque de Jeux de Société</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-100">
    <div class="container mx-auto px-4 py-8">
        <div class="max-w-6xl mx-auto space-y-6">
            <!-- En-tête -->
            <div class="bg-white rounded-lg shadow-lg p-6">
                <div class="flex justify-between items-center">
                    <h1 class="text-3xl font-bold text-red-600">Gestion des Alertes</h1>
                    <a href="/" class="bg-gray-500 hover:bg-gray-600 text-white px-4 py-2 rounded">Retour à l'accueil</a>
                </div>
                <p class="text-gray-600 mt-2">Alertes actives : %d</p>
            </div>

            <!-- Actions rapides -->
            <div class="bg-white rounded-lg shadow-lg p-6">
                <h2 class="text-xl font-semibold text-red-600 mb-4">⚡ Actions Rapides</h2>
                <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
                    <form action="/alerts/generate-overdue" method="POST" style="display: inline;">
                        <button type="submit" class="w-full bg-red-500 hover:bg-red-600 text-white font-medium py-2 px-4 rounded-md transition-colors">
                            ⚠️ Générer Alertes Retard
                        </button>
                    </form>
                    <form action="/alerts/generate-reminders" method="POST" style="display: inline;">
                        <button type="submit" class="w-full bg-yellow-500 hover:bg-yellow-600 text-white font-medium py-2 px-4 rounded-md transition-colors">
                            ⏰ Générer Rappels
                        </button>
                    </form>
                    <form action="/alerts/cleanup" method="POST" style="display: inline;">
                        <button type="submit" class="w-full bg-blue-500 hover:bg-blue-600 text-white font-medium py-2 px-4 rounded-md transition-colors">
                            🧹 Nettoyer Alertes Résolues
                        </button>
                    </form>
                </div>
            </div>

            <!-- Formulaire de nouvelle alerte personnalisée -->
            <div class="bg-white rounded-lg shadow-lg p-6">
                <h2 class="text-xl font-semibold text-red-600 mb-4">📝 Créer une Alerte Personnalisée</h2>
                <form action="/alerts/create" method="POST" class="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div>
                        <label for="user_id" class="block text-sm font-medium text-gray-700 mb-1">Utilisateur *</label>
                        <select id="user_id" name="user_id" required 
                                class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-red-500">
                            <option value="">Sélectionner un utilisateur</option>
                            %s
                        </select>
                    </div>
                    <div>
                        <label for="game_id" class="block text-sm font-medium text-gray-700 mb-1">Jeu *</label>
                        <select id="game_id" name="game_id" required 
                                class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-red-500">
                            <option value="">Sélectionner un jeu</option>
                            %s
                        </select>
                    </div>
                    <div class="md:col-span-2">
                        <label for="message" class="block text-sm font-medium text-gray-700 mb-1">Message *</label>
                        <textarea id="message" name="message" rows="3" required
                                  class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-red-500"
                                  placeholder="Saisir le message de l'alerte personnalisée"></textarea>
                    </div>
                    <div class="md:col-span-2">
                        <button type="submit" 
                                class="w-full md:w-auto bg-red-500 hover:bg-red-600 text-white font-medium py-2 px-6 rounded-md transition-colors">
                            ➕ Créer l'Alerte
                        </button>
                    </div>
                </form>
            </div>

            <!-- Liste des alertes -->
            <div class="bg-white rounded-lg shadow-lg p-6">
                <h2 class="text-xl font-semibold text-red-600 mb-4">🚨 Alertes Actives</h2>
                %s
            </div>

            <!-- Informations API -->
            <div class="bg-gray-50 rounded-lg p-4 text-center">
                <h3 class="text-lg font-semibold mb-2">Accès API</h3>
                <div class="space-x-4 text-sm">
                    <span class="text-gray-500">API REST disponible pour les opérations avancées</span>
                </div>
            </div>
        </div>
    </div>
</body>
</html>`, len(alerts), usersOptions, gamesOptions, alertsHTML)
	})
}

// setupFormRoutes configures form submission routes
func setupFormRoutes(router *gin.Engine, gameService *services.GameService, userService *services.UserService, borrowingService *services.BorrowingService, alertService *services.AlertService) {
	// Create new game
	router.POST("/games/create", func(c *gin.Context) {
		name := strings.TrimSpace(c.PostForm("name"))
		description := strings.TrimSpace(c.PostForm("description"))
		category := strings.TrimSpace(c.PostForm("category"))
		condition := strings.TrimSpace(c.PostForm("condition"))
		
		if name == "" || category == "" || condition == "" {
			c.Header("Content-Type", "text/html")
			c.String(http.StatusBadRequest, `
<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Erreur - Bibliothèque de Jeux de Société</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-100">
    <div class="container mx-auto px-4 py-8">
        <div class="max-w-4xl mx-auto bg-white rounded-lg shadow-lg p-6">
            <h1 class="text-3xl font-bold text-red-600 mb-4">❌ Erreur</h1>
            <p class="text-gray-600 mb-4">Veuillez remplir tous les champs obligatoires (Nom, Catégorie et État).</p>
            <a href="/games" class="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded">← Retour aux Jeux</a>
        </div>
    </div>
</body>
</html>`)
			return
		}
		
		game, err := gameService.AddGame(name, description, category, condition)
		if err != nil {
			c.Header("Content-Type", "text/html")
			c.String(http.StatusInternalServerError, `
<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Erreur - Bibliothèque de Jeux de Société</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-100">
    <div class="container mx-auto px-4 py-8">
        <div class="max-w-4xl mx-auto bg-white rounded-lg shadow-lg p-6">
            <h1 class="text-3xl font-bold text-red-600 mb-4">❌ Erreur</h1>
            <p class="text-gray-600 mb-4">Échec de la création du jeu : %s</p>
            <a href="/games" class="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded">← Retour aux Jeux</a>
        </div>
    </div>
</body>
</html>`, err.Error())
			return
		}
		
		c.Header("Content-Type", "text/html")
		c.String(http.StatusOK, `
<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Succès - Bibliothèque de Jeux de Société</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <meta http-equiv="refresh" content="3;url=/games">
</head>
<body class="bg-gray-100">
    <div class="container mx-auto px-4 py-8">
        <div class="max-w-4xl mx-auto bg-white rounded-lg shadow-lg p-6">
            <h1 class="text-3xl font-bold text-green-600 mb-4">✅ Succès !</h1>
            <p class="text-gray-600 mb-4">Le jeu "%s" a été ajouté avec succès à la bibliothèque !</p>
            <div class="bg-green-50 border border-green-200 rounded-lg p-4 mb-4">
                <h3 class="font-semibold text-green-800">Détails du Jeu :</h3>
                <ul class="text-green-700 mt-2">
                    <li><strong>Nom :</strong> %s</li>
                    <li><strong>Catégorie :</strong> %s</li>
                    <li><strong>État :</strong> %s</li>
                    <li><strong>Description :</strong> %s</li>
                </ul>
            </div>
            <p class="text-sm text-gray-500 mb-4">Vous serez redirigé vers la page des jeux dans 3 secondes...</p>
            <a href="/games" class="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded">← Retour aux Jeux</a>
        </div>
    </div>
</body>
</html>`, game.Name, game.Name, game.Category, game.Condition, game.Description)
	})
	
	// Delete game
	router.POST("/games/:id/delete", func(c *gin.Context) {
		idParam := c.Param("id")
		gameID, err := strconv.Atoi(idParam)
		
		if err != nil {
			c.Header("Content-Type", "text/html")
			c.String(http.StatusBadRequest, `
<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Erreur - Bibliothèque de Jeux de Société</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-100">
    <div class="container mx-auto px-4 py-8">
        <div class="max-w-4xl mx-auto bg-white rounded-lg shadow-lg p-6">
            <h1 class="text-3xl font-bold text-red-600 mb-4">❌ Erreur</h1>
            <p class="text-gray-600 mb-4">ID de jeu invalide.</p>
            <a href="/games" class="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded">← Retour aux Jeux</a>
        </div>
    </div>
</body>
</html>`)
			return
		}
		
		err = gameService.DeleteGame(gameID)
		if err != nil {
			// Determine the appropriate error message and styling
			errorTitle := "❌ Erreur"
			errorClass := "text-red-600"
			
			// Check if it's a constraint-related error
			if strings.Contains(err.Error(), "borrowing history") || strings.Contains(err.Error(), "associated records") || strings.Contains(err.Error(), "currently borrowed") {
				errorTitle = "⚠️ Suppression Impossible"
				errorClass = "text-yellow-600"
			}
			
			c.Header("Content-Type", "text/html")
			c.String(http.StatusConflict, `
<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Erreur - Bibliothèque de Jeux de Société</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-100">
    <div class="container mx-auto px-4 py-8">
        <div class="max-w-4xl mx-auto bg-white rounded-lg shadow-lg p-6">
            <h1 class="text-3xl font-bold %s mb-4">%s</h1>
            <div class="bg-yellow-50 border border-yellow-200 rounded-lg p-4 mb-4">
                <p class="text-gray-800 mb-2"><strong>Raison :</strong></p>
                <p class="text-gray-700">%s</p>
            </div>
            <div class="bg-blue-50 border border-blue-200 rounded-lg p-4 mb-4">
                <h3 class="font-semibold text-blue-800 mb-2">💡 Que faire ?</h3>
                <ul class="text-blue-700 space-y-1">
                    <li>• Les jeux avec un historique d'emprunts ne peuvent pas être supprimés</li>
                    <li>• Cela préserve l'intégrité des données et l'historique</li>
                    <li>• Vous pouvez marquer le jeu comme indisponible si nécessaire</li>
                </ul>
            </div>
            <a href="/games" class="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded">← Retour aux Jeux</a>
        </div>
    </div>
</body>
</html>`, errorClass, errorTitle, err.Error())
			return
		}
		
		c.Header("Content-Type", "text/html")
		c.String(http.StatusOK, `
<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Succès - Bibliothèque de Jeux de Société</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <meta http-equiv="refresh" content="2;url=/games">
</head>
<body class="bg-gray-100">
    <div class="container mx-auto px-4 py-8">
        <div class="max-w-4xl mx-auto bg-white rounded-lg shadow-lg p-6">
            <h1 class="text-3xl font-bold text-green-600 mb-4">✅ Jeu Supprimé !</h1>
            <p class="text-gray-600 mb-4">Le jeu a été supprimé avec succès de la bibliothèque.</p>
            <p class="text-sm text-gray-500 mb-4">Vous serez redirigé vers la page des jeux dans 2 secondes...</p>
            <a href="/games" class="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded">← Retour aux Jeux</a>
        </div>
    </div>
</body>
</html>`)
	})
	
	// Create new user
	router.POST("/users/create", func(c *gin.Context) {
		name := strings.TrimSpace(c.PostForm("name"))
		email := strings.TrimSpace(c.PostForm("email"))
		
		if name == "" || email == "" {
			c.Header("Content-Type", "text/html")
			c.String(http.StatusBadRequest, `
<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Erreur - Bibliothèque de Jeux de Société</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-100">
    <div class="container mx-auto px-4 py-8">
        <div class="max-w-4xl mx-auto bg-white rounded-lg shadow-lg p-6">
            <h1 class="text-3xl font-bold text-red-600 mb-4">❌ Erreur</h1>
            <p class="text-gray-600 mb-4">Veuillez remplir tous les champs obligatoires (Nom et Email).</p>
            <a href="/users" class="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded">← Retour aux Utilisateurs</a>
        </div>
    </div>
</body>
</html>`)
			return
		}
		
		user, err := userService.RegisterUser(name, email)
		if err != nil {
			c.Header("Content-Type", "text/html")
			c.String(http.StatusInternalServerError, `
<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Erreur - Bibliothèque de Jeux de Société</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-100">
    <div class="container mx-auto px-4 py-8">
        <div class="max-w-4xl mx-auto bg-white rounded-lg shadow-lg p-6">
            <h1 class="text-3xl font-bold text-red-600 mb-4">❌ Erreur</h1>
            <p class="text-gray-600 mb-4">Échec de la création de l'utilisateur : %s</p>
            <a href="/users" class="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded">← Retour aux Utilisateurs</a>
        </div>
    </div>
</body>
</html>`, err.Error())
			return
		}
		
		c.Header("Content-Type", "text/html")
		c.String(http.StatusOK, `
<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Succès - Bibliothèque de Jeux de Société</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <meta http-equiv="refresh" content="3;url=/users">
</head>
<body class="bg-gray-100">
    <div class="container mx-auto px-4 py-8">
        <div class="max-w-4xl mx-auto bg-white rounded-lg shadow-lg p-6">
            <h1 class="text-3xl font-bold text-green-600 mb-4">✅ Succès !</h1>
            <p class="text-gray-600 mb-4">L'utilisateur "%s" a été inscrit avec succès !</p>
            <div class="bg-green-50 border border-green-200 rounded-lg p-4 mb-4">
                <h3 class="font-semibold text-green-800">Détails de l'Utilisateur :</h3>
                <ul class="text-green-700 mt-2">
                    <li><strong>Nom :</strong> %s</li>
                    <li><strong>Email :</strong> %s</li>
                    <li><strong>ID Utilisateur :</strong> %d</li>
                    <li><strong>Statut :</strong> Actif</li>
                </ul>
            </div>
            <p class="text-sm text-gray-500 mb-4">Vous serez redirigé vers la page des utilisateurs dans 3 secondes...</p>
            <a href="/users" class="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded">← Retour aux Utilisateurs</a>
        </div>
    </div>
</body>
</html>`, user.Name, user.Name, user.Email, user.ID)
	})
	
	// Delete user
	router.POST("/users/:id/delete", func(c *gin.Context) {
		idParam := c.Param("id")
		userID, err := strconv.Atoi(idParam)
		
		if err != nil {
			c.Header("Content-Type", "text/html")
			c.String(http.StatusBadRequest, `
<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Erreur - Bibliothèque de Jeux de Société</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-100">
    <div class="container mx-auto px-4 py-8">
        <div class="max-w-4xl mx-auto bg-white rounded-lg shadow-lg p-6">
            <h1 class="text-3xl font-bold text-red-600 mb-4">❌ Erreur</h1>
            <p class="text-gray-600 mb-4">ID d'utilisateur invalide.</p>
            <a href="/users" class="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded">← Retour aux Utilisateurs</a>
        </div>
    </div>
</body>
</html>`)
			return
		}
		
		err = userService.DeleteUser(userID)
		if err != nil {
			c.Header("Content-Type", "text/html")
			c.String(http.StatusInternalServerError, `
<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Erreur - Bibliothèque de Jeux de Société</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-100">
    <div class="container mx-auto px-4 py-8">
        <div class="max-w-4xl mx-auto bg-white rounded-lg shadow-lg p-6">
            <h1 class="text-3xl font-bold text-red-600 mb-4">❌ Erreur</h1>
            <p class="text-gray-600 mb-4">Échec de la suppression de l'utilisateur : %s</p>
            <a href="/users" class="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded">← Retour aux Utilisateurs</a>
        </div>
    </div>
</body>
</html>`, err.Error())
			return
		}
		
		c.Header("Content-Type", "text/html")
		c.String(http.StatusOK, `
<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Succès - Bibliothèque de Jeux de Société</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <meta http-equiv="refresh" content="2;url=/users">
</head>
<body class="bg-gray-100">
    <div class="container mx-auto px-4 py-8">
        <div class="max-w-4xl mx-auto bg-white rounded-lg shadow-lg p-6">
            <h1 class="text-3xl font-bold text-green-600 mb-4">✅ Utilisateur Supprimé !</h1>
            <p class="text-gray-600 mb-4">L'utilisateur a été supprimé avec succès du système.</p>
            <p class="text-sm text-gray-500 mb-4">Vous serez redirigé vers la page des utilisateurs dans 2 secondes...</p>
            <a href="/users" class="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded">← Retour aux Utilisateurs</a>
        </div>
    </div>
</body>
</html>`)
	})

	// Create new borrowing
	router.POST("/borrowings/create", func(c *gin.Context) {
		userIDStr := strings.TrimSpace(c.PostForm("user_id"))
		gameIDStr := strings.TrimSpace(c.PostForm("game_id"))
		durationDaysStr := strings.TrimSpace(c.PostForm("duration_days"))
		
		if userIDStr == "" || gameIDStr == "" || durationDaysStr == "" {
			c.Header("Content-Type", "text/html")
			c.String(http.StatusBadRequest, `
<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Erreur - Bibliothèque de Jeux de Société</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-100">
    <div class="container mx-auto px-4 py-8">
        <div class="max-w-4xl mx-auto bg-white rounded-lg shadow-lg p-6">
            <h1 class="text-3xl font-bold text-red-600 mb-4">❌ Erreur</h1>
            <p class="text-gray-600 mb-4">Veuillez sélectionner un utilisateur, un jeu et une durée d'emprunt.</p>
            <a href="/borrowings" class="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded">← Retour aux Emprunts</a>
        </div>
    </div>
</body>
</html>`)
			return
		}

		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			c.Header("Content-Type", "text/html")
			c.String(http.StatusBadRequest, `
<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Erreur - Bibliothèque de Jeux de Société</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-100">
    <div class="container mx-auto px-4 py-8">
        <div class="max-w-4xl mx-auto bg-white rounded-lg shadow-lg p-6">
            <h1 class="text-3xl font-bold text-red-600 mb-4">❌ Erreur</h1>
            <p class="text-gray-600 mb-4">ID utilisateur invalide.</p>
            <a href="/borrowings" class="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded">← Retour aux Emprunts</a>
        </div>
    </div>
</body>
</html>`)
			return
		}

		gameID, err := strconv.Atoi(gameIDStr)
		if err != nil {
			c.Header("Content-Type", "text/html")
			c.String(http.StatusBadRequest, `
<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Erreur - Bibliothèque de Jeux de Société</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-100">
    <div class="container mx-auto px-4 py-8">
        <div class="max-w-4xl mx-auto bg-white rounded-lg shadow-lg p-6">
            <h1 class="text-3xl font-bold text-red-600 mb-4">❌ Erreur</h1>
            <p class="text-gray-600 mb-4">ID jeu invalide.</p>
            <a href="/borrowings" class="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded">← Retour aux Emprunts</a>
        </div>
    </div>
</body>
</html>`)
			return
		}

		durationDays, err := strconv.Atoi(durationDaysStr)
		if err != nil || durationDays <= 0 {
			c.Header("Content-Type", "text/html")
			c.String(http.StatusBadRequest, `
<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Erreur - Bibliothèque de Jeux de Société</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-100">
    <div class="container mx-auto px-4 py-8">
        <div class="max-w-4xl mx-auto bg-white rounded-lg shadow-lg p-6">
            <h1 class="text-3xl font-bold text-red-600 mb-4">❌ Erreur</h1>
            <p class="text-gray-600 mb-4">Durée d'emprunt invalide.</p>
            <a href="/borrowings" class="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded">← Retour aux Emprunts</a>
        </div>
    </div>
</body>
</html>`)
			return
		}

		// Calculer la date d'échéance
		dueDate := time.Now().Add(time.Duration(durationDays) * 24 * time.Hour)
		borrowing, err := borrowingService.BorrowGame(userID, gameID, dueDate)
		if err != nil {
			c.Header("Content-Type", "text/html")
			c.String(http.StatusInternalServerError, `
<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Erreur - Bibliothèque de Jeux de Société</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-100">
    <div class="container mx-auto px-4 py-8">
        <div class="max-w-4xl mx-auto bg-white rounded-lg shadow-lg p-6">
            <h1 class="text-3xl font-bold text-red-600 mb-4">❌ Erreur</h1>
            <p class="text-gray-600 mb-4">Échec de la création de l'emprunt : %s</p>
            <a href="/borrowings" class="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded">← Retour aux Emprunts</a>
        </div>
    </div>
</body>
</html>`, err.Error())
			return
		}
		
		c.Header("Content-Type", "text/html")
		c.String(http.StatusOK, `
<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Succès - Bibliothèque de Jeux de Société</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <meta http-equiv="refresh" content="3;url=/borrowings">
</head>
<body class="bg-gray-100">
    <div class="container mx-auto px-4 py-8">
        <div class="max-w-4xl mx-auto bg-white rounded-lg shadow-lg p-6">
            <h1 class="text-3xl font-bold text-green-600 mb-4">✅ Emprunt Créé !</h1>
            <p class="text-gray-600 mb-4">L'emprunt a été créé avec succès !</p>
            <div class="bg-green-50 border border-green-200 rounded-lg p-4 mb-4">
                <h3 class="font-semibold text-green-800">Détails de l'Emprunt :</h3>
                <ul class="text-green-700 mt-2">
                    <li><strong>ID Emprunt :</strong> %d</li>
                    <li><strong>Utilisateur :</strong> %d</li>
                    <li><strong>Jeu :</strong> %d</li>
                    <li><strong>Date d'emprunt :</strong> %s</li>
                    <li><strong>Date d'échéance :</strong> %s</li>
                    <li><strong>Durée :</strong> %d jours</li>
                </ul>
            </div>
            <p class="text-sm text-gray-500 mb-4">Vous serez redirigé vers la page des emprunts dans 3 secondes...</p>
            <a href="/borrowings" class="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded">← Retour aux Emprunts</a>
        </div>
    </div>
</body>
</html>`, borrowing.ID, borrowing.UserID, borrowing.GameID, borrowing.BorrowedAt.Format("2006-01-02 15:04"), borrowing.DueDate.Format("2006-01-02"), durationDays)
	})

	// Return borrowing
	router.POST("/borrowings/:id/return", func(c *gin.Context) {
		idParam := c.Param("id")
		borrowingID, err := strconv.Atoi(idParam)
		
		if err != nil {
			c.Header("Content-Type", "text/html")
			c.String(http.StatusBadRequest, `
<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Erreur - Bibliothèque de Jeux de Société</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-100">
    <div class="container mx-auto px-4 py-8">
        <div class="max-w-4xl mx-auto bg-white rounded-lg shadow-lg p-6">
            <h1 class="text-3xl font-bold text-red-600 mb-4">❌ Erreur</h1>
            <p class="text-gray-600 mb-4">ID d'emprunt invalide.</p>
            <a href="/borrowings" class="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded">← Retour aux Emprunts</a>
        </div>
    </div>
</body>
</html>`)
			return
		}
		
		err = borrowingService.ReturnGame(borrowingID)
		if err != nil {
			c.Header("Content-Type", "text/html")
			c.String(http.StatusInternalServerError, `
<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Erreur - Bibliothèque de Jeux de Société</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-100">
    <div class="container mx-auto px-4 py-8">
        <div class="max-w-4xl mx-auto bg-white rounded-lg shadow-lg p-6">
            <h1 class="text-3xl font-bold text-red-600 mb-4">❌ Erreur</h1>
            <p class="text-gray-600 mb-4">Échec du retour de l'emprunt : %s</p>
            <a href="/borrowings" class="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded">← Retour aux Emprunts</a>
        </div>
    </div>
</body>
</html>`, err.Error())
			return
		}
		
		c.Header("Content-Type", "text/html")
		c.String(http.StatusOK, `
<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Succès - Bibliothèque de Jeux de Société</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <meta http-equiv="refresh" content="2;url=/borrowings">
</head>
<body class="bg-gray-100">
    <div class="container mx-auto px-4 py-8">
        <div class="max-w-4xl mx-auto bg-white rounded-lg shadow-lg p-6">
            <h1 class="text-3xl font-bold text-green-600 mb-4">✅ Jeu Retourné !</h1>
            <p class="text-gray-600 mb-4">Le jeu a été retourné avec succès.</p>
            <p class="text-sm text-gray-500 mb-4">Vous serez redirigé vers la page des emprunts dans 2 secondes...</p>
            <a href="/borrowings" class="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded">← Retour aux Emprunts</a>
        </div>
    </div>
</body>
</html>`)
	})

	// Create custom alert
	router.POST("/alerts/create", func(c *gin.Context) {
		userIDStr := strings.TrimSpace(c.PostForm("user_id"))
		gameIDStr := strings.TrimSpace(c.PostForm("game_id"))
		message := strings.TrimSpace(c.PostForm("message"))
		
		if userIDStr == "" || gameIDStr == "" || message == "" {
			c.Header("Content-Type", "text/html")
			c.String(http.StatusBadRequest, `
<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Erreur - Bibliothèque de Jeux de Société</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-100">
    <div class="container mx-auto px-4 py-8">
        <div class="max-w-4xl mx-auto bg-white rounded-lg shadow-lg p-6">
            <h1 class="text-3xl font-bold text-red-600 mb-4">❌ Erreur</h1>
            <p class="text-gray-600 mb-4">Veuillez remplir tous les champs obligatoires.</p>
            <a href="/alerts" class="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded">← Retour aux Alertes</a>
        </div>
    </div>
</body>
</html>`)
			return
		}

		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			c.Header("Content-Type", "text/html")
			c.String(http.StatusBadRequest, `
<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Erreur - Bibliothèque de Jeux de Société</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-100">
    <div class="container mx-auto px-4 py-8">
        <div class="max-w-4xl mx-auto bg-white rounded-lg shadow-lg p-6">
            <h1 class="text-3xl font-bold text-red-600 mb-4">❌ Erreur</h1>
            <p class="text-gray-600 mb-4">ID utilisateur invalide.</p>
            <a href="/alerts" class="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded">← Retour aux Alertes</a>
        </div>
    </div>
</body>
</html>`)
			return
		}

		gameID, err := strconv.Atoi(gameIDStr)
		if err != nil {
			c.Header("Content-Type", "text/html")
			c.String(http.StatusBadRequest, `
<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Erreur - Bibliothèque de Jeux de Société</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-100">
    <div class="container mx-auto px-4 py-8">
        <div class="max-w-4xl mx-auto bg-white rounded-lg shadow-lg p-6">
            <h1 class="text-3xl font-bold text-red-600 mb-4">❌ Erreur</h1>
            <p class="text-gray-600 mb-4">ID jeu invalide.</p>
            <a href="/alerts" class="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded">← Retour aux Alertes</a>
        </div>
    </div>
</body>
</html>`)
			return
		}
		
		alert, err := alertService.CreateCustomAlert(userID, gameID, "custom", message)
		if err != nil {
			c.Header("Content-Type", "text/html")
			c.String(http.StatusInternalServerError, `
<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Erreur - Bibliothèque de Jeux de Société</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-100">
    <div class="container mx-auto px-4 py-8">
        <div class="max-w-4xl mx-auto bg-white rounded-lg shadow-lg p-6">
            <h1 class="text-3xl font-bold text-red-600 mb-4">❌ Erreur</h1>
            <p class="text-gray-600 mb-4">Échec de la création de l'alerte : %s</p>
            <a href="/alerts" class="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded">← Retour aux Alertes</a>
        </div>
    </div>
</body>
</html>`, err.Error())
			return
		}
		
		c.Header("Content-Type", "text/html")
		c.String(http.StatusOK, `
<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Succès - Bibliothèque de Jeux de Société</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <meta http-equiv="refresh" content="3;url=/alerts">
</head>
<body class="bg-gray-100">
    <div class="container mx-auto px-4 py-8">
        <div class="max-w-4xl mx-auto bg-white rounded-lg shadow-lg p-6">
            <h1 class="text-3xl font-bold text-green-600 mb-4">✅ Alerte Créée !</h1>
            <p class="text-gray-600 mb-4">L'alerte personnalisée a été créée avec succès !</p>
            <div class="bg-green-50 border border-green-200 rounded-lg p-4 mb-4">
                <h3 class="font-semibold text-green-800">Détails de l'Alerte :</h3>
                <ul class="text-green-700 mt-2">
                    <li><strong>ID Alerte :</strong> %d</li>
                    <li><strong>Utilisateur :</strong> %d</li>
                    <li><strong>Jeu :</strong> %d</li>
                    <li><strong>Type :</strong> %s</li>
                    <li><strong>Message :</strong> %s</li>
                </ul>
            </div>
            <p class="text-sm text-gray-500 mb-4">Vous serez redirigé vers la page des alertes dans 3 secondes...</p>
            <a href="/alerts" class="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded">← Retour aux Alertes</a>
        </div>
    </div>
</body>
</html>`, alert.ID, alert.UserID, alert.GameID, alert.Type, alert.Message)
	})

	// Mark alert as read
	router.POST("/alerts/:id/read", func(c *gin.Context) {
		idParam := c.Param("id")
		alertID, err := strconv.Atoi(idParam)
		
		if err != nil {
			c.Header("Content-Type", "text/html")
			c.String(http.StatusBadRequest, `
<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Erreur - Bibliothèque de Jeux de Société</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-100">
    <div class="container mx-auto px-4 py-8">
        <div class="max-w-4xl mx-auto bg-white rounded-lg shadow-lg p-6">
            <h1 class="text-3xl font-bold text-red-600 mb-4">❌ Erreur</h1>
            <p class="text-gray-600 mb-4">ID d'alerte invalide.</p>
            <a href="/alerts" class="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded">← Retour aux Alertes</a>
        </div>
    </div>
</body>
</html>`)
			return
		}
		
		err = alertService.MarkAlertAsRead(alertID)
		if err != nil {
			c.Header("Content-Type", "text/html")
			c.String(http.StatusInternalServerError, `
<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Erreur - Bibliothèque de Jeux de Société</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-100">
    <div class="container mx-auto px-4 py-8">
        <div class="max-w-4xl mx-auto bg-white rounded-lg shadow-lg p-6">
            <h1 class="text-3xl font-bold text-red-600 mb-4">❌ Erreur</h1>
            <p class="text-gray-600 mb-4">Échec du marquage de l'alerte : %s</p>
            <a href="/alerts" class="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded">← Retour aux Alertes</a>
        </div>
    </div>
</body>
</html>`, err.Error())
			return
		}
		
		c.Header("Content-Type", "text/html")
		c.String(http.StatusOK, `
<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Succès - Bibliothèque de Jeux de Société</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <meta http-equiv="refresh" content="2;url=/alerts">
</head>
<body class="bg-gray-100">
    <div class="container mx-auto px-4 py-8">
        <div class="max-w-4xl mx-auto bg-white rounded-lg shadow-lg p-6">
            <h1 class="text-3xl font-bold text-green-600 mb-4">✅ Alerte Marquée !</h1>
            <p class="text-gray-600 mb-4">L'alerte a été marquée comme lue avec succès.</p>
            <p class="text-sm text-gray-500 mb-4">Vous serez redirigé vers la page des alertes dans 2 secondes...</p>
            <a href="/alerts" class="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded">← Retour aux Alertes</a>
        </div>
    </div>
</body>
</html>`)
	})

	// Delete alert
	router.POST("/alerts/:id/delete", func(c *gin.Context) {
		idParam := c.Param("id")
		alertID, err := strconv.Atoi(idParam)
		
		if err != nil {
			c.Header("Content-Type", "text/html")
			c.String(http.StatusBadRequest, `
<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Erreur - Bibliothèque de Jeux de Société</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-100">
    <div class="container mx-auto px-4 py-8">
        <div class="max-w-4xl mx-auto bg-white rounded-lg shadow-lg p-6">
            <h1 class="text-3xl font-bold text-red-600 mb-4">❌ Erreur</h1>
            <p class="text-gray-600 mb-4">ID d'alerte invalide.</p>
            <a href="/alerts" class="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded">← Retour aux Alertes</a>
        </div>
    </div>
</body>
</html>`)
			return
		}
		
		err = alertService.DeleteAlert(alertID)
		if err != nil {
			c.Header("Content-Type", "text/html")
			c.String(http.StatusInternalServerError, `
<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Erreur - Bibliothèque de Jeux de Société</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-100">
    <div class="container mx-auto px-4 py-8">
        <div class="max-w-4xl mx-auto bg-white rounded-lg shadow-lg p-6">
            <h1 class="text-3xl font-bold text-red-600 mb-4">❌ Erreur</h1>
            <p class="text-gray-600 mb-4">Échec de la suppression de l'alerte : %s</p>
            <a href="/alerts" class="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded">← Retour aux Alertes</a>
        </div>
    </div>
</body>
</html>`, err.Error())
			return
		}
		
		c.Header("Content-Type", "text/html")
		c.String(http.StatusOK, `
<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Succès - Bibliothèque de Jeux de Société</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <meta http-equiv="refresh" content="2;url=/alerts">
</head>
<body class="bg-gray-100">
    <div class="container mx-auto px-4 py-8">
        <div class="max-w-4xl mx-auto bg-white rounded-lg shadow-lg p-6">
            <h1 class="text-3xl font-bold text-green-600 mb-4">✅ Alerte Supprimée !</h1>
            <p class="text-gray-600 mb-4">L'alerte a été supprimée avec succès.</p>
            <p class="text-sm text-gray-500 mb-4">Vous serez redirigé vers la page des alertes dans 2 secondes...</p>
            <a href="/alerts" class="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded">← Retour aux Alertes</a>
        </div>
    </div>
</body>
</html>`)
	})

	// Generate overdue alerts
	router.POST("/alerts/generate-overdue", func(c *gin.Context) {
		err := alertService.GenerateOverdueAlerts()
		if err != nil {
			c.Header("Content-Type", "text/html")
			c.String(http.StatusInternalServerError, `
<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Erreur - Bibliothèque de Jeux de Société</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-100">
    <div class="container mx-auto px-4 py-8">
        <div class="max-w-4xl mx-auto bg-white rounded-lg shadow-lg p-6">
            <h1 class="text-3xl font-bold text-red-600 mb-4">❌ Erreur</h1>
            <p class="text-gray-600 mb-4">Échec de la génération des alertes de retard : %s</p>
            <a href="/alerts" class="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded">← Retour aux Alertes</a>
        </div>
    </div>
</body>
</html>`, err.Error())
			return
		}
		
		c.Header("Content-Type", "text/html")
		c.String(http.StatusOK, `
<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Succès - Bibliothèque de Jeux de Société</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <meta http-equiv="refresh" content="2;url=/alerts">
</head>
<body class="bg-gray-100">
    <div class="container mx-auto px-4 py-8">
        <div class="max-w-4xl mx-auto bg-white rounded-lg shadow-lg p-6">
            <h1 class="text-3xl font-bold text-green-600 mb-4">✅ Alertes Générées !</h1>
            <p class="text-gray-600 mb-4">Les alertes de retard ont été générées avec succès.</p>
            <p class="text-sm text-gray-500 mb-4">Vous serez redirigé vers la page des alertes dans 2 secondes...</p>
            <a href="/alerts" class="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded">← Retour aux Alertes</a>
        </div>
    </div>
</body>
</html>`)
	})

	// Generate reminder alerts
	router.POST("/alerts/generate-reminders", func(c *gin.Context) {
		err := alertService.GenerateReminderAlerts()
		if err != nil {
			c.Header("Content-Type", "text/html")
			c.String(http.StatusInternalServerError, `
<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Erreur - Bibliothèque de Jeux de Société</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-100">
    <div class="container mx-auto px-4 py-8">
        <div class="max-w-4xl mx-auto bg-white rounded-lg shadow-lg p-6">
            <h1 class="text-3xl font-bold text-red-600 mb-4">❌ Erreur</h1>
            <p class="text-gray-600 mb-4">Échec de la génération des rappels : %s</p>
            <a href="/alerts" class="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded">← Retour aux Alertes</a>
        </div>
    </div>
</body>
</html>`, err.Error())
			return
		}
		
		c.Header("Content-Type", "text/html")
		c.String(http.StatusOK, `
<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Succès - Bibliothèque de Jeux de Société</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <meta http-equiv="refresh" content="2;url=/alerts">
</head>
<body class="bg-gray-100">
    <div class="container mx-auto px-4 py-8">
        <div class="max-w-4xl mx-auto bg-white rounded-lg shadow-lg p-6">
            <h1 class="text-3xl font-bold text-green-600 mb-4">✅ Rappels Générés !</h1>
            <p class="text-gray-600 mb-4">Les rappels ont été générés avec succès.</p>
            <p class="text-sm text-gray-500 mb-4">Vous serez redirigé vers la page des alertes dans 2 secondes...</p>
            <a href="/alerts" class="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded">← Retour aux Alertes</a>
        </div>
    </div>
</body>
</html>`)
	})

	// Cleanup resolved alerts
	router.POST("/alerts/cleanup", func(c *gin.Context) {
		err := alertService.CleanupResolvedAlerts()
		if err != nil {
			c.Header("Content-Type", "text/html")
			c.String(http.StatusInternalServerError, `
<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Erreur - Bibliothèque de Jeux de Société</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-100">
    <div class="container mx-auto px-4 py-8">
        <div class="max-w-4xl mx-auto bg-white rounded-lg shadow-lg p-6">
            <h1 class="text-3xl font-bold text-red-600 mb-4">❌ Erreur</h1>
            <p class="text-gray-600 mb-4">Échec du nettoyage des alertes : %s</p>
            <a href="/alerts" class="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded">← Retour aux Alertes</a>
        </div>
    </div>
</body>
</html>`, err.Error())
			return
		}
		
		c.Header("Content-Type", "text/html")
		c.String(http.StatusOK, `
<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Succès - Bibliothèque de Jeux de Société</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <meta http-equiv="refresh" content="2;url=/alerts">
</head>
<body class="bg-gray-100">
    <div class="container mx-auto px-4 py-8">
        <div class="max-w-4xl mx-auto bg-white rounded-lg shadow-lg p-6">
            <h1 class="text-3xl font-bold text-green-600 mb-4">✅ Nettoyage Effectué !</h1>
            <p class="text-gray-600 mb-4">Les alertes résolues ont été nettoyées avec succès.</p>
            <p class="text-sm text-gray-500 mb-4">Vous serez redirigé vers la page des alertes dans 2 secondes...</p>
            <a href="/alerts" class="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded">← Retour aux Alertes</a>
        </div>
    </div>
</body>
</html>`)
	})
}

// setupAPIRoutes configures API routes
func setupAPIRoutes(router *gin.Engine,
	gameHandler *handlers.GameHandler,
	userHandler *handlers.UserHandler,
	borrowingHandler *handlers.BorrowingHandler,
	alertHandler *handlers.AlertHandler) {

	api := router.Group("/api/v1")
	{
		// Game API routes
		games := api.Group("/games")
		{
			games.POST("", gameHandler.AddGame)
			games.GET("", gameHandler.GetAllGames)
			games.GET("/search", gameHandler.SearchGames)
			games.GET("/:id", gameHandler.GetGame)
			games.PUT("/:id", gameHandler.UpdateGame)
			games.DELETE("/:id", gameHandler.DeleteGame)
			games.GET("/:id/borrowings", gameHandler.GetGameBorrowingHistory)
			games.GET("/:id/availability", gameHandler.GetGameAvailability)
		}

		// User API routes
		users := api.Group("/users")
		{
			users.POST("", userHandler.RegisterUser)
			users.GET("", userHandler.GetAllUsers)
			users.GET("/:id", userHandler.GetUser)
			users.PUT("/:id", userHandler.UpdateUser)
			users.GET("/:id/borrowings", userHandler.GetUserBorrowings)
			users.GET("/:id/current-loans", userHandler.GetUserCurrentLoans)
			users.GET("/:id/eligibility", userHandler.CheckUserEligibility)
		}

		// Borrowing API routes
		borrowings := api.Group("/borrowings")
		{
			borrowings.POST("", borrowingHandler.BorrowGame)
			borrowings.GET("/:id", borrowingHandler.GetBorrowingDetails)
			borrowings.PUT("/:id/return", borrowingHandler.ReturnGame)
			borrowings.PUT("/:id/extend", borrowingHandler.ExtendDueDate)
			borrowings.GET("/overdue", borrowingHandler.GetOverdueItems)
			borrowings.GET("/due-soon", borrowingHandler.GetItemsDueSoon)
			borrowings.GET("/user/:id", borrowingHandler.GetActiveBorrowingsByUser)
			borrowings.GET("/game/:id", borrowingHandler.GetBorrowingsByGame)
			borrowings.POST("/update-overdue", borrowingHandler.UpdateOverdueStatus)
		}

		// Alert API routes
		alerts := api.Group("/alerts")
		{
			alerts.GET("", alertHandler.GetActiveAlerts)
			alerts.GET("/user/:id", alertHandler.GetAlertsByUser)
			alerts.PUT("/:id/read", alertHandler.MarkAlertAsRead)
			alerts.PUT("/user/:id/read-all", alertHandler.MarkAllUserAlertsAsRead)
			alerts.DELETE("/:id", alertHandler.DeleteAlert)
			alerts.GET("/summary", alertHandler.GetAlertsSummary)
			alerts.GET("/dashboard", alertHandler.GetDashboard)
			alerts.POST("", alertHandler.CreateCustomAlert)
		}
	}
}

// setupTemplateFunctions configures template functions and loads templates
func setupTemplateFunctions(router *gin.Engine) {
	// Define custom template functions
	funcMap := template.FuncMap{
		"substr": func(s string, start, length int) string {
			if start < 0 || start >= len(s) {
				return ""
			}
			end := start + length
			if end > len(s) {
				end = len(s)
			}
			return s[start:end]
		},
		"title": strings.Title,
		"upper": strings.ToUpper,
		"lower": strings.ToLower,
		"contains": strings.Contains,
		"hasPrefix": strings.HasPrefix,
		"hasSuffix": strings.HasSuffix,
		"join": strings.Join,
		"split": strings.Split,
		"replace": strings.ReplaceAll,
		"trim": strings.TrimSpace,
	}

	// Load templates with custom functions
	tmpl := template.Must(template.New("").Funcs(funcMap).ParseGlob("web/templates/*/*.html"))
	router.SetHTMLTemplate(tmpl)
}