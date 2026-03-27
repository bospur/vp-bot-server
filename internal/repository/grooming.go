package repository

import (
	"database/sql"
	"errors"
)

// ErrGroomingConflict возвращается при пересечении записей по времени
var ErrGroomingConflict = errors.New("время занято")

// ErrGroomingNotWorkingDay возвращается если дата не рабочий день по шаблону
var ErrGroomingNotWorkingDay = errors.New("не рабочий день")

// ErrGroomingOutOfHours возвращается если запись выходит за рамки рабочих часов
var ErrGroomingOutOfHours = errors.New("запись за пределами рабочего времени")

// ── Модели ────────────────────────────────────────────────────────────────────

// GroomingBreed — порода/услуга из коллекции грумера
type GroomingBreed struct {
	ID          int      `json:"id"`
	ClinicID    int      `json:"clinic_id"`
	Breed       string   `json:"breed"`
	Duration    int      `json:"duration"` // минуты
	Price       *float64 `json:"price"`    // nil если не задана
	Description *string  `json:"description"`
}

// GroomingBreedInput — данные для создания/обновления породы
type GroomingBreedInput struct {
	Breed       string   `json:"breed"`
	Duration    int      `json:"duration"`
	Price       *float64 `json:"price"`
	Description *string  `json:"description"`
}

// GroomingTemplateSlot — один рабочий день в шаблоне недели
type GroomingTemplateSlot struct {
	ID        int    `json:"id"`
	ClinicID  int    `json:"clinic_id"`
	DayOfWeek int    `json:"day_of_week"`
	TimeFrom  string `json:"time_from"`
	TimeTo    string `json:"time_to"`
}

// GroomingTemplateInput — данные для upsert слота шаблона
type GroomingTemplateInput struct {
	DayOfWeek int    `json:"day_of_week"`
	TimeFrom  string `json:"time_from"`
	TimeTo    string `json:"time_to"`
}

// GroomingAppointment — запись животного в расписание
type GroomingAppointment struct {
	ID         int      `json:"id"`
	ClinicID   int      `json:"clinic_id"`
	BreedID    int      `json:"breed_id"`
	Breed      string   `json:"breed"`    // из JOIN с grooming_breeds
	Duration   int      `json:"duration"` // из JOIN
	Price      *float64 `json:"price"`    // из JOIN
	Date       string   `json:"date"`
	PetName    string   `json:"pet_name"`
	OwnerPhone string   `json:"owner_phone"`
	StartTime  string   `json:"start_time"`
	EndTime    string   `json:"end_time"`
}

// GroomingAppointmentInput — данные для создания записи
type GroomingAppointmentInput struct {
	BreedID    int    `json:"breed_id"`
	Date       string `json:"date"`
	PetName    string `json:"pet_name"`
	OwnerPhone string `json:"owner_phone"`
	StartTime  string `json:"start_time"`
}

// GroomingRepository — запросы к таблицам груминга
type GroomingRepository struct {
	db *sql.DB
}

func NewGroomingRepository(db *sql.DB) *GroomingRepository {
	return &GroomingRepository{db: db}
}

// ── Породы ────────────────────────────────────────────────────────────────────

