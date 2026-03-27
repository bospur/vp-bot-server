package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"go-server/internal/middleware"
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

	if user == nil || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
		http.Error(w, "неверный логин или пароль", http.StatusUnauthorized)
		return
	}

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
	claims := middleware.ClaimsFromContext(r)

	var input repository.AnimalInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "неверный формат запроса", http.StatusBadRequest)
		return
	}

	animal, err := h.animalRepo.Create(claims.ClinicID, input)
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
	claims := middleware.ClaimsFromContext(r)

	var input repository.CategoryInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "неверный формат запроса", http.StatusBadRequest)
		return
	}

	category, err := h.animalRepo.CreateCategory(claims.ClinicID, input)
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

// GetAdminArticles обрабатывает GET /api/admin/articles
func (h *AdminHandler) GetAdminArticles(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromContext(r)

	articles, err := h.articleRepo.GetAll(claims.ClinicID)
	if err != nil {
		log.Printf("ошибка получения статей: %v", err)
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
	if articles == nil {
		articles = []repository.Article{}
	}
	writeJSON(w, http.StatusOK, articles)
}

// GetAdminArticle обрабатывает GET /api/admin/articles/{id}
func (h *AdminHandler) GetAdminArticle(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "неверный запрос", http.StatusBadRequest)
		return
	}

	article, err := h.articleRepo.GetByID(id)
	if err != nil {
		log.Printf("ошибка получения статьи: %v", err)
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
	if article == nil {
		http.Error(w, "не найдено", http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, article)
}

// GetArticleCategories обрабатывает GET /api/admin/articles/{id}/categories
func (h *AdminHandler) GetArticleCategories(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "неверный запрос", http.StatusBadRequest)
		return
	}

	categories, err := h.articleRepo.GetCategories(id)
	if err != nil {
		log.Printf("ошибка получения категорий статьи: %v", err)
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
	if categories == nil {
		categories = []repository.Category{}
	}
	writeJSON(w, http.StatusOK, categories)
}

// CreateArticle обрабатывает POST /api/admin/articles
func (h *AdminHandler) CreateArticle(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromContext(r)

	var input repository.ArticleInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "неверный формат запроса", http.StatusBadRequest)
		return
	}

	article, err := h.articleRepo.Create(claims.ClinicID, input)
	if err != nil {
		log.Printf("ошибка создания статьи: %v", err)
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, article)
}

// UpdateArticle обрабатывает PUT /api/admin/articles/{id}
// Editor не может редактировать опубликованные статьи
func (h *AdminHandler) UpdateArticle(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "неверный запрос", http.StatusBadRequest)
		return
	}

	claims := middleware.ClaimsFromContext(r)
	if claims.Role == "editor" {
		status, err := h.articleRepo.GetStatus(id)
		if err != nil {
			http.Error(w, "не найдено", http.StatusNotFound)
			return
		}
		if status == "published" {
			http.Error(w, "нельзя редактировать опубликованную статью", http.StatusForbidden)
			return
		}
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

// UpdateArticleStatus обрабатывает PATCH /api/admin/articles/{id}/status
// Только admin может менять статус
func (h *AdminHandler) UpdateArticleStatus(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromContext(r)
	if claims.Role != "admin" {
		http.Error(w, "доступ запрещён", http.StatusForbidden)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "неверный запрос", http.StatusBadRequest)
		return
	}

	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "неверный формат запроса", http.StatusBadRequest)
		return
	}
	if body.Status != "draft" && body.Status != "published" {
		http.Error(w, "недопустимый статус", http.StatusBadRequest)
		return
	}

	article, err := h.articleRepo.UpdateStatus(id, body.Status)
	if err != nil {
		log.Printf("ошибка обновления статуса статьи: %v", err)
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
// Editor может удалять только черновики
func (h *AdminHandler) DeleteArticle(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "неверный запрос", http.StatusBadRequest)
		return
	}

	claims := middleware.ClaimsFromContext(r)
	if claims.Role == "editor" {
		status, err := h.articleRepo.GetStatus(id)
		if err != nil {
			http.Error(w, "не найдено", http.StatusNotFound)
			return
		}
		if status == "published" {
			http.Error(w, "нельзя удалить опубликованную статью", http.StatusForbidden)
			return
		}
	}

	if err := h.articleRepo.Delete(id); err != nil {
		log.Printf("ошибка удаления статьи: %v", err)
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ── Users CRUD (только admin) ─────────────────────────────────────────────────

// GetAdminUsers обрабатывает GET /api/admin/users
func (h *AdminHandler) GetAdminUsers(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromContext(r)
	if claims.Role != "admin" {
		http.Error(w, "доступ запрещён", http.StatusForbidden)
		return
	}

	users, err := h.userRepo.GetAll(claims.ClinicID)
	if err != nil {
		log.Printf("ошибка получения пользователей: %v", err)
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
	if users == nil {
		users = []repository.User{}
	}
	writeJSON(w, http.StatusOK, users)
}

// CreateAdminUser обрабатывает POST /api/admin/users
func (h *AdminHandler) CreateAdminUser(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromContext(r)
	if claims.Role != "admin" {
		http.Error(w, "доступ запрещён", http.StatusForbidden)
		return
	}

	var body struct {
		Login    string `json:"login"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "неверный формат запроса", http.StatusBadRequest)
		return
	}
	if body.Login == "" || body.Password == "" {
		http.Error(w, "логин и пароль обязательны", http.StatusBadRequest)
		return
	}
	if body.Role != "admin" && body.Role != "editor" && body.Role != "groomer" {
		http.Error(w, "недопустимая роль", http.StatusBadRequest)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	user, err := h.userRepo.Create(claims.ClinicID, body.Login, string(hash), body.Role)
	if err != nil {
		log.Printf("ошибка создания пользователя: %v", err)
		http.Error(w, "пользователь с таким логином уже существует", http.StatusConflict)
		return
	}

	writeJSON(w, http.StatusCreated, user)
}

// DeleteAdminUser обрабатывает DELETE /api/admin/users/{id}
func (h *AdminHandler) DeleteAdminUser(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromContext(r)
	if claims.Role != "admin" {
		http.Error(w, "доступ запрещён", http.StatusForbidden)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "неверный запрос", http.StatusBadRequest)
		return
	}

	if err := h.userRepo.Delete(id); err != nil {
		log.Printf("ошибка удаления пользователя: %v", err)
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
