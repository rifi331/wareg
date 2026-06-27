-- Wareg database schema.
-- Run with:  psql "<conn>" -f sql/schema.sql
--
-- WARNING: this script DROPs and recreates every table (required by the
-- RECREATE_PROMPT clean-room spec). Re-running it on an existing database
-- DESTROYS all data. For an already-populated database use idempotent
-- migrations (CREATE TABLE IF NOT EXISTS / ALTER) instead of this file.
--
-- Fixes the original: ingredients now carry a default `unit` column, and every
-- recipe/pantry row specifies its own per-row unit (see RECREATE_PROMPT 5.6).

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
    mandatory     BOOLEAN       NOT NULL DEFAULT TRUE,  -- optional ingredients are ignored by the match engine
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
    id               SERIAL PRIMARY KEY,
    ingredient_id    INTEGER       NOT NULL REFERENCES ingredients(id) ON DELETE CASCADE,
    current_quantity NUMERIC(10,2) NOT NULL DEFAULT 0,
    unit             VARCHAR(50)   NOT NULL,
    -- Same ingredient may be stocked in multiple forms (e.g. ml + pieces),
    -- so uniqueness is per (ingredient_id, unit), not ingredient_id alone.
    UNIQUE (ingredient_id, unit)
);

CREATE TABLE meal_plan (
    id         SERIAL PRIMARY KEY,
    recipe_id  INTEGER NOT NULL REFERENCES recipes(id) ON DELETE CASCADE,
    plan_date  DATE    NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (plan_date, recipe_id)
);

CREATE INDEX idx_recipe_ingredients_recipe     ON recipe_ingredients(recipe_id);
CREATE INDEX idx_recipe_ingredients_ingredient ON recipe_ingredients(ingredient_id);
CREATE INDEX idx_recipe_steps_recipe           ON recipe_steps(recipe_id);
CREATE INDEX idx_meal_plan_date                ON meal_plan(plan_date);
