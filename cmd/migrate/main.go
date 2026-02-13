package main

import (
	"flag"
	"log"

	"go-next-cms/internal/config"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	direction := flag.String("direction", "up", "up or down")
	flag.Parse()
	cfg := config.Load()
	m, err := migrate.New("file://migrations", cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer m.Close()
	if *direction == "up" {
		err = m.Up()
	} else {
		err = m.Down()
	}
	if err != nil && err != migrate.ErrNoChange {
		log.Fatal(err)
	}
}
