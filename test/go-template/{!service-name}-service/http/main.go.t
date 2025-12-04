package main

import (
	"▶service-name◀-service/shared/env"
	"▶service-name◀-service/shared/postgres/repositories"
	"log"
)

func main() {
	if err := env.ReadPath("../.env"); err != nil {
		log.Fatal(err)
	}

	_, db := repositories.New(env.Postgres())
	defer db.Close()
}
