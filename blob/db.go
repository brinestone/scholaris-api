package blob

import "encore.dev/storage/sqldb"

var db = sqldb.NewDatabase("blob_db", sqldb.DatabaseConfig{Migrations: "./migrations"})
