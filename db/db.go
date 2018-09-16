package db

import (
	"web/config"

	"github.com/jackc/pgx"
)

type DB struct {
	*pgx.ConnPool
	config    *config.Config
	pgxConfig pgx.ConnConfig
}

func NewDB(cnf *config.Config) (*DB, error) {
	d := &DB{
		config: cnf,
	}
	var err error
	d.pgxConfig, err = pgx.ParseConnectionString(cnf.ConnStr)
	if err != nil {
		return nil, err
	}
	pgxConnPoolConfig := pgx.ConnPoolConfig{
		ConnConfig:     d.pgxConfig,
		MaxConnections: d.config.MaxDBConnections,
	}
	d.ConnPool, err = pgx.NewConnPool(pgxConnPoolConfig)
	if err != nil {
		if err.(pgx.PgError).Code == "3D000" {
			err := d.createDB(pgxConnPoolConfig)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	return d, nil
}

func (d *DB) Auth(username, password string) (string, bool) {
	return "", true
}

func (d *DB) Session(session, ip string) bool {
	return true
}
