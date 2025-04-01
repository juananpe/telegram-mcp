package tg

import "github.com/gotd/td/telegram"

type Client struct {
	T *telegram.Client
}

func New(appID int, appHash, sessionPath string) *Client {
	client := telegram.NewClient(appID, appHash, telegram.Options{
		SessionStorage: &telegram.FileSessionStorage{
			Path: sessionPath,
		},
		NoUpdates: true,
	})

	return &Client{
		T: client,
	}
}
