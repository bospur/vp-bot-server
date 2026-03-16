package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"go-server/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// AdminHandler содержит зависимости для административных эндпоинтов
type AdminHandler struct {
	animalRepo  *repository.AnimalRepository
	articleRepo *repository.ArticleRepository
	userRepo    *repository.UserRepository
	jwtSecret   string
}

// NewAdminHandler создаёт новый хендлер
func NewAdminHandler(
	animalRepo *repository.AnimalRepository,
	articleRepo *repository.ArticleRepository,
	userRepo *repository.UserRepository,
	jwtSecret string,
) *AdminHandler {
	return &AdminHandler{
		animalRepo:  animalRepo,
		articleRepo: articleRepo,
		userRepo:    userRepo,
		jwtSecret:   jwtSecret,
	}
}

// ── Авторизация ──────────────────────────────────────────────────────────────

type loginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string `json:"token"`
}

// Login обрабатывает POST /api/admin/login
func (h *AdminHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "неверный формат запроса", http.StatusBadRequest)
		return
	}

	user, err := h.userRepo.GetByLogin(req.Login)
	if err != nil {
		log.Printf("ошибка получения пользователя: %v", err)
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	// Пользователь не найден или пароль неверный — одинаковое сообщение (безопасность)
	if user == nil || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
		http.Error(w, "неверный логин или пароль", http.StatusUnauthorized)
		return
	}

	// Создаём JWT с данными пользователя
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":   user.ID,
		"clinic_id": user.ClinicID,
		"role":      user.Role,
		"exp":       time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString([]byte(h.jwtSecret))
	if err != nil {
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, loginResponse{Token: tokenString})
}

// ── Animals CRUD ─────────────────────────────────────────────────────────────

// CreateAnimal обрабатывает POST /api/admin/animals
func (h *AdminHandler) CreateAnimal(w http.ResponseWriter, r *http.Request) {
	var input repository.AnimalInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "неверный формат запроса", http.StatusBadRequest)
		return
	}

	// TODO: брать clinicID из JWT токена
	animal, err := h.animalRepo.Create(1, input)
	if err != nil {
		log.Printf("ошибка создания животного: %v", err)
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, animal)
}

// UpdateAnimal обрабатывает PUT /api/admin/animals/{id}
func (h *AdminHandler) UpdateAnimal(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "неверный запрос", http.StatusBadRequest)
		return
	}

	var input repository.AnimalInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "неверный формат запроса", http.StatusBadRequest)
		return
	}

	animal, err := h.animalRepo.Update(id, input)
	if err != nil {
		log.Printf("ошибка обновления животного: %v", err)
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
	if animal == nil {
		http.Error(w, "не найдено", http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, animal)
}

// DeleteAnimal обрабатывает DELETE /api/admin/animals/{id}
func (h *AdminHandler) DeleteAnimal(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "неверный запрос", http.StatusBadRequest)
		return
	}

	if err := h.animalRepo.Delete(id); err != nil {
		log.Printf("ошибка удаления животного: %v", err)
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ── Categories CRUD ───────────────────────────────────────────────────────────

// CreateCategory обрабатывает POST /api/admin/categories
func (h *AdminHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var input repository.CategoryInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "неверный формат запроса", http.StatusBadRequest)
		return
	}

	// TODO: брать clinicID из JWT токена
	category, err := h.animalRepo.CreateCategory(1, input)
	if err != nil {
		log.Printf("ошибка создания категории: %v", err)
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, category)
}

// UpdateCategory обрабатывает PUT /api/admin/categories/{id}
func (h *AdminHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "неверный запрос", http.StatusBadRequest)
		return
	}

	var input repository.CategoryInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "неверный формат запроса", http.StatusBadRequest)
		return
	}

	category, err := h.animalRepo.UpdateCategory(id, input)
	if err != nil {
		log.Printf("ошибка обновления категории: %v", err)
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
	if category == nil {
		http.Error(w, "не найдено", http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, category)
}

// DeleteCategory обрабатывает DELETE /api/admin/categories/{id}
func (h *AdminHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "неверный запрос", http.StatusBadRequest)
		return
	}

	if err := h.animalRepo.DeleteCategory(id); err != nil {
		log.Printf("ошибка удаления категории: %v", err)
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ── Articles CRUD ─────────────────────────────────────────────────────────────

// CreateArticle обрабатывает POST /api/admin/articles
func (h *AdminHandler) CreateArticle(w http.ResponseWriter, r *http.Request) {
	var input repository.ArticleInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "неверный формат запроса", http.StatusBadRequest)
		return
	}

	// TODO: брать clinicID из JWT токена
	article, err := h.articleRepo.Create(1, input)
	if err != nil {
		log.Printf("ошибка создания статьи: %v", err)
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, article)
}

// UpdateArticle обрабатывает PUT /api/admin/articles/{id}
func (h *AdminHandler) UpdateArticle(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "неверный запрос", http.StatusBadRequest)
		return
	}

	var input repository.ArticleInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "неверный формат запроса", http.StatusBadRequest)
		return
	}

	article, err := h.articleRepo.Update(id, input)
	if err != nil {
		log.Printf("ошибка обновления статьи: %v", err)
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
	if article == nil {
		http.Error(w, "не найдено", http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, article)
}

// DeleteArticle обрабатывает DELETE /api/admin/articles/{id}
func (h *AdminHandler) DeleteArticle(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "неверный запрос", http.StatusBadRequest)
		return
	}

	if err := h.articleRepo.Delete(id); err != nil {
		log.Printf("ошибка удаления статьи: %v", err)
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// AssignArticleToCategory обрабатывает POST /api/admin/articles/{id}/categories/{categoryId}
func (h *AdminHandler) AssignArticleToCategory(w http.ResponseWriter, r *http.Request) {
	articleID := r.PathValue("id")
	categoryID := r.PathValue("categoryId")
	if articleID == "" || categoryID == "" {
		http.Error(w, "неверный запрос", http.StatusBadRequest)
		return
	}

	if err := h.articleRepo.AssignToCategory(articleID, categoryID); err != nil {
		log.Printf("ошибка привязки статьи к категории: %v", err)
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RemoveArticleFromCategory обрабатывает DELETE /api/admin/articles/{id}/categories/{categoryId}
func (h *AdminHandler) RemoveArticleFromCategory(w http.ResponseWriter, r *http.Request) {
	articleID := r.PathValue("id")
	categoryID := r.PathValue("categoryId")
	if articleID == "" || categoryID == "" {
		http.Error(w, "неверный запрос", http.StatusBadRequest)
		return
	}

	if err := h.articleRepo.RemoveFromCategory(articleID, categoryID); err != nil {
		log.Printf("ошибка отвязки статьи от категории: %v", err)
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
