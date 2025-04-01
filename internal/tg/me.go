package tg

import (
	"context"
	"encoding/json"

	mcp "github.com/metoro-io/mcp-golang"
	"github.com/pkg/errors"
)

type MeResponse struct {
	ID        int64  `json:"id" jsonschema:"required,description=User ID"`
	FirstName string `json:"first_name" jsonschema:"required,description=User's first name"`
	LastName  string `json:"last_name" jsonschema:"description=User's last name"`
	Username  string `json:"username" jsonschema:"description=User's username"`
}

type EmptyArguments struct{}

func (c *Client) GetMe(_ EmptyArguments) (*mcp.ToolResponse, error) {
	var toolResponse *mcp.ToolResponse

	client := c.T()
	if err := client.Run(context.Background(), func(ctx context.Context) error {
		self, err := client.Self(ctx)
		if err != nil {
			return errors.Wrap(err, "failed to get self info")
		}

		// Create response
		response := MeResponse{
			ID:        self.ID,
			FirstName: self.FirstName,
			LastName:  self.LastName,
			Username:  self.Username,
		}

		// Convert response to JSON
		jsonData, err := json.Marshal(response)
		if err != nil {
			return errors.Wrap(err, "failed to marshal response")
		}

		toolResponse = mcp.NewToolResponse(mcp.NewTextContent(string(jsonData)))

		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "invalid session")
	}

	return toolResponse, nil
}
