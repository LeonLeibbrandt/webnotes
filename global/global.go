package global

import (
	"web/config"
	"web/db"
)

type Global struct {
	Config *config.Config
	DB     *db.DB
}

func NewGlobal(cnf *config.Config) (*Global, error) {
	g := &Global{
		Config: cnf,
	}
	var err error
	g.DB, err = db.NewDB(g.Config)
	if err != nil {
		return nil, err
	}
	return g, err
}
