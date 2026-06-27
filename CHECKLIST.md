# Local Development & GitHub Upload Checklist

## Part 1: Local Debugging & Testing

### Step 1: Database Setup
- [ ] Ensure PostgreSQL is running at `your-db-host:5432`
- [ ] Connect to database and verify connection:
  ```bash
  psql -h your-db-host -p 5432 -U postgres -d postgres
  ```
- [ ] Create `wareg` schema if not exists:
  ```sql
  CREATE SCHEMA IF NOT EXISTS wareg;
  ```
- [ ] Run the schema file:
  ```bash
  psql -h your-db-host -p 5432 -U postgres -d postgres -f sql/schema.sql
  ```
- [ ] Verify tables were created:
  ```sql
  \dt wareg.*
  ```

### Step 2: Environment Configuration
- [ ] Copy environment template:
  ```bash
  cp .env.example .env
  ```
- [ ] Edit `.env` file with your database credentials:
  ```
  DATABASE_URL=postgres://postgres:YOUR_PASSWORD@your-db-host:5432/postgres?search_path=wareg
  PORT=7001
  ```
- [ ] Verify the format is correct:
  - Protocol: `postgres://`
  - Username: `postgres`
  - Password: Your actual password
  - Host: `your-db-host`
  - Port: `5432`
  - Database: `postgres`
  - Schema: `search_path=wareg`

### Step 3: Install Go Dependencies
- [ ] Install Go 1.22+ if not installed
- [ ] Verify Go version:
  ```bash
  go version
  ```
- [ ] Download dependencies:
  ```bash
  go mod download
  ```
- [ ] Verify no errors in go.mod:
  ```bash
  go mod tidy
  ```

### Step 4: Test Database Connection
- [ ] Run a quick connection test (optional):
  ```bash
  # Create a test file test_db.go
  cat > test_db.go << 'EOF'
  package main
  import (
      "context"
      "log"
      "os"
      "github.com/jackc/pgx/v5/pgxpool"
  )
  func main() {
      dbURL := os.Getenv("DATABASE_URL")
      if dbURL == "" {
          dbURL = "postgres://postgres:YOUR_PASSWORD@your-db-host:5432/postgres?search_path=wareg"
      }
      pool, err := pgxpool.New(context.Background(), dbURL)
      if err != nil {
          log.Fatalf("Failed to connect: %v", err)
      }
      defer pool.Close()
      log.Println("Database connection successful!")
  }
  EOF
  go run test_db.go
  rm test_db.go
  ```

### Step 5: Run the Application Locally
- [ ] Start the application:
  ```bash
  go run main.go
  ```
- [ ] Verify no startup errors in terminal
- [ ] Check that port 7001 is not already in use:
  ```bash
  # On Windows PowerShell
  Get-NetTCPConnection -LocalPort 7001 -ErrorAction SilentlyContinue

  # On Windows Command Prompt
  netstat -an | findstr :7001

  # On Linux/Mac
  lsof -i :7001 || echo "Port 7001 is free"
  ```
- [ ] If port is occupied, find and close the process, or change PORT in `.env`

### Step 6: Access the Application
- [ ] Open browser to: `http://localhost:7001`
- [ ] Verify the page loads correctly
- [ ] Check all navigation tabs work:
  - [ ] Recipes
  - [ ] Pantry
  - [ ] Meal Plan
  - [ ] What Can I Cook?

### Step 7: Test Features

#### Create Ingredients
- [ ] Navigate to "Pantry" tab
- [ ] Try to add a new ingredient (if UI allows)
- [ ] Verify ingredient appears in pantry list

#### Create a Recipe
- [ ] Navigate to "Recipes" tab
- [ ] Create a test recipe:
  - Title: "Test Recipe"
  - Description: "This is a test"
  - Image URL: Leave empty or use a placeholder
  - Video URL: Leave empty
  - Nutrition: Fill with test values
  - Ingredients: Skip if UI requires existing ingredients
- [ ] Click "Add Recipe"
- [ ] Verify recipe appears in list
- [ ] Check browser console for errors (F12)

#### Add to Pantry
- [ ] Navigate to "Pantry" tab
- [ ] Add some ingredients with quantities:
  - Example: "Chicken breast", quantity: 1.0
  - Example: "Rice", quantity: 2.0
