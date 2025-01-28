package migrations

import (
	"database/sql"
	"embed"
	"log"

	migrate "github.com/rubenv/sql-migrate"
)

//go:embed *
var migrationsFiles embed.FS

func Run(db *sql.DB) error {

	migrations := &migrate.EmbedFileSystemMigrationSource{
		FileSystem: migrationsFiles,
		Root:       ".",
	}

	count, err := migrate.Exec(db, "mysql", migrations, migrate.Up)
	if err != nil {
		return err
	}

	log.Println(count)

	return nil
}
