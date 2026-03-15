package main

import (
	"log"
	"net/http"
	"os"

	"go-server/internal/bot"
	"go-server/internal/db"
	"go-server/internal/handler"
	"go-server/internal/repository"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL не задан")
	}

	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN не задан")
	}

	database, err := db.Connect(databaseURL)
	if err != nil {
		log.Fatalf("не удалось подключиться к БД: %v", err)
	}
	defer database.Close()
	log.Println("подключение к БД установлено")

	if err := db.RunMigrations(database); err != nil {
		log.Fatalf("ошибка миграций: %v", err)
	}

	// Инициализируем репозитории
	animalRepo := repository.NewAnimalRepository(database)
	articleRepo := repository.NewArticleRepository(database)

	// HTTP хендлеры
	animalHandler := handler.NewAnimalHandler(animalRepo)
	articleHandler := handler.NewArticleHandler(articleRepo)

	// Роуты
	http.HandleFunc("/api/animals", animalHandler.GetAnimals)
	http.HandleFunc("/api/animals/{slug}/categories", animalHandler.GetCategories)
	http.HandleFunc("/api/animals/{animalSlug}/categories/{categorySlug}/articles", articleHandler.GetArticles)
	http.HandleFunc("/api/articles/{slug}", articleHandler.GetArticle)

	// Telegram бот запускается в отдельной горутине
	// (горутина — это лёгкий поток в Go, позволяет делать несколько вещей одновременно)
	tgBot, err := bot.New(botToken, animalRepo, articleRepo)
	if err != nil {
		log.Fatalf("ошибка инициализации бота: %v", err)
	}
	go tgBot.Start()

	log.Println("server started :8080")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
