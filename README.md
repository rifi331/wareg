# Wareg - Recipe Management and Meal Planning

A lightweight web application for recipe management, meal planning, and smart pantry ("What can I cook?") built with Go, Echo, HTMX, Tailwind CSS, and PostgreSQL.

## Deployment Options

This application supports multiple deployment methods:

1. **Local Development** - Run directly with `go run`
2. **Docker** - Containerized deployment
3. **TrueNAS SCALE** - Kubernetes/Helm deployment (detailed guide available)
4. **Kubernetes** - Standard Helm chart deployment

For detailed TrueNAS SCALE deployment instructions, see [DEPLOYMENT.md](DEPLOYMENT.md).

See the [Deployment](#deployment) section for details.

## Features

- **Recipe CRUD**: Create, read, and delete recipes with nutrition info (calories, protein, carbs, fats), images, and video URLs
- **Media Integration**: Save YouTube or Instagram URLs for video guides
- **Meal Planner**: Calendar view to assign recipes to specific dates
- **Smart Inventory**:
  - Track pantry items with quantities
  - "What can I cook?" engine suggests recipes where you have 80%+ of required ingredients
- **Mobile-friendly UI** built with Tailwind CSS and HTMX for interactivity

## Tech Stack

- **Backend**: Go (1.22+) with Echo framework and pgx (PostgreSQL driver)
- **Frontend**: HTML templates with HTMX and Tailwind CSS
- **Database**: PostgreSQL (remote instance supported)

**Note**: The app loads `.env` automatically via `godotenv.Load()` in `main.go`
(the dependency is already in `go.mod`). You can also export `DATABASE_URL`
and `PORT` directly in your shell.

## Project Structure

```
wareg/
├── sql/
│   ├── schema.sql             # PostgreSQL schema (ingredients carry a default `unit`)
│   └── seed_data.sql          # ~10 recipes + pantry (transactional, FK-safe)
├── frontend/templates/
│   ├── base.html              # Dashboard: tabbed single-page UI (Ingredients/Recipes/Pantry/Meal Plan/Matches)
│   └── recipes.html           # Standalone recipes page (own working nav)
├── static/                    # Static assets (favicon, uploads/.gitkeep) — copied into the Docker image
├── main.go                    # Go backend: routes, handlers, escaped render helpers
├── go.mod / go.sum            # Go module definition
├── Dockerfile                 # Multi-stage build (copies binary + frontend + static)
├── docker-compose.yml         # App + Postgres for local dev
├── helm/                      # Helm chart (deployment, service, helpers)
└── .github/workflows/         # docker-build.yml, helm-release.yml
```

## Setup Instructions

### Prerequisites

- Go 1.22 or higher
- PostgreSQL (running instance)
- Basic knowledge of SQL and Go

### 1. Database Setup

Connect to your PostgreSQL instance and run the schema:

```bash
psql -U your_user -d your_database -f sql/schema.sql
```

Or run it manually:

```sql
-- Tables will be created automatically
-- recipes, ingredients, recipe_ingredients, pantry, meal_plan
```

### 2. Configure Database Connection

Option 1: Copy the example environment file and update the password:
```bash
cp .env.example .env
# Edit .env and update your_password
```

Option 2: Set up `DATABASE_URL` environment variable:

```bash
# On Linux/Mac
export DATABASE_URL="postgres://postgres:your_password@your-db-host:5432/postgres?search_path=wareg"

# On Windows (PowerShell)
$env:DATABASE_URL="postgres://postgres:your_password@your-db-host:5432/postgres?search_path=wareg"

# On Windows (Command Prompt)
set DATABASE_URL=postgres://postgres:your_password@your-db-host:5432/postgres?search_path=wareg"
```

**Note**: The application uses the `wareg` schema in PostgreSQL. The schema will be created automatically when you run the schema.sql file.

### 3. Install Dependencies

```bash
go mod download
```

### 4. Run the Application

```bash
go run main.go
```

The server will start on port 7001 by default. You can change the port using the `PORT` environment variable:

```bash
export PORT=3000
go run main.go
```

### 5. Access the Application

Open your browser and navigate to: `http://localhost:7001`

## API Endpoints

All `/api/*` HTML endpoints return HTMX fragments; the `*/options` and `/api/recipes/:id`
endpoints return clean JSON for populating dropdowns and programmatic access.

### System
- `GET /healthz` - Liveness/readiness probe (returns `ok`, HTTP 200)

### Pages
- `GET /` - Dashboard (tabbed single-page UI)
- `GET /recipes` - Recipes page
- `GET /recipes/:id` - Recipe detail page (server-rendered HTML)

### Ingredients
- `GET /api/ingredients` - List (HTML rows)
- `POST /api/ingredients` - Create (fields: `name`, `unit`)
- `DELETE /api/ingredients/:id` - Delete
- `GET /api/ingredients/options` - List (JSON: `[{id,name,unit}]`)

### Recipes
- `GET /api/recipes` - List (HTML cards)
- `POST /api/recipes` - Create (title, description, image_url, video_url, `nutrition.*`, `ingredients.N.*`, `steps.N.*`) — runs in a single transaction
- `DELETE /api/recipes/:id` - Delete (responds with `HX-Redirect: /recipes`)
- `GET /api/recipes/options` - List (JSON: `[{id,title}]`)
- `GET /api/recipes/:id` - **Get recipe as JSON** (with nested ingredients + steps)

### Pantry
- `GET /api/pantry` - List pantry items
- `POST /api/pantry` - Add (UPSERT on `ingredient_id`; re-adding **adds** to the quantity)
- `DELETE /api/pantry/:id` - Remove
- `GET /api/pantry/matches` - "What can I cook?" (>=80% ingredient match, sorted desc)

### Meal Plan
- `GET /api/meal-plan` - List meal plan
- `POST /api/meal-plan` - Add meal (fields: `recipe_id`, `plan_date`)
- `DELETE /api/meal-plan/:id` - Remove

## Database Schema

All tables live in the `wareg` schema (see `sql/schema.sql`).

- `ingredients` - catalog: `id, name (unique), unit (default), created_at`
- `recipes` - `id, title, description, instructions, image_url, video_url, nutrition_json (JSONB), created_at, updated_at`
- `recipe_ingredients` - `id, recipe_id, ingredient_id, quantity, unit` — `UNIQUE(recipe_id, ingredient_id)`
- `recipe_steps` - `id, recipe_id, step_number, instruction` — `UNIQUE(recipe_id, step_number)`
- `pantry` - `id, ingredient_id, current_quantity, unit` — `UNIQUE(ingredient_id)`
- `meal_plan` - `id, recipe_id, plan_date, created_at` — `UNIQUE(plan_date, recipe_id)`

Each ingredient has a default unit; recipe and pantry rows specify their own per-row unit. Child rows use `ON DELETE CASCADE`.

## Sample Data

A ready-to-use dataset (~10 Indonesian/Thai/European recipes plus a starter pantry) is
provided in `sql/seed_data.sql`. Load it after the schema:

```bash
psql "<conn>" -f sql/seed_data.sql
```

The seed runs in a single transaction, inserts **every** referenced ingredient first
(including `Mussel` and `Saffron`), and reseeds sequences — so it loads cleanly with no
foreign-key violations. To add data by hand:

```sql
SET search_path TO wareg;

INSERT INTO ingredients (name, unit) VALUES ('Chicken', 'g'), ('Rice', 'g');

INSERT INTO recipes (title, description, nutrition_json)
VALUES ('Soto Betawi', 'Traditional Indonesian soup', '{"calories":350,"protein":25,"carbs":30,"fats":15}');

-- quantity + unit per row
INSERT INTO recipe_ingredients (recipe_id, ingredient_id, quantity, unit) VALUES
(1, 1, 500, 'g'),
(1, 2, 200, 'g');

INSERT INTO pantry (ingredient_id, current_quantity, unit) VALUES
(1, 500, 'g'),
(2, 1000, 'g');

INSERT INTO meal_plan (recipe_id, plan_date) VALUES (1, '2026-07-01');
```

## Development

### Code Style

- Clean, minimal code with standard library
- No comments in production code
- Mobile-first responsive design
- RESTful API conventions

### Adding New Features

1. Update database schema in `sql/schema.sql`
2. Add data models in `main.go`
3. Implement API handlers
4. Create/modify HTML templates
5. Test the feature

---

# Deployment

## Option 1: Docker / Docker Compose

### Using Docker Compose (Recommended)

1. Copy `.env.example` to `.env` and configure:

```bash
cp .env.example .env
# Edit .env with your database credentials
```

2. Start the application:

```bash
docker-compose up -d
```

3. Access at: `http://localhost:7001`

### Using Docker directly

1. Build the image:

```bash
docker build -t wareg:latest .
```

2. Run the container:

```bash
docker run -d \
  --name wareg-app \
  -p 7001:7001 \
  -e DATABASE_URL="postgres://user:password@host:5432/db?search_path=wareg" \
  -e PORT="7001" \
  wareg:latest
```

## Option 2: TrueNAS SCALE (Kubernetes)

### Prerequisites

- TrueNAS SCALE with Apps (Kubernetes) enabled
- PostgreSQL database (can be a separate app on TrueNAS)
- Container registry (GitHub Container Registry recommended)

### Step 1: Build and Push Container Image

1. Create a GitHub Container Registry (GHCR) account
2. Push your code to GitHub
3. Enable GHCR in your repository settings
4. Build and push the image:

```bash
# Login to GHCR
echo $GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin

# Build and tag
docker build -t ghcr.io/yourusername/wareg:latest .

# Push
docker push ghcr.io/yourusername/wareg:latest
```

### Step 2: Publish Helm Chart to GitHub Pages

1. Create a `gh-pages` branch or use GitHub Actions to auto-publish
2. Package the Helm chart:

```bash
helm package helm --destination .
```

3. Upload the `.tgz` file to your repository's GitHub Pages

### Step 3: Deploy on TrueNAS SCALE

1. In TrueNAS SCALE, go to **Apps** > **Settings**
2. Click **Add Catalog**
   - Name: `wareg`
   - URL: `https://your-username.github.io/wareg/` (your GitHub Pages URL)
   - Train: Stable

3. Go to **Apps** > **Install Application**
4. Configure:
   - Application Name: `wareg-cooking-app`
   - Catalog: `wareg`
   - Chart: `wareg-cooking-app`
   - Version: `latest`

5. In **Configuration** section:
   - **Image Settings**:
     - Repository: `ghcr.io/yourusername/wareg`
     - Tag: `latest`
   - **Database Configuration**:
     - URL: `postgres://youruser:yourpass@your-db-host:5432/postgres?search_path=wareg`
   - **Service Configuration**:
     - Type: `NodePort` or `LoadBalancer`
     - Port: `7001`
   - **Ingress** (optional, for domain access):
     - Enabled: `true`
     - Host: `cooking-app.yourdomain.com`

6. Click **Install**

### Step 4: Access the Application

- **NodePort**: Find the NodePort in TrueNAS SCALE Apps section, access via `http://node-ip:port`
- **LoadBalancer**: Access via the external IP assigned
- **Ingress**: Access via `https://cooking-app.yourdomain.com`

## Option 3: Standard Kubernetes

### Using Helm

1. Clone this repository:

```bash
git clone https://github.com/your-username/wareg.git
cd wareg/helm
```

2. Create a `values.yaml`:

```yaml
image:
  repository: ghcr.io/yourusername/wareg
  tag: "latest"

database:
  url: "postgres://user:password@db-host:5432/postgres?search_path=wareg"

service:
  type: LoadBalancer
  port: 7001
```

3. Install:

```bash
helm install wareg-cooking-app ./ -f values.yaml
```

4. Get the service URL:

```bash
kubectl get svc wareg-cooking-app
```

## Environment Variables

| Variable | Required | Description | Example |
|----------|----------|-------------|---------|
| `DATABASE_URL` | Yes | PostgreSQL connection URL | `postgres://user:pass@host:5432/db?search_path=wareg` |
| `PORT` | No | Server port (default: 7001) | `7001` |

## Security Considerations

1. **Never commit** `.env` file to version control
2. Use **secrets management** for production:
   - Kubernetes Secrets
   - External secret stores (Vault, AWS Secrets Manager)
3. Enable **HTTPS** in production (use Ingress with TLS)
4. Use **read-only** database user where possible
5. Configure **resource limits** to prevent resource exhaustion
6. Enable **network policies** to restrict pod communication

## Troubleshooting

### Database Connection Issues

- Check `DATABASE_URL` format
- Verify database server is reachable
- Ensure `wareg` schema exists

### Container Won't Start

- Check logs: `docker logs wareg-app` or `kubectl logs <pod-name>`
- Verify environment variables are set
- Check port conflicts

### TrueNAS SCALE Issues

- Verify catalog URL is accessible
- Check Helm chart values are properly formatted
- Review pod logs in Apps > Application > Logs

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Additional Documentation

- [CHECKLIST.md](CHECKLIST.md) - Step-by-step checklist for local testing and GitHub upload
- [DEPLOYMENT.md](DEPLOYMENT.md) - Complete TrueNAS SCALE deployment guide
- [CONTRIBUTING.md](CONTRIBUTING.md) - Guidelines for contributing
- [helm/README.md](helm/README.md) - Helm chart documentation

## Quick Start

For first-time setup and testing, follow the [CHECKLIST.md](CHECKLIST.md) guide which covers:
1. Local development and debugging
2. Preparing code for GitHub
3. Uploading to GitHub
4. Configuring for TrueNAS deployment
