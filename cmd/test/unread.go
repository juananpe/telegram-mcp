package main

import (
	"context"
	"sort"
	"time"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	cfg "github.com/spf13/pflag"
	"golang.org/x/time/rate"
)

const (
	defaultMessageLimit = 10
	maxDialogsLimit     = 100
	rateLimitPerSec     = 5
)

//nolint:gochecknoglobals // CLI flags must be global
var (
	messageLimit = cfg.Int("limit", defaultMessageLimit, "limit of unread messages to fetch")
)

//nolint:gochecknoglobals // Rate limiter should be global for consistent rate limiting across all functions
var telegramLimiter = rate.NewLimiter(rate.Limit(rateLimitPerSec), 1)

// UnreadMessage represents a simplified message structure
type UnreadMessage struct {
	ID        int
	Text      string
	Date      time.Time
	FromID    int64
	FromName  string
	ChatType  string
	ChatTitle string
}

// DialogWithUnread represents a dialog with its unread count and latest message ID
type DialogWithUnread struct {
	Dialog      *tg.Dialog
	UnreadCount int
	TopMessage  int
}

// getUnreadMessages fetches unread messages from different users
//
//nolint:gocognit,gocyclo // complexity is inherent to handling different types of Telegram messages and users
func getUnreadMessages(ctx context.Context, client *telegram.Client) ([]UnreadMessage, error) {
	if err := telegramLimiter.Wait(ctx); err != nil {
		return nil, errors.Wrap(err, "rate limiter wait")
	}

	api := client.API()
	dialogsClass, err := api.MessagesGetDialogs(ctx, &tg.MessagesGetDialogsRequest{
		OffsetPeer:    &tg.InputPeerEmpty{},
		OffsetDate:    0,
		OffsetID:      0,
		Limit:         maxDialogsLimit,
		Hash:          0,
		Flags:         0,
		ExcludePinned: false,
		FolderID:      0,
	})
	if err != nil {
		return nil, errors.Wrap(err, "get dialogs")
	}

	var dialogs *tg.MessagesDialogs
	switch d := dialogsClass.(type) {
	case *tg.MessagesDialogs:
		dialogs = d
	case *tg.MessagesDialogsSlice:
		dialogs = &tg.MessagesDialogs{
			Dialogs:  d.Dialogs,
			Messages: d.Messages,
			Chats:    d.Chats,
			Users:    d.Users,
		}
	default:
		return nil, errors.New("unexpected dialogs response type")
	}

	// Create a slice of dialogs with unread count
	dialogsWithUnread := make([]DialogWithUnread, 0, len(dialogs.Dialogs))
	for _, dialog := range dialogs.Dialogs {
		dialogItem, ok := dialog.(*tg.Dialog)
		if !ok {
			continue
		}

		if dialogItem.UnreadCount > 0 {
			dialogsWithUnread = append(dialogsWithUnread, DialogWithUnread{
				Dialog:      dialogItem,
				UnreadCount: dialogItem.UnreadCount,
				TopMessage:  dialogItem.TopMessage,
			})
		}
	}

	// Sort dialogs by TopMessage in descending order (newest first)
	sort.Slice(dialogsWithUnread, func(i, j int) bool {
		return dialogsWithUnread[i].TopMessage > dialogsWithUnread[j].TopMessage
	})

	// Map to store the latest message from each user
	userMessages := make(map[int64]UnreadMessage)
	processedCount := 0

	for _, dialogWithUnread := range dialogsWithUnread {
		dialogItem := dialogWithUnread.Dialog

		var inputPeer tg.InputPeerClass
		var chatType, chatTitle string
		var fromID int64
		var fromName string

		switch peer := dialogItem.Peer.(type) {
		case *tg.PeerUser:
			for _, userItem := range dialogs.Users {
				user, ok := userItem.(*tg.User)
				if !ok || user.ID != peer.UserID {
					continue
				}

				inputPeer = &tg.InputPeerUser{
					UserID:     user.ID,
					AccessHash: user.AccessHash,
				}
				chatType = "user"
				chatTitle = user.FirstName + " " + user.LastName
				fromID = user.ID
				fromName = chatTitle

				break
			}
		case *tg.PeerChat:
			inputPeer = &tg.InputPeerChat{
				ChatID: peer.ChatID,
			}
			chatType = "chat"
			for _, chatItem := range dialogs.Chats {
				chat, ok := chatItem.(*tg.Chat)
				if !ok || chat.ID != peer.ChatID {
					continue
				}

				chatTitle = chat.Title

				break
			}
		case *tg.PeerChannel:
			for _, channelItem := range dialogs.Chats {
				channel, ok := channelItem.(*tg.Channel)
				if !ok || channel.ID != peer.ChannelID {
					continue
				}

				inputPeer = &tg.InputPeerChannel{
					ChannelID:  channel.ID,
					AccessHash: channel.AccessHash,
				}
				chatType = "channel"
				chatTitle = channel.Title

				break
			}
		}

		if inputPeer == nil {
			continue
		}

		if err := telegramLimiter.Wait(ctx); err != nil {
			return nil, errors.Wrap(err, "rate limiter wait")
		}

		messagesClass, err := api.MessagesGetHistory(ctx, &tg.MessagesGetHistoryRequest{
			Peer:       inputPeer,
			OffsetID:   0,
			OffsetDate: 0,
			AddOffset:  0,
			Limit:      1, // We only need the latest message
			MaxID:      0,
			MinID:      0,
			Hash:       0,
		})
		if err != nil {
			log.Error().Err(err).Msg("failed to get messages")

			continue
		}

		var messages *tg.MessagesMessages
		switch m := messagesClass.(type) {
		case *tg.MessagesMessages:
			messages = m
		case *tg.MessagesMessagesSlice:
			messages = &tg.MessagesMessages{
				Messages: m.Messages,
				Chats:    m.Chats,
				Users:    m.Users,
			}
		case *tg.MessagesChannelMessages:
			messages = &tg.MessagesMessages{
				Messages: m.Messages,
				Chats:    m.Chats,
				Users:    m.Users,
			}
		default:
			log.Error().Msg("unexpected messages response type")

			continue
		}

		for _, msg := range messages.Messages {
			message, ok := msg.(*tg.Message)
			if !ok {
				continue
			}

			if message.Out {
				continue
			}

			if message.FromID != nil {
				if from, ok := message.FromID.(*tg.PeerUser); ok {
					for _, userItem := range messages.Users {
						user, ok := userItem.(*tg.User)
						if !ok || user.ID != from.UserID {
							continue
						}

						fromID = user.ID
						fromName = user.FirstName + " " + user.LastName

						break
					}
				}
			}

			unreadMsg := UnreadMessage{
				ID:        message.ID,
				Text:      message.Message,
				Date:      time.Unix(int64(message.Date), 0),
				FromID:    fromID,
				FromName:  fromName,
				ChatType:  chatType,
				ChatTitle: chatTitle,
			}

			// Only store if we haven't seen this user yet or if this message is newer
			if existingMsg, exists := userMessages[fromID]; !exists || unreadMsg.Date.After(existingMsg.Date) {
				userMessages[fromID] = unreadMsg
				processedCount++
			}

			break // We only need the latest message
		}

		if len(userMessages) >= *messageLimit {
			break
		}
	}

	// Convert map to slice and sort by date
	messages := make([]UnreadMessage, 0, len(userMessages))
	for _, msg := range userMessages {
		messages = append(messages, msg)
	}

	// Sort messages by date in descending order
	sort.Slice(messages, func(i, j int) bool {
		return messages[i].Date.After(messages[j].Date)
	})

	return messages, nil
}
