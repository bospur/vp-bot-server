package bot

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"go-server/internal/repository"

	xhtml "golang.org/x/net/html"
	tele "gopkg.in/telebot.v3"
)

// Bot — обёртка над telebot с нашими зависимостями
type Bot struct {
	tele        *tele.Bot
	clinicSlug  string
	publicURL   string
	appURL      string
	animalRepo  *repository.AnimalRepository
	articleRepo *repository.ArticleRepository
	doctorRepo  *repository.DoctorRepository
}

// New создаёт и настраивает Telegram бота
func New(token, clinicSlug, publicURL, appURL string, animalRepo *repository.AnimalRepository, articleRepo *repository.ArticleRepository, doctorRepo *repository.DoctorRepository) (*Bot, error) {
	pref := tele.Settings{
		Token:  token,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания бота: %w", err)
	}

	// A. Регистрируем команды в меню "/" Telegram
	err = b.SetCommands([]tele.Command{
		{Text: "start", Description: "Начать / Главное меню"},
		{Text: "menu", Description: "Выбрать животное"},
		{Text: "help", Description: "Помощь"},
	})
	if err != nil {
		log.Printf("не удалось установить команды бота: %v", err)
	}

	// B. Устанавливаем Menu Button — кнопка появляется слева от поля ввода
	// и как быстрый доступ в списке чатов (кнопка "Открыть")
	_, err = b.Raw("setChatMenuButton", map[string]interface{}{
		"menu_button": map[string]interface{}{
			"type": "web_app",
			"text": "Открыть",
			"web_app": map[string]string{
				"url": appURL,
			},
		},
	})
	if err != nil {
		log.Printf("не удалось установить menu button: %v", err)
	}

	bot := &Bot{
		tele:        b,
		clinicSlug:  clinicSlug,
		publicURL:   publicURL,
		appURL:      appURL,
		animalRepo:  animalRepo,
		articleRepo: articleRepo,
		doctorRepo:  doctorRepo,
	}

	bot.registerHandlers()

	return bot, nil
}

// Start запускает бота (блокирующий вызов)
func (b *Bot) Start() {
	log.Println("Telegram бот запущен")
	b.tele.Start()
}

// mainMenuKeyboard — B. Постоянная Reply-клавиатура под полем ввода
// Показывается всегда, как обычные кнопки телефона
func mainMenuKeyboard() *tele.ReplyMarkup {
	menu := &tele.ReplyMarkup{ResizeKeyboard: true}
	menu.Reply(
		menu.Row(menu.Text("🐾 Выбрать животное")),
		menu.Row(menu.Text("👨‍⚕️ Врачи"), menu.Text("📅 Расписание")),
		menu.Row(menu.Text("ℹ️ Помощь")),
	)
	return menu
}

// registerHandlers регистрирует все обработчики
func (b *Bot) registerHandlers() {
	b.tele.Handle("/start", b.handleStart)
	b.tele.Handle("/menu", b.handleMenu)
	b.tele.Handle("/help", b.handleHelp)

	// Обработчик Reply-кнопок (текстовые кнопки под полем ввода)
	b.tele.Handle("🐾 Выбрать животное", b.handleMenu)
	b.tele.Handle("👨‍⚕️ Врачи", b.handleDoctors)
	b.tele.Handle("📅 Расписание", b.handleSchedule)
	b.tele.Handle("ℹ️ Помощь", b.handleHelp)

	// Обработчик Inline-кнопок (кнопки прямо в сообщении)
	b.tele.Handle(tele.OnCallback, b.handleCallback)
}

// handleStart обрабатывает /start
func (b *Bot) handleStart(c tele.Context) error {
	text := "🏥 *Ветеринарная первая помощь*\n\n" +
		"Здесь вы можете получить информацию о первой помощи вашему питомцу в нерабочие часы клиники.\n\n" +
		"Используйте кнопки ниже для навигации."

	inline := &tele.ReplyMarkup{}
	inline.Inline(
		inline.Row(tele.Btn{
			Text:   "🏥 Открыть приложение",
			WebApp: &tele.WebApp{URL: b.appURL},
		}),
	)

	return c.Send(text, mainMenuKeyboard(), inline, tele.ModeMarkdown)
}

