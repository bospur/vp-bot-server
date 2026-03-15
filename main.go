package main

import (
	"log"
	"net/http"
	"os"

	"go-server/internal/bot"
	"go-server/internal/db"
	"go-server/internal/handler"
	"go-server/internal/middleware"
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

	adminLogin := os.Getenv("ADMIN_LOGIN")
	adminPass := os.Getenv("ADMIN_PASSWORD")
	jwtSecret := os.Getenv("JWT_SECRET")
	if adminLogin == "" || adminPass == "" || jwtSecret == "" {
		log.Fatal("ADMIN_LOGIN, ADMIN_PASSWORD, JWT_SECRET не заданы")
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

	// Репозитории
	animalRepo := repository.NewAnimalRepository(database)
	articleRepo := repository.NewArticleRepository(database)

	// Публичные хендлеры
	animalHandler := handler.NewAnimalHandler(animalRepo)
	articleHandler := handler.NewArticleHandler(articleRepo)

	// Админ хендлер
	adminHandler := handler.NewAdminHandler(animalRepo, articleRepo, adminLogin, adminPass, jwtSecret)

	// ── Публичные роуты ──────────────────────────────────────────────────────
	http.HandleFunc("/api/animals", animalHandler.GetAnimals)
	http.HandleFunc("/api/animals/{slug}/categories", animalHandler.GetCategories)
	http.HandleFunc("/api/animals/{animalSlug}/categories/{categorySlug}/articles", articleHandler.GetArticles)
	http.HandleFunc("/api/articles/{slug}", articleHandler.GetArticle)

	// ── Авторизация ──────────────────────────────────────────────────────────
	http.HandleFunc("POST /api/admin/login", adminHandler.Login)

	// ── Защищённые админ роуты ───────────────────────────────────────────────
	auth := func(h http.HandlerFunc) http.HandlerFunc {
		return middleware.Auth(jwtSecret, h)
	}

	// Animals
	http.HandleFunc("POST /api/admin/animals", auth(adminHandler.CreateAnimal))
	http.HandleFunc("PUT /api/admin/animals/{id}", auth(adminHandler.UpdateAnimal))
	http.HandleFunc("DELETE /api/admin/animals/{id}", auth(adminHandler.DeleteAnimal))

	// Articles
	http.HandleFunc("POST /api/admin/articles", auth(adminHandler.CreateArticle))
	http.HandleFunc("PUT /api/admin/articles/{id}", auth(adminHandler.UpdateArticle))
	http.HandleFunc("DELETE /api/admin/articles/{id}", auth(adminHandler.DeleteArticle))
	http.HandleFunc("POST /api/admin/articles/{id}/categories/{categoryId}", auth(adminHandler.AssignArticleToCategory))
	http.HandleFunc("DELETE /api/admin/articles/{id}/categories/{categoryId}", auth(adminHandler.RemoveArticleFromCategory))

	// Telegram бот в отдельной горутине
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
