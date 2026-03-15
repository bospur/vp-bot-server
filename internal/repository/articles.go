package repository

import (
	"database/sql"
)

// Article — структура статьи
type Article struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
	Slug    string `json:"slug"`
}

// ArticleRepository — отвечает за запросы к таблице articles
type ArticleRepository struct {
	db *sql.DB
}

// NewArticleRepository создаёт новый репозиторий
func NewArticleRepository(db *sql.DB) *ArticleRepository {
	return &ArticleRepository{db: db}
}

// GetByCategory возвращает список статей для конкретной категории животного
func (r *ArticleRepository) GetByCategory(animalSlug, categorySlug string) ([]Article, error) {
	rows, err := r.db.Query(`
		SELECT a.id, a.title, a.content, a.slug
		FROM articles a
		JOIN article_categories ac ON ac.article_id = a.id
		JOIN categories c ON c.id = ac.category_id
		JOIN animals an ON an.id = c.animal_id
		WHERE an.slug = $1 AND c.slug = $2
		ORDER BY a.title
	`, animalSlug, categorySlug)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var articles []Article
	for rows.Next() {
		var a Article
		if err := rows.Scan(&a.ID, &a.Title, &a.Content, &a.Slug); err != nil {
			return nil, err
		}
		articles = append(articles, a)
	}

	return articles, nil
}

// GetBySlug возвращает одну статью по slug
func (r *ArticleRepository) GetBySlug(slug string) (*Article, error) {
	var a Article
	err := r.db.QueryRow(`
		SELECT id, title, content, slug
		FROM articles
		WHERE slug = $1
	`, slug).Scan(&a.ID, &a.Title, &a.Content, &a.Slug)

	if err == sql.ErrNoRows {
		return nil, nil // статья не найдена
	}
	if err != nil {
		return nil, err
	}

	return &a, nil
}
