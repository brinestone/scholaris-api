package tenants

import "encore.dev/storage/sqldb"

var tenantDb = sqldb.NewDatabase("tenants", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})
