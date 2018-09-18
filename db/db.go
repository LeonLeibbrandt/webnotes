package db

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"strings"

	"web/config"

	"github.com/jackc/pgx"
)

type DB struct {
	*pgx.ConnPool
	config    *config.Config
	pgxConfig pgx.ConnConfig
	mac       hash.Hash
}

const (
	SECRET = "secret key 12345"
)

func NewDB(cnf *config.Config) (*DB, error) {
	d := &DB{
		config: cnf,
		mac:    hmac.New(sha256.New, []byte(SECRET)),
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
	err = d.migrateDB()
	if err != nil {
		d.ConnPool.Close()
		return nil, err
	}
	return d, nil
}

func (d *DB) Auth(username, password, ip string) (string, bool) {
	fmt.Printf("%s -> %s\n", username, password)
	/*
		insert into webuser(email, session)
		values('leon', '[]' || jsonb_build_object(
			'ip', '192.168.1.101',
			'cookie', crypt('leonpassword192.168.1.101', gen_salt('bf')),
			'valid', now() + interval '24 hours')
		);

		update webuser set session = session ||	jsonb_build_object(
			'ip', '192.168.1.103',
			'cookie', crypt('leonpassword192.168.1.103', gen_salt('bf')),
			'valid', now() + interval '24 hours')
		where _id=6
		select w._id, (r->>'valid')::timestamp from webuser w, jsonb_array_elements(w.session) r where r->>'cookie' = crypt('leonpassword192.168.1.101', r->>'cookie')

		select * from webuser
	*/

	var id int64
	token := ""
	err := d.QueryRow("SELECT _id FROM webuser WHERE email=$1 and password=crypt($2, password)",
		username, password).Scan(&id)
	if err != nil {
		fmt.Printf("At 1 %v\n", err)
		return token, false
	}
	// Generate a cookie
	d.mac.Reset()
	var b bytes.Buffer
	b.WriteString(username)
	b.WriteString(password)
	b.WriteString(ip)
	d.mac.Write(b.Bytes())
	token = hex.EncodeToString(d.mac.Sum(nil))
	token = strings.ToUpper(token)
	_, err = d.Exec(fmt.Sprintf("UPDATE webuser SET session = session || jsonb_build_object('ip', '%s', 'cookie', '%s') WHERE _id=%v;", ip, token, id))
	if err != nil {
		fmt.Printf("At 2 %v\n", err)
		return "", false
	}
	return token, true
}

func (d *DB) Session(session, ip string) bool {
	var id int64
	err := d.QueryRow("SELECT w._id FROM webuser w, jsonb_array_elements(w.session) r where r->>'cookie' = $1 AND r->>'ip' = $2;", session, ip).Scan(&id)
	if err != nil {
		return false
	}
	return true
}
