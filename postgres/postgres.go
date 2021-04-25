package postgres

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/kelseyhightower/envconfig"
	_ "github.com/lib/pq"
	"github.com/mager/cafebean-api/config"
)

// ProvidePostgres provides a postgres client
func ProvidePostgres() *sql.DB {
	// Initialize config
	var conf config.Config

	err := envconfig.Process("cafebean", &conf)
	if err != nil {
		log.Fatal(err.Error())
	}
	var (
		host     = conf.PostgresHostname
		port     = 5432
		user     = "postgres"
		password = conf.PostgresPassword
		dbname   = "postgres"
	)

	psqlInfo := fmt.Sprintf(
		"host=%s port=%d user=%s "+"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		fmt.Println(err.Error())
		return &sql.DB{}
	}
	// defer db.Close()
	err = db.Ping()
	if err != nil {
		fmt.Println(err.Error())
		return &sql.DB{}
	}
	fmt.Printf("Successfully connected to Postgres db (%s)\n", host)

	return db
}

var Options = ProvidePostgres
