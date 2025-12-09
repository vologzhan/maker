MIGRATION=go run ./cmd/migration/main.go

up_migration:
	$(MIGRATION) up

down_migration:
	$(MIGRATION) down

create_migration:
	$(MIGRATION) create
