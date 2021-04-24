package postgres

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

const (
	host     = "35.225.71.58"
	port     = 5432
	user     = "postgres"
	password = "TODO"
	dbname   = "postgres"
)

// ProvidePostgres provides a postgres client
func ProvidePostgres() *sql.DB {
	psqlInfo := fmt.Sprintf(
		"host=%s port=%d user=%s "+"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		panic(err)
	}

	return db
}

var Options = ProvidePostgres
