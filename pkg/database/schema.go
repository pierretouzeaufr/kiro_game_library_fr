package database

// getInitialMigrations returns the initial database schema migrations
func getInitialMigrations() []Migration {
	return []Migration{
		{
			Version: 1,
			Name:    "create_users_table",
			Up: `
				CREATE TABLE users (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					name TEXT NOT NULL,
					email TEXT UNIQUE NOT NULL,
					registered_at DATETIME DEFAULT CURRENT_TIMESTAMP,
					is_active BOOLEAN DEFAULT TRUE
				);
			`,
			Down: "DROP TABLE users;",
		},
		{
			Version: 2,
			Name:    "create_games_table",
			Up: `
				CREATE TABLE games (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					name TEXT NOT NULL,
					description TEXT,
					category TEXT,
					entry_date DATETIME DEFAULT CURRENT_TIMESTAMP,
					condition TEXT DEFAULT 'good',
					is_available BOOLEAN DEFAULT TRUE
				);
			`,
			Down: "DROP TABLE games;",
		},
		{
			Version: 3,
			Name:    "create_borrowings_table",
			Up: `
				CREATE TABLE borrowings (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					user_id INTEGER NOT NULL,
					game_id INTEGER NOT NULL,
					borrowed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
					due_date DATETIME NOT NULL,
					returned_at DATETIME,
					is_overdue BOOLEAN DEFAULT FALSE,
					FOREIGN KEY (user_id) REFERENCES users(id),
					FOREIGN KEY (game_id) REFERENCES games(id)
				);
			`,
			Down: "DROP TABLE borrowings;",
		},
		{
			Version: 4,
			Name:    "create_alerts_table",
			Up: `
				CREATE TABLE alerts (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					user_id INTEGER NOT NULL,
					game_id INTEGER NOT NULL,
					type TEXT NOT NULL,
					message TEXT NOT NULL,
					created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
					is_read BOOLEAN DEFAULT FALSE,
					FOREIGN KEY (user_id) REFERENCES users(id),
					FOREIGN KEY (game_id) REFERENCES games(id)
				);
			`,
			Down: "DROP TABLE alerts;",
		},
		{
			Version: 5,
			Name:    "create_indexes",
			Up: `
				CREATE INDEX idx_borrowings_user_id ON borrowings(user_id);
				CREATE INDEX idx_borrowings_game_id ON borrowings(game_id);
				CREATE INDEX idx_borrowings_due_date ON borrowings(due_date);
				CREATE INDEX idx_alerts_user_id ON alerts(user_id);
				CREATE INDEX idx_alerts_is_read ON alerts(is_read);
			`,
			Down: `
				DROP INDEX idx_borrowings_user_id;
				DROP INDEX idx_borrowings_game_id;
				DROP INDEX idx_borrowings_due_date;
				DROP INDEX idx_alerts_user_id;
				DROP INDEX idx_alerts_is_read;
			`,
		},
	}
}