package tg

import "github.com/gotd/td/telegram"

type Client struct {
	appID       int
	appHash     string
	sessionPath string
}

func New(appID int, appHash, sessionPath string) *Client {
	return &Client{
		appID:       appID,
		appHash:     appHash,
		sessionPath: sessionPath,
	}
}

func (c *Client) T() *telegram.Client {
	opts := telegram.Options{
		SessionStorage: &telegram.FileSessionStorage{
			Path: c.sessionPath,
		},
		NoUpdates: true,
	}
	opts, _ = telegram.OptionsFromEnvironment(opts)
	return telegram.NewClient(c.appID, c.appHash, opts)
}
