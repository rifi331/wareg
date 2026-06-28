package main

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	echolib "github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// matchThreshold: an ingredient counts as "available" when the pantry holds at
// least this fraction of the required quantity. A recipe matches when at least
// this fraction of its ingredients are available. (see RECREATE_PROMPT 5.15)
const matchThreshold = 0.8

// Server holds shared application dependencies.
type Server struct {
	db *pgxpool.Pool
}

// ---------------------------------------------------------------------------
// Domain types
//
// Every struct that is bound from request data carries BOTH `json` and `form`
// tags. Echo matches HTML form fields by the `form` tag (then the field name),
// NOT the `json` tag, so the original app silently bound zero values and then
// hit foreign-key violations. (see RECREATE_PROMPT 5.3)
// ---------------------------------------------------------------------------

type Ingredient struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Unit string `json:"unit"`
}

type RecipeIngredient struct {
	ID             int     `json:"id"`
	RecipeID       int     `json:"-"`
	IngredientID   int     `json:"ingredient_id"`
	Quantity       float64 `json:"quantity"`
	Unit           string  `json:"unit"`
	IngredientName string  `json:"ingredient_name"`
	Mandatory      bool    `json:"mandatory"`
}

type RecipeStep struct {
	ID          int    `json:"id"`
	RecipeID    int    `json:"-"`
	StepNumber  int    `json:"step_number"`
	Instruction string `json:"instruction"`
}

type NutritionInfo struct {
	Calories float64 `json:"calories,omitempty"`
	Protein  float64 `json:"protein,omitempty"`
	Carbs    float64 `json:"carbs,omitempty"`
	Fats     float64 `json:"fats,omitempty"`
}

type Recipe struct {
	ID           int                `json:"id"`
	Title        string             `json:"title"`
	Description  string             `json:"description"`
	Instructions string             `json:"instructions"`
	ImageURL     string             `json:"image_url"`
	VideoURL     string             `json:"video_url"`
	Nutrition    *NutritionInfo     `json:"nutrition,omitempty"`
	Ingredients  []RecipeIngredient `json:"ingredients"`
	Steps        []RecipeStep       `json:"steps"`
}

type PantryItem struct {
	ID              int     `json:"id"`
	IngredientID    int     `json:"ingredient_id"`
	CurrentQuantity float64 `json:"current_quantity"`
	Unit            string  `json:"unit"`
	IngredientName  string  `json:"ingredient_name"`
}

type PantryItemCreate struct {
	IngredientID    int     `json:"ingredient_id" form:"ingredient_id"`
	CurrentQuantity float64 `json:"current_quantity" form:"current_quantity"`
	Unit            string  `json:"unit" form:"unit"`
}

type MealPlan struct {
	ID          int    `json:"id"`
	RecipeID    int    `json:"recipe_id"`
	RecipeTitle string `json:"recipe_title"`
	PlanDate    string `json:"plan_date"`
}

type MealPlanCreate struct {
	RecipeID int    `json:"recipe_id" form:"recipe_id"`
	PlanDate string `json:"plan_date" form:"plan_date"`
}

type RecipeMatch struct {
	Recipe             Recipe   `json:"recipe"`
	MatchPercentage    float64  `json:"match_percentage"`
	MissingIngredients []string `json:"missing_ingredients"`
}

type RecipeOption struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

// ---------------------------------------------------------------------------
// main
// ---------------------------------------------------------------------------

