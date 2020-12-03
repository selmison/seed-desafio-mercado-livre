package mercadolivre

import "database/sql"

// Config is used to configure the server
type Config struct {
	// Host defines the network addresses we bind to.
	Host string
	// Port defines the network port we bind to.
	Port int
	// DB defines the DB we connect to.
	DB *sql.DB
	// DriverName defines the database driver name.
	DriverName string
}
