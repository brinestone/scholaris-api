package settings

import "encore.dev/storage/sqldb"

var db = sqldb.NewDatabase("settings_db", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})
