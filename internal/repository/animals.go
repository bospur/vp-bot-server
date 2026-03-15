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

// AnimalInput — данные для создания/обновления животного
type AnimalInput struct {
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	Icon      string `json:"icon"`
	SortOrder int    `json:"sort_order"`
}

// CategoryInput — данные для создания/обновления категории
type CategoryInput struct {
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

// GetAllByClinic возвращает животных конкретной клиники
func (r *AnimalRepository) GetAllByClinic(clinicSlug string) ([]Animal, error) {
	rows, err := r.db.Query(`
		SELECT a.id, a.name, a.slug, COALESCE(a.icon, ''), a.sort_order
		FROM animals a
		JOIN clinics c ON c.id = a.clinic_id
		WHERE c.slug = $1
		ORDER BY a.sort_order, a.name
	`, clinicSlug)
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

// Create создаёт новое животное
func (r *AnimalRepository) Create(clinicID int, input AnimalInput) (*Animal, error) {
	var a Animal
	err := r.db.QueryRow(`
		INSERT INTO animals (clinic_id, name, slug, icon, sort_order)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, name, slug, COALESCE(icon, ''), sort_order
	`, clinicID, input.Name, input.Slug, input.Icon, input.SortOrder).
		Scan(&a.ID, &a.Name, &a.Slug, &a.Icon, &a.SortOrder)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

// Update обновляет животное по id
func (r *AnimalRepository) Update(id string, input AnimalInput) (*Animal, error) {
	var a Animal
	err := r.db.QueryRow(`
		UPDATE animals SET name=$1, slug=$2, icon=$3, sort_order=$4, updated_at=NOW()
		WHERE id=$5
		RETURNING id, name, slug, COALESCE(icon, ''), sort_order
	`, input.Name, input.Slug, input.Icon, input.SortOrder, id).
		Scan(&a.ID, &a.Name, &a.Slug, &a.Icon, &a.SortOrder)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

// Delete удаляет животное по id
func (r *AnimalRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM animals WHERE id=$1`, id)
	return err
}

// GetCategoriesByAnimalSlug возвращает категории животного в рамках клиники
func (r *AnimalRepository) GetCategoriesByAnimalSlug(clinicSlug, animalSlug string) ([]Category, error) {
	rows, err := r.db.Query(`
		SELECT c.id, c.animal_id, c.name, c.slug, COALESCE(c.icon, ''), c.sort_order
		FROM categories c
		JOIN animals a ON a.id = c.animal_id
		JOIN clinics cl ON cl.id = a.clinic_id
		WHERE cl.slug = $1 AND a.slug = $2
		ORDER BY c.sort_order, c.name
	`, clinicSlug, animalSlug)
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

// CreateCategory создаёт новую категорию
func (r *AnimalRepository) CreateCategory(clinicID int, input CategoryInput) (*Category, error) {
	var c Category
	err := r.db.QueryRow(`
		INSERT INTO categories (clinic_id, animal_id, name, slug, icon, sort_order)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, animal_id, name, slug, COALESCE(icon, ''), sort_order
	`, clinicID, input.AnimalID, input.Name, input.Slug, input.Icon, input.SortOrder).
		Scan(&c.ID, &c.AnimalID, &c.Name, &c.Slug, &c.Icon, &c.SortOrder)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// UpdateCategory обновляет категорию по id
func (r *AnimalRepository) UpdateCategory(id string, input CategoryInput) (*Category, error) {
	var c Category
	err := r.db.QueryRow(`
		UPDATE categories SET name=$1, slug=$2, icon=$3, sort_order=$4, updated_at=NOW()
		WHERE id=$5
		RETURNING id, animal_id, name, slug, COALESCE(icon, ''), sort_order
	`, input.Name, input.Slug, input.Icon, input.SortOrder, id).
		Scan(&c.ID, &c.AnimalID, &c.Name, &c.Slug, &c.Icon, &c.SortOrder)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// DeleteCategory удаляет категорию по id
func (r *AnimalRepository) DeleteCategory(id string) error {
	_, err := r.db.Exec(`DELETE FROM categories WHERE id=$1`, id)
	return err
}
