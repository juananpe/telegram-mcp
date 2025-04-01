# Telegram MCP server

The server is a bridge between the Telegram API and the AI assistants and is based on the [Model Context Protocol](https://modelcontextprotocol.io).

> [!IMPORTANT]
> Ensure that you have read and understood the [Telegram API Terms of Service](https://core.telegram.org/api/terms) before using this server.
> Any misuse of the Telegram API may result in the suspension of your account.

## What is MCP?

The Model Context Protocol (MCP) is a system that lets AI apps, like Claude Desktop or Cursor, connect to external tools and data sources. It gives a clear and safe way for AI assistants to work with local services and APIs while keeping the user in control.

## What does this server do?

- [x] Get current user data
- [x] Get the list of dialogs (chats, channels, groups)
- [ ] Get the list of (unread) messages in the given dialog
- [ ] Mark chanel as read
- [ ] Retrieve messages by date and time
- [ ] Get the list of contacts
- [ ] Draft a message

## Installation

```bash
go install github.com/chaindead/telegram-mcp@latest
```

## Configuration

### Telegram API Configuration

Before you can use the server, you need to connect to the Telegram API.

1. Get the API ID and hash from [Telegram API](https://my.telegram.org/auth)
2. Run the following command:

   ```bash
   telegram-mcp auth --app-id <your-api-id> --api-hash <your-api-hash> --phone <your-phone-number>
   ```

   Enter the code you received from Telegram to connect to the API.

   The password may be required if you have two-factor authentication enabled.

### Claude Desktop Configuration

Configure Claude Desktop to recognize the Exa MCP server.

1. Open the Claude Desktop configuration file:
    - in MacOS, the configuration file is located at `~/Library/Application Support/Claude/claude_desktop_config.json`
    - in Windows, the configuration file is located at `%APPDATA%\Claude\claude_desktop_config.json`

   > __Note:__
   > You can also find claude_desktop_config.json inside the settings of Claude Desktop app

2. Add the server configuration

    ```json
    {
      "mcpServers": {
        "telegram-mcp": {
          "command": "telegram-mcp",
          "env": {
            "TG_APP_ID": "<your-app-id>",
            "TG_API_HASH": "<your-api-hash>"
          }
        }
      }
    }
    ```