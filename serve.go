package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/chaindead/telegram-mcp/internal/tg"

	mcp "github.com/metoro-io/mcp-golang"
	"github.com/metoro-io/mcp-golang/transport/stdio"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"
)

func serve(ctx context.Context, cmd *cli.Command) error {
	appID := cmd.Int("app-id")
	appHash := cmd.String("api-hash")
	sessionPath := cmd.String("session")
	dryRun := cmd.Bool("dry")

	_, err := os.Stat(sessionPath)
	if err != nil {
		return fmt.Errorf("session file not found(%s): %w", sessionPath, err)
	}

	server := mcp.NewServer(stdio.NewStdioServerTransport())
	client := tg.New(int(appID), appHash, sessionPath)

	if dryRun {
		answer, err := client.GetMe(tg.EmptyArguments{})
		if err != nil {
			return fmt.Errorf("get user: %w", err)
		}

		data, err := json.MarshalIndent(answer, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal: %w", err)
		}

		log.Info().RawJSON("answer", data).Msg("Check GetMe: OK")

		return nil
	}

	err = server.RegisterTool("me", "Get current Telegram account info", client.GetMe)
	if err != nil {
		return fmt.Errorf("register tool: %w", err)
	}

	if err := server.Serve(); err != nil {
		return fmt.Errorf("serve: %w", err)
	}

	<-ctx.Done()

	return nil
}
