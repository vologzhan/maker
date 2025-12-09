package env

import (
	"fmt"
	"▶service-name◀-service/shared/postgres"

	"github.com/spf13/viper"
)

func ReadPath(path string) error {
	viper.SetConfigType("env")
	viper.SetConfigFile(path)

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	return nil
}

func Postgres() postgres.Config {
	return postgres.Config{
		viper.GetString("POSTGRES_HOST"),
		viper.GetString("POSTGRES_PORT"),
		viper.GetString("POSTGRES_DATABASE"),
		viper.GetString("POSTGRES_USER"),
		viper.GetString("POSTGRES_PASSWORD"),
	}
}

func MigrationsDir() string {
	return viper.GetString("MIGRATIONS_DIR")
}
