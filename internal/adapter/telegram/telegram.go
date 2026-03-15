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

// NewTelegramAdapter builds an adapter from the values printed by the session tool:
// authKeyHex   — TG_SESSION_AUTH_KEY
// authKeyIDHex — TG_SESSION_AUTH_KEY_ID
// salt         — TG_SESSION_SALT
// dc           — TG_DC
func NewTelegramAdapter(appID int, appHash, authKeyHex, authKeyIDHex string, salt int64, dc int) (*TelegramAdapter, error) {
	authKey, err := hex.DecodeString(authKeyHex)
	if err != nil {
		return nil, fmt.Errorf("decoding auth key: %w", err)
	}
	if len(authKey) != 256 {
		return nil, fmt.Errorf("auth key must be 256 bytes, got %d", len(authKey))
	}

	authKeyID, err := hex.DecodeString(authKeyIDHex)
	if err != nil {
		return nil, fmt.Errorf("decoding auth key ID: %w", err)
	}

	return &TelegramAdapter{
		appID:   appID,
		appHash: appHash,
		session: &session.Data{
			DC:        dc,
			AuthKey:   authKey,
			AuthKeyID: authKeyID,
			Salt:      salt,
		},
	}, nil
}

func (a *TelegramAdapter) Fetch(ctx context.Context, channels []string) ([]domain.Message, error) {
	storage := &session.StorageMemory{}

	loader := session.Loader{Storage: storage}
	if err := loader.Save(ctx, a.session); err != nil {
		return nil, fmt.Errorf("restoring session: %w", err)
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
