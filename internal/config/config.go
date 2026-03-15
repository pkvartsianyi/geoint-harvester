package config

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	MongoURI        string
	TGSessionString string
	TGAPIID         int
	TGAPIHash       string
	PinpointAuthToken string
	Channels        []string
}

func Load() *Config {
	mongoURI := os.Getenv("MONGODB_URI")
	tgSession := os.Getenv("TG_SESSION_STRING")
	apiIDStr := os.Getenv("TG_API_ID")
	apiHash := os.Getenv("TG_API_HASH")
	pinpointToken := os.Getenv("PINPOINT_AUTH_TOKEN")

	if mongoURI == "" || tgSession == "" || apiIDStr == "" || apiHash == "" || pinpointToken == "" {
		log.Fatal("Missing required environment variables: MONGODB_URI, TG_SESSION_STRING, TG_API_ID, TG_API_HASH, PINPOINT_AUTH_TOKEN")
	}

	apiID, err := strconv.Atoi(apiIDStr)
	if err != nil {
		log.Fatalf("Invalid TG_API_ID: %v", err)
	}

	return &Config{
		MongoURI:        mongoURI,
		TGSessionString: tgSession,
		TGAPIID:         apiID,
		TGAPIHash:       apiHash,
		PinpointAuthToken: pinpointToken,
		Channels:        []string{"telegram", "durov"},
	}
}
