# Prompt: Rebuild the "Wareg" Cooking App (Recipe Management + Meal Planning + Smart Pantry)

You are an expert full-stack engineer. Rebuild a lightweight, mobile-friendly web application called **Wareg** (a cooking app) from scratch. The original app exists but contains several defects; this prompt is a clean-room specification that reproduces all intended features while fixing every known bug. Follow this spec precisely.

---

## 1. Overview

A single-binary web app for:

- Managing a **catalog of ingredients**.
- **Recipe CRUD** with title, description, image URL, video URL (YouTube/Instagram), nutrition facts (calories, protein, carbs, fats), an **ordered list of ingredients with quantities + units**, and an **ordered list of cooking steps**.
- A **Pantry** that tracks how much of each ingredient you currently have.
- A **"What can I cook?"** matching engine that compares the pantry against every recipe's ingredient list and suggests recipes you can mostly make (>= 80% ingredient match), listing which ingredients are missing.
- A **Meal Planner** that assigns a recipe to a date.
- A clean, responsive UI with HTMX-style interactivity (no heavy SPA framework).

---

## 2. Tech Stack (non-negotiable)

- **Backend:** Go 1.22+, Echo v4 (`github.com/labstack/echo/v4`), `github.com/jackc/pgx/v5/pgxpool` for PostgreSQL.
- **Env loading:** `github.com/joho/godotenv`.
- **Frontend:** Server-rendered HTML + HTMX + Tailwind CSS (Tailwind via CDN is acceptable for dev). Mobile-first, responsive.
- **Database:** PostgreSQL 14+. All tables live in a dedicated schema named `wareg`.
- **Deployment artifacts:** `Dockerfile` (multi-stage), `docker-compose.yml`, and a Helm chart for Kubernetes.

---

## 3. Database Schema (corrected — use exactly this)

Put this in `sql/schema.sql`. It fixes the original's missing `unit` on ingredients and aligns with the Go code. Use `ON DELETE CASCADE` for child rows. Do **not** mix `unit` semantics across tables: each ingredient has a default unit; recipe/pantry rows specify their own unit per row.

```sql
CREATE SCHEMA IF NOT EXISTS wareg;
SET search_path TO wareg;

DROP TABLE IF EXISTS meal_plan CASCADE;
DROP TABLE IF EXISTS pantry CASCADE;
DROP TABLE IF EXISTS recipe_steps CASCADE;
DROP TABLE IF EXISTS recipe_ingredients CASCADE;
DROP TABLE IF EXISTS recipes CASCADE;
DROP TABLE IF EXISTS ingredients CASCADE;

CREATE TABLE ingredients (
    id         SERIAL PRIMARY KEY,
    name       VARCHAR(255) NOT NULL UNIQUE,
    unit       VARCHAR(50)  NOT NULL DEFAULT '',        -- default unit (e.g. g, ml, pc)
    created_at TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE recipes (
    id             SERIAL PRIMARY KEY,
    title          VARCHAR(255) NOT NULL,
    description    TEXT,
    instructions   TEXT,
    image_url      TEXT,
    video_url      TEXT,
    nutrition_json JSONB,
    created_at     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE recipe_ingredients (
    id            SERIAL PRIMARY KEY,
    recipe_id     INTEGER       NOT NULL REFERENCES recipes(id)    ON DELETE CASCADE,
    ingredient_id INTEGER       NOT NULL REFERENCES ingredients(id) ON DELETE CASCADE,
    quantity      NUMERIC(10,2) NOT NULL,
    unit          VARCHAR(50)   NOT NULL,
    UNIQUE (recipe_id, ingredient_id)
);

CREATE TABLE recipe_steps (
    id          SERIAL PRIMARY KEY,
    recipe_id   INTEGER     NOT NULL REFERENCES recipes(id) ON DELETE CASCADE,
    step_number INTEGER     NOT NULL,
    instruction TEXT        NOT NULL,
    UNIQUE (recipe_id, step_number)
);

CREATE TABLE pantry (
    id              SERIAL PRIMARY KEY,
    ingredient_id   INTEGER       NOT NULL REFERENCES ingredients(id) ON DELETE CASCADE,
    current_quantity NUMERIC(10,2) NOT NULL DEFAULT 0,
    unit            VARCHAR(50)   NOT NULL,
    UNIQUE (ingredient_id)
);

CREATE TABLE meal_plan (
    id        SERIAL PRIMARY KEY,
    recipe_id INTEGER NOT NULL REFERENCES recipes(id) ON DELETE CASCADE,
    plan_date DATE    NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (plan_date, recipe_id)
);

CREATE INDEX idx_recipe_ingredients_recipe     ON recipe_ingredients(recipe_id);
CREATE INDEX idx_recipe_ingredients_ingredient ON recipe_ingredients(ingredient_id);
CREATE INDEX idx_recipe_steps_recipe           ON recipe_steps(recipe_id);
CREATE INDEX idx_pantry_ingredient             ON pantry(ingredient_id);
CREATE INDEX idx_meal_plan_date                ON meal_plan(plan_date);
```

