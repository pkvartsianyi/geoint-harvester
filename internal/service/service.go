package service

import (
	"context"
	"log"

	"github.com/pkvartsianyi/geoint-harvester/internal/domain"
)

// Repository is a driven port for persisting messages.
type Repository interface {
	UpsertMessage(ctx context.Context, msg domain.Message) error
	Exists(ctx context.Context, channel string, msgID int) (bool, error)
}

// Extractor is a driven port for extracting geolocation from text.
type Extractor interface {
	Extract(ctx context.Context, text string) (*domain.GeoPoint, error)
}

// Source is a driven port for fetching messages from external sources.
type Source interface {
	Fetch(ctx context.Context, channels []string) ([]domain.Message, error)
}

// ScraperService is the application service (core) that orchestrates scraping logic.
type ScraperService struct {
	repo      Repository
	extractor Extractor
}

func NewScraperService(repo Repository, extractor Extractor) *ScraperService {
	return &ScraperService{
		repo:      repo,
		extractor: extractor,
	}
}

func (s *ScraperService) Run(ctx context.Context, channels []string, sources ...Source) error {
	var allMsgs []domain.Message
	for _, source := range sources {
		msgs, err := source.Fetch(ctx, channels)
		if err != nil {
			log.Printf("Source fetch failed: %v", err)
			continue
		}
		allMsgs = append(allMsgs, msgs...)
	}

	var newMsgs []domain.Message
	for _, msg := range allMsgs {
		exists, err := s.repo.Exists(ctx, msg.Channel, msg.MsgID)
		if err != nil {
			log.Printf("Failed to check message existence: %v", err)
		}
		if !exists {
			newMsgs = append(newMsgs, msg)
		}
	}

	for i := range newMsgs {
		if s.extractor != nil && newMsgs[i].Content != "" {
			geo, err := s.extractor.Extract(ctx, newMsgs[i].Content)
			if err == nil && geo.Coordinates != nil {
				newMsgs[i].Geolocation = geo
			} else {
				log.Printf("Extraction failed for msg %d: %s", newMsgs[i].MsgID, err)
				continue
			}
		}

		if err := s.repo.UpsertMessage(ctx, newMsgs[i]); err != nil {
			log.Printf("Failed to store message %d: %v", newMsgs[i].MsgID, err)
		}
	}

	return nil
}