func main() {
	_ = godotenv.Load() // optional .env; ignore error when absent

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	srv := &Server{db: pool}

	e := echolib.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Whole-site HTTP Basic Auth, driven by AUTH_USERS="user:pass,user:pass".
	// When unset, no auth (local dev). /healthz is skipped so probes still pass.
	if users := parseAuthUsers(os.Getenv("AUTH_USERS")); len(users) > 0 {
		e.Use(middleware.BasicAuthWithConfig(middleware.BasicAuthConfig{
			Skipper: func(c echolib.Context) bool {
				return c.Request().URL.Path == "/healthz"
			},
			Validator: func(user, pass string, _ echolib.Context) (bool, error) {
				expected, ok := users[user]
				if !ok {
					return false, nil
				}
				return subtle.ConstantTimeCompare([]byte(pass), []byte(expected)) == 1, nil
			},
		}))
	}

	e.Use(middleware.CORS())
	e.Use(middleware.RequestID())

	// Static files — served with no-cache so browsers always re-download JS/CSS
	// after updates (prevents stale recipe-form.js). (see RECREATE_PROMPT 5.16)
	e.Group("/static", func(next echolib.HandlerFunc) echolib.HandlerFunc {
		return func(c echolib.Context) error {
			c.Response().Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
			return next(c)
		}
	}).Static("/", "static")

	// Health endpoint used by Docker/k8s probes (accept GET and HEAD — probes
	// like `wget --spider` issue HEAD requests).
	e.Match([]string{http.MethodGet, http.MethodHead}, "/healthz", func(c echolib.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	// Pages.
	e.GET("/", srv.dashboardPage)
	e.GET("/recipes", srv.recipesPage)
	e.GET("/recipes/:id", srv.recipeDetailPage)
	e.GET("/recipes/:id/edit", srv.recipeEditPage)

	// Ingredients.
	e.GET("/api/ingredients", srv.listIngredients)
	e.POST("/api/ingredients", srv.createIngredient)
	e.DELETE("/api/ingredients/:id", srv.deleteIngredient)
	e.GET("/api/ingredients/options", srv.ingredientOptions)

	// Recipes.
	e.GET("/api/recipes", srv.listRecipes)
	e.POST("/api/recipes", srv.createRecipe)
	e.PUT("/api/recipes/:id", srv.updateRecipe)
	e.DELETE("/api/recipes/:id", srv.deleteRecipe)
	e.GET("/api/recipes/options", srv.recipeOptions)
	e.GET("/api/recipes/:id", srv.getRecipeJSON) // JSON (fixes 5.8)

	// Pantry.
	e.GET("/api/pantry", srv.listPantry)
	e.POST("/api/pantry", srv.addToPantry)
	e.PUT("/api/pantry/:id", srv.updatePantry)
	e.DELETE("/api/pantry/:id", srv.removeFromPantry)
	e.GET("/api/pantry/matches", srv.findMatchingRecipes)

	// Meal plan.
	e.GET("/api/meal-plan", srv.listMealPlan)
	e.POST("/api/meal-plan", srv.addToMealPlan)
	e.DELETE("/api/meal-plan/:id", srv.removeFromMealPlan)
	e.GET("/api/meal-plan/today", srv.mealPlanToday)

	port := os.Getenv("PORT")
	if port == "" {
		port = "7001"
	}

	log.Printf("Wareg starting on port %s", port)
	e.Logger.Fatal(e.Start(":" + port))
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// esc HTML-escapes user-supplied text before it is injected into markup,
// preventing XSS. (see RECREATE_PROMPT 5.4)
func esc(s string) string {
	return html.EscapeString(s)
}

// parseAuthUsers parses AUTH_USERS="user:pass,user:pass" into a map.
func parseAuthUsers(s string) map[string]string {
	users := make(map[string]string)
	for _, pair := range strings.Split(s, ",") {
		kv := strings.SplitN(strings.TrimSpace(pair), ":", 2)
		if len(kv) == 2 && kv[0] != "" && kv[1] != "" {
			users[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}
	return users
}

// safeURL returns the URL only if it uses an http(s) scheme. HTML-escaping alone
// does not stop `javascript:`/`data:` URLs from executing in href/src, so any
// user-controlled URL bound into an attribute must pass through here first.
func safeURL(u string) string {
	s := strings.TrimSpace(u)
	if s == "" {
		return ""
	}
	lower := strings.ToLower(s)
	if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") {
		return s
	}
	return ""
}

// videoEmbedHTML parses a YouTube or Instagram URL and returns a responsive
// iframe embed. Returns "" for empty/unknown URLs.
func videoEmbedHTML(rawURL string) string {
	u := safeURL(rawURL)
	if u == "" {
		return ""
	}
	lower := strings.ToLower(u)

	if strings.Contains(lower, "youtube.com") || strings.Contains(lower, "youtu.be") {
		id := extractYouTubeID(u)
		if id == "" {
			return ""
		}
		return `<div class="relative w-full mb-6" style="padding-bottom:56.25%"><iframe class="absolute inset-0 w-full h-full rounded-lg shadow-lg" src="https://www.youtube.com/embed/` + id + `" title="Recipe video" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe></div>`
	}

	if strings.Contains(lower, "instagram.com") {
		// Instagram supports /embed/ suffix on any post/reel URL
		embedURL := strings.TrimRight(u, "/") + "/embed/"
		return `<div class="relative w-full mb-6 flex justify-center"><iframe class="rounded-lg shadow-lg" src="` + esc(embedURL) + `" title="Recipe video" frameborder="0" allowfullscreen scrolling="no" width="400" height="720" loading="lazy"></iframe></div>`
	}

	return ""
}

// extractYouTubeID pulls the 11-char video ID from common YouTube URL formats.
func extractYouTubeID(u string) string {
	if i := strings.Index(u, "youtu.be/"); i >= 0 {
		rest := u[i+len("youtu.be/"):]
		return splitVideoID(rest)
	}
	if i := strings.Index(u, "v="); i >= 0 {
		rest := u[i+2:]
		return splitVideoID(rest)
	}
	if i := strings.Index(u, "/embed/"); i >= 0 {
		rest := u[i+len("/embed/"):]
		return splitVideoID(rest)
	}
	return ""
}

func splitVideoID(s string) string {
	for _, sep := range []string{"&", "?", "/"} {
		if i := strings.Index(s, sep); i >= 0 {
			s = s[:i]
		}
	}
	return s
}

func atoiParam(c echolib.Context, name string) (int, bool) {
	id, err := strconv.Atoi(c.Param(name))
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

// htmlError returns a short, user-friendly message and logs the detail
// server-side. Raw SQL/errors are never sent to the browser. The message is
// plain text so the front-end can show it as a toast. (see robustness notes)
func htmlError(c echolib.Context, code int, msg string, err error) error {
	if err != nil {
		log.Printf("%s: %v", msg, err)
	}
	return c.String(code, msg)
}

func parseFloatForm(c echolib.Context, key string) float64 {
	v, _ := strconv.ParseFloat(strings.TrimSpace(c.FormValue(key)), 64)
	return v
}

// formIndices returns the sorted set of integer indices present in the form for
// keys shaped like `<prefix><n><suffix>`. It tolerates gaps, so deleting a middle
// row in the UI (which is not re-indexed) no longer drops the rows after it.
func formIndices(form url.Values, prefix, suffix string) []int {
	seen := make(map[int]struct{})
	var idxs []int
	for k := range form {
		if !strings.HasPrefix(k, prefix) || !strings.HasSuffix(k, suffix) {
			continue
		}
		mid := strings.TrimSuffix(strings.TrimPrefix(k, prefix), suffix)
		n, err := strconv.Atoi(mid)
		if err != nil || n < 0 {
			continue
		}
		if _, ok := seen[n]; ok {
			continue
		}
		seen[n] = struct{}{}
		idxs = append(idxs, n)
	}
	sort.Ints(idxs)
	return idxs
}

func parseIngredients(c echolib.Context) []RecipeIngredient {
	form, err := c.FormParams()
	if err != nil {
		return nil
	}
	out := make([]RecipeIngredient, 0)
	for _, i := range formIndices(form, "ingredients.", ".ingredient_id") {
		ingredientID, err := strconv.Atoi(strings.TrimSpace(form.Get(fmt.Sprintf("ingredients.%d.ingredient_id", i))))
		if err != nil || ingredientID <= 0 {
			continue
		}
		q, _ := strconv.ParseFloat(strings.TrimSpace(form.Get(fmt.Sprintf("ingredients.%d.quantity", i))), 64)
		if q <= 0 {
			continue
		}
		out = append(out, RecipeIngredient{
			IngredientID: ingredientID,
			Quantity:     q,
			Unit:         strings.TrimSpace(form.Get(fmt.Sprintf("ingredients.%d.unit", i))),
			Mandatory:    form.Get(fmt.Sprintf("ingredients.%d.mandatory", i)) == "on",
		})
	}
	return out
}

func parseSteps(c echolib.Context) []RecipeStep {
	form, err := c.FormParams()
	if err != nil {
		return nil
	}
	out := make([]RecipeStep, 0)
	for _, i := range formIndices(form, "steps.", ".step_number") {
		num, err := strconv.Atoi(strings.TrimSpace(form.Get(fmt.Sprintf("steps.%d.step_number", i))))
		if err != nil || num <= 0 {
			continue
		}
		instr := strings.TrimSpace(form.Get(fmt.Sprintf("steps.%d.instruction", i)))
		if instr == "" {
			continue
		}
		out = append(out, RecipeStep{StepNumber: num, Instruction: instr})
	}
	return out
}

const baseRecipeSelect = `SELECT id, title, COALESCE(description,''), COALESCE(instructions,''),
	COALESCE(image_url,''), COALESCE(video_url,''), COALESCE(nutrition_json,'null'::jsonb)
	FROM recipes`

// loadAllRecipes fetches recipes and their children in three independent
// queries and assembles them in Go. This avoids the row multiplication caused
// by joining both child tables + two json_agg calls (ingredients and steps were
// each repeated). (see RECREATE_PROMPT 5.2)
func (s *Server) loadAllRecipes(ctx context.Context) ([]Recipe, error) {
	rows, err := s.db.Query(ctx, baseRecipeSelect+" ORDER BY created_at DESC, id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recipes []Recipe
	index := make(map[int]int)
	for rows.Next() {
		var r Recipe
		var nut []byte
		if err := rows.Scan(&r.ID, &r.Title, &r.Description, &r.Instructions, &r.ImageURL, &r.VideoURL, &nut); err != nil {
			return nil, err
		}
		decodeNutrition(nut, &r)
		r.Ingredients = []RecipeIngredient{}
		r.Steps = []RecipeStep{}
		index[r.ID] = len(recipes)
		recipes = append(recipes, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Ingredients are fetched separately and grouped in Go to avoid the
	// json_agg/JOIN row-multiplication bug. Steps are intentionally NOT loaded
	// here: neither caller (recipe cards, matches) renders steps, and the detail
	// page uses loadRecipe() which loads them. (see RECREATE_PROMPT 5.2)
	iRows, err := s.db.Query(ctx, `SELECT ri.id, ri.recipe_id, ri.ingredient_id, ri.quantity, ri.unit, ri.mandatory, i.name
		FROM recipe_ingredients ri JOIN ingredients i ON i.id = ri.ingredient_id
		ORDER BY ri.recipe_id, ri.id`)
	if err != nil {
		return nil, err
	}
	defer iRows.Close()
	for iRows.Next() {
		var ri RecipeIngredient
		if err := iRows.Scan(&ri.ID, &ri.RecipeID, &ri.IngredientID, &ri.Quantity, &ri.Unit, &ri.Mandatory, &ri.IngredientName); err != nil {
			return nil, err
		}
		if idx, ok := index[ri.RecipeID]; ok {
			recipes[idx].Ingredients = append(recipes[idx].Ingredients, ri)
		}
	}
	if err := iRows.Err(); err != nil {
		return nil, err
	}

	return recipes, nil
}

// loadRecipe fetches a single recipe with its children via separate queries.
func (s *Server) loadRecipe(ctx context.Context, id int) (*Recipe, error) {
	var r Recipe
	var nut []byte
	err := s.db.QueryRow(ctx, baseRecipeSelect+" WHERE id = $1", id).
		Scan(&r.ID, &r.Title, &r.Description, &r.Instructions, &r.ImageURL, &r.VideoURL, &nut)
	if err != nil {
		return nil, err
	}
	decodeNutrition(nut, &r)
	r.Ingredients = []RecipeIngredient{}
	r.Steps = []RecipeStep{}

	iRows, err := s.db.Query(ctx, `SELECT ri.id, ri.ingredient_id, ri.quantity, ri.unit, ri.mandatory, i.name
		FROM recipe_ingredients ri JOIN ingredients i ON i.id = ri.ingredient_id
		WHERE ri.recipe_id = $1 ORDER BY ri.id`, id)
	if err != nil {
		return nil, err
	}
	defer iRows.Close()
	for iRows.Next() {
		var ri RecipeIngredient
		if err := iRows.Scan(&ri.ID, &ri.IngredientID, &ri.Quantity, &ri.Unit, &ri.Mandatory, &ri.IngredientName); err != nil {
			return nil, err
		}
		r.Ingredients = append(r.Ingredients, ri)
	}
	if err := iRows.Err(); err != nil {
		return nil, err
	}

	sRows, err := s.db.Query(ctx, `SELECT id, step_number, instruction FROM recipe_steps WHERE recipe_id = $1 ORDER BY step_number`, id)
	if err != nil {
		return nil, err
	}
	defer sRows.Close()
	for sRows.Next() {
		var st RecipeStep
		if err := sRows.Scan(&st.ID, &st.StepNumber, &st.Instruction); err != nil {
			return nil, err
		}
		r.Steps = append(r.Steps, st)
	}
	if err := sRows.Err(); err != nil {
		return nil, err
	}

	return &r, nil
}

func decodeNutrition(nut []byte, r *Recipe) {
	if len(nut) == 0 || string(nut) == "null" {
		return
	}
	var n NutritionInfo
	if err := json.Unmarshal(nut, &n); err == nil {
		if n.Calories > 0 || n.Protein > 0 || n.Carbs > 0 || n.Fats > 0 {
			r.Nutrition = &n
		}
	}
}

func formatQty(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

// ---------------------------------------------------------------------------
// Pages
// ---------------------------------------------------------------------------

func (s *Server) dashboardPage(c echolib.Context) error {
	return c.File("frontend/templates/base.html")
}

func (s *Server) recipesPage(c echolib.Context) error {
	return c.File("frontend/templates/recipes.html")
}

func (s *Server) recipeDetailPage(c echolib.Context) error {
	id, ok := atoiParam(c, "id")
	if !ok {
		return c.HTML(http.StatusNotFound, pageShell("Not found", `<p class="text-center py-12 text-gray-700">Recipe not found.</p>`))
	}
	r, err := s.loadRecipe(c.Request().Context(), id)
	if err != nil {
		log.Printf("load recipe page %d: %v", id, err)
		return c.HTML(http.StatusNotFound, pageShell("Not found", `<p class="text-center py-12 text-gray-700">Recipe not found.</p>`))
	}
	return c.HTML(http.StatusOK, pageShell(esc(r.Title), renderRecipeDetail(*r)))
}

// ---------------------------------------------------------------------------
// Ingredients
// ---------------------------------------------------------------------------

func (s *Server) listIngredients(c echolib.Context) error {
	ctx := c.Request().Context()
	rows, err := s.db.Query(ctx, "SELECT id, name, COALESCE(unit,'') FROM ingredients ORDER BY name")
	if err != nil {
		return htmlError(c, http.StatusInternalServerError, "Could not load ingredients.", err)
	}
	defer rows.Close()

	var items []Ingredient
	for rows.Next() {
		var ing Ingredient
		if err := rows.Scan(&ing.ID, &ing.Name, &ing.Unit); err != nil {
			return htmlError(c, http.StatusInternalServerError, "Could not read ingredients.", err)
		}
		items = append(items, ing)
	}
	if err := rows.Err(); err != nil {
		return htmlError(c, http.StatusInternalServerError, "Could not read ingredients.", err)
	}
	return c.HTML(http.StatusOK, renderIngredientRows(items))
}

func (s *Server) ingredientOptions(c echolib.Context) error {
	ctx := c.Request().Context()
	rows, err := s.db.Query(ctx, "SELECT id, name, COALESCE(unit,'') FROM ingredients ORDER BY name")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "could not load ingredients"})
	}
	defer rows.Close()
	items := []Ingredient{}
	for rows.Next() {
		var ing Ingredient
		if err := rows.Scan(&ing.ID, &ing.Name, &ing.Unit); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "could not read ingredients"})
		}
		items = append(items, ing)
	}
	return c.JSON(http.StatusOK, items)
}

func (s *Server) createIngredient(c echolib.Context) error {
	ctx := c.Request().Context()
	name := strings.TrimSpace(c.FormValue("name"))
	if name == "" {
		return c.String(http.StatusBadRequest, "Name is required.")
	}
	unit := strings.TrimSpace(c.FormValue("unit"))

	ct, err := s.db.Exec(ctx,
		"INSERT INTO ingredients (name, unit) VALUES ($1, $2) ON CONFLICT (name) DO NOTHING",
		name, unit)
	if err != nil {
		return htmlError(c, http.StatusInternalServerError, "Could not create ingredient.", err)
	}
	if ct.RowsAffected() == 0 {
		return c.String(http.StatusConflict, "Ingredient already exists: "+name)
	}
	return s.listIngredients(c)
}

func (s *Server) deleteIngredient(c echolib.Context) error {
	ctx := c.Request().Context()
	id, ok := atoiParam(c, "id")
	if !ok {
		return htmlError(c, http.StatusBadRequest, "Invalid ingredient.", nil)
	}
	if _, err := s.db.Exec(ctx, "DELETE FROM ingredients WHERE id = $1", id); err != nil {
		return htmlError(c, http.StatusInternalServerError, "Could not delete ingredient.", err)
	}
	return s.listIngredients(c)
}

// ---------------------------------------------------------------------------
// Recipes
// ---------------------------------------------------------------------------

func (s *Server) listRecipes(c echolib.Context) error {
	recipes, err := s.loadAllRecipes(c.Request().Context())
	if err != nil {
		return htmlError(c, http.StatusInternalServerError, "Could not load recipes.", err)
	}
	if len(recipes) == 0 {
		return c.HTML(http.StatusOK, `<p class="text-gray-500 col-span-full text-center py-12">No recipes yet. Create one!</p>`)
	}
	var b strings.Builder
	for _, r := range recipes {
		b.WriteString(renderRecipeCard(r))
	}
	return c.HTML(http.StatusOK, b.String())
}

func (s *Server) recipeOptions(c echolib.Context) error {
	ctx := c.Request().Context()
	rows, err := s.db.Query(ctx, "SELECT id, title FROM recipes ORDER BY title")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "could not load recipes"})
	}
	defer rows.Close()
	opts := []RecipeOption{}
	for rows.Next() {
		var o RecipeOption
		if err := rows.Scan(&o.ID, &o.Title); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "could not read recipes"})
		}
		opts = append(opts, o)
	}
	return c.JSON(http.StatusOK, opts)
}

