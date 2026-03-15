package bot

import (
	"fmt"
	"log"
	"strings"
	"time"

	"go-server/internal/repository"

	tele "gopkg.in/telebot.v3"
)

// Bot — обёртка над telebot с нашими зависимостями
type Bot struct {
	tele        *tele.Bot
	clinicSlug  string
	animalRepo  *repository.AnimalRepository
	articleRepo *repository.ArticleRepository
}

// New создаёт и настраивает Telegram бота
func New(token, clinicSlug string, animalRepo *repository.AnimalRepository, articleRepo *repository.ArticleRepository) (*Bot, error) {
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

	bot := &Bot{
		tele:        b,
		clinicSlug:  clinicSlug,
		animalRepo:  animalRepo,
		articleRepo: articleRepo,
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
	b.tele.Handle("ℹ️ Помощь", b.handleHelp)

	// Обработчик Inline-кнопок (кнопки прямо в сообщении)
	b.tele.Handle(tele.OnCallback, b.handleCallback)
}

// handleStart обрабатывает /start
func (b *Bot) handleStart(c tele.Context) error {
	text := "🏥 *Ветеринарная первая помощь*\n\n" +
		"Здесь вы можете получить информацию о первой помощи вашему питомцу в нерабочие часы клиники.\n\n" +
		"Используйте кнопки ниже для навигации."

	return c.Send(text, mainMenuKeyboard(), tele.ModeMarkdown)
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
	case "back":
		// C. Кнопка "Назад" — возвращает к списку животных
		return b.showAnimalsInline(c)
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

// showArticle показывает содержимое статьи
func (b *Bot) showArticle(c tele.Context, slug string) error {
	article, err := b.articleRepo.GetBySlug(b.clinicSlug, slug)
	if err != nil || article == nil {
		log.Printf("ошибка получения статьи %s: %v", slug, err)
		return c.Edit("Статья не найдена.")
	}

	text := fmt.Sprintf("*%s*\n\n%s", article.Title, article.Content)

	backBtn := tele.Btn{Text: "⬅️ Назад", Data: "back:start"}
	menu := &tele.ReplyMarkup{}
	menu.Inline(tele.Row{backBtn})

	return c.Edit(text, menu, tele.ModeMarkdown)
}
