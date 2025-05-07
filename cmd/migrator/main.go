package main

import (
	"chalk/pkg/log"
	"chalk/pkg/migrator"
	"context"
	"flag"

	"github.com/jackc/pgx/v5"
)

func main() {
	postgresUri := flag.String("postgres_uri", "postgres://postgres:password@localhost:5432/postgres", "poatgres uri")
	migrationsPath := flag.String("migrations", "/etc/chalk/migrations", "path to the directory with migration files")
	up := flag.Bool("up", true, "up migrations")
	down := flag.Int("down", 0, "down migrations steps")
	flag.Parse()

	pcli, err := pgx.Connect(context.Background(), *postgresUri)
	if err != nil {
		log.Errorf("postgres connect: %v", err)
		return
	}
	defer pcli.Close(context.Background())

	migrator := migrator.NewPgx(pcli, *migrationsPath)

	if *up {
		err = migrator.Up()
		if err != nil {
			log.Errorf("up migrations error: %v", err)
		}
	}

	if *down > 0 {
		err = migrator.Down(1)
		if err != nil {
			log.Errorf("down migrations error: %v", err)
		}
	}
}
