package main

import (
	"hello-service/shared/env"
	"hello-service/shared/postgres/repositories"
	"log"
)

func main() {
	if err := env.ReadPath("../.env"); err != nil {
		log.Fatal(err)
	}

	_, db := repositories.New(env.Postgres())
	defer db.Close()
}