// handleMenu показывает список животных
func (b *Bot) handleMenu(c tele.Context) error {
	animals, err := b.animalRepo.GetAllByClinic(b.clinicSlug)
	if err != nil {
		log.Printf("ошибка получения животных: %v", err)
		return c.Send("Произошла ошибка. Попробуйте позже.")
	}

	if len(animals) == 0 {
		return c.Send("Информация пока недоступна. Попробуйте позже.")
	}

	var rows []tele.Row
	for _, animal := range animals {
		icon := animal.Icon
		if icon == "" {
			icon = "🐾"
		}
		btn := tele.Btn{
			Text: fmt.Sprintf("%s %s", icon, animal.Name),
			Data: fmt.Sprintf("animal:%s", animal.Slug),
		}
		rows = append(rows, tele.Row{btn})
	}

	inlineMenu := &tele.ReplyMarkup{}
	inlineMenu.Inline(rows...)

	return c.Send("Выберите вид животного:", inlineMenu)
}

// handleHelp обрабатывает /help
func (b *Bot) handleHelp(c tele.Context) error {
	text := "ℹ️ *Как пользоваться ботом*\n\n" +
		"1. Нажмите *🐾 Выбрать животное*\n" +
		"2. Выберите вид животного\n" +
		"3. Выберите ситуацию\n" +
		"4. Прочитайте инструкцию по первой помощи\n\n" +
		"⚠️ Бот не заменяет визит к ветеринару. При серьёзных симптомах обратитесь в клинику."

	return c.Send(text, mainMenuKeyboard(), tele.ModeMarkdown)
}

// handleCallback обрабатывает нажатия Inline-кнопок
func (b *Bot) handleCallback(c tele.Context) error {
	data := c.Callback().Data

	// Парсим "тип:значение"
	idx := strings.Index(data, ":")
	if idx == -1 {
		return c.Respond()
	}
	cbType := data[:idx]
	cbValue := data[idx+1:]

	switch cbType {
	case "animal":
		return b.showCategories(c, cbValue)
	case "category":
		// Формат cbValue: "animalSlug|categorySlug"
		parts := strings.SplitN(cbValue, "|", 2)
		if len(parts) != 2 {
			return c.Respond()
		}
		return b.showArticles(c, parts[0], parts[1])
	case "article":
		return b.showArticle(c, cbValue)
	case "ap":
		// Формат cbValue: "slug:pageIndex"
		parts := strings.SplitN(cbValue, ":", 2)
		if len(parts) != 2 {
			return c.Respond()
		}
		pageIdx, err := strconv.Atoi(parts[1])
		if err != nil {
			return c.Respond()
		}
		return b.showArticlePageBySlug(c, parts[0], pageIdx)
	case "doctor":
		return b.showDoctor(c, cbValue)
	case "sp":
		pageIdx, err := strconv.Atoi(cbValue)
		if err != nil {
			return c.Respond()
		}
		return b.showSchedulePage(c, pageIdx)
	case "back":
		switch cbValue {
		case "doctors":
			return b.showDoctorsList(c)
		default:
			return b.showAnimalsInline(c)
		}
	}

	return c.Respond()
}

// showAnimalsInline редактирует текущее сообщение показывая список животных
func (b *Bot) showAnimalsInline(c tele.Context) error {
	animals, err := b.animalRepo.GetAllByClinic(b.clinicSlug)
	if err != nil {
		return c.Edit("Произошла ошибка. Попробуйте позже.")
	}

	var rows []tele.Row
	for _, animal := range animals {
		icon := animal.Icon
		if icon == "" {
			icon = "🐾"
		}
		btn := tele.Btn{
			Text: fmt.Sprintf("%s %s", icon, animal.Name),
			Data: fmt.Sprintf("animal:%s", animal.Slug),
		}
		rows = append(rows, tele.Row{btn})
	}

	menu := &tele.ReplyMarkup{}
	menu.Inline(rows...)

	return c.Edit("Выберите вид животного:", menu)
}

