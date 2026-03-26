package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go-server/internal/middleware"
	"go-server/internal/repository"
)

// DoctorHandler содержит зависимости для эндпоинтов врачей
type DoctorHandler struct {
	doctorRepo *repository.DoctorRepository
	uploadsDir string
}

func NewDoctorHandler(doctorRepo *repository.DoctorRepository, uploadsDir string) *DoctorHandler {
	return &DoctorHandler{doctorRepo: doctorRepo, uploadsDir: uploadsDir}
}

// ── Врачи (admin) ─────────────────────────────────────────────────────────────

// GetDoctors обрабатывает GET /api/admin/doctors
func (h *DoctorHandler) GetDoctors(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromContext(r)

	doctors, err := h.doctorRepo.GetAll(claims.ClinicID)
	if err != nil {
		log.Printf("ошибка получения врачей: %v", err)
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
	if doctors == nil {
		doctors = []repository.Doctor{}
	}
	writeJSON(w, http.StatusOK, doctors)
}

// GetDoctor обрабатывает GET /api/admin/doctors/{id}
func (h *DoctorHandler) GetDoctor(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	doctor, err := h.doctorRepo.GetByID(id)
	if err != nil {
		log.Printf("ошибка получения врача: %v", err)
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
	if doctor == nil {
		http.Error(w, "не найдено", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, doctor)
}

// CreateDoctor обрабатывает POST /api/admin/doctors
func (h *DoctorHandler) CreateDoctor(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromContext(r)

	var input repository.DoctorInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "неверный формат запроса", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(input.FullName) == "" {
		http.Error(w, "ФИО обязательно", http.StatusBadRequest)
		return
	}

	doctor, err := h.doctorRepo.Create(claims.ClinicID, input)
	if err != nil {
		log.Printf("ошибка создания врача: %v", err)
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, doctor)
}

// UpdateDoctor обрабатывает PUT /api/admin/doctors/{id}
func (h *DoctorHandler) UpdateDoctor(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var input repository.DoctorInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "неверный формат запроса", http.StatusBadRequest)
		return
	}

	doctor, err := h.doctorRepo.Update(id, input)
	if err != nil {
		log.Printf("ошибка обновления врача: %v", err)
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
	if doctor == nil {
		http.Error(w, "не найдено", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, doctor)
}

// UpdateDoctorStatus обрабатывает PATCH /api/admin/doctors/{id}/status (только admin)
func (h *DoctorHandler) UpdateDoctorStatus(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromContext(r)
	if claims.Role != "admin" {
		http.Error(w, "доступ запрещён", http.StatusForbidden)
		return
	}

	id := r.PathValue("id")
	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "неверный формат запроса", http.StatusBadRequest)
		return
	}
	if body.Status != "draft" && body.Status != "published" {
		http.Error(w, "недопустимый статус", http.StatusBadRequest)
		return
	}

	doctor, err := h.doctorRepo.UpdateStatus(id, body.Status)
	if err != nil {
		log.Printf("ошибка обновления статуса врача: %v", err)
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
	if doctor == nil {
		http.Error(w, "не найдено", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, doctor)
}

// UploadDoctorPhoto обрабатывает POST /api/admin/doctors/{id}/photo
func (h *DoctorHandler) UploadDoctorPhoto(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if err := r.ParseMultipartForm(5 << 20); err != nil { // 5 MB
		http.Error(w, "файл слишком большой (макс 5 МБ)", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("photo")
	if err != nil {
		http.Error(w, "файл не найден в запросе", http.StatusBadRequest)
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" {
		http.Error(w, "допустимые форматы: jpg, png, webp", http.StatusBadRequest)
		return
	}

	filename := fmt.Sprintf("doctor_%s_%d%s", id, time.Now().UnixMilli(), ext)
	dstPath := filepath.Join(h.uploadsDir, filename)

	dst, err := os.Create(dstPath)
	if err != nil {
		log.Printf("ошибка создания файла: %v", err)
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		log.Printf("ошибка записи файла: %v", err)
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	photoURL := "/uploads/" + filename
	if err := h.doctorRepo.UpdatePhoto(id, photoURL); err != nil {
		log.Printf("ошибка обновления photo_url: %v", err)
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"photo_url": photoURL})
}

// DeleteDoctor обрабатывает DELETE /api/admin/doctors/{id}
func (h *DoctorHandler) DeleteDoctor(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if err := h.doctorRepo.Delete(id); err != nil {
		log.Printf("ошибка удаления врача: %v", err)
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ── Расписание (admin) ────────────────────────────────────────────────────────

// GetDoctorSchedule обрабатывает GET /api/admin/doctors/{id}/schedule
func (h *DoctorHandler) GetDoctorSchedule(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	slots, err := h.doctorRepo.GetSchedule(id)
	if err != nil {
		log.Printf("ошибка получения расписания: %v", err)
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
	if slots == nil {
		slots = []repository.DoctorSchedule{}
	}
	writeJSON(w, http.StatusOK, slots)
}

// AddScheduleSlot обрабатывает POST /api/admin/doctors/{id}/schedule
func (h *DoctorHandler) AddScheduleSlot(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var body struct {
		DayOfWeek int    `json:"day_of_week"`
		TimeFrom  string `json:"time_from"`
		TimeTo    string `json:"time_to"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "неверный формат запроса", http.StatusBadRequest)
		return
	}

	slot, err := h.doctorRepo.AddScheduleSlot(id, body.DayOfWeek, body.TimeFrom, body.TimeTo)
	if err != nil {
		log.Printf("ошибка добавления слота: %v", err)
		http.Error(w, "слот уже существует или неверные данные", http.StatusConflict)
		return
	}
	writeJSON(w, http.StatusCreated, slot)
}

// DeleteScheduleSlot обрабатывает DELETE /api/admin/doctors/{id}/schedule/{slotId}
func (h *DoctorHandler) DeleteScheduleSlot(w http.ResponseWriter, r *http.Request) {
	slotID := r.PathValue("slotId")

	if err := h.doctorRepo.DeleteScheduleSlot(slotID); err != nil {
		log.Printf("ошибка удаления слота: %v", err)
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ── Исключения (admin) ────────────────────────────────────────────────────────

// GetExceptions обрабатывает GET /api/admin/doctors/{id}/schedule/exceptions
func (h *DoctorHandler) GetExceptions(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	exceptions, err := h.doctorRepo.GetExceptions(id)
	if err != nil {
		log.Printf("ошибка получения исключений: %v", err)
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
	if exceptions == nil {
		exceptions = []repository.DoctorScheduleException{}
	}
	writeJSON(w, http.StatusOK, exceptions)
}

// UpsertException обрабатывает PUT /api/admin/doctors/{id}/schedule/exceptions
func (h *DoctorHandler) UpsertException(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var body struct {
		Date     string  `json:"date"`
		IsDayOff bool    `json:"is_day_off"`
		TimeFrom *string `json:"time_from"`
		TimeTo   *string `json:"time_to"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "неверный формат запроса", http.StatusBadRequest)
		return
	}

	exception, err := h.doctorRepo.UpsertException(id, body.Date, body.IsDayOff, body.TimeFrom, body.TimeTo)
	if err != nil {
		log.Printf("ошибка сохранения исключения: %v", err)
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, exception)
}

// DeleteException обрабатывает DELETE /api/admin/doctors/{id}/schedule/exceptions/{exceptionId}
func (h *DoctorHandler) DeleteException(w http.ResponseWriter, r *http.Request) {
	exceptionID := r.PathValue("exceptionId")

	if err := h.doctorRepo.DeleteException(exceptionID); err != nil {
		log.Printf("ошибка удаления исключения: %v", err)
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ── Настройки клиники (admin) ─────────────────────────────────────────────────

// GetSettings обрабатывает GET /api/admin/settings
func (h *DoctorHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromContext(r)

	settings, err := h.doctorRepo.GetSettings(claims.ClinicID)
	if err != nil {
		log.Printf("ошибка получения настроек: %v", err)
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, settings)
}

// UpdateSettings обрабатывает PATCH /api/admin/settings (только admin)
func (h *DoctorHandler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromContext(r)
	if claims.Role != "admin" {
		http.Error(w, "доступ запрещён", http.StatusForbidden)
		return
	}

	var body struct {
		ScheduleDisplayWeeks int `json:"schedule_display_weeks"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "неверный формат запроса", http.StatusBadRequest)
		return
	}
	if body.ScheduleDisplayWeeks < 1 || body.ScheduleDisplayWeeks > 5 {
		http.Error(w, "допустимые значения: 1–5", http.StatusBadRequest)
		return
	}

	settings, err := h.doctorRepo.UpsertSettings(claims.ClinicID, body.ScheduleDisplayWeeks)
	if err != nil {
		log.Printf("ошибка обновления настроек: %v", err)
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, settings)
}

// ── Публичный API (для бота) ──────────────────────────────────────────────────

// GetPublicDoctors обрабатывает GET /api/clinics/{clinicSlug}/doctors
func (h *DoctorHandler) GetPublicDoctors(w http.ResponseWriter, r *http.Request) {
	clinicSlug := r.PathValue("clinicSlug")

	// Получаем clinicID по slug через doctorRepo (нужен вспомогательный метод)
	// Пока используем прямой запрос через репо
	doctors, err := h.doctorRepo.GetPublishedByClinicSlug(clinicSlug)
	if err != nil {
		log.Printf("ошибка получения врачей: %v", err)
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
	if doctors == nil {
		doctors = []repository.Doctor{}
	}
	writeJSON(w, http.StatusOK, doctors)
}

// GetPublicSchedule обрабатывает GET /api/clinics/{clinicSlug}/schedule
func (h *DoctorHandler) GetPublicSchedule(w http.ResponseWriter, r *http.Request) {
	clinicSlug := r.PathValue("clinicSlug")

	entries, settings, err := h.doctorRepo.GetScheduleByClinicSlug(clinicSlug)
	if err != nil {
		log.Printf("ошибка получения расписания: %v", err)
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
	if entries == nil {
		entries = []repository.ScheduleEntry{}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"entries":  entries,
		"settings": settings,
	})
}
