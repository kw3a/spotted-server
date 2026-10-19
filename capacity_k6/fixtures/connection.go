package main

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
)

// Los limites tambien cubren el saludo/autenticacion del protocolo MySQL,
// durante los que el driver no observa siempre la cancelacion del contexto.
func openDatabase(source string, connectTimeout, ioTimeout time.Duration) (*sql.DB, error) {
	config, err := mysql.ParseDSN(source)
	if err != nil {
		return nil, fmt.Errorf("DATABASE_URL no tiene formato DSN MySQL valido")
	}
	config.Timeout = connectTimeout
	if ioTimeout > connectTimeout {
		ioTimeout = connectTimeout
	}
	config.ReadTimeout, config.WriteTimeout = ioTimeout, ioTimeout
	connector, err := mysql.NewConnector(config)
	if err != nil {
		return nil, fmt.Errorf("configuracion MySQL invalida: %w", err)
	}
	return sql.OpenDB(connector), nil
}
