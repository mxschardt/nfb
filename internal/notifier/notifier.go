package notifier

import (
	"context"
	"fmt"
	"io"
	"library/internal/botkit/markup"
	"library/internal/model"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/go-shiori/go-readability"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type ArticleProvider interface {
	AllNotPosted(ctx context.Context, since time.Time, limit uint64) ([]model.Article, error)
	MarkPosted(ctx context.Context, id int64) error
}

type Notifier struct {
	articles         ArticleProvider
	bot              *tgbotapi.BotAPI
	sendInterval     time.Duration
	lookupTimeWindow time.Duration
	channelID        int64
}

func New(
	articles ArticleProvider,
	bot *tgbotapi.BotAPI,
	sendInterval time.Duration,
	lookupTimeWindow time.Duration,
	channelID int64,
) *Notifier {
	return &Notifier{
		articles:         articles,
		bot:              bot,
		sendInterval:     sendInterval,
		lookupTimeWindow: lookupTimeWindow,
		channelID:        channelID,
	}
}

func (n *Notifier) Start(ctx context.Context) error {
	ticker := time.NewTicker(n.sendInterval)
	defer ticker.Stop()

	if err := n.SelectAndSendArticle(ctx); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := n.SelectAndSendArticle(ctx); err != nil {
				return err
			}
		}
	}
}

func (n *Notifier) SelectAndSendArticle(ctx context.Context) error {
	log.Printf("[INFO] selecting new articles")
	// TODO: add transactions?
	topArticles, err := n.articles.AllNotPosted(ctx, time.Now().UTC().Add(-time.Hour*24), 1)
	if err != nil {
		// TODO: wrap
		return err
	}
	log.Printf("[INFO] got articles: %d", len(topArticles))

	if len(topArticles) == 0 {
		return nil
	}

	article := topArticles[0]

	log.Printf("[INFO] selected: %s", article.Link)

	summary, err := n.extractSummary(article)
	if err != nil {
		return err
	}

	log.Printf("[INFO] sending: %s", article.Link)
	if err := n.sendArticle(article, summary); err != nil {
		return err
	}
	log.Printf("[INFO] send: %s", article.Link)

	return n.articles.MarkPosted(ctx, article.ID)
}

func (n *Notifier) extractSummary(article model.Article) (string, error) {
	var r io.Reader

	if article.Summary != "" {
		r = strings.NewReader(article.Summary)
	}

	doc, err := readability.FromReader(r, nil)
	if err != nil {
		return "", err
	}

	return cleanText(doc.TextContent), nil
}

var redundantNewLines = regexp.MustCompile("\n{3,}")

func cleanText(text string) string {
	return redundantNewLines.ReplaceAllString(text, "\n")
}

func (n *Notifier) sendArticle(article model.Article, summary string) error {
	const msgFormat = "*%s*\n\n%s\n\n%s"

	msg := tgbotapi.NewMessage(n.channelID, fmt.Sprintf(
		msgFormat,
		markup.EscapeForMarkdown(article.Title),
		markup.EscapeForMarkdown(summary),
		markup.EscapeForMarkdown(article.Link),
	))

	msg.ParseMode = tgbotapi.ModeMarkdownV2

	_, err := n.bot.Send(msg)
	if err != nil {
		return err
	}

	return nil
}