// getRecipeJSON returns JSON for a single recipe. The original handler returned
// a full HTML page here AND shadowed the page route. (see RECREATE_PROMPT 5.8)
func (s *Server) getRecipeJSON(c echolib.Context) error {
	id, ok := atoiParam(c, "id")
	if !ok {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}
	r, err := s.loadRecipe(c.Request().Context(), id)
	if err != nil {
		log.Printf("getRecipeJSON %d: %v", id, err)
		return c.JSON(http.StatusNotFound, map[string]string{"error": "recipe not found"})
	}
	return c.JSON(http.StatusOK, r)
}

// createRecipe inserts the recipe and all of its children in a single
// transaction; any error rolls back so no partial recipe is left behind.
// (see RECREATE_PROMPT 5.11)
func (s *Server) createRecipe(c echolib.Context) error {
	ctx := c.Request().Context()
	title := strings.TrimSpace(c.FormValue("title"))
	if title == "" {
		return c.String(http.StatusBadRequest, "Title is required.")
	}

	n := NutritionInfo{
		Calories: parseFloatForm(c, "nutrition.calories"),
		Protein:  parseFloatForm(c, "nutrition.protein"),
		Carbs:    parseFloatForm(c, "nutrition.carbs"),
		Fats:     parseFloatForm(c, "nutrition.fats"),
	}
	var nutritionJSON []byte
	if n.Calories > 0 || n.Protein > 0 || n.Carbs > 0 || n.Fats > 0 {
		nutritionJSON, _ = json.Marshal(n)
	}

	ingredients := parseIngredients(c)
	steps := parseSteps(c)

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return htmlError(c, http.StatusInternalServerError, "Could not start transaction.", err)
	}
	defer tx.Rollback(ctx) // safe no-op after Commit

	var recipeID int
	err = tx.QueryRow(ctx, `INSERT INTO recipes (title, description, instructions, image_url, video_url, nutrition_json)
		VALUES ($1,$2,$3,$4,$5,$6) RETURNING id`,
		title,
		strings.TrimSpace(c.FormValue("description")),
		strings.TrimSpace(c.FormValue("instructions")),
		strings.TrimSpace(c.FormValue("image_url")),
		strings.TrimSpace(c.FormValue("video_url")),
		nutritionJSON,
	).Scan(&recipeID)
	if err != nil {
		return htmlError(c, http.StatusInternalServerError, "Could not create recipe.", err)
	}

	for _, ing := range ingredients {
		if _, err := tx.Exec(ctx,
			`INSERT INTO recipe_ingredients (recipe_id, ingredient_id, quantity, unit, mandatory) VALUES ($1,$2,$3,$4,$5)
			 ON CONFLICT (recipe_id, ingredient_id) DO UPDATE SET quantity = EXCLUDED.quantity, unit = EXCLUDED.unit, mandatory = EXCLUDED.mandatory`,
			recipeID, ing.IngredientID, ing.Quantity, ing.Unit, ing.Mandatory); err != nil {
			return htmlError(c, http.StatusInternalServerError, "Could not add recipe ingredients.", err)
		}
	}

	for _, st := range steps {
		if _, err := tx.Exec(ctx,
			`INSERT INTO recipe_steps (recipe_id, step_number, instruction) VALUES ($1,$2,$3)
			 ON CONFLICT (recipe_id, step_number) DO UPDATE SET instruction = EXCLUDED.instruction`,
			recipeID, st.StepNumber, st.Instruction); err != nil {
			return htmlError(c, http.StatusInternalServerError, "Could not add recipe steps.", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return htmlError(c, http.StatusInternalServerError, "Could not save recipe.", err)
	}
	return s.listRecipes(c)
}

// deleteRecipe redirects the browser back to the recipes list via the HTMX
// HX-Redirect header instead of wiping the page body. (see RECREATE_PROMPT 5.12)
func (s *Server) deleteRecipe(c echolib.Context) error {
	ctx := c.Request().Context()
	id, ok := atoiParam(c, "id")
	if !ok {
		return htmlError(c, http.StatusBadRequest, "Invalid recipe.", nil)
	}
	if _, err := s.db.Exec(ctx, "DELETE FROM recipes WHERE id = $1", id); err != nil {
		return htmlError(c, http.StatusInternalServerError, "Could not delete recipe.", err)
	}
	c.Response().Header().Set("HX-Redirect", "/recipes")
	return c.String(http.StatusOK, "")
}

// recipeEditPage renders the edit form pre-filled with the recipe's current
// values (fields, nutrition, ingredients with mandatory flags, steps).
func (s *Server) recipeEditPage(c echolib.Context) error {
	id, ok := atoiParam(c, "id")
	if !ok {
		return c.HTML(http.StatusNotFound, pageShell("Not found", `<p class="text-center py-12 text-gray-700">Recipe not found.</p>`))
	}
	r, err := s.loadRecipe(c.Request().Context(), id)
	if err != nil {
		log.Printf("load recipe edit %d: %v", id, err)
		return c.HTML(http.StatusNotFound, pageShell("Not found", `<p class="text-center py-12 text-gray-700">Recipe not found.</p>`))
	}
	return c.HTML(http.StatusOK, pageShell("Edit "+esc(r.Title), renderEditForm(*r)))
}

// updateRecipe replaces a recipe's fields + ingredients + steps in a single
// transaction (children are deleted and re-inserted so add/remove/reorder and
// the mandatory flag are handled cleanly). Responds with HX-Redirect to the
// detail page.
func (s *Server) updateRecipe(c echolib.Context) error {
	ctx := c.Request().Context()
	id, ok := atoiParam(c, "id")
	if !ok {
		return c.String(http.StatusBadRequest, "Invalid recipe.")
	}
	title := strings.TrimSpace(c.FormValue("title"))
	if title == "" {
		return c.String(http.StatusBadRequest, "Title is required.")
	}

	n := NutritionInfo{
		Calories: parseFloatForm(c, "nutrition.calories"),
		Protein:  parseFloatForm(c, "nutrition.protein"),
		Carbs:    parseFloatForm(c, "nutrition.carbs"),
		Fats:     parseFloatForm(c, "nutrition.fats"),
	}
	var nutritionJSON []byte
	if n.Calories > 0 || n.Protein > 0 || n.Carbs > 0 || n.Fats > 0 {
		nutritionJSON, _ = json.Marshal(n)
	}
	ingredients := parseIngredients(c)
	steps := parseSteps(c)

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return htmlError(c, http.StatusInternalServerError, "Could not start transaction.", err)
	}
	defer tx.Rollback(ctx)

	ct, err := tx.Exec(ctx, `UPDATE recipes SET title=$1, description=$2, instructions=$3,
		image_url=$4, video_url=$5, nutrition_json=$6, updated_at=CURRENT_TIMESTAMP WHERE id=$7`,
		title,
		strings.TrimSpace(c.FormValue("description")),
		strings.TrimSpace(c.FormValue("instructions")),
		strings.TrimSpace(c.FormValue("image_url")),
		strings.TrimSpace(c.FormValue("video_url")),
		nutritionJSON,
		id)
	if err != nil {
		return htmlError(c, http.StatusInternalServerError, "Could not update recipe.", err)
	}
	if ct.RowsAffected() == 0 {
		return c.String(http.StatusNotFound, "Recipe not found.")
	}

	// Replace children entirely.
	if _, err := tx.Exec(ctx, "DELETE FROM recipe_ingredients WHERE recipe_id = $1", id); err != nil {
		return htmlError(c, http.StatusInternalServerError, "Could not update ingredients.", err)
	}
	if _, err := tx.Exec(ctx, "DELETE FROM recipe_steps WHERE recipe_id = $1", id); err != nil {
		return htmlError(c, http.StatusInternalServerError, "Could not update steps.", err)
	}
	for _, ing := range ingredients {
		if _, err := tx.Exec(ctx,
			`INSERT INTO recipe_ingredients (recipe_id, ingredient_id, quantity, unit, mandatory) VALUES ($1,$2,$3,$4,$5)`,
			id, ing.IngredientID, ing.Quantity, ing.Unit, ing.Mandatory); err != nil {
			return htmlError(c, http.StatusInternalServerError, "Could not add recipe ingredients.", err)
		}
	}
	for _, st := range steps {
		if _, err := tx.Exec(ctx,
			`INSERT INTO recipe_steps (recipe_id, step_number, instruction) VALUES ($1,$2,$3)`,
			id, st.StepNumber, st.Instruction); err != nil {
			return htmlError(c, http.StatusInternalServerError, "Could not add recipe steps.", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return htmlError(c, http.StatusInternalServerError, "Could not save recipe.", err)
	}
	c.Response().Header().Set("HX-Redirect", "/recipes/"+strconv.Itoa(id))
	return c.String(http.StatusOK, "")
}

// ---------------------------------------------------------------------------
// Pantry
// ---------------------------------------------------------------------------

func (s *Server) listPantry(c echolib.Context) error {
	ctx := c.Request().Context()
	rows, err := s.db.Query(ctx, `SELECT p.id, p.ingredient_id, p.current_quantity, p.unit, i.name
		FROM pantry p JOIN ingredients i ON i.id = p.ingredient_id ORDER BY i.name`)
	if err != nil {
		return htmlError(c, http.StatusInternalServerError, "Could not load pantry.", err)
	}
	defer rows.Close()
	items := []PantryItem{}
	for rows.Next() {
		var it PantryItem
		if err := rows.Scan(&it.ID, &it.IngredientID, &it.CurrentQuantity, &it.Unit, &it.IngredientName); err != nil {
			return htmlError(c, http.StatusInternalServerError, "Could not read pantry.", err)
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return htmlError(c, http.StatusInternalServerError, "Could not read pantry.", err)
	}
	return c.HTML(http.StatusOK, renderPantryTable(items))
}

// addToPantry stores pantry stock per (ingredient_id, unit) so the same
// ingredient can be tracked in multiple forms (e.g. "Beef Stock" as 1000 ml
// and as 10 cubes). Re-adding the same ingredient+unit ADDS to that quantity.
// Values come from form tags so they bind correctly. (see 5.3)
func (s *Server) addToPantry(c echolib.Context) error {
	ctx := c.Request().Context()
	var req PantryItemCreate
	if err := c.Bind(&req); err != nil {
		return c.String(http.StatusBadRequest, "Invalid pantry entry.")
	}
	if req.IngredientID <= 0 {
		return c.String(http.StatusBadRequest, "Select an ingredient.")
	}
	if req.CurrentQuantity <= 0 {
		return c.String(http.StatusBadRequest, "Enter a quantity greater than zero.")
	}
	unit := strings.TrimSpace(req.Unit)
	if unit == "" {
		return c.String(http.StatusBadRequest, "Select a unit.")
	}

	_, err := s.db.Exec(ctx, `INSERT INTO pantry (ingredient_id, current_quantity, unit) VALUES ($1,$2,$3)
		ON CONFLICT (ingredient_id, unit) DO UPDATE SET current_quantity = pantry.current_quantity + EXCLUDED.current_quantity`,
		req.IngredientID, req.CurrentQuantity, unit)
	if err != nil {
		return htmlError(c, http.StatusInternalServerError, "Could not add to pantry.", err)
	}
	return s.listPantry(c)
}

func (s *Server) removeFromPantry(c echolib.Context) error {
	ctx := c.Request().Context()
	id, ok := atoiParam(c, "id")
	if !ok {
		return htmlError(c, http.StatusBadRequest, "Invalid pantry item.", nil)
	}
	if _, err := s.db.Exec(ctx, "DELETE FROM pantry WHERE id = $1", id); err != nil {
		return htmlError(c, http.StatusInternalServerError, "Could not remove pantry item.", err)
	}
	return s.listPantry(c)
}

// updatePantry sets the quantity of a single pantry row (live edit). The row is
// addressed by id; the unit is unchanged. Responds with an empty body because
// the front-end uses hx-swap="none" (keeps the input value/focus).
func (s *Server) updatePantry(c echolib.Context) error {
	ctx := c.Request().Context()
	id, ok := atoiParam(c, "id")
	if !ok {
		return c.String(http.StatusBadRequest, "Invalid pantry item.")
	}
	qty := parseFloatForm(c, "current_quantity")
	if qty <= 0 {
		return c.String(http.StatusBadRequest, "Enter a quantity greater than zero, or delete the row.")
	}
	ct, err := s.db.Exec(ctx, "UPDATE pantry SET current_quantity = $2 WHERE id = $1", id, qty)
	if err != nil {
		return htmlError(c, http.StatusInternalServerError, "Could not update pantry.", err)
	}
	if ct.RowsAffected() == 0 {
		return c.String(http.StatusNotFound, "Pantry item not found.")
	}
	return c.String(http.StatusOK, "")
}

// ---------------------------------------------------------------------------
// "What can I cook?" matching engine
// ---------------------------------------------------------------------------

func (s *Server) findMatchingRecipes(c echolib.Context) error {
	ctx := c.Request().Context()

	pRows, err := s.db.Query(ctx, "SELECT ingredient_id, current_quantity FROM pantry")
	if err != nil {
		return htmlError(c, http.StatusInternalServerError, "Could not load pantry.", err)
	}
	defer pRows.Close()
	// Matching is by ingredient NAME (id) only — units and quantities are the
	// user's problem to convert. An ingredient counts as "available" if the
	// pantry holds ANY of it. A recipe matches when at least matchThreshold of
	// its ingredients are present.
	pantry := make(map[int]bool)
	for pRows.Next() {
		var ingID int
		var qty float64
		if err := pRows.Scan(&ingID, &qty); err != nil {
			return htmlError(c, http.StatusInternalServerError, "Could not read pantry.", err)
		}
		if qty > 0 {
			pantry[ingID] = true
		}
	}
	if err := pRows.Err(); err != nil {
		return htmlError(c, http.StatusInternalServerError, "Could not read pantry.", err)
	}

	recipes, err := s.loadAllRecipes(ctx)
	if err != nil {
		return htmlError(c, http.StatusInternalServerError, "Could not load recipes.", err)
	}

	matches := []RecipeMatch{}
	for _, r := range recipes {
		mandatoryTotal, available := 0, 0
		missing := []string{}
		for _, ing := range r.Ingredients {
			if !ing.Mandatory {
				continue // optional ingredients never count against the match
			}
			mandatoryTotal++
			if pantry[ing.IngredientID] {
				available++
			} else {
				missing = append(missing, ing.IngredientName)
			}
		}
		if mandatoryTotal == 0 {
			continue // skip recipes with no mandatory ingredients
		}
		pct := float64(available) / float64(mandatoryTotal) * 100
		matches = append(matches, RecipeMatch{Recipe: r, MatchPercentage: pct, MissingIngredients: missing})
	}

	sort.SliceStable(matches, func(i, j int) bool {
		return matches[i].MatchPercentage > matches[j].MatchPercentage
	})

	return c.HTML(http.StatusOK, renderMatches(matches))
}

// ---------------------------------------------------------------------------
// Meal plan
// ---------------------------------------------------------------------------

// mealPlanToday returns today's planned meals as JSON (used by the Android
// reminder notification). Always returns a JSON array (empty if none today).
func (s *Server) mealPlanToday(c echolib.Context) error {
	ctx := c.Request().Context()
	rows, err := s.db.Query(ctx, `SELECT mp.id, mp.recipe_id, r.title
		FROM meal_plan mp JOIN recipes r ON r.id = mp.recipe_id
		WHERE mp.plan_date = CURRENT_DATE ORDER BY mp.id`)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "could not load meal plan"})
	}
	defer rows.Close()
	type mealItem struct {
		ID          int    `json:"id"`
		RecipeID    int    `json:"recipe_id"`
		RecipeTitle string `json:"recipe_title"`
	}
	out := []mealItem{}
	for rows.Next() {
		var m mealItem
		if err := rows.Scan(&m.ID, &m.RecipeID, &m.RecipeTitle); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "could not read meal plan"})
		}
		out = append(out, m)
	}
	if err := rows.Err(); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "could not read meal plan"})
	}
	return c.JSON(http.StatusOK, out)
}