// showCategories показывает категории животного
func (b *Bot) showCategories(c tele.Context, animalSlug string) error {
	categories, err := b.animalRepo.GetCategoriesByAnimalSlug(b.clinicSlug, animalSlug)
	if err != nil {
		log.Printf("ошибка получения категорий: %v", err)
		return c.Edit("Произошла ошибка. Попробуйте позже.")
	}

	if len(categories) == 0 {
		return c.Edit("Категории пока не добавлены.")
	}

	var rows []tele.Row
	for _, cat := range categories {
		icon := cat.Icon
		if icon == "" {
			icon = "📋"
		}
		btn := tele.Btn{
			Text: fmt.Sprintf("%s %s", icon, cat.Name),
			Data: fmt.Sprintf("category:%s|%s", animalSlug, cat.Slug),
		}
		rows = append(rows, tele.Row{btn})
	}

	backBtn := tele.Btn{Text: "⬅️ Назад", Data: "back:start"}
	rows = append(rows, tele.Row{backBtn})

	menu := &tele.ReplyMarkup{}
	menu.Inline(rows...)

	return c.Edit("Выберите ситуацию:", menu)
}

// showArticles показывает список статей категории
func (b *Bot) showArticles(c tele.Context, animalSlug, categorySlug string) error {
	articles, err := b.articleRepo.GetByCategory(b.clinicSlug, animalSlug, categorySlug)
	if err != nil {
		log.Printf("ошибка получения статей: %v", err)
		return c.Edit("Произошла ошибка. Попробуйте позже.")
	}

	if len(articles) == 0 {
		return c.Edit("Статьи пока не добавлены.")
	}

	var rows []tele.Row
	for _, art := range articles {
		btn := tele.Btn{
			Text: art.Title,
			Data: fmt.Sprintf("article:%s", art.Slug),
		}
		rows = append(rows, tele.Row{btn})
	}

	backBtn := tele.Btn{Text: "⬅️ Назад", Data: fmt.Sprintf("animal:%s", animalSlug)}
	rows = append(rows, tele.Row{backBtn})

	menu := &tele.ReplyMarkup{}
	menu.Inline(rows...)

	return c.Edit("Выберите статью:", menu)
}

// showArticle показывает первую страницу статьи
func (b *Bot) showArticle(c tele.Context, slug string) error {
	article, err := b.articleRepo.GetBySlug(b.clinicSlug, slug)
	if err != nil || article == nil {
		log.Printf("ошибка получения статьи %s: %v", slug, err)
		return c.Edit("Статья не найдена.")
	}

	pages := paginateText(htmlToTelegram(article.Content))

	if err := c.Delete(); err != nil {
		log.Printf("не удалось удалить сообщение меню: %v", err)
	}
	return b.renderArticlePage(c, article.Title, slug, pages, 0, false)
}

// showArticlePageBySlug загружает статью и показывает нужную страницу (редактируя текущее сообщение)
func (b *Bot) showArticlePageBySlug(c tele.Context, slug string, pageIdx int) error {
	article, err := b.articleRepo.GetBySlug(b.clinicSlug, slug)
	if err != nil || article == nil {
		return c.Edit("Статья не найдена.")
	}

	pages := paginateText(htmlToTelegram(article.Content))
	if pageIdx < 0 {
		pageIdx = 0
	}
	if pageIdx >= len(pages) {
		pageIdx = len(pages) - 1
	}

	return b.renderArticlePage(c, article.Title, slug, pages, pageIdx, true)
}

// renderArticlePage формирует и отправляет/редактирует сообщение со страницей статьи
func (b *Bot) renderArticlePage(c tele.Context, title, slug string, pages []string, pageIdx int, edit bool) error {
	total := len(pages)

	var text string
	if total == 1 {
		text = fmt.Sprintf("<b>%s</b>\n\n%s", xhtml.EscapeString(title), pages[0])
	} else {
		text = fmt.Sprintf("<b>%s</b>  <i>(%d/%d)</i>\n\n%s", xhtml.EscapeString(title), pageIdx+1, total, pages[pageIdx])
	}

	var rows []tele.Row

	// Кнопки листания
	if total > 1 {
		var navRow tele.Row
		if pageIdx > 0 {
			navRow = append(navRow, tele.Btn{Text: "← Назад", Data: fmt.Sprintf("ap:%s:%d", slug, pageIdx-1)})
		}
		if pageIdx < total-1 {
			navRow = append(navRow, tele.Btn{Text: "Далее →", Data: fmt.Sprintf("ap:%s:%d", slug, pageIdx+1)})
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
