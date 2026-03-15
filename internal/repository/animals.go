package repository

import (
	"database/sql"
)

// Animal — структура данных животного
type Animal struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	Icon      string `json:"icon"`
	SortOrder int    `json:"sort_order"`
}

// Category — структура категории
type Category struct {
	ID        int    `json:"id"`
	AnimalID  int    `json:"animal_id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	Icon      string `json:"icon"`
	SortOrder int    `json:"sort_order"`
}

// AnimalRepository — отвечает за все запросы к таблицам animals и categories
type AnimalRepository struct {
	db *sql.DB
}

// NewAnimalRepository создаёт новый репозиторий
func NewAnimalRepository(db *sql.DB) *AnimalRepository {
	return &AnimalRepository{db: db}
}

// GetAll возвращает список всех животных отсортированных по sort_order
func (r *AnimalRepository) GetAll() ([]Animal, error) {
	rows, err := r.db.Query(`
		SELECT id, name, slug, COALESCE(icon, ''), sort_order
		FROM animals
		ORDER BY sort_order, name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var animals []Animal
	for rows.Next() {
		var a Animal
		if err := rows.Scan(&a.ID, &a.Name, &a.Slug, &a.Icon, &a.SortOrder); err != nil {
			return nil, err
		}
		animals = append(animals, a)
	}

	return animals, nil
}

// GetCategoriesByAnimalSlug возвращает категории для конкретного животного
func (r *AnimalRepository) GetCategoriesByAnimalSlug(slug string) ([]Category, error) {
	rows, err := r.db.Query(`
		SELECT c.id, c.animal_id, c.name, c.slug, COALESCE(c.icon, ''), c.sort_order
		FROM categories c
		JOIN animals a ON a.id = c.animal_id
		WHERE a.slug = $1
		ORDER BY c.sort_order, c.name
	`, slug)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.ID, &c.AnimalID, &c.Name, &c.Slug, &c.Icon, &c.SortOrder); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}

	return categories, nil
}
