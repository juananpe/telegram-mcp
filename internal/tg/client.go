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
	return telegram.NewClient(c.appID, c.appHash, telegram.Options{
		SessionStorage: &telegram.FileSessionStorage{
			Path: c.sessionPath,
		},
		NoUpdates: true,
	})
}
