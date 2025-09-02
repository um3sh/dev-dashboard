# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a GitHub Dashboard application built with Wails (Go + React) for managing monorepo microservices and Kubernetes resources. The application provides:

- Repository management for monorepos and Kubernetes resource repositories
- Microservice discovery and build/deployment tracking
- Kubernetes resource management with overlay support
- Background synchronization with GitHub API
- SQLite database for local data storage

## Development Commands

### Prerequisites
- Go 1.23+ installed
- Node.js 18+ and npm installed
- Git configured
- SSH key configured for Git access (for SSH authentication)
- GITHUB_TOKEN environment variable set for API access (optional, for GitHub API features)

### Common Commands

```bash
# Install dependencies
go mod tidy
cd frontend && npm install

# Development - Desktop Application
wails dev
# Note: During development, Wails may offer a browser option, but the app 
# is configured to run as a native desktop application

# Build for production
wails build

# Build for macOS (must be run on macOS)
wails build -platform darwin/universal

# Run the built desktop application
# Linux/Windows:
./build/bin/gh-dashboard
# macOS (.app bundle created when built on macOS):
open ./build/bin/GitHub\ Dashboard.app
# or double-click the .app bundle in Finder

# Create macOS DMG installer (macOS only, after building .app)
# Install create-dmg tool: brew install create-dmg
# Then create DMG:
create-dmg \
  --volname "GitHub Dashboard" \
  --window-pos 200 120 \
  --window-size 600 300 \
  --icon-size 100 \
  --app-drop-link 425 120 \
  "GitHub-Dashboard.dmg" \
  "./build/bin/GitHub Dashboard.app"

# Build frontend only
cd frontend && npm run build

# Run tests (when available)
go test ./...
cd frontend && npm test

# Generate Wails bindings (after adding new backend methods)
wails generate module

# Check Go code formatting
go fmt ./...

# Check for Go vulnerabilities
govulncheck ./...
```

### Environment Setup

Create a `.env` file or set environment variables:
```bash
# Optional: GitHub API token for enhanced features
export GITHUB_TOKEN=your_github_personal_access_token

# Optional: Custom SSH key path (defaults to ~/.ssh/id_rsa, ~/.ssh/id_ed25519, ~/.ssh/id_ecdsa)
export SSH_KEY_PATH=/path/to/your/ssh/key

# Optional: SSH key passphrase if required
export SSH_PASSPHRASE=your_ssh_key_passphrase
```

## Architecture

### Backend (Go)
- **Wails App**: Main application entry point with exposed methods
- **Database Layer**: SQLite with schema management
- **Models**: Repository, Microservice, KubernetesResource, Action models
- **GitHub Client**: API integration for repository discovery and workflow tracking
- **Sync Service**: Background service for periodic GitHub data synchronization

### Frontend (React + Tailwind CSS)
- **Pages**: Dashboard, Repositories, Microservices, KubernetesResources
- **Components**: Reusable UI components with Tailwind styling
- **Router**: React Router for navigation
- **Icons**: Lucide React for consistent iconography

### Key Directories
- `internal/`: Go backend code (models, database, GitHub client, sync service)
- `pkg/types/`: Shared type definitions
- `frontend/src/`: React frontend application
- `frontend/src/components/`: Reusable React components
- `frontend/src/pages/`: Main application pages

### Database Schema
- `repositories`: Stores repository information (monorepo/kubernetes type)
- `microservices`: Services discovered in monorepos
- `kubernetes_resources`: K8s resources found in resource repositories
- `actions`: Build and deployment actions tracked from GitHub workflows

## Key Features

### Repository Management
- Add monorepo and Kubernetes resource repositories via SSH or HTTPS URLs
- Specify custom service name and location for monorepos
- SSH key authentication for private repositories
- Automatic service/resource discovery
- Manual and automatic sync with GitHub

### Microservice Tracking
- Discovers services in `services/` directory of monorepos
- Tracks build and deployment actions
- Shows recent activity and status

### Kubernetes Resources
- Discovers YAML files in common K8s directories
- Tracks deployment PR creation and overlay updates
- Organizes by namespace

### Background Sync
- Periodic GitHub API synchronization
- Workflow run tracking
- Automatic service/resource discovery updates

## Development Workflow

1. **Backend Changes**: Modify Go code in `internal/` or `pkg/`
2. **Frontend Changes**: Modify React components in `frontend/src/`
3. **Database Changes**: Update schema in `internal/database/schema.sql`
4. **API Changes**: Add methods to `app.go` and regenerate bindings with `wails generate`
5. **Testing**: Use `wails dev` for hot reloading during development

## Configuration

### Desktop Application
- Window size: 1200x800 (resizable, min: 800x600, max: 1920x1080)
- Application title: "GitHub Dashboard"
- Native desktop application with system window controls
- Light gray background optimized for the dashboard interface

### Database and API
- SQLite database created at `~/.gh-dashboard/database.db`
- Requires GitHub personal access token for API access
- Background sync service runs every 5 minutes when token is provided

### Development vs Production
- Development (`wails dev`): Hot reloading enabled, may show browser option but runs as desktop app
- Production (`wails build`): Creates native executable in `build/bin/gh-dashboard`
- macOS Production: Creates `.app` bundle in `build/bin/GitHub Dashboard.app` (when built on macOS)
- macOS DMG: Use `create-dmg` tool to create installable `.dmg` file from `.app` bundle
- Cross-compilation: macOS builds must be done on macOS systems (cross-compilation not supported)
- Built application includes all frontend assets and is completely self-contained