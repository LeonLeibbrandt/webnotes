package db

import (
	"fmt"
	"strings"

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
	return nil
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
		fmt.Println("Creating version")
		_, err := conn.Exec(versiontable)
		if err != nil {
			return err
		}
	}

	tableVersion := make(map[string]int)
	rows, err := conn.Query("select name, version from version")
	if err != nil && err != pgx.ErrNoRows {
		return err
	}
	for rows.Next() {
		var tableName string
		var version int
		err = rows.Scan(&tableName, &version)
		if err != nil {
			return err
		}
		tableVersion[tableName] = version
	}
	for tableName := range tables {
		version, ok := tableVersion[tableName]
		if !ok {
			version = -1
		}
		err = d.migrateTable(tableName, version, conn)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *DB) migrateTable(tableName string, currentVersion int, conn *pgx.Conn) error {
	queries := tables[tableName]
	if currentVersion < len(queries)-1 {
		fmt.Printf("Migrating table %s from %d to %d\n", tableName, currentVersion, len(queries)-1)
		var builder strings.Builder
		builder.Reset()
		if currentVersion == -1 {
			builder.WriteString(fmt.Sprintf("INSERT INTO version(name, version) values('%s', -1);", tableName))
		}
		for indx, query := range queries {
			if indx > currentVersion {
				builder.WriteString(query)
				currentVersion = indx
			}
		}
		builder.WriteString(fmt.Sprintf("UPDATE version SET version=%d WHERE name='%s';", currentVersion, tableName))
		_, err := conn.Exec(builder.String())
		if err != nil {
			fmt.Println(builder.String())
			return err
		}
	}
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

var tables map[string][]string = map[string][]string{
	"webuser": []string{`CREATE TABLE public.webuser
(
    _id bigserial NOT NULL,
    email text COLLATE pg_catalog."default" NOT NULL,
    session jsonb DEFAULT '[]'::jsonb,
    CONSTRAINT webuser_pkey PRIMARY KEY (_id),
    CONSTRAINT unq_webuser_email UNIQUE (email)

)
WITH (
    OIDS = FALSE
)
TABLESPACE pg_default;
ALTER TABLE public.webuser
    OWNER to leon;`,
	},
}
