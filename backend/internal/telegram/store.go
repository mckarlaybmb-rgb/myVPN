package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	TelegramID                            int64
	Username, FirstName, LastName, UserID string
}
type Store interface {
	Upsert(context.Context, User) error
	LinkByEmail(context.Context, int64, string) error
	Get(context.Context, int64) (User, error)
	RecordNotification(context.Context, int64, string, string) (bool, error)
}
type PGStore struct{ pool *pgxpool.Pool }

func NewPGStore(pool *pgxpool.Pool) *PGStore { return &PGStore{pool: pool} }
func (s *PGStore) Upsert(ctx context.Context, user User) error {
	_, err := s.pool.Exec(ctx, `INSERT INTO telegram_users (telegram_id, username, first_name, last_name) VALUES ($1,$2,$3,$4) ON CONFLICT (telegram_id) DO UPDATE SET username=$2, first_name=$3, last_name=$4, updated_at=NOW()`, user.TelegramID, user.Username, user.FirstName, user.LastName)
	return err
}
func (s *PGStore) LinkByEmail(ctx context.Context, telegramID int64, email string) error {
	result, err := s.pool.Exec(ctx, `UPDATE telegram_users SET user_id = (SELECT id FROM users WHERE email = $2), updated_at=NOW() WHERE telegram_id=$1 AND EXISTS (SELECT 1 FROM users WHERE email=$2)`, telegramID, strings.TrimSpace(email))
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("account not found")
	}
	return nil
}
func (s *PGStore) Get(ctx context.Context, telegramID int64) (User, error) {
	var user User
	err := s.pool.QueryRow(ctx, `SELECT telegram_id, username, first_name, last_name, COALESCE(user_id::text,'') FROM telegram_users WHERE telegram_id=$1`, telegramID).Scan(&user.TelegramID, &user.Username, &user.FirstName, &user.LastName, &user.UserID)
	return user, err
}
func (s *PGStore) RecordNotification(ctx context.Context, telegramID int64, notificationType, referenceID string) (bool, error) {
	result, err := s.pool.Exec(ctx, `INSERT INTO telegram_notifications (telegram_id, notification_type, reference_id) VALUES ($1,$2,NULLIF($3,'')::uuid) ON CONFLICT DO NOTHING`, telegramID, notificationType, referenceID)
	return result.RowsAffected() == 1, err
}

type Notifier struct {
	pool *pgxpool.Pool
	bot  *Bot
}

func NewNotifier(pool *pgxpool.Pool, bot *Bot) *Notifier { return &Notifier{pool: pool, bot: bot} }
func (n *Notifier) Notify(ctx context.Context, notificationType, referenceID string) error {
	var chatID int64
	err := n.pool.QueryRow(ctx, `SELECT tu.telegram_id FROM telegram_users tu JOIN subscriptions s ON s.user_id=tu.user_id WHERE s.id=$1`, referenceID).Scan(&chatID)
	if err != nil {
		return err
	}
	inserted, err := n.bot.store.RecordNotification(ctx, chatID, notificationType, referenceID)
	if err != nil || !inserted {
		return err
	}
	return n.bot.Send(ctx, chatID, "Subscription update: "+notificationType)
}

type Bot struct {
	token   string
	client  *http.Client
	store   Store
	apiBase string
}

func NewBot(token string, store Store) *Bot {
	return &Bot{token: token, client: &http.Client{Timeout: 15 * time.Second}, store: store, apiBase: "https://api.telegram.org/bot" + token}
}

type Message struct {
	Chat struct {
		ID int64 `json:"id"`
	} `json:"chat"`
	From struct {
		ID                            int64 `json:"id"`
		Username, FirstName, LastName string
	} `json:"from"`
	Text string `json:"text"`
}
type update struct {
	UpdateID int      `json:"update_id"`
	Message  *Message `json:"message"`
}

func (b *Bot) call(ctx context.Context, method string, values url.Values, target any) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, b.apiBase+"/"+method, strings.NewReader(values.Encode()))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := b.client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("telegram HTTP %d", response.StatusCode)
	}
	return json.NewDecoder(response.Body).Decode(target)
}
func (b *Bot) Send(ctx context.Context, chatID int64, text string) error {
	var response struct {
		OK          bool   `json:"ok"`
		Description string `json:"description"`
	}
	err := b.call(ctx, "sendMessage", url.Values{"chat_id": {strconv.FormatInt(chatID, 10)}, "text": {text}}, &response)
	if err != nil {
		return err
	}
	if !response.OK {
		return fmt.Errorf("telegram send failed: %s", response.Description)
	}
	return nil
}
func (b *Bot) Poll(ctx context.Context) error {
	offset := 0
	for {
		values := url.Values{"timeout": {"30"}, "offset": {strconv.Itoa(offset)}}
		var response struct {
			OK     bool     `json:"ok"`
			Result []update `json:"result"`
		}
		if err := b.call(ctx, "getUpdates", values, &response); err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			continue
		}
		for _, item := range response.Result {
			offset = item.UpdateID + 1
			if item.Message != nil {
				_ = b.handle(ctx, item.Message)
			}
		}
	}
}
func (b *Bot) handle(ctx context.Context, message *Message) error {
	user := User{TelegramID: message.From.ID, Username: message.From.Username, FirstName: message.From.FirstName, LastName: message.From.LastName}
	if err := b.store.Upsert(ctx, user); err != nil {
		return err
	}
	command := strings.Fields(message.Text)
	if len(command) == 0 {
		return nil
	}
	switch command[0] {
	case "/start":
		if len(command) > 1 {
			if err := b.store.LinkByEmail(ctx, message.From.ID, command[1]); err != nil {
				return b.Send(ctx, message.Chat.ID, "Account not found.")
			}
			return b.Send(ctx, message.Chat.ID, "Account linked.")
		}
		return b.Send(ctx, message.Chat.ID, "Send /start followed by your account email to link.")
	case "/account", "/vpn", "/status":
		linked, err := b.store.Get(ctx, message.From.ID)
		if err != nil || linked.UserID == "" {
			return b.Send(ctx, message.Chat.ID, "Link an account with /start first.")
		}
		return b.Send(ctx, message.Chat.ID, "Account linked. Contact support for account details.")
	case "/renew":
		return b.Send(ctx, message.Chat.ID, "Renewal is handled by an administrator.")
	case "/support":
		return b.Send(ctx, message.Chat.ID, "Please contact support.")
	default:
		return nil
	}
}