- [ ] Verify items appear in pantry list

#### Test "What Can I Cook?"
- [ ] Navigate to "What Can I Cook?" tab
- [ ] Wait for matches to load
- [ ] Verify recipes with matching ingredients appear
- [ ] Check match percentages are displayed

#### Test Meal Plan
- [ ] Navigate to "Meal Plan" tab
- [ ] Add a meal for a future date:
  - Select a recipe
  - Choose a date (e.g., tomorrow)
- [ ] Click "Add"
- [ ] Verify meal appears in meal plan
- [ ] Try deleting the meal
- [ ] Verify meal is removed

### Step 8: Test Database Persistence
- [ ] Stop the application (Ctrl+C)
- [ ] Restart the application:
  ```bash
  go run main.go
  ```
- [ ] Refresh browser page
- [ ] Verify all data (recipes, pantry, meal plan) is still there
- [ ] Try creating and deleting records to verify CRUD works

### Step 9: Check Logs and Errors
- [ ] Monitor terminal output for:
  - [ ] Database connection errors
  - [ ] SQL query errors
  - [ ] HTTP request errors
  - [ ] Template rendering errors
- [ ] Check browser console (F12) for:
  - [ ] JavaScript errors
  - [ ] HTMX request failures (check Network tab)
  - [ ] 404 or 500 errors

### Step 10: Test Responsive Design
- [ ] Open DevTools (F12) and toggle device toolbar
- [ ] Test on mobile view:
  - [ ] All content is readable
  - [ ] Buttons are clickable
  - [ ] Forms are usable
- [ ] Test on tablet view
- [ ] Test on desktop view

### Step 11: Cleanup Test Data (Optional)
- [ ] Delete test recipes:
  ```sql
  psql -h your-db-host -p 5432 -U postgres -d postgres -c "DELETE FROM wareg.recipes WHERE title = 'Test Recipe';"
  ```
- [ ] Clear test pantry items:
  ```sql
  psql -h your-db-host -p 5432 -U postgres -d postgres -c "DELETE FROM wareg.pantry;"
  ```

---

## Part 2: Prepare for GitHub Upload

### Step 12: Verify Code Quality
- [ ] Review `main.go` for any hardcoded credentials
- [ ] Ensure `.env` file is NOT in code (should be ignored)
- [ ] Check `.gitignore` includes:
  - [ ] `.env`
  - [ ] `.env.local`
  - [ ] `*.db`
  - [ ] `*.sqlite`
  - [ ] `*.exe` and compiled binaries
- [ ] Run `go vet` for code issues:
  ```bash
  go vet ./...
  ```
- [ ] Run `go fmt` for formatting:
  ```bash
  go fmt ./...
  ```

### Step 13: Test Docker Build (Optional but Recommended)
- [ ] Build Docker image locally:
  ```bash
  docker build -t wareg:test .
  ```
- [ ] Test running in Docker:
  ```bash
  docker run -d \
    --name wareg-test \
    -p 7002:7001 \
    -e DATABASE_URL="postgres://postgres:YOUR_PASSWORD@host.docker.internal:5432/postgres?search_path=wareg" \
    wareg:test
  ```
- [ ] Access at `http://localhost:7002`
- [ ] Stop and remove test container:
  ```bash
  docker stop wareg-test
  docker rm wareg-test
  docker rmi wareg:test
  ```
- [ ] If you encounter Docker issues, try using host IP instead of `host.docker.internal`

### Step 14: Verify All Files Are Ready
- [ ] Check all required files exist:
  - [ ] `main.go` (Go backend)
  - [ ] `go.mod` and `go.sum` (Go modules)
  - [ ] `sql/schema.sql` (Database schema — `ingredients` has a `unit` column)
  - [ ] `sql/seed_data.sql` (transactional seed data)
  - [ ] `frontend/templates/base.html` (Dashboard UI)
  - [ ] `frontend/templates/recipes.html` (Recipes page)
  - [ ] `static/` (favicon + `uploads/.gitkeep`)
  - [ ] `Dockerfile` (Container build — copies binary + frontend + static)
  - [ ] `docker-compose.yml` (Local deployment — app + Postgres)
  - [ ] `.env.example` (Environment template)
  - [ ] `README.md` (Documentation)
  - [ ] `DEPLOYMENT.md` (Deployment guide)
  - [ ] `CONTRIBUTING.md` (Contributing guide)
  - [ ] `LICENSE` (License)
  - [ ] `.gitignore` (Git ignore rules)
