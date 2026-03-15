package handler

import (
	"net/http"
	"strings"

	"go-server/internal/repository"
)

// ArticleHandler содержит зависимости для HTTP хендлеров статей
type ArticleHandler struct {
	repo *repository.ArticleRepository
}

// NewArticleHandler создаёт новый хендлер
func NewArticleHandler(repo *repository.ArticleRepository) *ArticleHandler {
	return &ArticleHandler{repo: repo}
}

// GetArticles обрабатывает GET /api/animals/{animalSlug}/categories/{categorySlug}/articles
func (h *ArticleHandler) GetArticles(w http.ResponseWriter, r *http.Request) {
	// URL: /api/animals/cat/categories/poisoning/articles
	// parts: ["api", "animals", "cat", "categories", "poisoning", "articles"]
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 6 {
		http.Error(w, "неверный запрос", http.StatusBadRequest)
		return
	}
	animalSlug := parts[2]
	categorySlug := parts[4]

	articles, err := h.repo.GetByCategory(animalSlug, categorySlug)
	if err != nil {
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	if articles == nil {
		articles = []repository.Article{}
	}

	writeJSON(w, http.StatusOK, articles)
}

// GetArticle обрабатывает GET /api/articles/{slug}
func (h *ArticleHandler) GetArticle(w http.ResponseWriter, r *http.Request) {
	// URL: /api/articles/cat-poisoning-first-aid
	// parts: ["api", "articles", "cat-poisoning-first-aid"]
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 3 {
		http.Error(w, "неверный запрос", http.StatusBadRequest)
		return
	}
	slug := parts[2]

	article, err := h.repo.GetBySlug(slug)
	if err != nil {
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
	if article == nil {
		http.Error(w, "статья не найдена", http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, article)
}
