package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "termsOfService": "http://swagger.io/terms/",
        "contact": {
            "name": "Support API",
            "url": "http://www.swagger.io/support",
            "email": "support@swagger.io"
        },
        "license": {
            "name": "MIT",
            "url": "https://opensource.org/licenses/MIT"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/games": {
            "get": {
                "description": "Récupère la liste de tous les jeux avec recherche optionnelle",
                "consumes": ["application/json"],
                "produces": ["application/json"],
                "tags": ["games"],
                "summary": "Lister tous les jeux",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Terme de recherche",
                        "name": "search",
                        "in": "query"
                    },
                    {
                        "type": "boolean",
                        "description": "Filtrer par disponibilité",
                        "name": "available",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Liste des jeux",
                        "schema": {"type": "object"}
                    },
                    "500": {
                        "description": "Erreur serveur",
                        "schema": {"type": "object"}
                    }
                }
            },
            "post": {
                "description": "Ajoute un nouveau jeu à la bibliothèque",
                "consumes": ["application/json"],
                "produces": ["application/json"],
                "tags": ["games"],
                "summary": "Ajouter un nouveau jeu",
                "parameters": [
                    {
                        "description": "Informations du jeu",
                        "name": "game",
                        "in": "body",
                        "required": true,
                        "schema": {"$ref": "#/definitions/handlers.AddGameRequest"}
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Jeu créé avec succès",
                        "schema": {"type": "object"}
                    },
                    "400": {
                        "description": "Données invalides",
                        "schema": {"type": "object"}
                    },
                    "500": {
                        "description": "Erreur serveur",
                        "schema": {"type": "object"}
                    }
                }
            }
        },
        "/games/{id}": {
            "get": {
                "description": "Récupère les détails d'un jeu spécifique",
                "consumes": ["application/json"],
                "produces": ["application/json"],
                "tags": ["games"],
                "summary": "Obtenir un jeu par ID",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "ID du jeu",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Détails du jeu",
                        "schema": {"type": "object"}
                    },
                    "404": {
                        "description": "Jeu non trouvé",
                        "schema": {"type": "object"}
                    }
                }
            },
            "put": {
                "description": "Met à jour les informations d'un jeu",
                "consumes": ["application/json"],
                "produces": ["application/json"],
                "tags": ["games"],
                "summary": "Mettre à jour un jeu",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "ID du jeu",
                        "name": "id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "Nouvelles informations du jeu",
                        "name": "game",
                        "in": "body",
                        "required": true,
                        "schema": {"$ref": "#/definitions/handlers.AddGameRequest"}
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Jeu mis à jour avec succès",
                        "schema": {"type": "object"}
                    },
                    "400": {
                        "description": "Données invalides",
                        "schema": {"type": "object"}
                    },
                    "404": {
                        "description": "Jeu non trouvé",
                        "schema": {"type": "object"}
                    }
                }
            },
            "delete": {
                "description": "Supprime un jeu de la bibliothèque",
                "consumes": ["application/json"],
                "produces": ["application/json"],
                "tags": ["games"],
                "summary": "Supprimer un jeu",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "ID du jeu",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Jeu supprimé avec succès",
                        "schema": {"type": "object"}
                    },
                    "404": {
                        "description": "Jeu non trouvé",
                        "schema": {"type": "object"}
                    },
                    "409": {
                        "description": "Impossible de supprimer (contraintes)",
                        "schema": {"type": "object"}
                    }
                }
            }
        },
        "/users": {
            "get": {
                "description": "Récupère la liste de tous les utilisateurs",
                "consumes": ["application/json"],
                "produces": ["application/json"],
                "tags": ["users"],
                "summary": "Lister tous les utilisateurs",
                "responses": {
                    "200": {
                        "description": "Liste des utilisateurs",
                        "schema": {"type": "object"}
                    },
                    "500": {
                        "description": "Erreur serveur",
                        "schema": {"type": "object"}
                    }
                }
            },
            "post": {
                "description": "Inscrit un nouvel utilisateur",
                "consumes": ["application/json"],
                "produces": ["application/json"],
                "tags": ["users"],
                "summary": "Inscrire un nouvel utilisateur",
                "parameters": [
                    {
                        "description": "Informations de l'utilisateur",
                        "name": "user",
                        "in": "body",
                        "required": true,
                        "schema": {"$ref": "#/definitions/handlers.RegisterUserRequest"}
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Utilisateur créé avec succès",
                        "schema": {"type": "object"}
                    },
                    "400": {
                        "description": "Données invalides",
                        "schema": {"type": "object"}
                    },
                    "500": {
                        "description": "Erreur serveur",
                        "schema": {"type": "object"}
                    }
                }
            }
        },
        "/users/{id}": {
            "get": {
                "description": "Récupère les détails d'un utilisateur spécifique",
                "consumes": ["application/json"],
                "produces": ["application/json"],
                "tags": ["users"],
                "summary": "Obtenir un utilisateur par ID",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "ID de l'utilisateur",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Détails de l'utilisateur",
                        "schema": {"type": "object"}
                    },
                    "404": {
                        "description": "Utilisateur non trouvé",
                        "schema": {"type": "object"}
                    }
                }
            },
            "put": {
                "description": "Met à jour les informations d'un utilisateur",
                "consumes": ["application/json"],
                "produces": ["application/json"],
                "tags": ["users"],
                "summary": "Mettre à jour un utilisateur",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "ID de l'utilisateur",
                        "name": "id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "Nouvelles informations de l'utilisateur",
                        "name": "user",
                        "in": "body",
                        "required": true,
                        "schema": {"$ref": "#/definitions/handlers.RegisterUserRequest"}
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Utilisateur mis à jour avec succès",
                        "schema": {"type": "object"}
                    },
                    "400": {
                        "description": "Données invalides",
                        "schema": {"type": "object"}
                    },
                    "404": {
                        "description": "Utilisateur non trouvé",
                        "schema": {"type": "object"}
                    }
                }
            }
        },
        "/borrowings": {
            "get": {
                "description": "Récupère la liste de tous les emprunts",
                "consumes": ["application/json"],
                "produces": ["application/json"],
                "tags": ["borrowings"],
                "summary": "Lister tous les emprunts",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "ID de l'utilisateur pour filtrer",
                        "name": "user_id",
                        "in": "query"
                    },
                    {
                        "type": "boolean",
                        "description": "Filtrer par emprunts actifs seulement",
                        "name": "active_only",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Liste des emprunts",
                        "schema": {"type": "object"}
                    },
                    "500": {
                        "description": "Erreur serveur",
                        "schema": {"type": "object"}
                    }
                }
            },
            "post": {
                "description": "Crée un nouvel emprunt",
                "consumes": ["application/json"],
                "produces": ["application/json"],
                "tags": ["borrowings"],
                "summary": "Créer un nouvel emprunt",
                "parameters": [
                    {
                        "description": "Informations de l'emprunt",
                        "name": "borrowing",
                        "in": "body",
                        "required": true,
                        "schema": {"$ref": "#/definitions/handlers.BorrowGameRequest"}
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Emprunt créé avec succès",
                        "schema": {"type": "object"}
                    },
                    "400": {
                        "description": "Données invalides",
                        "schema": {"type": "object"}
                    },
                    "500": {
                        "description": "Erreur serveur",
                        "schema": {"type": "object"}
                    }
                }
            }
        },
        "/borrowings/{id}": {
            "get": {
                "description": "Récupère les détails d'un emprunt spécifique",
                "consumes": ["application/json"],
                "produces": ["application/json"],
                "tags": ["borrowings"],
                "summary": "Obtenir un emprunt par ID",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "ID de l'emprunt",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Détails de l'emprunt",
                        "schema": {"type": "object"}
                    },
                    "404": {
                        "description": "Emprunt non trouvé",
                        "schema": {"type": "object"}
                    }
                }
            }
        },
        "/borrowings/{id}/return": {
            "post": {
                "description": "Marque un emprunt comme retourné",
                "consumes": ["application/json"],
                "produces": ["application/json"],
                "tags": ["borrowings"],
                "summary": "Retourner un jeu emprunté",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "ID de l'emprunt",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Jeu retourné avec succès",
                        "schema": {"type": "object"}
                    },
                    "404": {
                        "description": "Emprunt non trouvé",
                        "schema": {"type": "object"}
                    },
                    "400": {
                        "description": "Emprunt déjà retourné",
                        "schema": {"type": "object"}
                    }
                }
            }
        },
        "/alerts": {
            "get": {
                "description": "Récupère la liste de toutes les alertes",
                "consumes": ["application/json"],
                "produces": ["application/json"],
                "tags": ["alerts"],
                "summary": "Lister toutes les alertes",
                "parameters": [
                    {
                        "type": "boolean",
                        "description": "Filtrer par alertes non lues seulement",
                        "name": "unread_only",
                        "in": "query"
                    },
                    {
                        "type": "integer",
                        "description": "ID de l'utilisateur pour filtrer",
                        "name": "user_id",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Liste des alertes",
                        "schema": {"type": "object"}
                    },
                    "500": {
                        "description": "Erreur serveur",
                        "schema": {"type": "object"}
                    }
                }
            },
            "post": {
                "description": "Crée une nouvelle alerte personnalisée",
                "consumes": ["application/json"],
                "produces": ["application/json"],
                "tags": ["alerts"],
                "summary": "Créer une nouvelle alerte",
                "parameters": [
                    {
                        "description": "Informations de l'alerte",
                        "name": "alert",
                        "in": "body",
                        "required": true,
                        "schema": {"$ref": "#/definitions/handlers.CreateAlertRequest"}
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Alerte créée avec succès",
                        "schema": {"type": "object"}
                    },
                    "400": {
                        "description": "Données invalides",
                        "schema": {"type": "object"}
                    },
                    "500": {
                        "description": "Erreur serveur",
                        "schema": {"type": "object"}
                    }
                }
            }
        },
        "/alerts/{id}": {
            "delete": {
                "description": "Supprime une alerte",
                "consumes": ["application/json"],
                "produces": ["application/json"],
                "tags": ["alerts"],
                "summary": "Supprimer une alerte",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "ID de l'alerte",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Alerte supprimée avec succès",
                        "schema": {"type": "object"}
                    },
                    "404": {
                        "description": "Alerte non trouvée",
                        "schema": {"type": "object"}
                    }
                }
            }
        },
        "/alerts/{id}/read": {
            "post": {
                "description": "Marque une alerte comme lue",
                "consumes": ["application/json"],
                "produces": ["application/json"],
                "tags": ["alerts"],
                "summary": "Marquer une alerte comme lue",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "ID de l'alerte",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Alerte marquée comme lue",
                        "schema": {"type": "object"}
                    },
                    "404": {
                        "description": "Alerte non trouvée",
                        "schema": {"type": "object"}
                    }
                }
            }
        }
    },
    "definitions": {
        "handlers.AddGameRequest": {
            "type": "object",
            "required": ["name"],
            "properties": {
                "category": {"type": "string", "description": "Catégorie du jeu (ex: Stratégie, Famille)"},
                "condition": {"type": "string", "description": "État du jeu (excellent, good, fair, poor)"},
                "description": {"type": "string", "description": "Description du jeu"},
                "name": {"type": "string", "description": "Nom du jeu"}
            }
        },
        "handlers.RegisterUserRequest": {
            "type": "object",
            "required": ["name", "email"],
            "properties": {
                "name": {"type": "string", "description": "Nom complet de l'utilisateur"},
                "email": {"type": "string", "description": "Adresse email de l'utilisateur"}
            }
        },
        "handlers.BorrowGameRequest": {
            "type": "object",
            "required": ["user_id", "game_id"],
            "properties": {
                "user_id": {"type": "integer", "description": "ID de l'utilisateur qui emprunte"},
                "game_id": {"type": "integer", "description": "ID du jeu à emprunter"},
                "due_date": {"type": "string", "format": "date", "description": "Date d'échéance (optionnel, 14 jours par défaut)"}
            }
        },
        "handlers.CreateAlertRequest": {
            "type": "object",
            "required": ["user_id", "game_id", "type", "message"],
            "properties": {
                "user_id": {"type": "integer", "description": "ID de l'utilisateur concerné"},
                "game_id": {"type": "integer", "description": "ID du jeu concerné"},
                "type": {"type": "string", "enum": ["overdue", "reminder", "custom"], "description": "Type d'alerte"},
                "message": {"type": "string", "description": "Message de l'alerte"}
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "localhost:8080",
	BasePath:         "/api/v1",
	Schemes:          []string{"http"},
	Title:            "Board Game Library API",
	Description:      "API pour la gestion d'une bibliothèque de jeux de société",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}