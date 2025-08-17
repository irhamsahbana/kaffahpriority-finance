package adapter

import (
	// "log"

	"codebase-app/internal/infrastructure/config"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
)

func WithPostgres() Option {
	return func(a *Adapter) {
		dbUser := config.Envs.Postgres.Username
		dbPassword := config.Envs.Postgres.Password
		dbName := config.Envs.Postgres.Database
		dbHost := config.Envs.Postgres.Host
		dbEnv := config.Envs.Postgres.Env
		dbSSLMode := config.Envs.Postgres.SslMode
		dbPort := config.Envs.Postgres.Port

		dbMaxPoolSize := config.Envs.DB.MaxOpenCons
		dbMaxIdleConns := config.Envs.DB.MaxIdleCons
		dbConnMaxLifetime := config.Envs.DB.ConnMaxLifetime

		connectionString := "user=" + dbUser + " password=" + dbPassword + " host=" + dbHost + " port=" + dbPort + " dbname=" + dbName + " sslmode=" + dbSSLMode + " TimeZone=UTC"
		db, err := sqlx.Connect("postgres", connectionString)
		if err != nil {
			log.Fatal().Err(err).Msg("Error connecting to Postgres")
		}

		db.SetMaxOpenConns(dbMaxPoolSize)
		db.SetMaxIdleConns(dbMaxIdleConns)
		db.SetConnMaxLifetime(time.Duration(dbConnMaxLifetime) * time.Second)

		// check connection
		err = db.Ping()
		if err != nil {
			log.Fatal().Err(err).Msg("Error connecting to Digihub Postgres")
		}

		a.Postgres = db
		log.Info().Msg("Postgres connected")

		if strings.EqualFold(dbEnv, "production") {
			lines := []string{
				"⚠ PRODUCTION DATABASE",
				fmt.Sprintf("Host: %s:%s", dbHost, dbPort),
				fmt.Sprintf("DB:   %s", dbName),
				fmt.Sprintf("User: %s", dbUser),
			}
			printBox(lines)
		}
	}
}

func printBox(lines []string) {
	max := 0
	for _, s := range lines {
		if l := utf8.RuneCountInString(s); l > max {
			max = l
		}
	}
	pad := 1
	width := max + pad*2

	fmt.Println("┌" + strings.Repeat("─", width) + "┐")
	for _, s := range lines {
		l := utf8.RuneCountInString(s)
		right := width - pad - l // = pad + (max - l)
		fmt.Printf("│%s%s%s│\n",
			strings.Repeat(" ", pad),
			s,
			strings.Repeat(" ", right),
		)
	}
	fmt.Println("└" + strings.Repeat("─", width) + "┘")
}
