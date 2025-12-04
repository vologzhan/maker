package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/uptrace/bun/migrate"
	"os"
	"path/filepath"
	"strconv"
	"▶service-name◀-service/shared/env"
	"▶service-name◀-service/shared/postgres"
)

type command string

const (
	up     command = "up"
	down   command = "down"
	create command = "create"
)

func main() {
	cmd := getCommandOrExitWithHint()

	err := env.ReadPath(".env")
	if err != nil {
		printErrorAndExit(err)
	}

	dir, err := getMigrationsDir()
	if err != nil {
		printErrorAndExit(err)
	}

	ctx := context.Background()

	migrator, err := newMigrator(ctx, dir)
	if err != nil {
		printErrorAndExit(err)
	}

	switch cmd {
	case up:
		_, err = migrator.Migrate(ctx)
	case down:
		_, err = migrator.Rollback(ctx)
	case create:
		err = createMigrationFiles(ctx, migrator, dir)
	}

	if err != nil {
		printErrorAndExit(err)
	}
}

func getCommandOrExitWithHint() command {
	if len(os.Args) == 2 {
		c := command(os.Args[1])
		if c == up || c == down || c == create {
			return c
		}
	}

	fmt.Print(`Usage: migration COMMAND

Commands:
  up             Migrate to the latest available version
  down           Rollback the last migration
  create         Create new up and down migration files
`)

	os.Exit(0)

	return ""
}

func printErrorAndExit(err error) {
	fmt.Println(err)
	os.Exit(1)
}

func newMigrator(ctx context.Context, dir string) (*migrate.Migrator, error) {
	db := postgres.New(env.Postgres())
	migrations := migrate.NewMigrations(migrate.WithMigrationsDirectory(dir))

	err := migrations.Discover(os.DirFS(dir))
	if err != nil {
		return nil, err
	}

	migrator := migrate.NewMigrator(db, migrations)

	err = migrator.Init(ctx)
	if err != nil {
		return nil, err
	}

	return migrator, nil
}

func getMigrationsDir() (string, error) {
	projectDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	migrationDir := env.MigrationsDir()
	if migrationDir == "" {
		return "", errors.New("migrations directory not set")
	}

	migrationPath := filepath.Join(projectDir, migrationDir)

	info, err := os.Stat(migrationPath)
	if os.IsNotExist(err) {
		err = os.Mkdir(migrationPath, 0755)
		if err != nil {
			return "", err
		}

		return migrationPath, nil
	}

	if !info.IsDir() {
		return "", fmt.Errorf("'%s' exists but is not a directory", migrationPath)
	}

	return migrationPath, nil
}

func createMigrationFiles(ctx context.Context, migrator *migrate.Migrator, dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		printErrorAndExit(err)
	}
	migrationName := strconv.Itoa(len(entries) / 2)

	files, err := migrator.CreateSQLMigrations(ctx, migrationName)
	if err != nil {
		return err
	}

	for _, file := range files {
		err = os.WriteFile(file.Path, []byte{}, 0644) // erase file content
		if err != nil {
			return err
		}
	}

	return nil
}