- [ ] Check Helm chart files:
  - [ ] `helm/Chart.yaml`
  - [ ] `helm/values.yaml`
  - [ ] `helm/templates/deployment.yaml`
  - [ ] `helm/templates/service.yaml`
  - [ ] `helm/templates/_helpers.tpl`
  - [ ] `helm/README.md`
- [ ] Check CI/CD files:
  - [ ] `.github/workflows/docker-build.yml`
  - [ ] `.github/workflows/helm-release.yml`

### Step 15: Final Verification
- [ ] Make sure `.env` is NOT committed (check `.gitignore`)
- [ ] No temporary or test files should be committed:
  ```bash
  # Check for temp files
  git status
  ```
- [ ] All credentials should be placeholders:
  - [ ] No real passwords in `main.go`
  - [ ] No real URLs in code files
  - [ ] Only `.env.example` contains placeholder values

---

## Part 3: Upload to GitHub

### Step 16: Initialize Git Repository
- [ ] Navigate to project directory:
  ```bash
  cd /path/to/wareg
  ```
- [ ] Initialize git:
  ```bash
  git init
  ```

### Step 17: Create GitHub Repository
- [ ] Log in to GitHub
- [ ] Click "New repository"
- [ ] Name it: `wareg` (or your preferred name)
- [ ] Description: "Recipe Management and Meal Planning App"
- [ ] Choose: Public or Private
- [ ] DO NOT initialize with:
  - [ ] README
  - [ ] .gitignore
  - [ ] License
  (We already have these files)

### Step 18: Connect Local Repository to GitHub
- [ ] Add remote:
  ```bash
  git remote add origin https://github.com/YOUR_USERNAME/wareg.git
  ```
- [ ] Verify remote was added:
  ```bash
  git remote -v
  ```

### Step 19: Create Initial Commit
- [ ] Stage all files:
  ```bash
  git add .
  ```
- [ ] Review staged files:
  ```bash
  git status
  ```
- [ ] Create commit:
  ```bash
  git commit -m "Initial commit: Cooking app with Go/Echo/HTMX"
  ```

### Step 20: Push to GitHub
- [ ] Set main branch:
  ```bash
  git branch -M main
  ```
- [ ] Push to GitHub:
  ```bash
  git push -u origin main
  ```
- [ ] Verify files are on GitHub:
  - [ ] Go to your repository on GitHub
  - [ ] Check all files are present
  - [ ] Verify no `.env` file was uploaded (should be ignored)

---

## Part 4: Post-GitHub Configuration

### Step 21: Enable GitHub Container Registry (GHCR)
- [ ] Go to repository on GitHub
- [ ] Click "Settings" tab
- [ ] Scroll to "Features" section
- [ ] Confirm "Packages" is enabled

### Step 22: Enable GitHub Pages (for Helm Chart)
- [ ] Go to repository "Settings"
- [ ] Click "Pages" on left sidebar
- [ ] Source: Select "GitHub Actions"
- [ ] Click "Save"
- [ ] Wait for GitHub Pages to be ready (may take a few minutes)

### Step 23: Trigger GitHub Actions
- [ ] Go to "Actions" tab on GitHub
- [ ] You should see 2 workflows:
  - [ ] Build and Push Docker Image
  - [ ] Release Helm Chart
- [ ] If workflows didn't run automatically, trigger manually:
  - [ ] Click "Build and Push Docker Image"
  - [ ] Click "Run workflow"
  - [ ] Click "Release Helm Chart"
  - [ ] Click "Run workflow"

### Step 24: Verify GitHub Actions Success
- [ ] Check "Build and Push Docker Image" workflow:
  - [ ] Status: ✅ Completed (green checkmark)
  - [ ] Review logs for errors if failed
- [ ] Check "Release Helm Chart" workflow:
  - [ ] Status: ✅ Completed
  - [ ] Review logs for errors if failed

