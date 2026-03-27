package bot

import (
	"fmt"
	"log"
	"strings"
	"time"

	"go-server/internal/repository"

	xhtml "golang.org/x/net/html"
	tele "gopkg.in/telebot.v3"
)

var ruMonths = []string{
	"", "янв", "фев", "мар", "апр", "май", "июн",
	"июл", "авг", "сен", "окт", "ноя", "дек",
}

var ruDays = []string{"Вс", "Пн", "Вт", "Ср", "Чт", "Пт", "Сб"}

// handleDoctors обрабатывает нажатие Reply-кнопки "👨‍⚕️ Врачи"
func (b *Bot) handleDoctors(c tele.Context) error {
	doctors, err := b.doctorRepo.GetPublishedByClinicSlug(b.clinicSlug)
	if err != nil {
		log.Printf("ошибка получения врачей: %v", err)
		return c.Send("Произошла ошибка. Попробуйте позже.")
	}

	if len(doctors) == 0 {
		return c.Send("Информация о врачах пока недоступна.")
	}

	menu := &tele.ReplyMarkup{}
	menu.Inline(buildDoctorRows(doctors)...)
	return c.Send("👨‍⚕️ Наши врачи:", menu)
}

// showDoctorsList редактирует текущее inline-сообщение показывая список врачей
func (b *Bot) showDoctorsList(c tele.Context) error {
	doctors, err := b.doctorRepo.GetPublishedByClinicSlug(b.clinicSlug)
	if err != nil {
		return c.Edit("Произошла ошибка. Попробуйте позже.")
	}
	if len(doctors) == 0 {
		return c.Edit("Информация о врачах пока недоступна.")
	}

	menu := &tele.ReplyMarkup{}
	menu.Inline(buildDoctorRows(doctors)...)
	return c.Edit("👨‍⚕️ Наши врачи:", menu)
}

// showDoctor показывает карточку врача
func (b *Bot) showDoctor(c tele.Context, idStr string) error {
	doctor, err := b.doctorRepo.GetByID(idStr)
	if err != nil || doctor == nil {
		return c.Edit("Врач не найден.")
	}

	text := formatDoctorCard(doctor)

	menu := &tele.ReplyMarkup{}
	menu.Inline(
		menu.Row(tele.Btn{Text: "⬅️ К списку врачей", Data: "back:doctors"}),
	)

	if doctor.PhotoURL != "" {
		photo := &tele.Photo{
			File:    tele.FromURL(b.publicURL + doctor.PhotoURL),
			Caption: text,
		}
		if err := c.Send(photo, menu, tele.ModeHTML); err == nil {
			_ = c.Delete()
			return nil
		}
		log.Printf("не удалось отправить фото врача %s: пробуем текст", idStr)
	}

	return c.Edit(text, menu, tele.ModeHTML)
}

// handleSchedule обрабатывает нажатие Reply-кнопки "📅 Расписание"
func (b *Bot) handleSchedule(c tele.Context) error {
	text, err := b.buildScheduleText()
	if err != nil {
		log.Printf("ошибка получения расписания: %v", err)
		return c.Send("Произошла ошибка. Попробуйте позже.")
	}

	pages := paginateText(text)
	return b.renderSchedulePage(c, pages, 0, false)
}

// showSchedulePage редактирует сообщение показывая нужную страницу расписания
func (b *Bot) showSchedulePage(c tele.Context, pageIdx int) error {
	text, err := b.buildScheduleText()
	if err != nil {
		return c.Edit("Произошла ошибка. Попробуйте позже.")
	}

	pages := paginateText(text)
	if pageIdx < 0 {
		pageIdx = 0
	}
	if pageIdx >= len(pages) {
		pageIdx = len(pages) - 1
	}
	return b.renderSchedulePage(c, pages, pageIdx, true)
}

func (b *Bot) buildScheduleText() (string, error) {
	entries, _, err := b.doctorRepo.GetScheduleByClinicSlug(b.clinicSlug)
	if err != nil {
		return "", err
	}

	if len(entries) == 0 {
		return "Расписание пока не добавлено.", nil
	}

	var sb strings.Builder
	sb.WriteString("📅 <b>Расписание клиники</b>\n")

	currentDate := ""
	for _, e := range entries {
		if e.Date != currentDate {
			currentDate = e.Date
			if t, err := time.Parse("2006-01-02", e.Date); err == nil {
				sb.WriteString(fmt.Sprintf("\n<b>%s, %d %s</b>\n",
					ruDays[t.Weekday()], t.Day(), ruMonths[int(t.Month())]))
			}
		}
		name := xhtml.EscapeString(e.FullName)
		if e.Specialty != "" {
			name += " (" + xhtml.EscapeString(e.Specialty) + ")"
		}
		sb.WriteString(fmt.Sprintf("• %s  %s–%s\n", name, trimTime(e.TimeFrom), trimTime(e.TimeTo)))
	}

	return strings.TrimSpace(sb.String()), nil
}

func (b *Bot) renderSchedulePage(c tele.Context, pages []string, pageIdx int, edit bool) error {
	total := len(pages)
	text := pages[pageIdx]
	if total > 1 {
		text = fmt.Sprintf("%s\n\n<i>(%d/%d)</i>", text, pageIdx+1, total)
	}

	var rows []tele.Row
	if total > 1 {
		var navRow tele.Row
		if pageIdx > 0 {
			navRow = append(navRow, tele.Btn{Text: "← Назад", Data: fmt.Sprintf("sp:%d", pageIdx-1)})
		}
		if pageIdx < total-1 {
			navRow = append(navRow, tele.Btn{Text: "Далее →", Data: fmt.Sprintf("sp:%d", pageIdx+1)})
		}
		if len(navRow) > 0 {
			rows = append(rows, navRow)
		}
	}
	rows = append(rows, tele.Row{tele.Btn{Text: "⬅️ В меню", Data: "back:start"}})

	menu := &tele.ReplyMarkup{}
	menu.Inline(rows...)

	if edit {
		return c.Edit(text, menu, tele.ModeHTML)
	}
	return c.Send(text, menu, tele.ModeHTML)
}

// ── helpers ───────────────────────────────────────────────────────────────────

func buildDoctorRows(doctors []repository.Doctor) []tele.Row {
	var rows []tele.Row
	for _, d := range doctors {
		label := d.FullName
		if d.Specialty != "" {
			label += " — " + d.Specialty
		}
		rows = append(rows, tele.Row{
			tele.Btn{Text: label, Data: fmt.Sprintf("doctor:%d", d.ID)},
		})
	}
	return rows
}

func formatDoctorCard(d *repository.Doctor) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("<b>%s</b>", xhtml.EscapeString(d.FullName)))
	if d.Specialty != "" {
		sb.WriteString("\n<i>" + xhtml.EscapeString(d.Specialty) + "</i>")
	}
	if d.Description != "" {
		sb.WriteString("\n\n" + htmlToTelegram(d.Description))
	}
	if d.Contacts != "" {
		sb.WriteString("\n\n📞 " + xhtml.EscapeString(d.Contacts))
	}
	return sb.String()
}

// trimTime берёт "HH:MM:SS" или "HH:MM" → "HH:MM"
func trimTime(t string) string {
	if len(t) >= 5 {
		return t[:5]
	}
	return t
}