Provide `sql/seed_data.sql` with ~10 sample recipes (Indonesian/Thai/European) and a pantry. **IMPORTANT bug to avoid:** every `recipe_ingredients` row must reference an ingredient that is actually inserted. The original seed referenced non-existent ingredients (`Mussel`, `Saffron`) which violated the FK and aborted the whole seed. Either insert all referenced ingredients first, or insert every needed ingredient (including `Mussel`, `Saffron`, etc.).

---

## 4. Application Behavior & API

Config from env: `DATABASE_URL` (required, format `postgres://user:pass@host:5432/db?search_path=wareg`) and `PORT` (default `7001`). Load `.env` via godotenv.

Routes:

Pages
- `GET /`              -> main dashboard (single-page tabbed UI)
- `GET /recipes`       -> recipes page
- `GET /recipes/:id`   -> recipe detail page (server-rendered HTML)

Ingredients
- `GET    /api/ingredients`
- `POST   /api/ingredients`
- `DELETE /api/ingredients/:id`

Recipes
- `GET    /api/recipes`
- `POST   /api/recipes`
- `DELETE /api/recipes/:id`
- `GET    /api/recipes/:id`  -> **return JSON** (see bug fix below)

Pantry
- `GET    /api/pantry`
- `POST   /api/pantry`
- `DELETE /api/pantry/:id`
- `GET    /api/pantry/matches`  -> "What can I cook?" results

Meal Plan
- `GET    /api/meal-plan`
- `POST   /api/meal-plan`
- `DELETE /api/meal-plan/:id`

Static files: `GET /static/*` from a `static/` directory (create the dir and a `static/uploads/.gitkeep`).

---

## 5. CRITICAL: Bugs to avoid (these existed in the original)

These are real defects found in the original codebase. Your implementation MUST not reproduce them.

### 5.1 SQL syntax error in meal_plan insert (always 500s)
Original had an unbalanced parenthesis:
```go
"INSERT INTO meal_plan (recipe_id, plan_date) VALUES ($1, $2) RETURNING id, recipe_id, plan_date)"
//                                                                                              ^ stray ')'
```
Fix: `RETURNING id, recipe_id, plan_date` (no trailing paren). Adding a meal plan must work.

### 5.2 Row duplication from two LEFT JOINs + json_agg
The original `listRecipes`/`getRecipe`/`findMatchingRecipes` query joins BOTH `recipe_ingredients` and `recipe_steps` in one query and does two `json_agg` calls. Because the joins multiply (N ingredients x M steps rows), `json_agg` returns each ingredient repeated M times and each step repeated N times. Also `json_agg` returns `[null]` when a recipe has zero ingredients.

Fix: fetch ingredients and steps in **separate queries** (or use `jsonb_agg(...) FILTER (WHERE ...)` with proper null filtering, and aggregate each child table independently via subqueries/LATERAL). Recipes with no ingredients must read as an empty list, not `[null]`.

### 5.3 Echo form binding: structs only had `json` tags
`PantryItemCreate`/`MealPlanCreate`/`PantryItem`/etc. only had `json:"..."` tags, but HTML forms send `ingredient_id`, `current_quantity`, `recipe_id`, `plan_date`. Echo's binder matches the `form` tag (then field name), not the `json` tag, so values failed to bind (zero values -> FK violations).

Fix: add `form:"..."` tags (keep `json:"..."` too) on every struct bound from request data, e.g.:
```go
type PantryItemCreate struct {
    IngredientID    int     `json:"ingredient_id" form:"ingredient_id"`
    CurrentQuantity float64 `json:"current_quantity" form:"current_quantity"`
    Unit            string  `json:"unit" form:"unit"`
}
```
Alternatively read values with `c.FormValue(...)`. Do the same for `MealPlanCreate` and the recipe creation form fields.

