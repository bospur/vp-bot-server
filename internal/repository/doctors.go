package repository

import (
	"database/sql"
	"time"
)

// Doctor — карточка врача
type Doctor struct {
	ID          int    `json:"id"`
	ClinicID    int    `json:"clinic_id"`
	FullName    string `json:"full_name"`
	Specialty   string `json:"specialty"`
	Description string `json:"description"`
	Contacts    string `json:"contacts"`
	PhotoURL    string `json:"photo_url"`
	Status      string `json:"status"`
	SortOrder   int    `json:"sort_order"`
}

// DoctorInput — данные для создания/обновления карточки
type DoctorInput struct {
	FullName    string `json:"full_name"`
	Specialty   string `json:"specialty"`
	Description string `json:"description"`
	Contacts    string `json:"contacts"`
	SortOrder   int    `json:"sort_order"`
}

// DoctorSchedule — слот еженедельного расписания
type DoctorSchedule struct {
	ID        int    `json:"id"`
	DoctorID  int    `json:"doctor_id"`
	DayOfWeek int    `json:"day_of_week"`
	TimeFrom  string `json:"time_from"`
	TimeTo    string `json:"time_to"`
}

// DoctorScheduleException — исключение на конкретную дату
type DoctorScheduleException struct {
	ID       int     `json:"id"`
	DoctorID int     `json:"doctor_id"`
	Date     string  `json:"date"`
	IsDayOff bool    `json:"is_day_off"`
	TimeFrom *string `json:"time_from"`
	TimeTo   *string `json:"time_to"`
}

// ClinicSettings — настройки клиники
type ClinicSettings struct {
	ClinicID             int `json:"clinic_id"`
	ScheduleDisplayWeeks int `json:"schedule_display_weeks"`
}

// DoctorRepository — запросы к таблицам врачей и расписания
type DoctorRepository struct {
	db *sql.DB
}

func NewDoctorRepository(db *sql.DB) *DoctorRepository {
	return &DoctorRepository{db: db}
}

// ── Врачи ────────────────────────────────────────────────────────────────────

