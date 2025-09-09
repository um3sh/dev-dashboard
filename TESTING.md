# Testing the Dev Dashboard Desktop Application

## Quick Start

### 1. Run the Built Application (Recommended)
```bash
./build/bin/dev-dashboard
```
This runs the production build as a native desktop application.

### 2. Development Mode
```bash
export PATH=$PATH:$(go env GOPATH)/bin
wails dev
```

**Important Notes about `wails dev`:**
- Wails development mode may show browser-related messages in the terminal
- The actual application will open as a **native desktop window**
- If you see WebKit messages, this is normal - Wails uses WebKit for rendering but in a desktop window
- The application is configured to run as a desktop app, not in a browser

### 3. With GitHub Integration (Optional)
```bash
export GITHUB_TOKEN=your_github_personal_access_token
./build/bin/dev-dashboard
```

## What to Expect

### Desktop Application Features
- **Window**: Native desktop window (1200x800, resizable)
- **Title**: "Dev Dashboard" in the window title bar
- **Controls**: Standard window controls (minimize, maximize, close)
- **Behavior**: Runs like any other desktop application

### Initial State
- Dashboard shows zero repositories, microservices, and Kubernetes resources
- Navigation sidebar with Dashboard, Repositories, Microservices, and Kubernetes sections
- Modern UI with Tailwind CSS styling
- All data will be mock data initially (until repositories are added)

### Adding Repositories
1. Click "Repositories" in the sidebar
2. Click "Add Repository" button
3. Fill in repository details:
   - Name: e.g., "my-monorepo"
   - URL: GitHub repository URL
   - Type: "Monorepo" or "Kubernetes"
   - Description: Optional description

## Troubleshooting

### Database Issues
If you see database-related errors:
- The app will still run but with limited functionality
- Check logs for database path: `~/.dev-dashboard/database.db`
- Delete the database file to reset: `rm -rf ~/.dev-dashboard/`

### WebKit Messages
Messages like "Overriding existing handler for signal 10" are normal WebKit messages and don't indicate problems.

### Not Opening as Desktop App
If the app doesn't open a window:
1. Check the terminal output for error messages
2. Ensure you have required system dependencies
3. Try running the built executable directly: `./build/bin/dev-dashboard`

## System Requirements

- Linux with X11 or Wayland
- WebKit libraries (usually pre-installed)
- For development: Go 1.23+, Node.js 18+, npm