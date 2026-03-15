package db

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

// Connect открывает соединение с PostgreSQL и проверяет что оно работает
func Connect(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("ошибка открытия БД: %w", err)
	}

	// Проверяем что соединение реально работает
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("БД недоступна: %w", err)
	}

	return db, nil
}

// RunMigrations применяет все новые миграции из папки migrations/
func RunMigrations(db *sql.DB) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("ошибка создания драйвера миграций: %w", err)
	}

	// "file://migrations" — путь к папке с .sql файлами
	m, err := migrate.NewWithDatabaseInstance("file://migrations", "postgres", driver)
	if err != nil {
		return fmt.Errorf("ошибка инициализации миграций: %w", err)
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("ошибка применения миграций: %w", err)
	}

	if err == migrate.ErrNoChange {
		log.Println("миграции: нет новых изменений")
	} else {
		log.Println("миграции: успешно применены")
	}

	return nil
}
