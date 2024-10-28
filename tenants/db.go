package tenants

import "encore.dev/storage/sqldb"

var tenantDb = sqldb.NewDatabase("tenants_db", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})