func (s *Server) listMealPlan(c echolib.Context) error {
	ctx := c.Request().Context()
	rows, err := s.db.Query(ctx, `SELECT mp.id, mp.recipe_id, mp.plan_date::text, r.title
		FROM meal_plan mp JOIN recipes r ON r.id = mp.recipe_id ORDER BY mp.plan_date ASC, mp.id`)
	if err != nil {
		return htmlError(c, http.StatusInternalServerError, "Could not load meal plan.", err)
	}
	defer rows.Close()
	plans := []MealPlan{}
	for rows.Next() {
		var p MealPlan
		if err := rows.Scan(&p.ID, &p.RecipeID, &p.PlanDate, &p.RecipeTitle); err != nil {
			return htmlError(c, http.StatusInternalServerError, "Could not read meal plan.", err)
		}
		plans = append(plans, p)
	}
	if err := rows.Err(); err != nil {
		return htmlError(c, http.StatusInternalServerError, "Could not read meal plan.", err)
	}
	return c.HTML(http.StatusOK, renderMealPlanTable(plans))
}

// addToMealPlan uses the corrected RETURNING clause (no stray paren) so adding
// a meal actually works. (see RECREATE_PROMPT 5.1)
func (s *Server) addToMealPlan(c echolib.Context) error {
	ctx := c.Request().Context()
	var req MealPlanCreate
	if err := c.Bind(&req); err != nil {
		return htmlError(c, http.StatusBadRequest, "Invalid meal plan entry.", err)
	}
	if req.RecipeID <= 0 || strings.TrimSpace(req.PlanDate) == "" {
		return c.String(http.StatusBadRequest, "Select a recipe and a date.")
	}

	var plan MealPlan
	err := s.db.QueryRow(ctx,
		"INSERT INTO meal_plan (recipe_id, plan_date) VALUES ($1, $2) RETURNING id, recipe_id, plan_date::text",
		req.RecipeID, strings.TrimSpace(req.PlanDate)).
		Scan(&plan.ID, &plan.RecipeID, &plan.PlanDate)
	if err != nil {
		return htmlError(c, http.StatusInternalServerError, "Could not add meal plan.", err)
	}
	return s.listMealPlan(c)
}

