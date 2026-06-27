# Quick Start Guide

This is a condensed version of the setup process. For detailed steps, see [CHECKLIST.md](CHECKLIST.md).

## TL;DR - Fast Path to TrueNAS Deployment

```bash
# 1. Setup local database
psql -h 192.168.100.126 -p 5432 -U postgres -d postgres -f sql/schema.sql

# 2. Configure environment
cp .env.example .env
# Edit .env with your password

# 3. Install dependencies and run
go mod download
go run main.go

# 4. Test at http://localhost:7001
# - Create a recipe
# - Add to pantry
# - Test "What can I cook?"
# - Add to meal plan

# 5. Push to GitHub
git init
git add .
git commit -m "Initial commit"
git branch -M main
git remote add origin https://github.com/YOUR_USERNAME/wareg.git
git push -u origin main

# 6. On GitHub:
# - Enable Packages
# - Enable Pages
# - Check Actions (should auto-run)

# 7. Deploy to TrueNAS SCALE
# - Add Helm catalog from your GitHub Pages
# - Install app with ghcr.io/YOUR_USERNAME/wareg:latest
# - Configure DATABASE_URL
```

## Common Commands

### Development
```bash
# Run application
go run main.go

# Build binary
go build -o wareg

# Format code
go fmt ./...

# Check code
go vet ./...
```

### Database
```bash
# Connect to database
psql -h 192.168.100.126 -p 5432 -U postgres -d postgres

# Run schema
psql -h 192.168.100.126 -p 5432 -U postgres -d postgres -f sql/schema.sql

# Check tables
psql -h 192.168.100.126 -p 5432 -U postgres -d postgres -c "\dt wareg.*"
```

### Git
```bash
# Status
git status

# Add files
git add .

# Commit
git commit -m "message"

# Push
git push
```

### Docker
```bash
# Build
docker build -t wareg:latest .

# Run
docker run -d -p 7001:7001 -e DATABASE_URL="..." wareg:latest

# Stop
docker stop wareg-app
```

## Environment Variables

| Variable | Default | Description |
|----------|----------|-------------|
| `DATABASE_URL` | Required | PostgreSQL connection string |
| `PORT` | `7001` | Server port |

Example DATABASE_URL:
```
postgres://postgres:PASSWORD@192.168.100.126:5432/postgres?search_path=wareg
```

## URLs After GitHub Setup

- Repository: `https://github.com/YOUR_USERNAME/wareg`
- Docker Image: `ghcr.io/YOUR_USERNAME/wareg:latest`
- Helm Chart: `https://YOUR_USERNAME.github.io/wareg/`

## Need Help?

- Detailed checklist: [CHECKLIST.md](CHECKLIST.md)
- TrueNAS deployment: [DEPLOYMENT.md](DEPLOYMENT.md)
- Contributing: [CONTRIBUTING.md](CONTRIBUTING.md)
- Full docs: [README.md](README.md)

## Troubleshooting

### Port Already in Use

```bash
# Windows
netstat -ano | findstr :7001

# Linux/Mac
lsof -i :7001
```

### Database Connection Failed

1. Check PostgreSQL is running at `192.168.100.126:5432`
2. Verify credentials in `.env`
3. Test with psql directly

### Go Dependencies Missing

```bash
go mod download
go mod tidy
```

### Application Won't Start

1. Check terminal for errors
2. Verify `.env` exists
3. Check port availability
