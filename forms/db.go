package forms

import "encore.dev/storage/sqldb"

var formsDb = sqldb.NewDatabase("forms_db", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})