func (s *Server) removeFromMealPlan(c echolib.Context) error {
	ctx := c.Request().Context()
	id, ok := atoiParam(c, "id")
	if !ok {
		return htmlError(c, http.StatusBadRequest, "Invalid meal plan entry.", nil)
	}
	if _, err := s.db.Exec(ctx, "DELETE FROM meal_plan WHERE id = $1", id); err != nil {
		return htmlError(c, http.StatusInternalServerError, "Could not remove meal plan.", err)
	}
	return s.listMealPlan(c)
}

// ---------------------------------------------------------------------------
// Render helpers (all user-supplied text is escaped via esc())
// ---------------------------------------------------------------------------

func renderIngredientRows(items []Ingredient) string {
	if len(items) == 0 {
		return `<tr><td colspan="3" class="px-6 py-8 text-center text-gray-500">No ingredients yet. Add one above.</td></tr>`
	}
	var b strings.Builder
	for _, ing := range items {
		b.WriteString(`<tr id="ingredient-` + strconv.Itoa(ing.ID) + `">`)
		b.WriteString(`<td class="px-6 py-3 whitespace-nowrap text-sm font-medium text-gray-900">` + esc(ing.Name) + `</td>`)
		b.WriteString(`<td class="px-6 py-3 whitespace-nowrap text-sm text-gray-500">` + esc(ing.Unit) + `</td>`)
		b.WriteString(`<td class="px-6 py-3 whitespace-nowrap text-right text-sm font-medium">`)
		b.WriteString(`<button hx-delete="/api/ingredients/` + strconv.Itoa(ing.ID) + `" hx-target="#ingredients-list-body" hx-swap="innerHTML" class="text-red-600 hover:text-red-900">Delete</button>`)
		b.WriteString(`</td></tr>`)
	}
	return b.String()
}

