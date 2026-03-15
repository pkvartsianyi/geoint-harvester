package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/gotd/td/session"
	"github.com/pkvartsianyi/geoint-harvester/internal/adapter/db"
	"github.com/pkvartsianyi/geoint-harvester/internal/adapter/pinpoint"
	"github.com/pkvartsianyi/geoint-harvester/internal/adapter/telegram"
	"github.com/pkvartsianyi/geoint-harvester/internal/config"
	"github.com/pkvartsianyi/geoint-harvester/internal/service"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	cfg := config.Load()

	mongoAdapter, err := db.NewMongoAdapter(ctx, cfg.MongoURI)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer func() {
		if err := mongoAdapter.Close(ctx); err != nil {
			log.Printf("Failed to close MongoDB connection: %v", err)
		}
	}()

	pinpointAdapter := pinpoint.NewPinpointAdapter(cfg.PinpointAuthToken)
	scraperService := service.NewScraperService(mongoAdapter, pinpointAdapter)

	data, err := session.TelethonSession(cfg.TGSessionString)
	if err != nil {
		log.Fatalf("Failed to parse Telegram session string: %v", err)
	}

	tgAdapter := telegram.NewTelegramAdapter(cfg.TGAPIID, cfg.TGAPIHash, data)
	if err := scraperService.Run(ctx, cfg.Channels, tgAdapter); err != nil {
		log.Fatalf("Scraping failed: %v", err)
	}

	log.Println("Scraping completed successfully.")
}
