package users

import "encore.dev/storage/sqldb"

var userDb = sqldb.NewDatabase("users_db", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})