func nutritionSummary(n *NutritionInfo) string {
	if n == nil {
		return ""
	}
	var parts []string
	if n.Calories > 0 {
		parts = append(parts, formatQty(n.Calories)+" kcal")
	}
	if n.Protein > 0 {
		parts = append(parts, formatQty(n.Protein)+"g protein")
	}
	if n.Carbs > 0 {
		parts = append(parts, formatQty(n.Carbs)+"g carbs")
	}
	if n.Fats > 0 {
		parts = append(parts, formatQty(n.Fats)+"g fats")
	}
	if len(parts) == 0 {
		return ""
	}
	return `<div class="text-xs text-gray-500 mb-3 flex gap-3 flex-wrap">` + strings.Join(parts, `<span aria-hidden="true">·</span>`) + `</div>`
}

func renderRecipeCard(r Recipe) string {
	image := recipeImage(r.ImageURL, r.Title, "h-48")
	nut := nutritionSummary(r.Nutrition)
	desc := ""
	if r.Description != "" {
		desc = `<p class="text-gray-600 text-sm mb-3">` + esc(r.Description) + `</p>`
	}
	id := strconv.Itoa(r.ID)
	return `<div id="recipe-` + id + `" class="bg-white rounded-lg shadow-md overflow-hidden hover:shadow-xl transition-shadow cursor-pointer flex flex-col">` +
		// Image area with hover overlay buttons
		`<div class="relative group">` +
			image +
			`<div class="absolute inset-0 bg-black/0 group-hover:bg-black/50 transition-all duration-200 flex items-center justify-center gap-3 opacity-0 group-hover:opacity-100">` +
				`<a href="/recipes/` + id + `" class="bg-emerald-600 text-white p-3 rounded-full hover:bg-emerald-700 transition shadow-lg" title="View recipe">` +
					`<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"/></svg>` +
				`</a>` +
				`<button hx-delete="/api/recipes/` + id + `" hx-target="#recipes-list" hx-swap="innerHTML" hx-confirm="Delete this recipe?" ` +
					`hx-on::after-request="if(event.detail.xhr.status<300){showToast('Recipe deleted.','success');}" ` +
					`class="bg-red-600 text-white p-3 rounded-full hover:bg-red-700 transition shadow-lg" title="Delete recipe">` +
					`<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/></svg>` +
				`</button>` +
			`</div>` +
		`</div>` +
		// Card body
		`<div class="p-4 flex-1 flex flex-col">` +
		`<h3 class="text-xl font-bold text-gray-800 mb-2">` + esc(r.Title) + `</h3>` +
		desc + nut +
		`<div class="mt-auto text-xs text-emerald-600 font-semibold">View Recipe →</div>` +
		`</div></div>`
}

func recipeImage(imageURL, alt, height string) string {
	u := safeURL(imageURL)
	if u == "" {
		return `<div class="w-full ` + height + ` bg-gray-200 flex items-center justify-center text-gray-400">` +
			`<svg class="w-16 h-16" fill="none" stroke="currentColor" viewBox="0 0 24 24">` +
			`<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z"></path>` +
			`</svg></div>`
	}
	return `<img src="` + esc(u) + `" alt="` + esc(alt) + `" class="w-full ` + height + ` object-cover">`
}

func renderPantryTable(items []PantryItem) string {
	if len(items) == 0 {
		return emptyCard("Your pantry is empty. Add some ingredients!")
	}
	var b strings.Builder
	b.WriteString(`<table class="w-full"><thead class="bg-gray-50"><tr>`)
	b.WriteString(`<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Ingredient</th>`)
	b.WriteString(`<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Quantity</th>`)
	b.WriteString(`<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Unit</th>`)
	b.WriteString(`<th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Actions</th>`)
	b.WriteString(`</tr></thead><tbody class="bg-white divide-y divide-gray-200">`)
	for _, it := range items {
		b.WriteString(`<tr id="pantry-` + strconv.Itoa(it.ID) + `">`)
		b.WriteString(`<td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">` + esc(it.IngredientName) + `</td>`)
		b.WriteString(`<td class="px-6 py-4 whitespace-nowrap">`)
		b.WriteString(`<input type="number" step="0.01" min="0" name="current_quantity" value="` + formatQty(it.CurrentQuantity) + `" ` +
			`class="w-24 px-2 py-1 border rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500" ` +
			`hx-put="/api/pantry/` + strconv.Itoa(it.ID) + `" hx-trigger="input changed delay:400ms" ` +
			`hx-include="this" hx-swap="none" ` +
			`hx-on::after-request="if(event.detail.xhr.status<300){ showToast('Quantity updated.','success'); }">`)
		b.WriteString(`</td>`)
		b.WriteString(`<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">` + esc(it.Unit) + `</td>`)
		b.WriteString(`<td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">`)
		b.WriteString(`<button hx-delete="/api/pantry/` + strconv.Itoa(it.ID) + `" hx-target="#pantry-list" hx-swap="innerHTML" hx-confirm="Remove this pantry entry?" class="text-red-600 hover:text-red-900">Delete</button>`)
		b.WriteString(`</td></tr>`)
	}
	b.WriteString(`</tbody></table>`)
	return b.String()
}

func renderMealPlanTable(plans []MealPlan) string {
	if len(plans) == 0 {
		return emptyCard("Your meal plan is empty. Add some meals!")
	}
	var b strings.Builder
	b.WriteString(`<table class="w-full"><thead class="bg-gray-50"><tr>`)
	b.WriteString(`<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Date</th>`)
	b.WriteString(`<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Recipe</th>`)
	b.WriteString(`<th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Actions</th>`)
	b.WriteString(`</tr></thead><tbody class="bg-white divide-y divide-gray-200">`)
	for _, p := range plans {
		b.WriteString(`<tr id="meal-` + strconv.Itoa(p.ID) + `">`)
		b.WriteString(`<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">` + esc(p.PlanDate) + `</td>`)
		b.WriteString(`<td class="px-6 py-4 whitespace-nowrap"><a href="/recipes/` + strconv.Itoa(p.RecipeID) + `" class="text-sm font-medium text-gray-900 hover:text-emerald-600">` + esc(p.RecipeTitle) + `</a></td>`)
		b.WriteString(`<td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">`)
		b.WriteString(`<button hx-delete="/api/meal-plan/` + strconv.Itoa(p.ID) + `" hx-target="#meal-plan-list" hx-swap="innerHTML" class="text-red-600 hover:text-red-900">Remove</button>`)
		b.WriteString(`</td></tr>`)
	}
	b.WriteString(`</tbody></table>`)
	return b.String()
}

func emptyCard(msg string) string {
	return `<div class="bg-white rounded-lg shadow-md p-6"><p class="text-center text-gray-500">` + esc(msg) + `</p></div>`
}

