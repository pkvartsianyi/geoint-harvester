package telegram

import (
	"context"
	"log"
	"time"

	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
	"github.com/pkvartsianyi/geoint-harvester/internal/domain"
)

type TelegramAdapter struct {
	appID   int
	appHash string
	session *session.Data
}

func NewTelegramAdapter(appID int, appHash string, sessionData *session.Data) *TelegramAdapter {
	return &TelegramAdapter{
		appID:   appID,
		appHash: appHash,
		session: sessionData,
	}
}

func (a *TelegramAdapter) Fetch(ctx context.Context, channels []string) ([]domain.Message, error) {
	storage := &session.StorageMemory{}
	client := telegram.NewClient(a.appID, a.appHash, telegram.Options{
		SessionStorage: storage,
	})

	var allMsgs []domain.Message
	err := client.Run(ctx, func(ctx context.Context) error {
		loader := session.Loader{Storage: storage}
		if err := loader.Save(ctx, a.session); err != nil {
			return err
		}

		api := client.API()

		for _, username := range channels {
			log.Printf("Scraping channel: %s", username)

			peer, err := api.ContactsResolveUsername(ctx, &tg.ContactsResolveUsernameRequest{
				Username: username,
			})
			if err != nil {
				log.Printf("Failed to resolve username %s: %v", username, err)
				continue
			}

			channel, ok := peer.Peer.(*tg.PeerChannel)
			if !ok {
				log.Printf("Username %s is not a channel", username)
				continue
			}

			var inputChannel tg.InputChannel
			for _, c := range peer.Chats {
				if ch, ok := c.(*tg.Channel); ok && ch.ID == channel.ChannelID {
					inputChannel.ChannelID = ch.ID
					inputChannel.AccessHash = ch.AccessHash
					break
				}
			}

			history, err := api.MessagesGetHistory(ctx, &tg.MessagesGetHistoryRequest{
				Peer:  &tg.InputPeerChannel{ChannelID: inputChannel.ChannelID, AccessHash: inputChannel.AccessHash},
				Limit: 20,
			})
			if err != nil {
				log.Printf("Failed to get history for %s: %v", username, err)
				continue
			}

			messages, ok := history.(*tg.MessagesChannelMessages)
			if !ok {
				log.Printf("Unexpected history type for %s", username)
				continue
			}

			for _, m := range messages.Messages {
				msg, ok := m.(*tg.Message)
				if !ok {
					continue
				}

				allMsgs = append(allMsgs, domain.Message{
					Channel:   username,
					Content:   msg.Message,
					MsgID:     msg.ID,
					Timestamp: time.Unix(int64(msg.Date), 0),
				})
			}
			log.Printf("Finished scraping %s", username)
		}
		return nil
	})

	return allMsgs, err
}
