package main

import (
	"bye-service/shared/env"
	"bye-service/shared/postgres/repositories"
	"log"
)

func main() {
	if err := env.ReadPath("../.env"); err != nil {
		log.Fatal(err)
	}

	_, db := repositories.New(env.Postgres())
	defer db.Close()
}
