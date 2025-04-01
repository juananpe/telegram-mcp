package tg

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
	"github.com/rs/zerolog/log"
)

func Auth(phone string, appID int64, appHash string, sessionPath string) error {
	client := telegram.NewClient(int(appID), appHash, telegram.Options{
		SessionStorage: &telegram.FileSessionStorage{
			Path: sessionPath,
		},
	})

	sessionDir := filepath.Dir(sessionPath)
	if err := os.MkdirAll(sessionDir, 0700); err != nil {
		return fmt.Errorf("mkdir(%s): %w", sessionDir, err)
	}

	if err := client.Run(context.Background(), func(ctx context.Context) error {
		// Authenticate if needed
		flow := auth.NewFlow(auth.Constant(phone, "", auth.CodeAuthenticatorFunc(func(ctx context.Context, _ *tg.AuthSentCode) (string, error) {
			fmt.Print("Enter code: ")
			code, err := bufio.NewReader(os.Stdin).ReadString('\n')
			if err != nil {
				return "", err
			}
			return strings.TrimSpace(code), nil
		})), auth.SendCodeOptions{})

		if err := client.Auth().IfNecessary(ctx, flow); err != nil {
			return fmt.Errorf("auth: %w", err)
		}

		// Get current user info
		self, err := client.Self(ctx)
		if err != nil {
			return fmt.Errorf("get self info: %w", err)
		}

		log.Info().
			Str("first_name", self.FirstName).
			Str("last_name", self.LastName).
			Str("username", self.Username).
			Int64("id", self.ID).
			Msg("Logged in as")

		return nil
	}); err != nil {
		return fmt.Errorf("client error: %w", err)
	}

	return nil
}
