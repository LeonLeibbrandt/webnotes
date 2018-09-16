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
	if err != nil {
		return err
	}
	err = d.migrateDB()
	return err
}

func (d *DB) migrateDB() error {
	conn, err := d.ConnPool.Acquire()
	if err != nil {
		return err
	}
	defer d.ConnPool.Release(conn)

	var exists bool
	err = conn.QueryRow("SELECT to_regclass('public.version')::text IS NOT NULL AS exists").Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		_, err := conn.Exec(versiontable)
		if err != nil {
			return err
		}
	}
	// select to_regclass('public.version')::text is not null as exists
	tx, err := conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	return nil
}

var versiontable string = `
CREATE TABLE public.version
(
    name text NOT NULL,
    version bigint NOT NULL,
    CONSTRAINT pk_version PRIMARY KEY (name),
    CONSTRAINT unq_version_name UNIQUE (name)
)
WITH (
    OIDS = FALSE
);
ALTER TABLE public.version
    OWNER to leon;`