### 5.4 XSS / HTML injection in all render helpers
Every render function concatenated user-controlled strings (`recipe.Title`, `recipe.Description`, `ingredient_name`, step instructions, etc.) directly into HTML. A recipe named `<script>...</script>` executes.

Fix: HTML-escape ALL user-supplied text before injecting. Use `html.EscapeString(...)` (stdlib `html`) or `template.HTMLEscapeString`, or render through `html/template`. Nutrition numbers and IDs may be formatted directly.

### 5.5 Seed data references non-existent ingredients
`Mussel` and `Saffron` (Paella recipe) were never inserted, so the FK constraint aborted seeding Recipe 9 (and, without a transaction, left partial data). Fix: insert ALL referenced ingredients before any recipe_ingredients rows; wrap the seed in a transaction (or at least ensure ordering).

### 5.6 Migration scripts referenced a non-existent `unit` column
`sql/alter_ingredients_unique_constraint.sql` operated on `ingredients.unit`, which the original `schema.sql` never defined, so the migration failed. Fix: the corrected schema above defines `unit` up front; re-introduce the composite unique constraint only if you actually want `(name, unit)` uniqueness (the Go app uses name-only uniqueness). Pick one and keep schema + migrations + Go code consistent.

### 5.7 Dead/broken helper script
`deduplicate_ingredients.py` read a file `sql/seed_ingredients.sql` that does not exist and assumed `INSERT INTO ingredients (name, unit)`. Either delete it or rewrite it to match the real seed file and schema.

### 5.8 `GET /api/recipes/:id` returned a full HTML page
The README documented JSON but the handler returned a full HTML document, and the route was also shadowed by the page route. Fix: `/api/recipes/:id` returns JSON (`Recipe` with nested ingredients + steps). The HTML detail page is served only by `GET /recipes/:id`.

### 5.9 Inconsistent dual navigation
The recipe detail inline HTML used `<a href="/" onclick="showSection('pantry')">` links, but `showSection` was not defined in that context, so clicking navigated away and broke. Fix: pick ONE navigation model. Recommended: a single-page dashboard at `/` with tab sections, plus standalone `/recipes` and `/recipes/:id` pages that have their own working nav (no reliance on a global `showSection`). Navigation links must always point to real routes.

### 5.10 Dead code: parsed templates never used
The original parsed `frontend/templates/*.html` into a `templates` var but never executed any template — rendering was done with string concatenation, and several template files (e.g. `recipe-detail.html`) were orphaned. Fix: decide on ONE rendering approach. If using `html/template`, render through it consistently and escape by default; if using string helpers, do not parse unused templates and remove orphaned template files.

### 5.11 No transaction for recipe creation
`createRecipe` inserted the recipe, then each ingredient, then each step as independent statements. A mid-way failure left a partial recipe. Fix: insert the recipe and all its children in a single transaction; on any error, roll back and return a clean error.

### 5.12 deleteRecipe wiped the page
`deleteRecipe` returned `c.HTML(200, "")` with `hx-target="body" hx-swap="innerHTML"`, replacing the whole body with nothing. Fix: after deleting from the detail page, return a proper response (e.g. an HTMX redirect via `HX-Redirect: /recipes` header) so the user lands on the recipes list. Provide clear, working delete UX for recipes, pantry items, ingredients, and meal plans.

### 5.13 Secrets/binaries in repo
The original committed a real `.env` (with a DB password) and compiled `wareg.exe` binaries. Fix: never commit `.env`; ensure `.gitignore` covers `.env`, `*.exe`, `*.exe~`, build output. `.env.example` must contain only placeholders.

### 5.14 Empty static dir not copied into Docker image
`e.Static("/static", "static")` pointed at an empty dir, and the Dockerfile only copied `frontend/`, so `/static/*` 404'd in the container. Fix: copy `static/` into the image (and create `static/uploads/.gitkeep`).

### 5.15 Match engine clarity
The original mixed two different "80%" thresholds: an ingredient counted as "available" only if pantry quantity >= 80% of required quantity, AND a recipe matched only if >= 80% of its ingredients were "available". Keep this behavior but make it explicit and documented in the UI ("shows recipes where you have at least 80% of the ingredients"). Sort matches by match percentage descending. A recipe with zero ingredients should be skipped (not falsely shown as 100%).

---

## 6. Frontend Requirements

