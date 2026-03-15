package bot

import (
	"fmt"
	"log"
	"time"

	"go-server/internal/repository"

	tele "gopkg.in/telebot.v3"
)

// Bot — обёртка над telebot с нашими зависимостями
type Bot struct {
	tele       *tele.Bot
	animalRepo *repository.AnimalRepository
	articleRepo *repository.ArticleRepository
}

// New создаёт и настраивает Telegram бота
func New(token string, animalRepo *repository.AnimalRepository, articleRepo *repository.ArticleRepository) (*Bot, error) {
	pref := tele.Settings{
		Token:  token,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания бота: %w", err)
	}

	bot := &Bot{
		tele:        b,
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

// registerHandlers регистрирует все обработчики команд и кнопок
func (b *Bot) registerHandlers() {
	// /start — приветствие и список животных
	b.tele.Handle("/start", b.handleStart)

	// Кнопка выбора животного — показывает категории
	b.tele.Handle(tele.OnCallback, b.handleCallback)
}

// handleStart обрабатывает команду /start
func (b *Bot) handleStart(c tele.Context) error {
	animals, err := b.animalRepo.GetAll()
	if err != nil {
		log.Printf("ошибка получения животных: %v", err)
		return c.Send("Произошла ошибка. Попробуйте позже.")
	}

	if len(animals) == 0 {
		return c.Send("Информация пока недоступна. Попробуйте позже.")
	}

	// Строим клавиатуру из списка животных
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

	return c.Send("🏥 *Ветеринарная помощь*\n\nВыберите вид животного:", menu, tele.ModeMarkdown)
}

// handleCallback обрабатывает нажатия на кнопки
func (b *Bot) handleCallback(c tele.Context) error {
	data := c.Callback().Data

	// Парсим данные кнопки: "тип:значение"
	var cbType, cbValue string
	fmt.Sscanf(data, "%s", &data)

	// Разбираем вручную через разделитель ":"
	for i, ch := range data {
		if ch == ':' {
			cbType = data[:i]
			cbValue = data[i+1:]
			break
		}
	}

	switch cbType {
	case "animal":
		return b.showCategories(c, cbValue)
	case "category":
		// Формат: "animalSlug|categorySlug"
		var animalSlug, categorySlug string
		for i, ch := range cbValue {
			if ch == '|' {
				animalSlug = cbValue[:i]
				categorySlug = cbValue[i+1:]
				break
			}
		}
		return b.showArticles(c, animalSlug, categorySlug)
	case "article":
		return b.showArticle(c, cbValue)
	}

	return c.Respond()
}

// showCategories показывает категории для выбранного животного
func (b *Bot) showCategories(c tele.Context, animalSlug string) error {
	categories, err := b.animalRepo.GetCategoriesByAnimalSlug(animalSlug)
	if err != nil {
		log.Printf("ошибка получения категорий: %v", err)
		return c.Send("Произошла ошибка. Попробуйте позже.")
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

	// Кнопка "Назад"
	backBtn := tele.Btn{Text: "⬅️ Назад", Data: "back:start"}
	rows = append(rows, tele.Row{backBtn})

	menu := &tele.ReplyMarkup{}
	menu.Inline(rows...)

	return c.Edit("Выберите ситуацию:", menu)
}

// showArticles показывает список статей категории
func (b *Bot) showArticles(c tele.Context, animalSlug, categorySlug string) error {
	articles, err := b.articleRepo.GetByCategory(animalSlug, categorySlug)
	if err != nil {
		log.Printf("ошибка получения статей: %v", err)
		return c.Send("Произошла ошибка. Попробуйте позже.")
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
	article, err := b.articleRepo.GetBySlug(slug)
	if err != nil || article == nil {
		log.Printf("ошибка получения статьи %s: %v", slug, err)
		return c.Send("Статья не найдена.")
	}

	text := fmt.Sprintf("*%s*\n\n%s", article.Title, article.Content)

	backBtn := tele.Btn{Text: "⬅️ Назад", Data: "back:start"}
	menu := &tele.ReplyMarkup{}
	menu.Inline(tele.Row{backBtn})

	return c.Edit(text, menu, tele.ModeMarkdown)
}
