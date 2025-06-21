# 📊 ToDo Info CLI

A beautiful command-line tool for analyzing your Microsoft ToDo tasks! Track task ages, identify procrastination patterns, and get insights into your productivity habits.


## 🚀 Quick Start

### 1. Install
```bash
git clone https://github.com/uchr/ToDoInfo.git
cd ToDoInfo
go build -o todoinfo cmd/cli/main.go
```

### 3. Authenticate
```bash
./todoinfo login
```

### 4. Analyze Your Tasks
```bash
./todoinfo stats
```


## 🎯 Task Rottenness Levels

- 😊 **Fresh** (0-2 days) - Recently created, still fresh
- 😏 **Ripe** (3-6 days) - Getting older, should be addressed soon  
- 🥱 **Tired** (7-13 days) - Procrastination setting in
- 🤢 **Zombie** (14+ days) - Seriously overdue, needs immediate attention

## ⚙️ Configuration

### Option 1: .env File (Recommended)
```bash
# Create .env file in project root
echo "AZURE_CLIENT_ID=your_azure_client_id" > .env
./todoinfo stats
```

### Option 2: Command Line Flag
```bash
./todoinfo stats --client-id YOUR_AZURE_CLIENT_ID
```

### Option 3: Environment Variable
```bash
export AZURE_CLIENT_ID=your_azure_client_id
./todoinfo stats
```

### Option 4: Config File
Create `~/.todoinfo.yaml`:
```yaml
client-id: your_azure_client_id
```

## 🛠️ Development

### Setup Azure App Registration
1. Go to [Azure Portal](https://portal.azure.com) → App Registrations
2. Create new registration with redirect URI: `http://localhost:8080`
3. Note the Application (client) ID
4. Grant `Tasks.ReadWrite` and `User.Read` permissions

### Build
```bash
go build -o todoinfo cmd/cli/main.go
```

### Test
```bash
go test ./...
```