func (r *DoctorRepository) GetAll(clinicID int) ([]Doctor, error) {
	rows, err := r.db.Query(`
		SELECT id, clinic_id, full_name, specialty, description, contacts, photo_url, status, sort_order
		FROM doctors WHERE clinic_id = $1
		ORDER BY sort_order, full_name
	`, clinicID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var doctors []Doctor
	for rows.Next() {
		var d Doctor
		if err := rows.Scan(&d.ID, &d.ClinicID, &d.FullName, &d.Specialty, &d.Description, &d.Contacts, &d.PhotoURL, &d.Status, &d.SortOrder); err != nil {
			return nil, err
		}
		doctors = append(doctors, d)
	}
	return doctors, nil
}

func (r *DoctorRepository) GetPublished(clinicID int) ([]Doctor, error) {
	rows, err := r.db.Query(`
		SELECT id, clinic_id, full_name, specialty, description, contacts, photo_url, status, sort_order
		FROM doctors WHERE clinic_id = $1 AND status = 'published'
		ORDER BY sort_order, full_name
	`, clinicID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var doctors []Doctor
	for rows.Next() {
		var d Doctor
		if err := rows.Scan(&d.ID, &d.ClinicID, &d.FullName, &d.Specialty, &d.Description, &d.Contacts, &d.PhotoURL, &d.Status, &d.SortOrder); err != nil {
			return nil, err
		}
		doctors = append(doctors, d)
	}
	return doctors, nil
}

func (r *DoctorRepository) GetByID(id string) (*Doctor, error) {
	var d Doctor
	err := r.db.QueryRow(`
		SELECT id, clinic_id, full_name, specialty, description, contacts, photo_url, status, sort_order
		FROM doctors WHERE id = $1
	`, id).Scan(&d.ID, &d.ClinicID, &d.FullName, &d.Specialty, &d.Description, &d.Contacts, &d.PhotoURL, &d.Status, &d.SortOrder)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *DoctorRepository) Create(clinicID int, input DoctorInput) (*Doctor, error) {
	var d Doctor
	err := r.db.QueryRow(`
		INSERT INTO doctors (clinic_id, full_name, specialty, description, contacts, sort_order)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, clinic_id, full_name, specialty, description, contacts, photo_url, status, sort_order
	`, clinicID, input.FullName, input.Specialty, input.Description, input.Contacts, input.SortOrder).
		Scan(&d.ID, &d.ClinicID, &d.FullName, &d.Specialty, &d.Description, &d.Contacts, &d.PhotoURL, &d.Status, &d.SortOrder)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *DoctorRepository) Update(id string, input DoctorInput) (*Doctor, error) {
	var d Doctor
	err := r.db.QueryRow(`
		UPDATE doctors SET full_name=$1, specialty=$2, description=$3, contacts=$4, sort_order=$5, updated_at=NOW()
		WHERE id=$6
		RETURNING id, clinic_id, full_name, specialty, description, contacts, photo_url, status, sort_order
	`, input.FullName, input.Specialty, input.Description, input.Contacts, input.SortOrder, id).
		Scan(&d.ID, &d.ClinicID, &d.FullName, &d.Specialty, &d.Description, &d.Contacts, &d.PhotoURL, &d.Status, &d.SortOrder)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *DoctorRepository) UpdateStatus(id, status string) (*Doctor, error) {
	var d Doctor
	err := r.db.QueryRow(`
		UPDATE doctors SET status=$1, updated_at=NOW()
		WHERE id=$2
		RETURNING id, clinic_id, full_name, specialty, description, contacts, photo_url, status, sort_order
	`, status, id).
		Scan(&d.ID, &d.ClinicID, &d.FullName, &d.Specialty, &d.Description, &d.Contacts, &d.PhotoURL, &d.Status, &d.SortOrder)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *DoctorRepository) UpdatePhoto(id, photoURL string) error {
	_, err := r.db.Exec(`UPDATE doctors SET photo_url=$1, updated_at=NOW() WHERE id=$2`, photoURL, id)
	return err
}

func (r *DoctorRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM doctors WHERE id=$1`, id)
	return err
}

// ── Расписание ───────────────────────────────────────────────────────────────

func (r *DoctorRepository) GetSchedule(doctorID string) ([]DoctorSchedule, error) {
	rows, err := r.db.Query(`
		SELECT id, doctor_id, day_of_week, time_from::text, time_to::text
		FROM doctor_schedules WHERE doctor_id = $1
		ORDER BY day_of_week, time_from
	`, doctorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var slots []DoctorSchedule
	for rows.Next() {
		var s DoctorSchedule
		if err := rows.Scan(&s.ID, &s.DoctorID, &s.DayOfWeek, &s.TimeFrom, &s.TimeTo); err != nil {
			return nil, err
		}
		slots = append(slots, s)
	}
	return slots, nil
}

func (r *DoctorRepository) AddScheduleSlot(doctorID string, dayOfWeek int, timeFrom, timeTo string) (*DoctorSchedule, error) {
	var s DoctorSchedule
	err := r.db.QueryRow(`
		INSERT INTO doctor_schedules (doctor_id, day_of_week, time_from, time_to)
		VALUES ($1, $2, $3, $4)
		RETURNING id, doctor_id, day_of_week, time_from::text, time_to::text
	`, doctorID, dayOfWeek, timeFrom, timeTo).
		Scan(&s.ID, &s.DoctorID, &s.DayOfWeek, &s.TimeFrom, &s.TimeTo)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *DoctorRepository) DeleteScheduleSlot(slotID string) error {
	_, err := r.db.Exec(`DELETE FROM doctor_schedules WHERE id=$1`, slotID)
	return err
}

// ── Исключения ───────────────────────────────────────────────────────────────

func (r *DoctorRepository) GetExceptions(doctorID string) ([]DoctorScheduleException, error) {
	rows, err := r.db.Query(`
		SELECT id, doctor_id, date::text, is_day_off, time_from::text, time_to::text
		FROM doctor_schedule_exceptions WHERE doctor_id = $1
		ORDER BY date
	`, doctorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var exceptions []DoctorScheduleException
	for rows.Next() {
		var e DoctorScheduleException
		if err := rows.Scan(&e.ID, &e.DoctorID, &e.Date, &e.IsDayOff, &e.TimeFrom, &e.TimeTo); err != nil {
			return nil, err
		}
		exceptions = append(exceptions, e)
	}
	return exceptions, nil
}

func (r *DoctorRepository) UpsertException(doctorID, date string, isDayOff bool, timeFrom, timeTo *string) (*DoctorScheduleException, error) {
	var e DoctorScheduleException
	err := r.db.QueryRow(`
		INSERT INTO doctor_schedule_exceptions (doctor_id, date, is_day_off, time_from, time_to)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (doctor_id, date) DO UPDATE
			SET is_day_off=$3, time_from=$4, time_to=$5
		RETURNING id, doctor_id, date::text, is_day_off, time_from::text, time_to::text
	`, doctorID, date, isDayOff, timeFrom, timeTo).
		Scan(&e.ID, &e.DoctorID, &e.Date, &e.IsDayOff, &e.TimeFrom, &e.TimeTo)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *DoctorRepository) DeleteException(exceptionID string) error {
	_, err := r.db.Exec(`DELETE FROM doctor_schedule_exceptions WHERE id=$1`, exceptionID)
	return err
}

// ── Расписание на период (для бота) ──────────────────────────────────────────

// ScheduleEntry — один день в развёрнутом расписании
type ScheduleEntry struct {
	DoctorID  int    `json:"doctor_id"`
	FullName  string `json:"full_name"`
	Specialty string `json:"specialty"`
	PhotoURL  string `json:"photo_url"`
	Date      string `json:"date"`
	TimeFrom  string `json:"time_from"`
	TimeTo    string `json:"time_to"`
}

// GetScheduleForPeriod возвращает расписание всех опубликованных врачей на указанный период
func (r *DoctorRepository) GetScheduleForPeriod(clinicID int, from, to time.Time) ([]ScheduleEntry, error) {
	rows, err := r.db.Query(`
		WITH RECURSIVE dates AS (
			SELECT $1::date AS d
			UNION ALL
			SELECT d + 1 FROM dates WHERE d < $2::date
		),
		base AS (
			SELECT
				d.d AS date,
				ds.doctor_id,
				ds.time_from,
				ds.time_to
			FROM dates d
			JOIN doctor_schedules ds ON ds.day_of_week = EXTRACT(DOW FROM d.d)
			JOIN doctors doc ON doc.id = ds.doctor_id
			WHERE doc.clinic_id = $3 AND doc.status = 'published'
		),
		with_exceptions AS (
			SELECT
				b.date::text,
				b.doctor_id,
				COALESCE(e.time_from, b.time_from)::text AS time_from,
				COALESCE(e.time_to,   b.time_to)::text   AS time_to,
				COALESCE(e.is_day_off, false)             AS is_day_off
			FROM base b
			LEFT JOIN doctor_schedule_exceptions e
				ON e.doctor_id = b.doctor_id AND e.date = b.date
			UNION ALL
			-- Дополнительные смены из исключений (когда нет базового слота)
			SELECT
				e.date::text,
				e.doctor_id,
				e.time_from::text,
				e.time_to::text,
				false AS is_day_off
			FROM doctor_schedule_exceptions e
			JOIN doctors doc ON doc.id = e.doctor_id
			WHERE doc.clinic_id = $3 AND doc.status = 'published'
				AND e.is_day_off = false
				AND e.time_from IS NOT NULL
				AND NOT EXISTS (
					SELECT 1 FROM doctor_schedules ds
					WHERE ds.doctor_id = e.doctor_id
					  AND ds.day_of_week = EXTRACT(DOW FROM e.date)
				)
				AND e.date BETWEEN $1 AND $2
		)
		SELECT we.date, doc.id, doc.full_name, doc.specialty, doc.photo_url,
			   we.time_from, we.time_to
		FROM with_exceptions we
		JOIN doctors doc ON doc.id = we.doctor_id
		WHERE we.is_day_off = false
		ORDER BY we.date, doc.sort_order, doc.full_name
	`, from.Format("2006-01-02"), to.Format("2006-01-02"), clinicID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []ScheduleEntry
	for rows.Next() {
		var e ScheduleEntry
		if err := rows.Scan(&e.Date, &e.DoctorID, &e.FullName, &e.Specialty, &e.PhotoURL, &e.TimeFrom, &e.TimeTo); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, nil
}

// GetPublishedByClinicSlug возвращает опубликованных врачей по slug клиники
func (r *DoctorRepository) GetPublishedByClinicSlug(clinicSlug string) ([]Doctor, error) {
	rows, err := r.db.Query(`
		SELECT d.id, d.clinic_id, d.full_name, d.specialty, d.description, d.contacts, d.photo_url, d.status, d.sort_order
		FROM doctors d
		JOIN clinics c ON c.id = d.clinic_id
		WHERE c.slug = $1 AND d.status = 'published'
		ORDER BY d.sort_order, d.full_name
	`, clinicSlug)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var doctors []Doctor
	for rows.Next() {
		var d Doctor
		if err := rows.Scan(&d.ID, &d.ClinicID, &d.FullName, &d.Specialty, &d.Description, &d.Contacts, &d.PhotoURL, &d.Status, &d.SortOrder); err != nil {
			return nil, err
		}
		doctors = append(doctors, d)
	}
	return doctors, nil
}

// GetScheduleByClinicSlug возвращает расписание и настройки по slug клиники
func (r *DoctorRepository) GetScheduleByClinicSlug(clinicSlug string) ([]ScheduleEntry, *ClinicSettings, error) {
	var clinicID int
	err := r.db.QueryRow(`SELECT id FROM clinics WHERE slug = $1`, clinicSlug).Scan(&clinicID)
	if err != nil {
		return nil, nil, err
	}

	settings, err := r.GetSettings(clinicID)
	if err != nil {
		return nil, nil, err
	}

	weeks := settings.ScheduleDisplayWeeks
	days := weeks * 7
	if weeks == 5 { // "месяц" — 31 день
		days = 31
	}

	from := time.Now().Truncate(24 * time.Hour)
	to := from.Add(time.Duration(days-1) * 24 * time.Hour)

	entries, err := r.GetScheduleForPeriod(clinicID, from, to)
	return entries, settings, err
}

// ── Настройки клиники ─────────────────────────────────────────────────────────

func (r *DoctorRepository) GetSettings(clinicID int) (*ClinicSettings, error) {
	var s ClinicSettings
	err := r.db.QueryRow(`
		SELECT clinic_id, schedule_display_weeks FROM clinic_settings WHERE clinic_id = $1
	`, clinicID).Scan(&s.ClinicID, &s.ScheduleDisplayWeeks)
	if err == sql.ErrNoRows {
		return &ClinicSettings{ClinicID: clinicID, ScheduleDisplayWeeks: 2}, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *DoctorRepository) UpsertSettings(clinicID, weeks int) (*ClinicSettings, error) {
	var s ClinicSettings
	err := r.db.QueryRow(`
		INSERT INTO clinic_settings (clinic_id, schedule_display_weeks)
		VALUES ($1, $2)
		ON CONFLICT (clinic_id) DO UPDATE SET schedule_display_weeks = $2
		RETURNING clinic_id, schedule_display_weeks
	`, clinicID, weeks).Scan(&s.ClinicID, &s.ScheduleDisplayWeeks)
	if err != nil {
		return nil, err
	}
	return &s, nil
}