- Tailwind + HTMX. Single dashboard at `/` with tabs: Ingredients, Recipes, Pantry, Meal Plan, "What can I cook?".
- Forms:
  - Add ingredient (name + unit).
  - Add recipe: title, description, image URL, video URL, nutrition (calories/protein/carbs/fats), a dynamic **add/remove** list of ingredients (select existing ingredient + quantity + unit), and a dynamic **add/remove** ordered list of steps (auto-incrementing step number + instruction).
  - Add pantry item (select ingredient + quantity + unit). Note: re-adding an existing ingredient should add to its quantity (UPSERT on `ingredient_id`).
  - Add meal-plan entry (select recipe + date).
- Recipe cards: image (or placeholder), title, description, nutrition summary, link to detail.
- Recipe detail: full image, description, video button, nutrition grid, ingredient list with quantities/units, ordered steps, and a delete button.
- Pantry list, meal-plan list, and ingredient list: tables with remove buttons; live HTMX updates; toast notifications on success/error.
- Ingredient/recipe `<select>` dropdowns must be populated from the API (return clean JSON lists for these, e.g. `GET /api/ingredients` may return JSON when `Accept: application/json` or via dedicated endpoints like `GET /api/ingredients/options`).
- Mobile-first responsive layout. No XSS (escape all dynamic text).

---

## 7. Concurrency / SQL robustness

- Use `pgxpool.Pool` shared via request context (as the original did) or via a clean app struct.
- `defer rows.Close()` on every query (the original did this — keep it).
- Transactions for multi-row writes (recipe create, recipe update, seed).
- Handle DB errors gracefully: return a user-friendly HTML/JSON error and log the detail; never leak raw SQL/stack to the browser.

---

## 8. Deployment

- `Dockerfile`: multi-stage (`golang:1.22-alpine` builder -> `alpine` runtime). Copy the compiled binary, `frontend/`, AND `static/`. Add `ca-certificates`. Healthcheck must use a tool actually present in the image (alpine ships busybox `wget`, so `wget --spider http://localhost:${PORT:-7001}/` is acceptable; or add a `/healthz` route returning 200).
- `docker-compose.yml`: app service wiring `DATABASE_URL` and `PORT`, port mapping, restart policy, healthcheck.
- `helm/` chart: `Chart.yaml`, `values.yaml`, `templates/deployment.yaml`, `templates/service.yaml`, `templates/_helpers.tpl`, `helm/README.md`.
- `.github/workflows/`: `docker-build.yml` (build+push to GHCR), `helm-release.yml` (publish chart to GitHub Pages).
- `.env.example` (placeholders only), `.gitignore` (env + binaries), `README.md`, `DEPLOYMENT.md`, `CONTRIBUTING.md`, `LICENSE`, `CHECKLIST.md`.

---

## 9. Acceptance Criteria

A reviewer should be able to:

1. `psql -f sql/schema.sql` then `psql -f sql/seed_data.sql` with NO errors (all referenced ingredients exist).
2. `go run main.go` (with `DATABASE_URL` set) starts on port 7001 and connects to PostgreSQL.
3. Create, view, list, and delete recipes; ingredients and steps show **exactly once** (no duplicates).
4. Add/remove pantry items (UPSERT adds quantity for existing ingredients).
5. Add a meal to the meal plan (this MUST succeed — the original was broken).
6. "What can I cook?" returns recipes with correct, sorted match percentages and accurate missing-ingredient lists.
7. No XSS: user text with `<script>` is escaped.
8. `docker compose up --build` works and `/static/*` is reachable inside the container.
9. `go vet ./...` and `gofmt -l .` are clean.
10. No `.env`, no binaries, no real secrets are committed.

---

## 10. Deliverables (file tree)

```
wareg/
├── main.go                         # Echo app: routes, handlers, render helpers
├── go.mod / go.sum
├── sql/
│   ├── schema.sql
│   ├── seed_data.sql
│   └── (optional) migrations/
├── frontend/templates/             # base, recipes, recipe-detail, partials (if using html/template)
├── static/                         # uploads/.gitkeep, favicon, etc.
├── Dockerfile
├── docker-compose.yml
├── helm/                           # Chart.yaml, values.yaml, templates/*
├── .github/workflows/              # docker-build.yml, helm-release.yml
├── .env.example
├── .gitignore
├── README.md / DEPLOYMENT.md / CONTRIBUTING.md / CHECKLIST.md
└── LICENSE
```

Prioritize clean code, minimal dependencies, consistency between schema/migrations/code, proper error handling, and a polished mobile-friendly UI.
