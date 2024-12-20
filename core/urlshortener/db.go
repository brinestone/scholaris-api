package urlshortener

import "encore.dev/storage/sqldb"

var linksDb = sqldb.NewDatabase("links_db", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})
