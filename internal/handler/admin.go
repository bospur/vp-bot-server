package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"go-server/internal/repository"

	"github.com/golang-jwt/jwt/v5"
)

// AdminHandler содержит зависимости для административных эндпоинтов
type AdminHandler struct {
	animalRepo  *repository.AnimalRepository
	articleRepo *repository.ArticleRepository
	adminLogin  string
	adminPass   string
	jwtSecret   string
}

// NewAdminHandler создаёт новый хендлер
func NewAdminHandler(
	animalRepo *repository.AnimalRepository,
	articleRepo *repository.ArticleRepository,
	adminLogin, adminPass, jwtSecret string,
) *AdminHandler {
	return &AdminHandler{
		animalRepo:  animalRepo,
		articleRepo: articleRepo,
		adminLogin:  adminLogin,
		adminPass:   adminPass,
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

	if req.Login != h.adminLogin || req.Password != h.adminPass {
		http.Error(w, "неверный логин или пароль", http.StatusUnauthorized)
		return
	}

	// Создаём JWT токен с истечением через 24 часа
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"login": req.Login,
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
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

	// TODO: брать clinicID из JWT токена когда заменим на таблицу users
	animal, err := h.animalRepo.Create(1, input)
	if err != nil {
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

	// TODO: брать clinicID из JWT токена когда заменим на таблицу users
	article, err := h.articleRepo.Create(1, input)
	if err != nil {
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
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
