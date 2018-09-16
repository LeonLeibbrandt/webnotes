package db

import (
	"fmt"

	"github.com/jackc/pgx"
)

func (d *DB) createDB(pgxConnPoolConfig pgx.ConnPoolConfig) error {
	connConfig := d.pgxConfig
	connConfig.Database = "postgres"
	conn, err := pgx.Connect(connConfig)
	if err != nil {
		return err
	}
	_, err = conn.Exec(fmt.Sprintf("CREATE DATABASE %s OWNER %s", d.pgxConfig.Database, d.pgxConfig.User))
	if err != nil {
		return err
	}
	conn.Close()
	d.ConnPool, err = pgx.NewConnPool(pgxConnPoolConfig)
	return err
}
