package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"go-server/internal/middleware"
	"go-server/internal/repository"
)

// GroomingHandler содержит зависимости для эндпоинтов грумера
type GroomingHandler struct {
	groomingRepo *repository.GroomingRepository
}

func NewGroomingHandler(groomingRepo *repository.GroomingRepository) *GroomingHandler {
	return &GroomingHandler{groomingRepo: groomingRepo}
}

// ── Породы ────────────────────────────────────────────────────────────────────

// GetBreeds обрабатывает GET /api/admin/grooming/breeds
func (h *GroomingHandler) GetBreeds(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromContext(r)

	breeds, err := h.groomingRepo.GetAllBreeds(claims.ClinicID)
	if err != nil {
		log.Printf("ошибка получения пород груминга: %v", err)
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
	if breeds == nil {
		breeds = []repository.GroomingBreed{}
	}
	writeJSON(w, http.StatusOK, breeds)
}

// CreateBreed обрабатывает POST /api/admin/grooming/breeds
func (h *GroomingHandler) CreateBreed(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromContext(r)

	var input repository.GroomingBreedInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "неверный формат запроса", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(input.Breed) == "" {
		http.Error(w, "порода обязательна", http.StatusBadRequest)
		return
	}
	if input.Duration <= 0 {
		http.Error(w, "продолжительность должна быть больше 0", http.StatusBadRequest)
		return
	}

	breed, err := h.groomingRepo.CreateBreed(claims.ClinicID, input)
	if err != nil {
		log.Printf("ошибка создания породы: %v", err)
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, breed)
}

// UpdateBreed обрабатывает PUT /api/admin/grooming/breeds/{id}
func (h *GroomingHandler) UpdateBreed(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var input repository.GroomingBreedInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "неверный формат запроса", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(input.Breed) == "" {
		http.Error(w, "порода обязательна", http.StatusBadRequest)
		return
	}
	if input.Duration <= 0 {
		http.Error(w, "продолжительность должна быть больше 0", http.StatusBadRequest)
		return
	}

	breed, err := h.groomingRepo.UpdateBreed(id, input)
	if err != nil {
		log.Printf("ошибка обновления породы: %v", err)
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
	if breed == nil {
		http.Error(w, "не найдено", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, breed)
}

// DeleteBreed обрабатывает DELETE /api/admin/grooming/breeds/{id}
func (h *GroomingHandler) DeleteBreed(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if err := h.groomingRepo.DeleteBreed(id); err != nil {
		log.Printf("ошибка удаления породы: %v", err)
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ── Шаблон недели ─────────────────────────────────────────────────────────────

// GetTemplate обрабатывает GET /api/admin/grooming/template
func (h *GroomingHandler) GetTemplate(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromContext(r)

	slots, err := h.groomingRepo.GetTemplate(claims.ClinicID)
	if err != nil {
		log.Printf("ошибка получения шаблона недели: %v", err)
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
	if slots == nil {
		slots = []repository.GroomingTemplateSlot{}
	}
	writeJSON(w, http.StatusOK, slots)
}

// UpsertTemplateSlot обрабатывает PUT /api/admin/grooming/template
func (h *GroomingHandler) UpsertTemplateSlot(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromContext(r)

	var input repository.GroomingTemplateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "неверный формат запроса", http.StatusBadRequest)
		return
	}
	if input.DayOfWeek < 0 || input.DayOfWeek > 6 {
		http.Error(w, "day_of_week должен быть от 0 до 6", http.StatusBadRequest)
		return
	}
	if input.TimeFrom == "" || input.TimeTo == "" {
		http.Error(w, "time_from и time_to обязательны", http.StatusBadRequest)
		return
	}

	slot, err := h.groomingRepo.UpsertTemplateSlot(claims.ClinicID, input)
	if err != nil {
		log.Printf("ошибка сохранения слота шаблона: %v", err)
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, slot)
}

// DeleteTemplateSlot обрабатывает DELETE /api/admin/grooming/template/{dayOfWeek}
func (h *GroomingHandler) DeleteTemplateSlot(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromContext(r)

	var body struct {
		DayOfWeek int `json:"day_of_week"`
	}
	// day_of_week берём из path
	dayStr := r.PathValue("dayOfWeek")
	var day int
	if _, err := parseID(dayStr, &day); err != nil {
		http.Error(w, "неверный day_of_week", http.StatusBadRequest)
		return
	}
	_ = body

	if err := h.groomingRepo.DeleteTemplateSlot(claims.ClinicID, day); err != nil {
		log.Printf("ошибка удаления слота шаблона: %v", err)
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ── Записи ────────────────────────────────────────────────────────────────────

// GetAppointments обрабатывает GET /api/admin/grooming/appointments?month=2026-04
func (h *GroomingHandler) GetAppointments(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromContext(r)

	month := r.URL.Query().Get("month")
	if month == "" {
		http.Error(w, "параметр month обязателен (формат: 2026-04)", http.StatusBadRequest)
		return
	}

	appointments, err := h.groomingRepo.GetAppointmentsByMonth(claims.ClinicID, month)
	if err != nil {
		log.Printf("ошибка получения записей: %v", err)
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
	if appointments == nil {
		appointments = []repository.GroomingAppointment{}
	}
	writeJSON(w, http.StatusOK, appointments)
}

// CreateAppointment обрабатывает POST /api/admin/grooming/appointments
func (h *GroomingHandler) CreateAppointment(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromContext(r)

	var input repository.GroomingAppointmentInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "неверный формат запроса", http.StatusBadRequest)
		return
	}
	if input.BreedID <= 0 {
		http.Error(w, "breed_id обязателен", http.StatusBadRequest)
		return
	}
	if input.Date == "" {
		http.Error(w, "date обязательна", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(input.PetName) == "" {
		http.Error(w, "кличка обязательна", http.StatusBadRequest)
		return
	}
	if input.StartTime == "" {
		http.Error(w, "start_time обязателен", http.StatusBadRequest)
		return
	}

	appointment, err := h.groomingRepo.CreateAppointment(claims.ClinicID, input)
	if err != nil {
		if errors.Is(err, repository.ErrGroomingNotWorkingDay) {
			http.Error(w, "не рабочий день", http.StatusBadRequest)
			return
		}
		if errors.Is(err, repository.ErrGroomingOutOfHours) {
			http.Error(w, "запись за пределами рабочего времени", http.StatusBadRequest)
			return
		}
		if errors.Is(err, repository.ErrGroomingConflict) {
			http.Error(w, "время занято другой записью", http.StatusConflict)
			return
		}
		log.Printf("ошибка создания записи: %v", err)
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, appointment)
}

// DeleteAppointment обрабатывает DELETE /api/admin/grooming/appointments/{id}
func (h *GroomingHandler) DeleteAppointment(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromContext(r)
	id := r.PathValue("id")

	if err := h.groomingRepo.DeleteAppointment(id, claims.ClinicID); err != nil {
		log.Printf("ошибка удаления записи: %v", err)
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// parseID парсит строку в int
func parseID(s string, out *int) (bool, error) {
	var n int
	_, err := scanInt(s, &n)
	if err != nil {
		return false, err
	}
	*out = n
	return true, nil
}

func scanInt(s string, out *int) (int, error) {
	n := 0
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return 0, errors.New("не число")
		}
		n = n*10 + int(ch-'0')
	}
	*out = n
	return n, nil
}
