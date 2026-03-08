# 📊 ToDo Info

Analyze your Microsoft ToDo tasks. Track task ages, spot procrastination patterns, get productivity insights — via CLI or Telegram bot.

## 🎯 Task Rottenness Levels

- 😊 **Fresh** (0-2 days)
- 😏 **Ripe** (3-6 days)
- 🥱 **Tired** (7-13 days)
- 🤢 **Zombie** (14+ days)

## 🚀 Quick Start

### 1. Setup Azure App
1. [Azure Portal](https://portal.azure.com) → App Registrations → New
2. Redirect URI: `http://localhost:8080`
3. Grant `Tasks.ReadWrite` and `User.Read` permissions
4. Note the Application (client) ID

### 2. Install
```bash
git clone https://github.com/uchr/ToDoInfo.git
cd ToDoInfo
echo "AZURE_CLIENT_ID=your_client_id" > .env
go build -o todoinfo cmd/cli/main.go
```

### 3. Use
```bash
./todoinfo login           # Authenticate via browser
./todoinfo stats           # Fetch and display task stats
./todoinfo stats --offline # Use stored data (no API call)
./todoinfo logout          # Clear credentials
```

## 🤖 Telegram Bot

Long-running bot with periodic data collection, daily summaries, and on-demand queries.

```bash
./todoinfo bot --telegram-token TOKEN --telegram-chat-id CHAT_ID
```

Bot commands: `/login`, `/stats`, `/zombies`, `/oldest`, `/chart`, `/refresh`

## ⚙️ Configuration

`AZURE_CLIENT_ID` can be provided via `.env` file, `--client-id` flag, environment variable, or `~/.todoinfo.yaml`.

## 🚢 Deploy to Coolify

1. Create a new service from **Docker Compose**, point to your repo
2. Set environment variables:
   ```
   AZURE_CLIENT_ID=your_azure_client_id
   TELEGRAM_BOT_TOKEN=your_bot_token
   TELEGRAM_CHAT_ID=your_chat_id
   DAILY_SUMMARY_TIME=09:00
   ```
3. Deploy — volumes for auth cache and data are handled automatically
4. Send `/login` to the bot to authenticate via device code

## 🛠️ Development

```bash
go build -o todoinfo cmd/cli/main.go
go test ./...
```
