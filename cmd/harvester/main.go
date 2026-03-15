package main

import (
	"context"
	"log"
	"os"
	"os/signal"

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

	tgAdapter, err := telegram.NewTelegramAdapter(cfg.TGAPIID, cfg.TGAPIHash, cfg.TGSessionAuthKey, cfg.TGSessionAuthKeyID, cfg.TGSessionSalt, cfg.TGDC)
	if err != nil {
		log.Fatalf("Failed to initialize Telegram adapter: %v", err)
	}

	if err := scraperService.Run(ctx, cfg.Channels, tgAdapter); err != nil {
		log.Fatalf("Scraping failed: %v", err)
	}

	log.Println("Scraping completed successfully.")
}