func (r *GroomingRepository) GetAllBreeds(clinicID int) ([]GroomingBreed, error) {
	rows, err := r.db.Query(`
		SELECT id, clinic_id, breed, duration, price, description
		FROM grooming_breeds WHERE clinic_id = $1
		ORDER BY breed
	`, clinicID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var breeds []GroomingBreed
	for rows.Next() {
		var b GroomingBreed
		if err := rows.Scan(&b.ID, &b.ClinicID, &b.Breed, &b.Duration, &b.Price, &b.Description); err != nil {
			return nil, err
		}
		breeds = append(breeds, b)
	}
	return breeds, nil
}

func (r *GroomingRepository) GetBreedByID(id string) (*GroomingBreed, error) {
	var b GroomingBreed
	err := r.db.QueryRow(`
		SELECT id, clinic_id, breed, duration, price, description
		FROM grooming_breeds WHERE id = $1
	`, id).Scan(&b.ID, &b.ClinicID, &b.Breed, &b.Duration, &b.Price, &b.Description)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *GroomingRepository) CreateBreed(clinicID int, input GroomingBreedInput) (*GroomingBreed, error) {
	var b GroomingBreed
	err := r.db.QueryRow(`
		INSERT INTO grooming_breeds (clinic_id, breed, duration, price, description)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, clinic_id, breed, duration, price, description
	`, clinicID, input.Breed, input.Duration, input.Price, input.Description).
		Scan(&b.ID, &b.ClinicID, &b.Breed, &b.Duration, &b.Price, &b.Description)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *GroomingRepository) UpdateBreed(id string, input GroomingBreedInput) (*GroomingBreed, error) {
	var b GroomingBreed
	err := r.db.QueryRow(`
		UPDATE grooming_breeds
		SET breed=$1, duration=$2, price=$3, description=$4, updated_at=NOW()
		WHERE id=$5
		RETURNING id, clinic_id, breed, duration, price, description
	`, input.Breed, input.Duration, input.Price, input.Description, id).
		Scan(&b.ID, &b.ClinicID, &b.Breed, &b.Duration, &b.Price, &b.Description)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *GroomingRepository) DeleteBreed(id string) error {
	_, err := r.db.Exec(`DELETE FROM grooming_breeds WHERE id=$1`, id)
	return err
}

// ── Шаблон недели ─────────────────────────────────────────────────────────────

func (r *GroomingRepository) GetTemplate(clinicID int) ([]GroomingTemplateSlot, error) {
	rows, err := r.db.Query(`
		SELECT id, clinic_id, day_of_week, time_from::text, time_to::text
		FROM grooming_weekly_template WHERE clinic_id = $1
		ORDER BY day_of_week
	`, clinicID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var slots []GroomingTemplateSlot
	for rows.Next() {
		var s GroomingTemplateSlot
		if err := rows.Scan(&s.ID, &s.ClinicID, &s.DayOfWeek, &s.TimeFrom, &s.TimeTo); err != nil {
			return nil, err
		}
		slots = append(slots, s)
	}
	return slots, nil
}

// UpsertTemplateSlot создаёт или обновляет слот шаблона для дня недели
func (r *GroomingRepository) UpsertTemplateSlot(clinicID int, input GroomingTemplateInput) (*GroomingTemplateSlot, error) {
	var s GroomingTemplateSlot
	err := r.db.QueryRow(`
		INSERT INTO grooming_weekly_template (clinic_id, day_of_week, time_from, time_to)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (clinic_id, day_of_week) DO UPDATE
			SET time_from=$3, time_to=$4
		RETURNING id, clinic_id, day_of_week, time_from::text, time_to::text
	`, clinicID, input.DayOfWeek, input.TimeFrom, input.TimeTo).
		Scan(&s.ID, &s.ClinicID, &s.DayOfWeek, &s.TimeFrom, &s.TimeTo)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *GroomingRepository) DeleteTemplateSlot(clinicID, dayOfWeek int) error {
	_, err := r.db.Exec(
		`DELETE FROM grooming_weekly_template WHERE clinic_id=$1 AND day_of_week=$2`,
		clinicID, dayOfWeek,
	)
	return err
}

// ── Записи ────────────────────────────────────────────────────────────────────

// GetAppointmentsByMonth возвращает все записи за указанный месяц (формат "2026-04")
func (r *GroomingRepository) GetAppointmentsByMonth(clinicID int, month string) ([]GroomingAppointment, error) {
	rows, err := r.db.Query(`
		SELECT
			a.id, a.clinic_id, a.breed_id,
			b.breed, b.duration, b.price,
			a.date::text, a.pet_name, a.owner_phone,
			a.start_time::text, a.end_time::text
		FROM grooming_appointments a
		JOIN grooming_breeds b ON b.id = a.breed_id
		WHERE a.clinic_id = $1
		  AND TO_CHAR(a.date, 'YYYY-MM') = $2
		ORDER BY a.date, a.start_time
	`, clinicID, month)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var appointments []GroomingAppointment
	for rows.Next() {
		var a GroomingAppointment
		if err := rows.Scan(
			&a.ID, &a.ClinicID, &a.BreedID,
			&a.Breed, &a.Duration, &a.Price,
			&a.Date, &a.PetName, &a.OwnerPhone,
			&a.StartTime, &a.EndTime,
		); err != nil {
			return nil, err
		}
		appointments = append(appointments, a)
	}
	return appointments, nil
}

// CreateAppointment создаёт запись, проверяя рабочий день и пересечения
func (r *GroomingRepository) CreateAppointment(clinicID int, input GroomingAppointmentInput) (*GroomingAppointment, error) {
	// Получаем продолжительность породы
	var duration int
	err := r.db.QueryRow(
		`SELECT duration FROM grooming_breeds WHERE id=$1 AND clinic_id=$2`,
		input.BreedID, clinicID,
	).Scan(&duration)
	if err == sql.ErrNoRows {
		return nil, errors.New("порода не найдена")
	}
	if err != nil {
		return nil, err
	}

	// Проверяем что дата — рабочий день по шаблону
	// EXTRACT(DOW FROM date) → 0=Вс, 1=Пн … 6=Сб
	var timeFrom, timeTo string
	err = r.db.QueryRow(`
		SELECT time_from::text, time_to::text
		FROM grooming_weekly_template
		WHERE clinic_id=$1 AND day_of_week = EXTRACT(DOW FROM $2::date)
	`, clinicID, input.Date).Scan(&timeFrom, &timeTo)
	if err == sql.ErrNoRows {
		return nil, ErrGroomingNotWorkingDay
	}
	if err != nil {
		return nil, err
	}

	// Вычисляем end_time = start_time + duration
	// Используем PostgreSQL для простоты
	var endTime string
	err = r.db.QueryRow(
		`SELECT ($1::time + ($2 * INTERVAL '1 minute'))::time::text`,
		input.StartTime, duration,
	).Scan(&endTime)
	if err != nil {
		return nil, err
	}

	// Проверяем что запись в рамках рабочего времени
	var inHours bool
	err = r.db.QueryRow(`
		SELECT $1::time >= $2::time AND $3::time <= $4::time
	`, input.StartTime, timeFrom, endTime, timeTo).Scan(&inHours)
	if err != nil {
		return nil, err
	}
	if !inHours {
		return nil, ErrGroomingOutOfHours
	}

	// Проверяем пересечение с существующими записями
	var hasConflict bool
	err = r.db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM grooming_appointments
			WHERE clinic_id=$1 AND date=$2::date
			  AND start_time < $3::time
			  AND end_time   > $4::time
		)
	`, clinicID, input.Date, endTime, input.StartTime).Scan(&hasConflict)
	if err != nil {
		return nil, err
	}
	if hasConflict {
		return nil, ErrGroomingConflict
	}

	// Создаём запись
	var a GroomingAppointment
	err = r.db.QueryRow(`
		INSERT INTO grooming_appointments
			(clinic_id, breed_id, date, pet_name, owner_phone, start_time, end_time)
		VALUES ($1, $2, $3::date, $4, $5, $6::time, $7::time)
		RETURNING id, clinic_id, breed_id, date::text, pet_name, owner_phone,
		          start_time::text, end_time::text
	`, clinicID, input.BreedID, input.Date, input.PetName, input.OwnerPhone,
		input.StartTime, endTime).
		Scan(&a.ID, &a.ClinicID, &a.BreedID, &a.Date, &a.PetName, &a.OwnerPhone,
			&a.StartTime, &a.EndTime)
	if err != nil {
		return nil, err
	}

	// Дополняем breed-поля из уже известных данных
	err = r.db.QueryRow(
		`SELECT breed, duration, price FROM grooming_breeds WHERE id=$1`, input.BreedID,
	).Scan(&a.Breed, &a.Duration, &a.Price)
	if err != nil {
		return nil, err
	}

	return &a, nil
}

func (r *GroomingRepository) DeleteAppointment(id string, clinicID int) error {
	_, err := r.db.Exec(
		`DELETE FROM grooming_appointments WHERE id=$1 AND clinic_id=$2`,
		id, clinicID,
	)
	return err
}
