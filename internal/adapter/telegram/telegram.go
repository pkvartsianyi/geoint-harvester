package telegram

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"github.com/gotd/td/session"
	tdtg "github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
	"github.com/pkvartsianyi/geoint-harvester/internal/domain"
)

type TelegramAdapter struct {
	appID   int
	appHash string
	session *session.Data
}

// NewTelegramAdapter builds an adapter from the values printed at first login:
// authKeyHex — the AuthKey (Hex) string
// dc         — the DC integer
// addr       — the Addr string (optional)
func NewTelegramAdapter(appID int, appHash, authKeyHex string, dc int, addr string) (*TelegramAdapter, error) {
	authKey, err := hex.DecodeString(authKeyHex)
	if err != nil {
		return nil, fmt.Errorf("decoding auth key: %w", err)
	}
	if len(authKey) != 256 {
		return nil, fmt.Errorf("auth key must be 256 bytes, got %d", len(authKey))
	}

	if addr == "" {
		// Default DC addresses for production if not provided.
		switch dc {
		case 1:
			addr = "149.154.175.53:443"
		case 2:
			addr = "149.154.167.51:443"
		case 3:
			addr = "149.154.175.100:443"
		case 4:
			addr = "149.154.167.91:443"
		case 5:
			addr = "91.108.56.130:443"
		}
	}

	return &TelegramAdapter{
		appID:   appID,
		appHash: appHash,
		session: &session.Data{
			DC:      dc,
			Addr:    addr,
			AuthKey: authKey,
		},
	}, nil
}

func (a *TelegramAdapter) Fetch(ctx context.Context, channels []string) ([]domain.Message, error) {
	storage := &session.StorageMemory{}

	log.Printf("Connecting to Telegram (DC %d, Addr %s, Key len %d)...", a.session.DC, a.session.Addr, len(a.session.AuthKey))

	// Pre-populate storage with the session data to ensure the client
	// connects to the correct DC from the start.
	loader := session.Loader{Storage: storage}
	if err := loader.Save(ctx, a.session); err != nil {
		return nil, fmt.Errorf("storing session: %w", err)
	}

	client := tdtg.NewClient(a.appID, a.appHash, tdtg.Options{
		SessionStorage: storage,
	})

	var allMsgs []domain.Message
	err := client.Run(ctx, func(ctx context.Context) error {
		api := client.API()

		for _, username := range channels {
			if err := a.scrapeChannel(ctx, api, username, &allMsgs); err != nil {
				log.Printf("Error scraping channel %s: %v", username, err)
			}
			time.Sleep(500 * time.Millisecond)
		}
		return nil
	})

	return allMsgs, err
}

func (a *TelegramAdapter) scrapeChannel(ctx context.Context, api *tg.Client, username string, out *[]domain.Message) error {
	log.Printf("Scraping channel: %s", username)

	peer, err := api.ContactsResolveUsername(ctx, &tg.ContactsResolveUsernameRequest{
		Username: username,
	})
	if err != nil {
		return fmt.Errorf("resolving username: %w", err)
	}

	peerChannel, ok := peer.Peer.(*tg.PeerChannel)
	if !ok {
		return fmt.Errorf("username %s is not a channel", username)
	}

	var inputChannel *tg.InputChannel
	for _, c := range peer.Chats {
		if ch, ok := c.(*tg.Channel); ok && ch.ID == peerChannel.ChannelID {
			inputChannel = &tg.InputChannel{
				ChannelID:  ch.ID,
				AccessHash: ch.AccessHash,
			}
			break
		}
	}
	if inputChannel == nil {
		return fmt.Errorf("channel %s not found in resolved chats", username)
	}

	history, err := api.MessagesGetHistory(ctx, &tg.MessagesGetHistoryRequest{
		Peer:  &tg.InputPeerChannel{ChannelID: inputChannel.ChannelID, AccessHash: inputChannel.AccessHash},
		Limit: 20,
	})
	if err != nil {
		return fmt.Errorf("getting history: %w", err)
	}

	messages, ok := history.(*tg.MessagesChannelMessages)
	if !ok {
		return fmt.Errorf("unexpected history type for %s", username)
	}

	for _, m := range messages.Messages {
		msg, ok := m.(*tg.Message)
		if !ok {
			continue
		}
		*out = append(*out, domain.Message{
			Channel:   username,
			Content:   msg.Message,
			MsgID:     msg.ID,
			Timestamp: time.Unix(int64(msg.Date), 0),
		})
	}

	log.Printf("Finished scraping %s (%d messages)", username, len(messages.Messages))
	return nil
}