func renderMatches(matches []RecipeMatch) string {
	if len(matches) == 0 {
		return emptyCard("No recipes found. Add some recipes first!")
	}
	var b strings.Builder
	for _, m := range matches {
		image := ""
		if m.Recipe.ImageURL != "" {
			image = recipeImage(m.Recipe.ImageURL, m.Recipe.Title, "h-40")
		}
		badge := "bg-yellow-100 text-yellow-800"
		if m.MatchPercentage >= 100 {
			badge = "bg-green-100 text-green-800"
		} else if m.MatchPercentage < 50 {
			badge = "bg-red-100 text-red-800"
		}
		missing := `<p class="text-sm text-green-600 mb-2">All ingredients available!</p>`
		if len(m.MissingIngredients) > 0 {
			escaped := make([]string, len(m.MissingIngredients))
			for i, name := range m.MissingIngredients {
				escaped[i] = esc(name)
			}
			missing = `<div class="text-sm text-gray-600 mb-2"><strong class="text-orange-600">Missing:</strong> <span class="text-gray-500">` + strings.Join(escaped, ", ") + `</span></div>`
		}
		desc := ""
		if m.Recipe.Description != "" {
			desc = `<p class="text-gray-600 text-sm mb-3">` + esc(m.Recipe.Description) + `</p>`
		}
		b.WriteString(`<div class="match-card bg-white rounded-lg shadow-md overflow-hidden flex flex-col" data-title="` + esc(strings.ToLower(m.Recipe.Title)) + `" data-missing="` + esc(strings.ToLower(strings.Join(m.MissingIngredients, " "))) + `">`)
		b.WriteString(image)
		b.WriteString(`<div class="p-4 flex-1 flex flex-col">`)
		b.WriteString(`<div class="flex justify-between items-start mb-2 gap-2">`)
		b.WriteString(`<h3 class="text-xl font-bold text-gray-800">` + esc(m.Recipe.Title) + `</h3>`)
		b.WriteString(`<span class="px-3 py-1 rounded-full text-sm font-bold whitespace-nowrap ` + badge + `">` + formatQty(m.MatchPercentage) + `%</span>`)
		b.WriteString(`</div>`)
		b.WriteString(desc)
		b.WriteString(missing)
		b.WriteString(`<a href="/recipes/` + strconv.Itoa(m.Recipe.ID) + `" class="mt-auto text-xs text-emerald-600 font-semibold">View Recipe →</a>`)
		b.WriteString(`</div></div>`)
	}
	return b.String()
}

func renderRecipeDetail(r Recipe) string {
	var b strings.Builder
	b.WriteString(`<div class="bg-white rounded-lg shadow-lg overflow-hidden mb-6">`)
	b.WriteString(recipeImage(r.ImageURL, r.Title, "h-72 sm:h-96"))
	b.WriteString(`<div class="p-6 sm:p-8">`)
	b.WriteString(`<h1 class="text-3xl sm:text-4xl font-bold text-gray-800 mb-3">` + esc(r.Title) + `</h1>`)
	if r.Description != "" {
		b.WriteString(`<p class="text-gray-600 text-base sm:text-lg mb-4">` + esc(r.Description) + `</p>`)
	}
	if embed := videoEmbedHTML(r.VideoURL); embed != "" {
		b.WriteString(embed)
	} else if v := safeURL(r.VideoURL); v != "" {
		b.WriteString(`<a href="` + esc(v) + `" target="_blank" rel="noopener" class="inline-flex items-center gap-2 bg-red-600 text-white px-5 py-2.5 rounded-lg hover:bg-red-700 transition mb-6">`)
		b.WriteString(`<svg class="w-5 h-5" fill="currentColor" viewBox="0 0 24 24"><path d="M8 5v14l11-7z"/></svg> Watch Video</a>`)
	}
	b.WriteString(`</div></div>`)

	if r.Nutrition != nil {
		b.WriteString(`<div class="bg-emerald-50 border-l-4 border-emerald-500 p-4 rounded mb-6">`)
		b.WriteString(`<h3 class="font-bold text-emerald-800 mb-3 text-lg">Nutrition</h3><div class="grid grid-cols-2 md:grid-cols-4 gap-4">`)
		if r.Nutrition.Calories > 0 {
			b.WriteString(statCell(formatQty(r.Nutrition.Calories), "Calories"))
		}
		if r.Nutrition.Protein > 0 {
			b.WriteString(statCell(formatQty(r.Nutrition.Protein)+"g", "Protein"))
		}
		if r.Nutrition.Carbs > 0 {
			b.WriteString(statCell(formatQty(r.Nutrition.Carbs)+"g", "Carbs"))
		}
		if r.Nutrition.Fats > 0 {
			b.WriteString(statCell(formatQty(r.Nutrition.Fats)+"g", "Fats"))
		}
		b.WriteString(`</div></div>`)
	}

	if len(r.Ingredients) > 0 {
		b.WriteString(`<div class="bg-blue-50 border-l-4 border-blue-500 p-4 rounded mb-6">`)
		b.WriteString(`<h3 class="font-bold text-blue-800 mb-3 text-lg">Ingredients</h3><ul class="space-y-2">`)
		for _, ing := range r.Ingredients {
			opt := ""
			if !ing.Mandatory {
				opt = ` <span class="text-xs text-gray-400">(optional)</span>`
			}
			b.WriteString(`<li class="flex items-center gap-2"><span class="text-gray-700"><strong>` + esc(ing.IngredientName) + `:</strong> ` + formatQty(ing.Quantity) + ` ` + esc(ing.Unit) + opt + `</span></li>`)
		}
		b.WriteString(`</ul></div>`)
	}

	if len(r.Steps) > 0 {
		b.WriteString(`<div class="bg-amber-50 border-l-4 border-amber-500 p-4 rounded mb-6">`)
		b.WriteString(`<h3 class="font-bold text-amber-800 mb-3 text-lg">Cooking Steps</h3><ol class="space-y-3">`)
		for _, st := range r.Steps {
			b.WriteString(`<li class="flex gap-3"><div class="flex-shrink-0 w-8 h-8 bg-amber-500 text-white rounded-full flex items-center justify-center font-bold text-sm">` + strconv.Itoa(st.StepNumber) + `</div><div class="flex-1 text-gray-700 pt-1">` + esc(st.Instruction) + `</div></li>`)
		}
		b.WriteString(`</ol></div>`)
	}

	if strings.TrimSpace(r.Instructions) != "" {
		b.WriteString(`<div class="bg-purple-50 border-l-4 border-purple-500 p-4 rounded mb-6">`)
		b.WriteString(`<h3 class="font-bold text-purple-800 mb-3 text-lg">Notes</h3><p class="text-gray-700 whitespace-pre-line">` + esc(r.Instructions) + `</p></div>`)
	}

	b.WriteString(`<div class="flex gap-4 mt-8 flex-wrap">`)
	b.WriteString(`<a href="/recipes/` + strconv.Itoa(r.ID) + `/edit" class="bg-emerald-600 text-white px-6 py-2 rounded-lg hover:bg-emerald-700 transition">Edit Recipe</a>`)
	b.WriteString(`<button hx-delete="/api/recipes/` + strconv.Itoa(r.ID) + `" hx-swap="none" hx-confirm="Delete this recipe?" class="bg-red-600 text-white px-6 py-2 rounded-lg hover:bg-red-700 transition">Delete Recipe</button>`)
	b.WriteString(`<a href="/recipes" class="px-6 py-2 rounded-lg bg-gray-200 text-gray-700 hover:bg-gray-300 transition">Back to Recipes</a>`)
	b.WriteString(`</div>`)
	return b.String()
}

func statCell(value, label string) string {
	return `<div class="text-center"><div class="text-2xl font-bold text-emerald-700">` + value + `</div><div class="text-sm text-gray-600">` + label + `</div></div>`
}

