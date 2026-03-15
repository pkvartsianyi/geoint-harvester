package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	MongoURI               string
	TGSessionAuthKey       string
	TGSessionAuthKeyID     string
	TGSessionSalt          int64
	TGAPIID                int
	TGDC                   int
	TGAPIHash              string
	PinpointAuthToken      string
	Channels               []string
}

func Load() *Config {
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Println("No .env file found, reading from environment variables")
	}

	required := []string{"MONGODB_URI", "TG_API_ID", "TG_API_HASH", "TG_SESSION_AUTH_KEY", "TG_SESSION_AUTH_KEY_ID", "TG_SESSION_SALT", "TG_DC", "PINPOINT_AUTH_TOKEN"}
	for _, key := range required {
		if !viper.IsSet(key) {
			log.Fatalf("Missing required environment variable: %s", key)
		}
	}

	return &Config{
		MongoURI:           viper.GetString("MONGODB_URI"),
		TGSessionAuthKey:   viper.GetString("TG_SESSION_AUTH_KEY"),
		TGSessionAuthKeyID: viper.GetString("TG_SESSION_AUTH_KEY_ID"),
		TGSessionSalt:      viper.GetInt64("TG_SESSION_SALT"),
		TGAPIID:            viper.GetInt("TG_API_ID"),
		TGDC:               viper.GetInt("TG_DC"),
		TGAPIHash:          viper.GetString("TG_API_HASH"),
		PinpointAuthToken:  viper.GetString("PINPOINT_AUTH_TOKEN"),
		Channels:           []string{"telegram", "durov"},
	}
}
