package main

import (
	"bufio"
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
	"github.com/spf13/viper"
)

// termAuth implements auth.UserAuthenticator
type termAuth struct {
}

func (a termAuth) Phone(ctx context.Context) (string, error) {
	fmt.Print("Enter phone number: ")
	phone, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	return strings.TrimSpace(phone), nil
}

func (a termAuth) Password(ctx context.Context) (string, error) {
	fmt.Print("Enter password (if any): ")
	password, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	return strings.TrimSpace(password), nil
}

func (a termAuth) Code(ctx context.Context, sentCode *tg.AuthSentCode) (string, error) {
	fmt.Print("Enter code: ")
	code, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	return strings.TrimSpace(code), nil
}

func (a termAuth) AcceptTermsOfService(ctx context.Context, tos tg.HelpTermsOfService) error {
	return nil
}

func (a termAuth) SignUp(ctx context.Context) (auth.UserInfo, error) {
	return auth.UserInfo{}, fmt.Errorf("sign up not supported")
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Println("No .env file found, reading from environment variables")
	}

	appID := viper.GetInt("TG_API_ID")
	appHash := viper.GetString("TG_API_HASH")

	if appID == 0 || appHash == "" {
		log.Fatal("Missing required environment variables: TG_API_ID, TG_API_HASH")
	}

	storage := &session.StorageMemory{}
	client := telegram.NewClient(appID, appHash, telegram.Options{
		SessionStorage: storage,
	})

	if err := client.Run(ctx, func(ctx context.Context) error {
		if err := auth.NewFlow(termAuth{}, auth.SendCodeOptions{}).Run(ctx, client.Auth()); err != nil {
			return err
		}

		loader := session.Loader{Storage: storage}
		data, err := loader.Load(ctx)
		if err != nil {
			return fmt.Errorf("load session: %w", err)
		}

		fmt.Println("\n--- Telegram Session Data (Copy to .env) ---")
		fmt.Printf("TG_SESSION_AUTH_KEY=%s\n", hex.EncodeToString(data.AuthKey))
		fmt.Printf("TG_DC=%d\n", data.DC)
		fmt.Printf("TG_SESSION_ADDR=%s\n", data.Addr)
		
		// If address is empty, we can't restore easily unless we know it.
		// DC 4 production address is often 149.154.167.91:443
		if data.Addr == "" {
			fmt.Println("# Note: Address was not saved. Suggested address for DC", data.DC, "is:")
			switch data.DC {
			case 1: fmt.Println("# TG_SESSION_ADDR=149.154.175.53:443")
			case 2: fmt.Println("# TG_SESSION_ADDR=149.154.167.51:443")
			case 3: fmt.Println("# TG_SESSION_ADDR=149.154.175.100:443")
			case 4: fmt.Println("# TG_SESSION_ADDR=149.154.167.91:443")
			case 5: fmt.Println("# TG_SESSION_ADDR=91.108.56.130:443")
			}
		}
		fmt.Println("--------------------------------------------")

		return nil
	}); err != nil {
		log.Fatalf("Execution failed: %v", err)
	}
}