// renderEditForm builds the pre-filled edit form. Ingredient rows reuse the
// same .ingredient-row / #recipe-ingredients-list structure (and .ingredient-combo)
// as the create form, so the shared recipe-form.js add/remove/reindex + combobox
// logic works unchanged.
func renderEditForm(r Recipe) string {
	var b strings.Builder
	b.WriteString(`<div class="bg-white rounded-lg shadow-md p-6 mb-6">`)
	b.WriteString(`<div class="flex justify-between items-center mb-4"><h2 class="text-2xl font-bold text-gray-800">Edit Recipe</h2>` +
		`<a href="/recipes/` + strconv.Itoa(r.ID) + `" class="text-gray-500 hover:text-gray-700 text-sm">Cancel</a></div>`)

	b.WriteString(`<form hx-put="/api/recipes/` + strconv.Itoa(r.ID) + `" hx-swap="none" ` +
		`hx-on::after-request="if(event.detail.xhr.status>=300) showToast('Could not save recipe.','error');" class="space-y-4">`)

	b.WriteString(fieldInput("Title *", "title", r.Title, "text", true))
	b.WriteString(`<div><label class="block text-sm font-medium mb-1">Description</label>` +
		`<textarea name="description" rows="2" class="w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-emerald-500">` + esc(r.Description) + `</textarea></div>`)

	b.WriteString(`<div class="grid grid-cols-1 sm:grid-cols-2 gap-4">`)
	b.WriteString(fieldInput("Image URL", "image_url", r.ImageURL, "text", false))
	b.WriteString(fieldInput("Video URL (YouTube/Instagram)", "video_url", r.VideoURL, "text", false))
	b.WriteString(`</div>`)

	cal, pro, car, fat := "", "", "", ""
	if r.Nutrition != nil {
		cal, pro, car, fat = numOrEmpty(r.Nutrition.Calories), numOrEmpty(r.Nutrition.Protein), numOrEmpty(r.Nutrition.Carbs), numOrEmpty(r.Nutrition.Fats)
	}
	b.WriteString(`<div class="grid grid-cols-2 sm:grid-cols-4 gap-4">`)
	b.WriteString(fieldInput("Calories", "nutrition.calories", cal, "number", false))
	b.WriteString(fieldInput("Protein (g)", "nutrition.protein", pro, "number", false))
	b.WriteString(fieldInput("Carbs (g)", "nutrition.carbs", car, "number", false))
	b.WriteString(fieldInput("Fats (g)", "nutrition.fats", fat, "number", false))
	b.WriteString(`</div>`)

	// Ingredients
	b.WriteString(`<div class="border-t pt-4"><div class="flex justify-between items-center mb-3"><h4 class="text-lg font-semibold">Ingredients</h4>` +
		`<button type="button" onclick="addIngredientRow()" class="bg-gray-600 text-white px-3 py-1.5 rounded-md hover:bg-gray-700 text-sm">+ Add</button></div>`)
	b.WriteString(`<div id="recipe-ingredients-list" class="space-y-2">`)
	if len(r.Ingredients) == 0 {
		b.WriteString(editIngredientRow(0, RecipeIngredient{}))
	} else {
		for i, ing := range r.Ingredients {
			b.WriteString(editIngredientRow(i, ing))
		}
	}
	b.WriteString(`</div></div>`)

	// Steps
	b.WriteString(`<div class="border-t pt-4"><div class="flex justify-between items-center mb-3"><h4 class="text-lg font-semibold">Cooking Steps</h4>` +
		`<button type="button" onclick="addStepRow()" class="bg-gray-600 text-white px-3 py-1.5 rounded-md hover:bg-gray-700 text-sm">+ Add</button></div>`)
	b.WriteString(`<div id="recipe-steps-list" class="space-y-2">`)
	if len(r.Steps) == 0 {
		b.WriteString(editStepRow(0, 1, ""))
	} else {
		for i, st := range r.Steps {
			b.WriteString(editStepRow(i, st.StepNumber, st.Instruction))
		}
	}
	b.WriteString(`</div></div>`)

	b.WriteString(`<div class="flex gap-3 pt-2">`)
	b.WriteString(`<button type="submit" class="bg-emerald-600 text-white px-6 py-2 rounded-md hover:bg-emerald-700 transition">Save Changes</button>`)
	b.WriteString(`<a href="/recipes/` + strconv.Itoa(r.ID) + `" class="px-6 py-2 rounded-lg bg-gray-200 text-gray-700 hover:bg-gray-300 transition">Cancel</a>`)
	b.WriteString(`</div>`)

	b.WriteString(`</form></div>`)
	// Wire the comboboxes (loads ingredient list + attaches type-ahead).
	b.WriteString(`<script>document.addEventListener('DOMContentLoaded', function(){ fillIngredientSelects(); });</script>`)
	return b.String()
}

func fieldInput(label, name, value, inputType string, required bool) string {
	req := ""
	if required {
		req = " required"
	}
	return `<div><label class="block text-sm font-medium mb-1">` + label + `</label>` +
		`<input type="` + inputType + `"` + req + ` name="` + name + `" value="` + esc(value) + `" step="0.1" class="w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-emerald-500"></div>`
}

func numOrEmpty(f float64) string {
	if f == 0 {
		return ""
	}
	return formatQty(f)
}

// editIngredientRow renders one pre-filled ingredient row using a combobox
// (text + hidden id) so the user can type or re-pick.
func editIngredientRow(idx int, ing RecipeIngredient) string {
	checked := ""
	if ing.Mandatory {
		checked = " checked"
	}
	i := strconv.Itoa(idx)
	return `<div class="ingredient-row flex flex-wrap gap-2 items-end">` +
		`<div class="ingredient-combo relative flex-1 min-w-[180px]">` +
		`<input type="text" class="ic-input w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-emerald-500" value="` + esc(ing.IngredientName) + `" placeholder="Type or search ingredient..." autocomplete="off">` +
		`<input type="hidden" name="ingredients.` + i + `.ingredient_id" class="ic-id" value="` + strconv.Itoa(ing.IngredientID) + `">` +
		`<div class="ic-list absolute z-20 left-0 right-0 mt-1 bg-white border rounded-md shadow-lg max-h-60 overflow-auto hidden"></div>` +
		`</div>` +
		`<div class="w-24"><label class="block text-xs font-medium mb-1">Qty</label><input type="number" step="0.01" min="0.01" name="ingredients.` + i + `.quantity" required value="` + formatQty(ing.Quantity) + `" class="w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-emerald-500"></div>` +
		`<div class="w-24"><label class="block text-xs font-medium mb-1">Unit</label>` + unitOptionsGo("ingredients."+i+".unit", ing.Unit, true) + `</div>` +
		`<div class="w-24 flex flex-col"><label class="block text-xs font-medium mb-1">Required</label><input type="checkbox" name="ingredients.` + i + `.mandatory"` + checked + ` class="mt-2 h-5 w-5" title="Uncheck to mark optional"></div>` +
		`<button type="button" onclick="removeIngredientRow(this)" class="bg-red-600 text-white px-3 py-2 rounded-md hover:bg-red-700">\u00d7</button>` +
		`</div>`
}

func editStepRow(idx, num int, instr string) string {
	i := strconv.Itoa(idx)
	return `<div class="step-row flex gap-2 items-start">` +
		`<div class="w-16"><label class="block text-xs font-medium mb-1">Step</label><input type="number" name="steps.` + i + `.step_number" required readonly value="` + strconv.Itoa(num) + `" class="w-full px-3 py-2 border rounded-md bg-gray-100 focus:outline-none"></div>` +
		`<div class="flex-1"><label class="block text-xs font-medium mb-1">Instruction</label><input type="text" name="steps.` + i + `.instruction" required value="` + esc(instr) + `" class="w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-emerald-500"></div>` +
		`<button type="button" onclick="removeStepRow(this)" class="bg-red-600 text-white px-3 py-2 rounded-md hover:bg-red-700 mt-5">\u00d7</button>` +
		`</div>`
}

// unitOptionsGo renders a unit <select> with the given option selected
// (server-side counterpart of the JS unitOptionsHTML).
func unitOptionsGo(name, selected string, required bool) string {
	units := []string{"g", "kg", "ml", "l", "pc", "clove", "tbsp", "tsp", "cup", "lb", "oz", "stalk", "leaf", "sprig"}
	req := ""
	if required {
		req = " required"
	}
	var b strings.Builder
	b.WriteString(`<select name="` + name + `"` + req + ` class="w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-emerald-500"><option value="">Unit</option>`)
	for _, u := range units {
		sel := ""
		if u == selected {
			sel = " selected"
		}
		b.WriteString(`<option value="` + u + `"` + sel + `>` + u + `</option>`)
	}
	b.WriteString(`</select>`)
	return b.String()
}

// assetVersion is appended to static asset URLs to bust browser caches after
// updates. Bump this when recipe-form.js changes.
const assetVersion = "4"

// pageShell wraps content in a full HTML document with a working standalone
// navigation (no reliance on a global showSection). It reuses the shared
// static/recipe-form.js for toasts (single source of truth). (see 5.9)
func pageShell(title, content string) string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<link rel="icon" type="image/svg+xml" href="/static/favicon.svg">
<title>` + title + ` - Wareg</title>
<script src="https://cdn.tailwindcss.com"></script>
<script src="https://unpkg.com/htmx.org@1.9.10"></script>
</head>
<body class="bg-gray-50 min-h-screen">
<div id="toast-container" class="fixed top-4 right-4 z-50 space-y-2"></div>
<nav class="bg-emerald-600 text-white shadow-lg">
  <div class="max-w-6xl mx-auto px-4">
    <div class="flex justify-between items-center py-4">
      <a href="/" class="text-2xl font-bold hover:text-emerald-100">Wareg</a>
      <div class="flex gap-2 sm:gap-4 flex-wrap">
        <a href="/" class="px-3 py-2 hover:bg-emerald-700 rounded">Dashboard</a>
        <a href="/recipes" class="px-3 py-2 bg-emerald-700 rounded">Recipes</a>
      </div>
    </div>
  </div>
</nav>
<main class="max-w-4xl mx-auto px-4 py-8">
` + content + `
</main>
<script src="/static/recipe-form.js?v=` + assetVersion + `"></script>
</body>
</html>`
}