### Step 25: Verify Docker Image
- [ ] Go to repository "Packages" tab (or click link from settings)
- [ ] Look for `wareg` package
- [ ] Verify image tag `latest` exists
- [ ] Note the image URL: `ghcr.io/YOUR_USERNAME/wareg:latest`

### Step 26: Verify Helm Chart Published
- [ ] GitHub Pages URL should be: `https://YOUR_USERNAME.github.io/wareg/`
- [ ] Visit the URL in browser
- [ ] You should see:
  - [ ] `wareg-cooking-app-*.tgz` file
  - [ ] `index.yaml` file

---

## Part 5: Ready for TrueNAS Deployment

### Step 27: Prepare TrueNAS SCALE
- [ ] TrueNAS SCALE is running and accessible
- [ **Apps** feature is enabled
- [ ] PostgreSQL is either:
  - [ ] External database at `your-db-host:5432`, OR
  - [ ] Deployed as TrueNAS app with known connection string

### Step 28: Note Database Connection String for TrueNAS
- [ ] For external PostgreSQL (your current setup):
  ```
  postgres://postgres:YOUR_PASSWORD@your-db-host:5432/postgres?search_path=wareg
  ```
- [ ] For TrueNAS PostgreSQL app:
  ```
  postgres://postgres:YOUR_PASSWORD@wareg-db-postgresql:5432/postgres?search_path=wareg
  ```
  (Replace `wareg-db` with your PostgreSQL app name)

### Step 29: Follow DEPLOYMENT.md for TrueNAS Setup
- [ ] Read through `DEPLOYMENT.md`
- [ ] Add Helm chart catalog in TrueNAS
- [ ] Deploy application using `ghcr.io/YOUR_USERNAME/wareg:latest`
- [ ] Configure with correct DATABASE_URL
- [ ] Access application via NodePort or Ingress

---

## Troubleshooting Checklist

### Can't Connect to Database
- [ ] PostgreSQL is running
- [ ] Host IP is correct (your-db-host)
- [ ] Port is correct (5432)
- [ ] Username/password is correct
- [ ] Database name is correct (postgres)
- [ ] Schema exists (wareg)
- [ ] Test with `psql` directly

### Application Won't Start
- [ ] Port 7001 is not in use
- [ ] Go dependencies are installed (`go mod download`)
- [ ] DATABASE_URL is set in environment or `.env` file
- [ ] Check terminal error messages

### GitHub Actions Failed
- [ ] Repository has "Packages" enabled
- [ ] Repository has "Pages" enabled
- [ ] GitHub Token has correct permissions
- [ ] Review workflow logs for specific errors

### Docker Build Fails
- [ ] Dockerfile syntax is correct
- [ ] All required files exist in repository
- [ ] Go modules are properly defined in `go.mod`

### TrueNAS Deployment Issues
- [ ] Helm chart catalog URL is correct
- [ ] Image repository URL is correct (`ghcr.io/YOUR_USERNAME/wareg`)
- [ ] DATABASE_URL is properly formatted
- [ ] Check TrueNAS SCALE logs

---

## Quick Reference Commands

```bash
# Local development
go run main.go

# Build Go binary
go build -o wareg

# Format code
go fmt ./...

# Vet code
go vet ./...

# Download dependencies
go mod download

# Git operations
git init
git add .
git commit -m "message"
git push -u origin main

# Docker build
docker build -t wareg:latest .

# Docker run
docker run -d -p 7001:7001 -e DATABASE_URL="..." wareg:latest

# Database connection test
psql -h your-db-host -p 5432 -U postgres -d postgres
```

---

## Success Criteria

You're ready to deploy to TrueNAS when:

✅ Application runs locally on http://localhost:7001
✅ All features work (Create recipe, Add to pantry, Meal plan, Matches)
✅ Database persists data across restarts
✅ No hardcoded credentials in code
✅ `.gitignore` properly configured
✅ All files pushed to GitHub
✅ GitHub Actions workflows completed successfully
✅ Docker image available in GitHub Packages
✅ Helm chart published to GitHub Pages

---

## Notes

- **Save your progress**: Check off items as you complete them
- **Take your time**: Thoroughly test each feature before moving to next step
- **Document issues**: If something fails, write down the error message and what you tried
- **Ask for help**: Refer to `DEPLOYMENT.md` or create GitHub issues if stuck
