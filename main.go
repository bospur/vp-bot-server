package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"go-server/internal/db"
)

type Response struct {
	Message string `json:"message"`
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	res := Response{
		Message: "Hello from Go server",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func main() {
	// Читаем строку подключения из переменной окружения
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL не задан")
	}

	// Подключаемся к БД
	database, err := db.Connect(databaseURL)
	if err != nil {
		log.Fatalf("не удалось подключиться к БД: %v", err)
	}
	defer database.Close()
	log.Println("подключение к БД установлено")

	// Применяем миграции
	if err := db.RunMigrations(database); err != nil {
		log.Fatalf("ошибка миграций: %v", err)
	}

	http.HandleFunc("/hello", helloHandler)

	log.Println("server started :8080")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
