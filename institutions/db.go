package institutions

import "encore.dev/storage/sqldb"

var db = sqldb.NewDatabase("institutions_db", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})
