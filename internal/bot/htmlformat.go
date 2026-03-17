package bot

import (
	"strings"

	"golang.org/x/net/html"
)

// htmlToTelegram конвертирует HTML от TipTap в текст с Telegram HTML разметкой.
// Telegram поддерживает: <b>, <i>, <s>, <u>, <code>, <pre>
// Заголовки → жирный текст, списки → • / 1. 2. 3.
func htmlToTelegram(src string) string {
	doc, err := html.Parse(strings.NewReader(src))
	if err != nil {
		// Если парсинг не удался — возвращаем исходник без тегов
		return stripTags(src)
	}

	var sb strings.Builder
	walkNode(&sb, doc, false, false)

	// Убираем лишние переносы строк
	result := strings.TrimSpace(sb.String())
	for strings.Contains(result, "\n\n\n") {
		result = strings.ReplaceAll(result, "\n\n\n", "\n\n")
	}
	return result
}

func walkNode(sb *strings.Builder, n *html.Node, inOL bool, firstLI bool) {
	if n.Type == html.TextNode {
		sb.WriteString(html.EscapeString(n.Data))
		return
	}

	if n.Type != html.ElementNode {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walkNode(sb, c, inOL, false)
		}
		return
	}

	tag := n.Data
	switch tag {
	case "h1", "h2", "h3":
		sb.WriteString("<b>")
		walkChildren(sb, n, false)
		sb.WriteString("</b>\n\n")

	case "p":
		walkChildren(sb, n, false)
		sb.WriteString("\n")

	case "strong", "b":
		sb.WriteString("<b>")
		walkChildren(sb, n, false)
		sb.WriteString("</b>")

	case "em", "i":
		sb.WriteString("<i>")
		walkChildren(sb, n, false)
		sb.WriteString("</i>")

	case "s":
		sb.WriteString("<s>")
		walkChildren(sb, n, false)
		sb.WriteString("</s>")

	case "u":
		sb.WriteString("<u>")
		walkChildren(sb, n, false)
		sb.WriteString("</u>")

	case "code":
		sb.WriteString("<code>")
		walkChildren(sb, n, false)
		sb.WriteString("</code>")

	case "pre":
		sb.WriteString("<pre>")
		walkChildren(sb, n, false)
		sb.WriteString("</pre>\n")

	case "ul":
		walkListItems(sb, n, false)
		sb.WriteString("\n")

	case "ol":
		walkListItems(sb, n, true)
		sb.WriteString("\n")

	case "li":
		// Обрабатывается внутри walkListItems
		walkChildren(sb, n, inOL)

	case "br":
		sb.WriteString("\n")

	default:
		walkChildren(sb, n, inOL)
	}
}

func walkChildren(sb *strings.Builder, n *html.Node, inOL bool) {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		walkNode(sb, c, inOL, false)
	}
}

func walkListItems(sb *strings.Builder, n *html.Node, ordered bool) {
	counter := 1
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type != html.ElementNode || c.Data != "li" {
			continue
		}
		if ordered {
			sb.WriteString(strings.Repeat(" ", 0))
			sb.WriteString(itoa(counter))
			sb.WriteString(". ")
			counter++
		} else {
			sb.WriteString("• ")
		}
		// Содержимое <li> — убираем обёртку <p> чтобы не было лишних переносов
		writeLiContent(sb, c)
		sb.WriteString("\n")
	}
}

func writeLiContent(sb *strings.Builder, li *html.Node) {
	for c := li.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == "p" {
			walkChildren(sb, c, false)
		} else {
			walkNode(sb, c, false, false)
		}
	}
}

// pageCharLimit — максимальное количество символов на одну страницу статьи
const pageCharLimit = 1200

// paginateText разбивает текст на страницы по границам абзацев
func paginateText(text string) []string {
	if len([]rune(text)) <= pageCharLimit {
		return []string{text}
	}

	paragraphs := strings.Split(text, "\n\n")
	var pages []string
	var current strings.Builder

	for i, para := range paragraphs {
		addition := para
		if i > 0 {
			addition = "\n\n" + para
		}
		if current.Len() > 0 && len([]rune(current.String()))+len([]rune(addition)) > pageCharLimit {
			pages = append(pages, strings.TrimSpace(current.String()))
			current.Reset()
			current.WriteString(para)
		} else {
			current.WriteString(addition)
		}
	}
	if current.Len() > 0 {
		pages = append(pages, strings.TrimSpace(current.String()))
	}
	if len(pages) == 0 {
		return []string{text}
	}
	return pages
}

func stripTags(s string) string {
	var sb strings.Builder
	inTag := false
	for _, r := range s {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag:
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	digits := []byte{}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}
