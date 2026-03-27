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

	uploadsDir := os.Getenv("UPLOADS_DIR")
	if uploadsDir == "" {
		uploadsDir = "./uploads"
	}
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		log.Fatalf("не удалось создать папку uploads: %v", err)
	}

	// Репозитории
	animalRepo := repository.NewAnimalRepository(database)
	articleRepo := repository.NewArticleRepository(database)
	userRepo := repository.NewUserRepository(database)
	doctorRepo := repository.NewDoctorRepository(database)
	groomingRepo := repository.NewGroomingRepository(database)

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
	doctorHandler := handler.NewDoctorHandler(doctorRepo, uploadsDir)
	groomingHandler := handler.NewGroomingHandler(groomingRepo)

	// ── Публичные роуты ──────────────────────────────────────────────────────
	http.HandleFunc("/api/clinics/{clinicSlug}/animals", animalHandler.GetAnimals)
	http.HandleFunc("/api/clinics/{clinicSlug}/animals/{slug}/categories", animalHandler.GetCategories)
	http.HandleFunc("/api/clinics/{clinicSlug}/animals/{animalSlug}/categories/{categorySlug}/articles", articleHandler.GetArticles)
	http.HandleFunc("/api/clinics/{clinicSlug}/articles/{slug}", articleHandler.GetArticle)
	http.HandleFunc("GET /api/clinics/{clinicSlug}/doctors", doctorHandler.GetPublicDoctors)
	http.HandleFunc("GET /api/clinics/{clinicSlug}/schedule", doctorHandler.GetPublicSchedule)

	// Статические файлы (фото врачей)
	http.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir(uploadsDir))))

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

	// Users (только admin)
	http.HandleFunc("GET /api/admin/users", auth(adminHandler.GetAdminUsers))
	http.HandleFunc("POST /api/admin/users", auth(adminHandler.CreateAdminUser))
	http.HandleFunc("DELETE /api/admin/users/{id}", auth(adminHandler.DeleteAdminUser))

	// Articles
	http.HandleFunc("GET /api/admin/articles", auth(adminHandler.GetAdminArticles))
	http.HandleFunc("GET /api/admin/articles/{id}", auth(adminHandler.GetAdminArticle))
	http.HandleFunc("GET /api/admin/articles/{id}/categories", auth(adminHandler.GetArticleCategories))
	http.HandleFunc("POST /api/admin/articles", auth(adminHandler.CreateArticle))
	http.HandleFunc("PUT /api/admin/articles/{id}", auth(adminHandler.UpdateArticle))
	http.HandleFunc("PATCH /api/admin/articles/{id}/status", auth(adminHandler.UpdateArticleStatus))
	http.HandleFunc("DELETE /api/admin/articles/{id}", auth(adminHandler.DeleteArticle))
	http.HandleFunc("POST /api/admin/articles/{id}/categories/{categoryId}", auth(adminHandler.AssignArticleToCategory))
	http.HandleFunc("DELETE /api/admin/articles/{id}/categories/{categoryId}", auth(adminHandler.RemoveArticleFromCategory))

	// Doctors
	http.HandleFunc("GET /api/admin/doctors", auth(doctorHandler.GetDoctors))
	http.HandleFunc("GET /api/admin/doctors/{id}", auth(doctorHandler.GetDoctor))
	http.HandleFunc("POST /api/admin/doctors", auth(doctorHandler.CreateDoctor))
	http.HandleFunc("PUT /api/admin/doctors/{id}", auth(doctorHandler.UpdateDoctor))
	http.HandleFunc("PATCH /api/admin/doctors/{id}/status", auth(doctorHandler.UpdateDoctorStatus))
	http.HandleFunc("POST /api/admin/doctors/{id}/photo", auth(doctorHandler.UploadDoctorPhoto))
	http.HandleFunc("DELETE /api/admin/doctors/{id}", auth(doctorHandler.DeleteDoctor))

	// Schedule
	http.HandleFunc("GET /api/admin/doctors/{id}/schedule", auth(doctorHandler.GetDoctorSchedule))
	http.HandleFunc("POST /api/admin/doctors/{id}/schedule", auth(doctorHandler.AddScheduleSlot))
	http.HandleFunc("DELETE /api/admin/doctors/{id}/schedule/{slotId}", auth(doctorHandler.DeleteScheduleSlot))

	// Schedule exceptions
	http.HandleFunc("GET /api/admin/doctors/{id}/schedule/exceptions", auth(doctorHandler.GetExceptions))
	http.HandleFunc("PUT /api/admin/doctors/{id}/schedule/exceptions", auth(doctorHandler.UpsertException))
	http.HandleFunc("DELETE /api/admin/doctors/{id}/schedule/exceptions/{exceptionId}", auth(doctorHandler.DeleteException))

	// Settings
	http.HandleFunc("GET /api/admin/settings", auth(doctorHandler.GetSettings))
	http.HandleFunc("PATCH /api/admin/settings", auth(doctorHandler.UpdateSettings))

	// Grooming breeds
	http.HandleFunc("GET /api/admin/grooming/breeds", auth(groomingHandler.GetBreeds))
	http.HandleFunc("POST /api/admin/grooming/breeds", auth(groomingHandler.CreateBreed))
	http.HandleFunc("PUT /api/admin/grooming/breeds/{id}", auth(groomingHandler.UpdateBreed))
	http.HandleFunc("DELETE /api/admin/grooming/breeds/{id}", auth(groomingHandler.DeleteBreed))

	// Grooming weekly template
	http.HandleFunc("GET /api/admin/grooming/template", auth(groomingHandler.GetTemplate))
	http.HandleFunc("PUT /api/admin/grooming/template", auth(groomingHandler.UpsertTemplateSlot))
	http.HandleFunc("DELETE /api/admin/grooming/template/{dayOfWeek}", auth(groomingHandler.DeleteTemplateSlot))

	// Grooming appointments
	http.HandleFunc("GET /api/admin/grooming/appointments", auth(groomingHandler.GetAppointments))
	http.HandleFunc("POST /api/admin/grooming/appointments", auth(groomingHandler.CreateAppointment))
	http.HandleFunc("DELETE /api/admin/grooming/appointments/{id}", auth(groomingHandler.DeleteAppointment))

	publicURL := os.Getenv("PUBLIC_URL")
	if publicURL == "" {
		publicURL = "https://api.snzbeachvolleyball25.ru"
	}
	appURL := os.Getenv("APP_URL")
	if appURL == "" {
		appURL = "https://app.snzbeachvolleyball25.ru"
	}

	// Telegram бот
	tgBot, err := bot.New(botToken, clinicSlug, publicURL, appURL, animalRepo, articleRepo, doctorRepo)
	if err != nil {
		log.Fatalf("ошибка инициализации бота: %v", err)
	}
	go tgBot.Start()

	log.Println("server started :8080")
	if err := http.ListenAndServe(":8080", middleware.CORS(http.DefaultServeMux)); err != nil {
		log.Fatal(err)
	}
}
