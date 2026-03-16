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

	"golang.org/x/crypto/bcrypt"
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
	clinicSlug := os.Getenv("CLINIC_SLUG")
	if clinicSlug == "" {
		log.Fatal("CLINIC_SLUG не задан")
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET не задан")
	}
	// Используются только для создания первого пользователя при первом запуске
	adminLogin := os.Getenv("ADMIN_LOGIN")
	adminPass := os.Getenv("ADMIN_PASSWORD")

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
	userRepo := repository.NewUserRepository(database)

	// Создаём первого admin пользователя если таблица users пустая
	if adminLogin != "" && adminPass != "" {
		count, err := userRepo.Count()
		if err != nil {
			log.Fatalf("ошибка проверки пользователей: %v", err)
		}
		if count == 0 {
			hash, err := bcrypt.GenerateFromPassword([]byte(adminPass), bcrypt.DefaultCost)
			if err != nil {
				log.Fatalf("ошибка хеширования пароля: %v", err)
			}
			if _, err := userRepo.Create(1, adminLogin, string(hash), "admin"); err != nil {
				log.Fatalf("ошибка создания admin пользователя: %v", err)
			}
			log.Printf("создан первый пользователь: %s", adminLogin)
		}
	}

	// Хендлеры
	animalHandler := handler.NewAnimalHandler(animalRepo)
	articleHandler := handler.NewArticleHandler(articleRepo)
	adminHandler := handler.NewAdminHandler(animalRepo, articleRepo, userRepo, jwtSecret)

	// ── Публичные роуты ──────────────────────────────────────────────────────
	http.HandleFunc("/api/clinics/{clinicSlug}/animals", animalHandler.GetAnimals)
	http.HandleFunc("/api/clinics/{clinicSlug}/animals/{slug}/categories", animalHandler.GetCategories)
	http.HandleFunc("/api/clinics/{clinicSlug}/animals/{animalSlug}/categories/{categorySlug}/articles", articleHandler.GetArticles)
	http.HandleFunc("/api/clinics/{clinicSlug}/articles/{slug}", articleHandler.GetArticle)

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

	// Categories
	http.HandleFunc("POST /api/admin/categories", auth(adminHandler.CreateCategory))
	http.HandleFunc("PUT /api/admin/categories/{id}", auth(adminHandler.UpdateCategory))
	http.HandleFunc("DELETE /api/admin/categories/{id}", auth(adminHandler.DeleteCategory))

	// Articles
	http.HandleFunc("GET /api/admin/articles", auth(adminHandler.GetAdminArticles))
	http.HandleFunc("GET /api/admin/articles/{id}", auth(adminHandler.GetAdminArticle))
	http.HandleFunc("GET /api/admin/articles/{id}/categories", auth(adminHandler.GetArticleCategories))
	http.HandleFunc("POST /api/admin/articles", auth(adminHandler.CreateArticle))
	http.HandleFunc("PUT /api/admin/articles/{id}", auth(adminHandler.UpdateArticle))
	http.HandleFunc("DELETE /api/admin/articles/{id}", auth(adminHandler.DeleteArticle))
	http.HandleFunc("POST /api/admin/articles/{id}/categories/{categoryId}", auth(adminHandler.AssignArticleToCategory))
	http.HandleFunc("DELETE /api/admin/articles/{id}/categories/{categoryId}", auth(adminHandler.RemoveArticleFromCategory))

	// Telegram бот
	tgBot, err := bot.New(botToken, clinicSlug, animalRepo, articleRepo)
	if err != nil {
		log.Fatalf("ошибка инициализации бота: %v", err)
	}
	go tgBot.Start()

	log.Println("server started :8080")
	if err := http.ListenAndServe(":8080", middleware.CORS(http.DefaultServeMux)); err != nil {
		log.Fatal(err)
	}
}
