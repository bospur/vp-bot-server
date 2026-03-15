package handler

import (
	"net/http"

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
	animalSlug := r.PathValue("animalSlug")
	categorySlug := r.PathValue("categorySlug")
	if animalSlug == "" || categorySlug == "" {
		http.Error(w, "неверный запрос", http.StatusBadRequest)
		return
	}

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
	slug := r.PathValue("slug")
	if slug == "" {
		http.Error(w, "неверный запрос", http.StatusBadRequest)
		return
	}

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
