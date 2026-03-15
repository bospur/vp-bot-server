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

// ArticleInput — данные для создания/обновления статьи
type ArticleInput struct {
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

// Create создаёт новую статью
func (r *ArticleRepository) Create(clinicID int, input ArticleInput) (*Article, error) {
	var a Article
	err := r.db.QueryRow(`
		INSERT INTO articles (clinic_id, title, content, slug)
		VALUES ($1, $2, $3, $4)
		RETURNING id, title, content, slug
	`, clinicID, input.Title, input.Content, input.Slug).
		Scan(&a.ID, &a.Title, &a.Content, &a.Slug)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

// Update обновляет статью по id
func (r *ArticleRepository) Update(id string, input ArticleInput) (*Article, error) {
	var a Article
	err := r.db.QueryRow(`
		UPDATE articles SET title=$1, content=$2, slug=$3, updated_at=NOW()
		WHERE id=$4
		RETURNING id, title, content, slug
	`, input.Title, input.Content, input.Slug, id).
		Scan(&a.ID, &a.Title, &a.Content, &a.Slug)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

// Delete удаляет статью по id
func (r *ArticleRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM articles WHERE id=$1`, id)
	return err
}

// AssignToCategory привязывает статью к категории
func (r *ArticleRepository) AssignToCategory(articleID, categoryID string) error {
	_, err := r.db.Exec(`
		INSERT INTO article_categories (article_id, category_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, articleID, categoryID)
	return err
}

// RemoveFromCategory отвязывает статью от категории
func (r *ArticleRepository) RemoveFromCategory(articleID, categoryID string) error {
	_, err := r.db.Exec(`
		DELETE FROM article_categories WHERE article_id=$1 AND category_id=$2
	`, articleID, categoryID)
	return err
}

// GetByCategory возвращает список статей для конкретной категории животного в рамках клиники
func (r *ArticleRepository) GetByCategory(clinicSlug, animalSlug, categorySlug string) ([]Article, error) {
	rows, err := r.db.Query(`
		SELECT a.id, a.title, a.content, a.slug
		FROM articles a
		JOIN article_categories ac ON ac.article_id = a.id
		JOIN categories c ON c.id = ac.category_id
		JOIN animals an ON an.id = c.animal_id
		JOIN clinics cl ON cl.id = an.clinic_id
		WHERE cl.slug = $1 AND an.slug = $2 AND c.slug = $3
		ORDER BY a.title
	`, clinicSlug, animalSlug, categorySlug)
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

// GetBySlug возвращает одну статью по slug в рамках клиники
func (r *ArticleRepository) GetBySlug(clinicSlug, slug string) (*Article, error) {
	var a Article
	err := r.db.QueryRow(`
		SELECT a.id, a.title, a.content, a.slug
		FROM articles a
		JOIN clinics c ON c.id = a.clinic_id
		WHERE c.slug = $1 AND a.slug = $2
	`, clinicSlug, slug).Scan(&a.ID, &a.Title, &a.Content, &a.Slug)

	if err == sql.ErrNoRows {
		return nil, nil // статья не найдена
	}
	if err != nil {
		return nil, err
	}

	return &a, nil
}
