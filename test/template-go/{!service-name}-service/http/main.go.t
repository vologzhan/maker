package main

import (
	"log"
	"▶service-name◀-service/shared/env"
	"▶service-name◀-service/shared/postgres/repositories"
)

func main() {
	if err := env.ReadPath("../.env"); err != nil {
		log.Fatal(err)
	}

	_, db := repositories.New(env.Postgres())
	defer db.Close()
}
