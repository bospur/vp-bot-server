package repository

import (
	"database/sql"
)

// User — структура пользователя
type User struct {
	ID           int    `json:"id"`
	ClinicID     int    `json:"clinic_id"`
	Login        string `json:"login"`
	PasswordHash string `json:"-"` // никогда не отдаём в JSON
	Role         string `json:"role"`
}

// UserRepository — запросы к таблице users
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository создаёт новый репозиторий
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// GetByLogin возвращает пользователя по логину
func (r *UserRepository) GetByLogin(login string) (*User, error) {
	var u User
	err := r.db.QueryRow(`
		SELECT id, clinic_id, login, password_hash, role
		FROM users
		WHERE login = $1
	`, login).Scan(&u.ID, &u.ClinicID, &u.Login, &u.PasswordHash, &u.Role)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// GetAll возвращает всех пользователей клиники
func (r *UserRepository) GetAll(clinicID int) ([]User, error) {
	rows, err := r.db.Query(`
		SELECT id, clinic_id, login, role FROM users
		WHERE clinic_id = $1
		ORDER BY login
	`, clinicID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.ClinicID, &u.Login, &u.Role); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

// Delete удаляет пользователя по id
func (r *UserRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM users WHERE id=$1`, id)
	return err
}

// Count возвращает количество пользователей
func (r *UserRepository) Count() (int, error) {
	var count int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&count)
	return count, err
}

// Create создаёт нового пользователя
func (r *UserRepository) Create(clinicID int, login, passwordHash, role string) (*User, error) {
	var u User
	err := r.db.QueryRow(`
		INSERT INTO users (clinic_id, login, password_hash, role)
		VALUES ($1, $2, $3, $4)
		RETURNING id, clinic_id, login, password_hash, role
	`, clinicID, login, passwordHash, role).
		Scan(&u.ID, &u.ClinicID, &u.Login, &u.PasswordHash, &u.Role)
	if err != nil {
		return nil, err
	}
	return &u, nil
}
