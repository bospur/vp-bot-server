package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"go-server/internal/repository"
)

// AnimalHandler содержит зависимости для HTTP хендлеров животных
type AnimalHandler struct {
	repo *repository.AnimalRepository
}

// NewAnimalHandler создаёт новый хендлер
func NewAnimalHandler(repo *repository.AnimalRepository) *AnimalHandler {
	return &AnimalHandler{repo: repo}
}

// GetAnimals обрабатывает GET /api/animals
// Возвращает список всех животных
func (h *AnimalHandler) GetAnimals(w http.ResponseWriter, r *http.Request) {
	animals, err := h.repo.GetAll()
	if err != nil {
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	// Если животных нет — возвращаем пустой массив, не null
	if animals == nil {
		animals = []repository.Animal{}
	}

	writeJSON(w, http.StatusOK, animals)
}

// GetCategories обрабатывает GET /api/animals/{slug}/categories
// Возвращает категории для конкретного животного
func (h *AnimalHandler) GetCategories(w http.ResponseWriter, r *http.Request) {
	// Извлекаем slug из URL вручную
	// URL вида: /api/animals/cat/categories
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 3 {
		http.Error(w, "неверный запрос", http.StatusBadRequest)
		return
	}
	slug := parts[2] // ["api", "animals", "cat", "categories"]

	categories, err := h.repo.GetCategoriesByAnimalSlug(slug)
	if err != nil {
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	if categories == nil {
		categories = []repository.Category{}
	}

	writeJSON(w, http.StatusOK, categories)
}

// writeJSON — вспомогательная функция для отправки JSON ответа
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
