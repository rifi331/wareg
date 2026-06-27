-- Wareg sample data.
-- Run AFTER schema.sql with:  psql "<conn>" -f sql/seed_data.sql
--
-- Bug fixes vs. the original seed (see RECREATE_PROMPT 5.5):
--   * Wrapped in a single transaction so a mid-seed failure leaves NO partial data.
--   * Every ingredient referenced by a recipe is inserted FIRST. The original
--     referenced `Mussel` and `Saffron` (Paella) that were never inserted, which
--     aborted the whole seed via the foreign-key constraint.
--   * Removed a duplicate `Nutmeg` that violated the ingredients.name UNIQUE
--     constraint.

SET search_path TO wareg;

BEGIN;

-- Clear existing data (order respects FKs; CASCADE on parents also covers it).
DELETE FROM meal_plan;
DELETE FROM pantry;
DELETE FROM recipe_steps;
DELETE FROM recipe_ingredients;
DELETE FROM recipes;
DELETE FROM ingredients;

-- Reset sequences so recipe/ingredient IDs line up with the seed below.
SELECT setval('ingredients_id_seq',           1, false);
SELECT setval('recipes_id_seq',               1, false);
SELECT setval('recipe_ingredients_id_seq',    1, false);
SELECT setval('recipe_steps_id_seq',          1, false);
SELECT setval('pantry_id_seq',                1, false);
SELECT setval('meal_plan_id_seq',             1, false);

-- ============================================
-- INGREDIENTS (default unit on the catalog; per-row units on recipes/pantry)
-- ============================================

-- Southeast Asian Proteins
INSERT INTO ingredients (name, unit) VALUES
('Chicken', 'g'),
('Beef', 'g'),
('Pork', 'g'),
('Fish', 'g'),
('Shrimp', 'g'),
('Tofu', 'g'),
('Tempeh', 'g'),
('Duck', 'g'),
('Squid', 'g'),
('Crab', 'g');

-- Southeast Asian Vegetables & Herbs
INSERT INTO ingredients (name, unit) VALUES
('Bok Choy', 'g'),
('Napa Cabbage', 'g'),
('Bean Sprouts', 'g'),
('Morning Glory', 'g'),
('Thai Basil', 'cup'),
('Kaffir Lime Leaves', 'leaf'),
('Lemongrass', 'stalk'),
('Galangal', 'g'),
('Ginger', 'g'),
('Garlic', 'clove'),
('Shallots', 'pc'),
('Thai Chilies', 'pc'),
('Bird''s Eye Chili', 'pc'),
('Thai Eggplant', 'g'),
('Bamboo Shoots', 'g'),
('Water Spinach', 'g'),
('Long Beans', 'g'),
('Cucumber', 'pc'),
('Tomato', 'pc'),
('Lime', 'pc'),
('Coriander', 'cup'),
('Mint', 'cup'),
('Spring Onions', 'cup');

-- Southeast Asian Rice & Noodles
INSERT INTO ingredients (name, unit) VALUES
('Jasmine Rice', 'g'),
('Sticky Rice', 'g'),
('Rice Noodles', 'g'),
('Egg Noodles', 'g'),
('Glass Noodles', 'g'),
('Rice Flour', 'g'),
('Coconut Milk', 'ml'),
('Coconut Cream', 'ml');

-- Southeast Asian Sauces & Condiments
INSERT INTO ingredients (name, unit) VALUES
('Fish Sauce', 'tbsp'),
('Soy Sauce', 'tbsp'),
('Oyster Sauce', 'tbsp'),
('Sweet Soy Sauce', 'tbsp'),
('Tamarind Paste', 'tbsp'),
('Shrimp Paste', 'tbsp'),
('Palm Sugar', 'tbsp'),
('Coconut Sugar', 'tbsp'),
('Sambal Oelek', 'tbsp'),
('Chili Paste', 'tbsp'),
('Rice Vinegar', 'tbsp'),
('Sesame Oil', 'tbsp');

-- Southeast Asian Spices
INSERT INTO ingredients (name, unit) VALUES
('Turmeric', 'g'),
('Coriander Seeds', 'tsp'),
('Cumin', 'tsp'),
('Cardamom', 'pc'),
('Star Anise', 'pc'),
('Cinnamon', 'stick'),
('Cloves', 'pc'),
('Nutmeg', 'tsp'),
('Black Pepper', 'tsp'),
('White Pepper', 'tsp'),
('Salt', 'tsp');

-- European Proteins
INSERT INTO ingredients (name, unit) VALUES
('Salmon', 'g'),
('Cod', 'g'),
('Bacon', 'g'),
('Ham', 'g'),
('Sausage', 'g'),
('Lamb', 'g'),
('Veal', 'g'),
('Turkey', 'g'),
('Eggs', 'pc');

-- European Vegetables
INSERT INTO ingredients (name, unit) VALUES
('Onion', 'pc'),
('Carrot', 'pc'),
('Celery', 'stalk'),
('Potato', 'pc'),
('Bell Pepper', 'pc'),
('Mushroom', 'g'),
('Spinach', 'g'),
('Zucchini', 'pc'),
('Eggplant', 'pc'),
('Leek', 'pc'),
('Asparagus', 'g'),
('Broccoli', 'g'),
('Cauliflower', 'g'),
('Green Beans', 'g'),
('Peas', 'g'),
('Corn', 'g'),
('Artichoke', 'pc'),
('Fennel', 'pc'),
('Radish', 'pc'),
('Turnip', 'pc');

-- European Grains & Baking
INSERT INTO ingredients (name, unit) VALUES
('Pasta', 'g'),
('Bread', 'pc'),
('All-Purpose Flour', 'g'),
('Bread Flour', 'g'),
('Oats', 'g'),
('Quinoa', 'g'),
('Couscous', 'g'),
('Polenta', 'g'),
('Breadcrumbs', 'g');

-- European Dairy
INSERT INTO ingredients (name, unit) VALUES
('Milk', 'ml'),
('Butter', 'g'),
('Parmesan Cheese', 'g'),
('Mozzarella Cheese', 'g'),
('Cheddar Cheese', 'g'),
('Cream', 'ml'),
('Heavy Cream', 'ml'),
('Sour Cream', 'g'),
('Yogurt', 'g'),
('Ricotta', 'g'),
('Feta Cheese', 'g'),
('Goat Cheese', 'g'),
('Cream Cheese', 'g');

-- European Oils & Fats
INSERT INTO ingredients (name, unit) VALUES
('Olive Oil', 'tbsp'),
('Vegetable Oil', 'tbsp'),
('Lard', 'g'),
('Duck Fat', 'g');

-- European Herbs
INSERT INTO ingredients (name, unit) VALUES
('Basil', 'cup'),
('Oregano', 'tsp'),
('Thyme', 'sprig'),
('Rosemary', 'sprig'),
('Parsley', 'cup'),
('Sage', 'leaf'),
('Bay Leaf', 'leaf'),
('Dill', 'cup'),
('Tarragon', 'cup'),
('Chives', 'cup');

-- European Spices & Seasonings (Nutmeg already inserted above -> not repeated)
INSERT INTO ingredients (name, unit) VALUES
('Paprika', 'tsp'),
('Smoked Paprika', 'tsp'),
('Cayenne Pepper', 'tsp'),
('Mustard', 'tsp'),
('Dijon Mustard', 'tsp'),
('Worcestershire Sauce', 'tbsp'),
('Vanilla', 'tsp'),
('Allspice', 'tsp');

-- Seafood & specialty (referenced by Paella; MUST exist before recipe rows)
INSERT INTO ingredients (name, unit) VALUES
('Mussel', 'g'),
('Saffron', 'g');

-- Common Ingredients (Both Cuisines)
INSERT INTO ingredients (name, unit) VALUES
('Sugar', 'tbsp'),
('Brown Sugar', 'tbsp'),
('Honey', 'tbsp'),
('Maple Syrup', 'tbsp'),
('Vinegar', 'tbsp'),
('Balsamic Vinegar', 'tbsp'),
('Apple Cider Vinegar', 'tbsp'),
('White Wine', 'ml'),
('Red Wine', 'ml'),
('Chicken Stock', 'ml'),
('Beef Stock', 'ml'),
('Vegetable Stock', 'ml'),
('Cornstarch', 'tbsp'),
('Baking Powder', 'tsp'),
('Baking Soda', 'tsp'),
('Yeast', 'tsp'),
('Water', 'ml');

-- Indonesian specialty ingredients (used by the Indonesian recipes below)
INSERT INTO ingredients (name, unit) VALUES
('Candlenut', 'pc'),
('Fennel Seeds', 'tsp'),
('Chayote', 'pc'),
('Pumpkin', 'g'),
('Keluak (Black Nut)', 'pc'),
('Fried Shallots', 'tbsp'),
('Peanuts', 'g'),
('Cabbage', 'g');

-- ============================================
-- RECIPES
-- ============================================

-- Recipe 1: Nasi Goreng (Indonesian Fried Rice)
INSERT INTO recipes (title, description, instructions, nutrition_json) VALUES
('Nasi Goreng', 'Indonesian fried rice with sweet soy sauce',
'Stir-fry garlic and shallots. Add chicken and cook. Add rice and sweet soy sauce. Push to side and fry egg. Mix well and serve with cucumber and tomato.',
'{"calories": 450, "protein": 20, "carbs": 55, "fats": 15}');

INSERT INTO recipe_ingredients (recipe_id, ingredient_id, quantity, unit) VALUES
(1, (SELECT id FROM ingredients WHERE name = 'Jasmine Rice'), 400, 'g'),
(1, (SELECT id FROM ingredients WHERE name = 'Chicken'), 150, 'g'),
(1, (SELECT id FROM ingredients WHERE name = 'Eggs'), 2, 'pc'),
(1, (SELECT id FROM ingredients WHERE name = 'Garlic'), 3, 'clove'),
(1, (SELECT id FROM ingredients WHERE name = 'Shallots'), 4, 'pc'),
(1, (SELECT id FROM ingredients WHERE name = 'Sweet Soy Sauce'), 3, 'tbsp'),
(1, (SELECT id FROM ingredients WHERE name = 'Soy Sauce'), 1, 'tbsp'),
(1, (SELECT id FROM ingredients WHERE name = 'Vegetable Oil'), 3, 'tbsp'),
(1, (SELECT id FROM ingredients WHERE name = 'Cucumber'), 0.5, 'pc'),
(1, (SELECT id FROM ingredients WHERE name = 'Tomato'), 1, 'pc');

INSERT INTO recipe_steps (recipe_id, step_number, instruction) VALUES
(1, 1, 'Heat oil in a wok over high heat'),
(1, 2, 'Add minced garlic and sliced shallots, stir-fry until fragrant'),
(1, 3, 'Add diced chicken and cook until browned'),
(1, 4, 'Add cold rice and stir-fry for 2-3 minutes'),
(1, 5, 'Add sweet soy sauce and soy sauce, mix well'),
(1, 6, 'Push rice to the side, crack egg into the wok and scramble'),
(1, 7, 'Mix egg with rice and serve with cucumber and tomato slices');

-- Recipe 2: Tom Yum Goong (Thai Shrimp Soup)
INSERT INTO recipes (title, description, instructions, nutrition_json) VALUES
('Tom Yum Goong', 'Spicy and sour Thai shrimp soup with lemongrass and galangal',
'Boil stock with lemongrass, galangal, and kaffir lime leaves. Add shrimp and mushrooms. Season with fish sauce and lime juice. Add chilies and serve.',
'{"calories": 180, "protein": 18, "carbs": 12, "fats": 6}');

INSERT INTO recipe_ingredients (recipe_id, ingredient_id, quantity, unit) VALUES
(2, (SELECT id FROM ingredients WHERE name = 'Shrimp'), 200, 'g'),
(2, (SELECT id FROM ingredients WHERE name = 'Mushroom'), 100, 'g'),
(2, (SELECT id FROM ingredients WHERE name = 'Lemongrass'), 2, 'stalk'),
(2, (SELECT id FROM ingredients WHERE name = 'Galangal'), 50, 'g'),
(2, (SELECT id FROM ingredients WHERE name = 'Kaffir Lime Leaves'), 4, 'leaf'),
(2, (SELECT id FROM ingredients WHERE name = 'Fish Sauce'), 3, 'tbsp'),
(2, (SELECT id FROM ingredients WHERE name = 'Lime'), 2, 'pc'),
(2, (SELECT id FROM ingredients WHERE name = 'Thai Chilies'), 3, 'pc'),
(2, (SELECT id FROM ingredients WHERE name = 'Chicken Stock'), 500, 'ml'),
(2, (SELECT id FROM ingredients WHERE name = 'Coriander'), 0.25, 'cup');

INSERT INTO recipe_steps (recipe_id, step_number, instruction) VALUES
(2, 1, 'Bring chicken stock to a boil'),
(2, 2, 'Add bruised lemongrass, sliced galangal, and kaffir lime leaves'),
(2, 3, 'Simmer for 5 minutes to infuse flavors'),
(2, 4, 'Add mushrooms and cook for 2 minutes'),
(2, 5, 'Add shrimp and cook until pink (2-3 minutes)'),
(2, 6, 'Season with fish sauce and lime juice'),
(2, 7, 'Add sliced Thai chilies and garnish with coriander');

-- Recipe 3: Rendang (Indonesian Beef Curry)
INSERT INTO recipes (title, description, instructions, nutrition_json) VALUES
('Rendang', 'Slow-cooked Indonesian beef in coconut milk and spices',
'Blend spice paste. Fry paste until fragrant. Add beef and coconut milk. Cook slowly until sauce thickens and meat is tender.',
'{"calories": 520, "protein": 35, "carbs": 8, "fats": 40}');

INSERT INTO recipe_ingredients (recipe_id, ingredient_id, quantity, unit) VALUES
(3, (SELECT id FROM ingredients WHERE name = 'Beef'), 500, 'g'),
(3, (SELECT id FROM ingredients WHERE name = 'Coconut Milk'), 400, 'ml'),
(3, (SELECT id FROM ingredients WHERE name = 'Garlic'), 6, 'clove'),
(3, (SELECT id FROM ingredients WHERE name = 'Shallots'), 8, 'pc'),
(3, (SELECT id FROM ingredients WHERE name = 'Ginger'), 50, 'g'),
(3, (SELECT id FROM ingredients WHERE name = 'Galangal'), 50, 'g'),
(3, (SELECT id FROM ingredients WHERE name = 'Turmeric'), 30, 'g'),
(3, (SELECT id FROM ingredients WHERE name = 'Cinnamon'), 1, 'stick'),
(3, (SELECT id FROM ingredients WHERE name = 'Star Anise'), 2, 'pc'),
(3, (SELECT id FROM ingredients WHERE name = 'Kaffir Lime Leaves'), 4, 'leaf'),
(3, (SELECT id FROM ingredients WHERE name = 'Thai Chilies'), 5, 'pc'),
(3, (SELECT id FROM ingredients WHERE name = 'Vegetable Oil'), 3, 'tbsp');

INSERT INTO recipe_steps (recipe_id, step_number, instruction) VALUES
(3, 1, 'Blend garlic, shallots, ginger, galangal, turmeric, and chilies into a paste'),
(3, 2, 'Heat oil and fry spice paste until fragrant (5-7 minutes)'),
(3, 3, 'Add cinnamon, star anise, and kaffir lime leaves'),
(3, 4, 'Add cubed beef and stir to coat with spices'),
(3, 5, 'Pour in coconut milk and bring to a boil'),
(3, 6, 'Reduce heat and simmer uncovered for 2-3 hours, stirring occasionally'),
(3, 7, 'Cook until sauce thickens and coats the meat, and oil separates');

-- Recipe 4: Spaghetti Carbonara (Italian)
INSERT INTO recipes (title, description, instructions, nutrition_json) VALUES
('Spaghetti Carbonara', 'Classic Italian pasta with eggs, cheese, and pancetta',
'Cook pasta. Fry bacon until crispy. Mix eggs with cheese. Combine hot pasta with egg mixture and bacon. Serve immediately.',
'{"calories": 650, "protein": 28, "carbs": 65, "fats": 30}');

INSERT INTO recipe_ingredients (recipe_id, ingredient_id, quantity, unit) VALUES
(4, (SELECT id FROM ingredients WHERE name = 'Pasta'), 400, 'g'),
(4, (SELECT id FROM ingredients WHERE name = 'Bacon'), 200, 'g'),
(4, (SELECT id FROM ingredients WHERE name = 'Eggs'), 4, 'pc'),
(4, (SELECT id FROM ingredients WHERE name = 'Parmesan Cheese'), 100, 'g'),
(4, (SELECT id FROM ingredients WHERE name = 'Black Pepper'), 1, 'tsp'),
(4, (SELECT id FROM ingredients WHERE name = 'Salt'), 0.5, 'tsp');

INSERT INTO recipe_steps (recipe_id, step_number, instruction) VALUES
(4, 1, 'Cook spaghetti in salted boiling water until al dente'),
(4, 2, 'While pasta cooks, cut bacon into small strips and fry until crispy'),
(4, 3, 'Whisk eggs with grated Parmesan and plenty of black pepper'),
(4, 4, 'Reserve 1 cup pasta water, then drain pasta'),
(4, 5, 'Add hot pasta to the pan with bacon (off heat)'),
(4, 6, 'Quickly pour egg mixture over pasta and toss vigorously'),
(4, 7, 'Add pasta water if needed to create creamy sauce. Serve immediately');

-- Recipe 5: Beef Bourguignon (French)
INSERT INTO recipes (title, description, instructions, nutrition_json) VALUES
('Beef Bourguignon', 'French beef stew with red wine, mushrooms, and bacon',
'Brown beef in batches. Cook bacon and vegetables. Add wine and stock. Simmer until tender. Add mushrooms and pearl onions.',
'{"calories": 580, "protein": 42, "carbs": 18, "fats": 32}');

INSERT INTO recipe_ingredients (recipe_id, ingredient_id, quantity, unit) VALUES
(5, (SELECT id FROM ingredients WHERE name = 'Beef'), 800, 'g'),
(5, (SELECT id FROM ingredients WHERE name = 'Bacon'), 150, 'g'),
(5, (SELECT id FROM ingredients WHERE name = 'Red Wine'), 750, 'ml'),
(5, (SELECT id FROM ingredients WHERE name = 'Beef Stock'), 500, 'ml'),
(5, (SELECT id FROM ingredients WHERE name = 'Onion'), 2, 'pc'),
(5, (SELECT id FROM ingredients WHERE name = 'Carrot'), 2, 'pc'),
(5, (SELECT id FROM ingredients WHERE name = 'Garlic'), 4, 'clove'),
(5, (SELECT id FROM ingredients WHERE name = 'Mushroom'), 250, 'g'),
(5, (SELECT id FROM ingredients WHERE name = 'Butter'), 50, 'g'),
(5, (SELECT id FROM ingredients WHERE name = 'Thyme'), 4, 'sprig'),
(5, (SELECT id FROM ingredients WHERE name = 'Bay Leaf'), 2, 'leaf');

INSERT INTO recipe_steps (recipe_id, step_number, instruction) VALUES
(5, 1, 'Cut beef into 2-inch cubes and pat dry'),
(5, 2, 'Brown beef in batches in a heavy pot, set aside'),
(5, 3, 'Cook diced bacon until crispy, add to beef'),
(5, 4, 'Sauté sliced onions and carrots in the same pot'),
(5, 5, 'Add minced garlic and cook for 1 minute'),
(5, 6, 'Return beef and bacon to pot, add wine and stock'),
(5, 7, 'Add thyme and bay leaves, bring to a boil'),
(5, 8, 'Cover and simmer in oven at 325°F for 2-3 hours'),
(5, 9, 'Sauté mushrooms in butter and add to stew before serving');

-- Recipe 6: Pad Thai
INSERT INTO recipes (title, description, instructions, nutrition_json) VALUES
('Pad Thai', 'Thai stir-fried rice noodles with shrimp and peanuts',
'Soak noodles. Stir-fry shrimp and tofu. Add noodles and sauce. Push to side, scramble egg. Mix and serve with peanuts and lime.',
'{"calories": 550, "protein": 22, "carbs": 70, "fats": 18}');

INSERT INTO recipe_ingredients (recipe_id, ingredient_id, quantity, unit) VALUES
(6, (SELECT id FROM ingredients WHERE name = 'Rice Noodles'), 200, 'g'),
(6, (SELECT id FROM ingredients WHERE name = 'Shrimp'), 150, 'g'),
(6, (SELECT id FROM ingredients WHERE name = 'Tofu'), 100, 'g'),
(6, (SELECT id FROM ingredients WHERE name = 'Eggs'), 2, 'pc'),
(6, (SELECT id FROM ingredients WHERE name = 'Bean Sprouts'), 100, 'g'),
(6, (SELECT id FROM ingredients WHERE name = 'Garlic'), 3, 'clove'),
(6, (SELECT id FROM ingredients WHERE name = 'Fish Sauce'), 2, 'tbsp'),
(6, (SELECT id FROM ingredients WHERE name = 'Tamarind Paste'), 2, 'tbsp'),
(6, (SELECT id FROM ingredients WHERE name = 'Palm Sugar'), 2, 'tbsp'),
(6, (SELECT id FROM ingredients WHERE name = 'Vegetable Oil'), 3, 'tbsp'),
(6, (SELECT id FROM ingredients WHERE name = 'Lime'), 1, 'pc'),
(6, (SELECT id FROM ingredients WHERE name = 'Spring Onions'), 0.25, 'cup');

INSERT INTO recipe_steps (recipe_id, step_number, instruction) VALUES
(6, 1, 'Soak rice noodles in warm water for 30 minutes'),
(6, 2, 'Mix fish sauce, tamarind paste, and palm sugar for the sauce'),
(6, 3, 'Heat oil in a wok over high heat'),
(6, 4, 'Stir-fry garlic, shrimp, and cubed tofu until shrimp is pink'),
(6, 5, 'Add drained noodles and sauce, stir-fry for 2 minutes'),
(6, 6, 'Push to side, add beaten egg and scramble'),
(6, 7, 'Mix everything together, add bean sprouts'),
(6, 8, 'Serve with crushed peanuts, lime wedges, and spring onions');

-- Recipe 7: Chicken Tikka Masala (British-Indian)
INSERT INTO recipes (title, description, instructions, nutrition_json) VALUES
('Chicken Tikka Masala', 'Grilled chicken in creamy tomato curry sauce',
'Marinate chicken in yogurt and spices. Grill or bake. Make sauce with tomatoes, cream, and spices. Combine and serve with rice.',
'{"calories": 620, "protein": 38, "carbs": 25, "fats": 40}');

INSERT INTO recipe_ingredients (recipe_id, ingredient_id, quantity, unit) VALUES
(7, (SELECT id FROM ingredients WHERE name = 'Chicken'), 600, 'g'),
(7, (SELECT id FROM ingredients WHERE name = 'Yogurt'), 150, 'g'),
(7, (SELECT id FROM ingredients WHERE name = 'Heavy Cream'), 200, 'ml'),
(7, (SELECT id FROM ingredients WHERE name = 'Tomato'), 400, 'g'),
(7, (SELECT id FROM ingredients WHERE name = 'Onion'), 2, 'pc'),
(7, (SELECT id FROM ingredients WHERE name = 'Garlic'), 4, 'clove'),
(7, (SELECT id FROM ingredients WHERE name = 'Ginger'), 30, 'g'),
(7, (SELECT id FROM ingredients WHERE name = 'Butter'), 50, 'g'),
(7, (SELECT id FROM ingredients WHERE name = 'Cumin'), 1, 'tsp'),
(7, (SELECT id FROM ingredients WHERE name = 'Turmeric'), 1, 'tsp'),
(7, (SELECT id FROM ingredients WHERE name = 'Paprika'), 1, 'tsp'),
(7, (SELECT id FROM ingredients WHERE name = 'Lime'), 1, 'pc'),
(7, (SELECT id FROM ingredients WHERE name = 'Coriander'), 0.25, 'cup');

INSERT INTO recipe_steps (recipe_id, step_number, instruction) VALUES
(7, 1, 'Mix yogurt with half the spices, marinate chicken for 2 hours'),
(7, 2, 'Thread chicken on skewers and grill or bake at 400°F for 20 minutes'),
(7, 3, 'Melt butter in a pan, sauté sliced onions until golden'),
(7, 4, 'Add minced garlic and ginger, cook for 2 minutes'),
(7, 5, 'Add remaining spices and cook for 1 minute'),
(7, 6, 'Add chopped tomatoes and simmer for 15 minutes'),
(7, 7, 'Blend sauce until smooth, return to pan'),
(7, 8, 'Stir in cream, add grilled chicken pieces'),
(7, 9, 'Simmer for 5 minutes, garnish with coriander and lime juice');

-- Recipe 8: Green Curry (Thai)
INSERT INTO recipes (title, description, instructions, nutrition_json) VALUES
('Green Curry', 'Thai green curry with chicken and vegetables',
'Make green curry paste. Fry paste in coconut cream. Add chicken and vegetables. Season with fish sauce and palm sugar.',
'{"calories": 480, "protein": 30, "carbs": 15, "fats": 35}');

INSERT INTO recipe_ingredients (recipe_id, ingredient_id, quantity, unit) VALUES
(8, (SELECT id FROM ingredients WHERE name = 'Chicken'), 400, 'g'),
(8, (SELECT id FROM ingredients WHERE name = 'Coconut Milk'), 400, 'ml'),
(8, (SELECT id FROM ingredients WHERE name = 'Thai Eggplant'), 100, 'g'),
(8, (SELECT id FROM ingredients WHERE name = 'Bamboo Shoots'), 100, 'g'),
(8, (SELECT id FROM ingredients WHERE name = 'Thai Basil'), 1, 'cup'),
(8, (SELECT id FROM ingredients WHERE name = 'Kaffir Lime Leaves'), 4, 'leaf'),
(8, (SELECT id FROM ingredients WHERE name = 'Thai Chilies'), 4, 'pc'),
(8, (SELECT id FROM ingredients WHERE name = 'Garlic'), 4, 'clove'),
(8, (SELECT id FROM ingredients WHERE name = 'Shallots'), 4, 'pc'),
(8, (SELECT id FROM ingredients WHERE name = 'Galangal'), 30, 'g'),
(8, (SELECT id FROM ingredients WHERE name = 'Lemongrass'), 1, 'stalk'),
(8, (SELECT id FROM ingredients WHERE name = 'Fish Sauce'), 2, 'tbsp'),
(8, (SELECT id FROM ingredients WHERE name = 'Palm Sugar'), 1, 'tbsp');

INSERT INTO recipe_steps (recipe_id, step_number, instruction) VALUES
(8, 1, 'Blend chilies, garlic, shallots, galangal, and lemongrass into a paste'),
(8, 2, 'Scoop thick cream from top of coconut milk into a hot wok'),
(8, 3, 'Fry curry paste in coconut cream until fragrant and oil separates'),
(8, 4, 'Add sliced chicken and stir-fry until sealed'),
(8, 5, 'Pour in remaining coconut milk, bring to a boil'),
(8, 6, 'Add Thai eggplant, bamboo shoots, and kaffir lime leaves'),
(8, 7, 'Simmer for 10 minutes until chicken is cooked'),
(8, 8, 'Season with fish sauce and palm sugar'),
(8, 9, 'Add Thai basil leaves and serve with jasmine rice');

-- Recipe 9: Paella (Spanish) -- now all ingredients exist (Mussel, Saffron inserted above)
INSERT INTO recipes (title, description, instructions, nutrition_json) VALUES
('Paella', 'Spanish rice dish with seafood, chicken, and saffron',
'Brown chicken and seafood. Sauté vegetables. Add rice and saffron. Pour in stock and cook without stirring until rice is tender.',
'{"calories": 680, "protein": 35, "carbs": 70, "fats": 25}');

INSERT INTO recipe_ingredients (recipe_id, ingredient_id, quantity, unit) VALUES
(9, (SELECT id FROM ingredients WHERE name = 'Jasmine Rice'), 400, 'g'),
(9, (SELECT id FROM ingredients WHERE name = 'Chicken'), 300, 'g'),
(9, (SELECT id FROM ingredients WHERE name = 'Shrimp'), 200, 'g'),
(9, (SELECT id FROM ingredients WHERE name = 'Squid'), 150, 'g'),
(9, (SELECT id FROM ingredients WHERE name = 'Mussel'), 200, 'g'),
(9, (SELECT id FROM ingredients WHERE name = 'Onion'), 1, 'pc'),
(9, (SELECT id FROM ingredients WHERE name = 'Garlic'), 4, 'clove'),
(9, (SELECT id FROM ingredients WHERE name = 'Tomato'), 2, 'pc'),
(9, (SELECT id FROM ingredients WHERE name = 'Bell Pepper'), 2, 'pc'),
(9, (SELECT id FROM ingredients WHERE name = 'Green Beans'), 100, 'g'),
(9, (SELECT id FROM ingredients WHERE name = 'Chicken Stock'), 800, 'ml'),
(9, (SELECT id FROM ingredients WHERE name = 'Olive Oil'), 4, 'tbsp'),
(9, (SELECT id FROM ingredients WHERE name = 'Saffron'), 0.5, 'g'),
(9, (SELECT id FROM ingredients WHERE name = 'Paprika'), 1, 'tsp');

INSERT INTO recipe_steps (recipe_id, step_number, instruction) VALUES
(9, 1, 'Heat olive oil in a large paella pan'),
(9, 2, 'Brown chicken pieces, set aside'),
(9, 3, 'Sauté shrimp, squid, and mussels until just cooked, set aside'),
(9, 4, 'Sauté diced onion and garlic until soft'),
(9, 5, 'Add grated tomato and cook for 2 minutes'),
(9, 6, 'Add rice, paprika, and saffron, stir to coat'),
(9, 7, 'Pour in hot chicken stock, arrange chicken on top'),
(9, 8, 'Cook on medium heat without stirring for 20 minutes'),
(9, 9, 'Add seafood and green beans, cook 5 more minutes'),
(9, 10, 'Let rest for 5 minutes before serving');

-- Recipe 10: Laksa (Malaysian Noodle Soup)
INSERT INTO recipes (title, description, instructions, nutrition_json) VALUES
('Laksa', 'Malaysian spicy noodle soup with coconut curry broth',
'Make spice paste. Fry paste in oil. Add coconut milk and stock. Add tofu puffs and shrimp. Serve over noodles with toppings.',
'{"calories": 520, "protein": 25, "carbs": 55, "fats": 24}');

INSERT INTO recipe_ingredients (recipe_id, ingredient_id, quantity, unit) VALUES
(10, (SELECT id FROM ingredients WHERE name = 'Rice Noodles'), 200, 'g'),
(10, (SELECT id FROM ingredients WHERE name = 'Shrimp'), 200, 'g'),
(10, (SELECT id FROM ingredients WHERE name = 'Tofu'), 150, 'g'),
(10, (SELECT id FROM ingredients WHERE name = 'Coconut Milk'), 400, 'ml'),
(10, (SELECT id FROM ingredients WHERE name = 'Chicken Stock'), 500, 'ml'),
(10, (SELECT id FROM ingredients WHERE name = 'Fish Sauce'), 2, 'tbsp'),
(10, (SELECT id FROM ingredients WHERE name = 'Lemongrass'), 2, 'stalk'),
(10, (SELECT id FROM ingredients WHERE name = 'Galangal'), 30, 'g'),
(10, (SELECT id FROM ingredients WHERE name = 'Turmeric'), 20, 'g'),
(10, (SELECT id FROM ingredients WHERE name = 'Shallots'), 6, 'pc'),
(10, (SELECT id FROM ingredients WHERE name = 'Garlic'), 4, 'clove'),
(10, (SELECT id FROM ingredients WHERE name = 'Thai Chilies'), 5, 'pc'),
(10, (SELECT id FROM ingredients WHERE name = 'Bean Sprouts'), 100, 'g'),
(10, (SELECT id FROM ingredients WHERE name = 'Cucumber'), 0.5, 'pc'),
(10, (SELECT id FROM ingredients WHERE name = 'Eggs'), 2, 'pc');

INSERT INTO recipe_steps (recipe_id, step_number, instruction) VALUES
(10, 1, 'Blend lemongrass, galangal, turmeric, shallots, garlic, and chilies into a paste'),
(10, 2, 'Heat oil and fry paste until fragrant and oil separates'),
(10, 3, 'Pour in coconut milk and chicken stock, bring to a boil'),
(10, 4, 'Add cubed tofu and simmer for 10 minutes'),
(10, 5, 'Season with fish sauce to taste'),
(10, 6, 'Cook rice noodles according to package instructions'),
(10, 7, 'Boil eggs for 7 minutes, peel and halve'),
(10, 8, 'Add shrimp to soup and cook until pink'),
(10, 9, 'Serve noodles in bowls, ladle soup over, top with bean sprouts, cucumber, and egg');

-- Recipe 11: Sate Padang
INSERT INTO recipes (title, description, instructions, nutrition_json) VALUES
('Sate Padang', 'Padang-style beef skewers in a thick, spicy yellow sauce',
'Grind spices with candlenut into a paste. Fry paste, add stock and beef, simmer. Thicken with rice flour. Thread beef on skewers and serve drenched in sauce.',
'{"calories": 520, "protein": 32, "carbs": 22, "fats": 30}');

INSERT INTO recipe_ingredients (recipe_id, ingredient_id, quantity, unit) VALUES
(11, (SELECT id FROM ingredients WHERE name = 'Beef'), 500, 'g'),
(11, (SELECT id FROM ingredients WHERE name = 'Beef Stock'), 750, 'ml'),
(11, (SELECT id FROM ingredients WHERE name = 'Rice Flour'), 40, 'g'),
(11, (SELECT id FROM ingredients WHERE name = 'Turmeric'), 30, 'g'),
(11, (SELECT id FROM ingredients WHERE name = 'Ginger'), 30, 'g'),
(11, (SELECT id FROM ingredients WHERE name = 'Galangal'), 30, 'g'),
(11, (SELECT id FROM ingredients WHERE name = 'Garlic'), 6, 'clove'),
(11, (SELECT id FROM ingredients WHERE name = 'Shallots'), 8, 'pc'),
(11, (SELECT id FROM ingredients WHERE name = 'Coriander Seeds'), 2, 'tsp'),
(11, (SELECT id FROM ingredients WHERE name = 'Cumin'), 1, 'tsp'),
(11, (SELECT id FROM ingredients WHERE name = 'Fennel Seeds'), 1, 'tsp'),
(11, (SELECT id FROM ingredients WHERE name = 'Candlenut'), 5, 'pc'),
(11, (SELECT id FROM ingredients WHERE name = 'Lemongrass'), 2, 'stalk'),
(11, (SELECT id FROM ingredients WHERE name = 'Kaffir Lime Leaves'), 4, 'leaf'),
(11, (SELECT id FROM ingredients WHERE name = 'Vegetable Oil'), 3, 'tbsp'),
(11, (SELECT id FROM ingredients WHERE name = 'Salt'), 1, 'tsp');

INSERT INTO recipe_steps (recipe_id, step_number, instruction) VALUES
(11, 1, 'Cut beef into thin strips and blanch briefly in beef stock'),
(11, 2, 'Blend turmeric, ginger, galangal, garlic, shallots, candlenut and spices into a smooth paste'),
(11, 3, 'Heat oil and fry the paste until fragrant'),
(11, 4, 'Pour in beef stock and the beef, then simmer for about 30 minutes'),
(11, 5, 'Dissolve rice flour in a little water and stir in to thicken the sauce'),
(11, 6, 'Thread the cooked beef onto skewers'),
(11, 7, 'Serve the skewers drenched in the thick yellow sauce');

-- Recipe 12: Sayur Lodeh
INSERT INTO recipes (title, description, instructions, nutrition_json) VALUES
('Sayur Lodeh', 'Indonesian mixed vegetables in a rich coconut milk broth',
'Boil coconut milk with spice paste and aromatics. Add hard vegetables first, then soft ones. Season with shrimp paste and palm sugar.',
'{"calories": 240, "protein": 6, "carbs": 22, "fats": 16}');

INSERT INTO recipe_ingredients (recipe_id, ingredient_id, quantity, unit) VALUES
(12, (SELECT id FROM ingredients WHERE name = 'Coconut Milk'), 400, 'ml'),
(12, (SELECT id FROM ingredients WHERE name = 'Green Beans'), 100, 'g'),
(12, (SELECT id FROM ingredients WHERE name = 'Long Beans'), 100, 'g'),
(12, (SELECT id FROM ingredients WHERE name = 'Eggplant'), 1, 'pc'),
(12, (SELECT id FROM ingredients WHERE name = 'Chayote'), 1, 'pc'),
(12, (SELECT id FROM ingredients WHERE name = 'Pumpkin'), 150, 'g'),
(12, (SELECT id FROM ingredients WHERE name = 'Corn'), 1, 'pc'),
(12, (SELECT id FROM ingredients WHERE name = 'Thai Eggplant'), 100, 'g'),
(12, (SELECT id FROM ingredients WHERE name = 'Lemongrass'), 2, 'stalk'),
(12, (SELECT id FROM ingredients WHERE name = 'Galangal'), 30, 'g'),
(12, (SELECT id FROM ingredients WHERE name = 'Kaffir Lime Leaves'), 4, 'leaf'),
(12, (SELECT id FROM ingredients WHERE name = 'Shallots'), 6, 'pc'),
(12, (SELECT id FROM ingredients WHERE name = 'Garlic'), 3, 'clove'),
(12, (SELECT id FROM ingredients WHERE name = 'Shrimp Paste'), 1, 'tsp'),
(12, (SELECT id FROM ingredients WHERE name = 'Palm Sugar'), 1, 'tbsp'),
(12, (SELECT id FROM ingredients WHERE name = 'Salt'), 1, 'tsp');

INSERT INTO recipe_steps (recipe_id, step_number, instruction) VALUES
(12, 1, 'Blend shallots, garlic and shrimp paste into a paste'),
(12, 2, 'Bring coconut milk to a gentle simmer with bruised lemongrass, galangal and lime leaves'),
(12, 3, 'Stir in the spice paste and cook until fragrant'),
(12, 4, 'Add the hard vegetables (chayote, pumpkin, corn) and cook for 5 minutes'),
(12, 5, 'Add eggplant, beans and Thai eggplant, cook until just tender'),
(12, 6, 'Season with palm sugar and salt, serve hot with rice');

-- Recipe 13: Rawon
INSERT INTO recipes (title, description, instructions, nutrition_json) VALUES
('Rawon', 'East Javanese black beef soup made with keluak nuts',
'Simmer beef in stock. Grind keluak with spices into the characteristic black paste. Combine and serve with bean sprouts and fried shallots.',
'{"calories": 420, "protein": 34, "carbs": 12, "fats": 24}');

INSERT INTO recipe_ingredients (recipe_id, ingredient_id, quantity, unit) VALUES
(13, (SELECT id FROM ingredients WHERE name = 'Beef'), 500, 'g'),
(13, (SELECT id FROM ingredients WHERE name = 'Keluak (Black Nut)'), 4, 'pc'),
(13, (SELECT id FROM ingredients WHERE name = 'Garlic'), 6, 'clove'),
(13, (SELECT id FROM ingredients WHERE name = 'Shallots'), 8, 'pc'),
(13, (SELECT id FROM ingredients WHERE name = 'Ginger'), 30, 'g'),
(13, (SELECT id FROM ingredients WHERE name = 'Galangal'), 30, 'g'),
(13, (SELECT id FROM ingredients WHERE name = 'Turmeric'), 20, 'g'),
(13, (SELECT id FROM ingredients WHERE name = 'Lemongrass'), 2, 'stalk'),
(13, (SELECT id FROM ingredients WHERE name = 'Kaffir Lime Leaves'), 4, 'leaf'),
(13, (SELECT id FROM ingredients WHERE name = 'Beef Stock'), 1000, 'ml'),
(13, (SELECT id FROM ingredients WHERE name = 'Bean Sprouts'), 100, 'g'),
(13, (SELECT id FROM ingredients WHERE name = 'Fried Shallots'), 3, 'tbsp'),
(13, (SELECT id FROM ingredients WHERE name = 'Lime'), 1, 'pc'),
(13, (SELECT id FROM ingredients WHERE name = 'Bird''s Eye Chili'), 3, 'pc'),
(13, (SELECT id FROM ingredients WHERE name = 'Salt'), 1, 'tsp');

INSERT INTO recipe_steps (recipe_id, step_number, instruction) VALUES
(13, 1, 'Cut beef into chunks and simmer in beef stock until tender'),
(13, 2, 'Crack open the keluak nuts and scoop out the black flesh'),
(13, 3, 'Blend keluak with garlic, shallots, ginger, galangal and turmeric'),
(13, 4, 'Fry the paste until dark and fragrant'),
(13, 5, 'Add the paste, lemongrass and lime leaves to the soup'),
(13, 6, 'Simmer for 20 minutes, season with salt'),
(13, 7, 'Serve topped with bean sprouts, fried shallots and chili');

-- Recipe 14: Gulai Padang
INSERT INTO recipes (title, description, instructions, nutrition_json) VALUES
('Gulai Padang', 'Rich, spicy Padang-style beef curry in coconut milk',
'Toast and grind the spice blend with candlenut. Fry paste, add beef and coconut milk, simmer until rich and thick.',
'{"calories": 620, "protein": 36, "carbs": 14, "fats": 46}');

INSERT INTO recipe_ingredients (recipe_id, ingredient_id, quantity, unit) VALUES
(14, (SELECT id FROM ingredients WHERE name = 'Beef'), 600, 'g'),
(14, (SELECT id FROM ingredients WHERE name = 'Coconut Milk'), 500, 'ml'),
(14, (SELECT id FROM ingredients WHERE name = 'Garlic'), 6, 'clove'),
(14, (SELECT id FROM ingredients WHERE name = 'Shallots'), 8, 'pc'),
(14, (SELECT id FROM ingredients WHERE name = 'Ginger'), 30, 'g'),
(14, (SELECT id FROM ingredients WHERE name = 'Galangal'), 40, 'g'),
(14, (SELECT id FROM ingredients WHERE name = 'Turmeric'), 30, 'g'),
(14, (SELECT id FROM ingredients WHERE name = 'Coriander Seeds'), 2, 'tsp'),
(14, (SELECT id FROM ingredients WHERE name = 'Cumin'), 1, 'tsp'),
(14, (SELECT id FROM ingredients WHERE name = 'Cardamom'), 3, 'pc'),
(14, (SELECT id FROM ingredients WHERE name = 'Cinnamon'), 1, 'stick'),
(14, (SELECT id FROM ingredients WHERE name = 'Cloves'), 4, 'pc'),
(14, (SELECT id FROM ingredients WHERE name = 'Star Anise'), 2, 'pc'),
(14, (SELECT id FROM ingredients WHERE name = 'Nutmeg'), 1, 'tsp'),
(14, (SELECT id FROM ingredients WHERE name = 'Candlenut'), 4, 'pc'),
(14, (SELECT id FROM ingredients WHERE name = 'Kaffir Lime Leaves'), 4, 'leaf'),
(14, (SELECT id FROM ingredients WHERE name = 'Lemongrass'), 2, 'stalk'),
(14, (SELECT id FROM ingredients WHERE name = 'Chili Paste'), 2, 'tbsp'),
(14, (SELECT id FROM ingredients WHERE name = 'Vegetable Oil'), 3, 'tbsp'),
(14, (SELECT id FROM ingredients WHERE name = 'Salt'), 1, 'tsp');

INSERT INTO recipe_steps (recipe_id, step_number, instruction) VALUES
(14, 1, 'Toast coriander, cumin, cardamom, cloves and star anise briefly'),
(14, 2, 'Blend the toasted spices with shallots, garlic, ginger, galangal, turmeric and candlenut'),
(14, 3, 'Heat oil and fry the paste with chili paste until fragrant'),
(14, 4, 'Add the beef and brown it in the paste'),
(14, 5, 'Pour in coconut milk, add cinnamon, lime leaves and lemongrass'),
(14, 6, 'Simmer covered for 1.5-2 hours until beef is tender and sauce thickens'),
(14, 7, 'Season with salt and nutmeg, serve with rice');

-- Recipe 15: Gulai Jawa
INSERT INTO recipes (title, description, instructions, nutrition_json) VALUES
('Gulai Jawa', 'Javanese yellow gulai, sweeter and aromatic with palm sugar',
'Make a yellow spice paste. Fry and add beef with coconut milk. Sweeten with palm sugar and kecap manis.',
'{"calories": 560, "protein": 34, "carbs": 16, "fats": 38}');

INSERT INTO recipe_ingredients (recipe_id, ingredient_id, quantity, unit) VALUES
(15, (SELECT id FROM ingredients WHERE name = 'Beef'), 500, 'g'),
(15, (SELECT id FROM ingredients WHERE name = 'Coconut Milk'), 500, 'ml'),
(15, (SELECT id FROM ingredients WHERE name = 'Garlic'), 5, 'clove'),
(15, (SELECT id FROM ingredients WHERE name = 'Shallots'), 6, 'pc'),
(15, (SELECT id FROM ingredients WHERE name = 'Ginger'), 25, 'g'),
(15, (SELECT id FROM ingredients WHERE name = 'Galangal'), 30, 'g'),
(15, (SELECT id FROM ingredients WHERE name = 'Turmeric'), 30, 'g'),
(15, (SELECT id FROM ingredients WHERE name = 'Coriander Seeds'), 1, 'tsp'),
(15, (SELECT id FROM ingredients WHERE name = 'Cumin'), 1, 'tsp'),
(15, (SELECT id FROM ingredients WHERE name = 'Lemongrass'), 2, 'stalk'),
(15, (SELECT id FROM ingredients WHERE name = 'Kaffir Lime Leaves'), 4, 'leaf'),
(15, (SELECT id FROM ingredients WHERE name = 'Palm Sugar'), 2, 'tbsp'),
(15, (SELECT id FROM ingredients WHERE name = 'Sweet Soy Sauce'), 1, 'tbsp'),
(15, (SELECT id FROM ingredients WHERE name = 'Vegetable Oil'), 3, 'tbsp'),
(15, (SELECT id FROM ingredients WHERE name = 'Salt'), 1, 'tsp');

INSERT INTO recipe_steps (recipe_id, step_number, instruction) VALUES
(15, 1, 'Blend shallots, garlic, ginger, galangal, turmeric, coriander and cumin'),
(15, 2, 'Heat oil and fry the paste until fragrant and golden'),
(15, 3, 'Add the beef and stir until coated'),
(15, 4, 'Pour in coconut milk with bruised lemongrass and lime leaves'),
(15, 5, 'Simmer until the beef is tender, about 1 hour'),
(15, 6, 'Stir in palm sugar and sweet soy sauce, season with salt');

-- Recipe 16: Sambal Goreng Cumi
INSERT INTO recipes (title, description, instructions, nutrition_json) VALUES
('Sambal Goreng Cumi', 'Spicy squid sauteed in sambal and coconut milk',
'Clean squid. Make a sambal of chili, shallots, garlic and tomato. Fry, add squid briefly, finish with coconut milk.',
'{"calories": 320, "protein": 28, "carbs": 10, "fats": 18}');

INSERT INTO recipe_ingredients (recipe_id, ingredient_id, quantity, unit) VALUES
(16, (SELECT id FROM ingredients WHERE name = 'Squid'), 400, 'g'),
(16, (SELECT id FROM ingredients WHERE name = 'Shallots'), 8, 'pc'),
(16, (SELECT id FROM ingredients WHERE name = 'Garlic'), 4, 'clove'),
(16, (SELECT id FROM ingredients WHERE name = 'Bird''s Eye Chili'), 8, 'pc'),
(16, (SELECT id FROM ingredients WHERE name = 'Thai Chilies'), 4, 'pc'),
(16, (SELECT id FROM ingredients WHERE name = 'Tomato'), 2, 'pc'),
(16, (SELECT id FROM ingredients WHERE name = 'Tamarind Paste'), 2, 'tbsp'),
(16, (SELECT id FROM ingredients WHERE name = 'Palm Sugar'), 1, 'tbsp'),
(16, (SELECT id FROM ingredients WHERE name = 'Shrimp Paste'), 1, 'tsp'),
(16, (SELECT id FROM ingredients WHERE name = 'Coconut Milk'), 200, 'ml'),
(16, (SELECT id FROM ingredients WHERE name = 'Vegetable Oil'), 3, 'tbsp'),
(16, (SELECT id FROM ingredients WHERE name = 'Salt'), 1, 'tsp'),
(16, (SELECT id FROM ingredients WHERE name = 'Lime'), 1, 'pc');

INSERT INTO recipe_steps (recipe_id, step_number, instruction) VALUES
(16, 1, 'Clean the squid and score the bodies in a crosshatch'),
(16, 2, 'Blend shallots, garlic, chilies, tomato and shrimp paste into a sambal'),
(16, 3, 'Heat oil and fry the sambal until the oil separates'),
(16, 4, 'Add the squid and stir-fry for 2-3 minutes only'),
(16, 5, 'Pour in coconut milk, tamarind and palm sugar'),
(16, 6, 'Simmer briefly until the sauce clings, season with salt and lime');

-- Recipe 17: Soto Ayam
INSERT INTO recipes (title, description, instructions, nutrition_json) VALUES
('Soto Ayam', 'Indonesian yellow chicken soup with noodles and herbs',
'Cook chicken in turmeric-spiced stock. Shred chicken. Serve over rice noodles with sprouts, egg and fried shallots.',
'{"calories": 430, "protein": 28, "carbs": 38, "fats": 18}');

INSERT INTO recipe_ingredients (recipe_id, ingredient_id, quantity, unit) VALUES
(17, (SELECT id FROM ingredients WHERE name = 'Chicken'), 500, 'g'),
(17, (SELECT id FROM ingredients WHERE name = 'Chicken Stock'), 1000, 'ml'),
(17, (SELECT id FROM ingredients WHERE name = 'Turmeric'), 20, 'g'),
(17, (SELECT id FROM ingredients WHERE name = 'Garlic'), 5, 'clove'),
(17, (SELECT id FROM ingredients WHERE name = 'Shallots'), 5, 'pc'),
(17, (SELECT id FROM ingredients WHERE name = 'Ginger'), 25, 'g'),
(17, (SELECT id FROM ingredients WHERE name = 'Galangal'), 30, 'g'),
(17, (SELECT id FROM ingredients WHERE name = 'Lemongrass'), 2, 'stalk'),
(17, (SELECT id FROM ingredients WHERE name = 'Kaffir Lime Leaves'), 4, 'leaf'),
(17, (SELECT id FROM ingredients WHERE name = 'Rice Noodles'), 150, 'g'),
(17, (SELECT id FROM ingredients WHERE name = 'Bean Sprouts'), 100, 'g'),
(17, (SELECT id FROM ingredients WHERE name = 'Eggs'), 2, 'pc'),
(17, (SELECT id FROM ingredients WHERE name = 'Celery'), 2, 'stalk'),
(17, (SELECT id FROM ingredients WHERE name = 'Fried Shallots'), 3, 'tbsp'),
(17, (SELECT id FROM ingredients WHERE name = 'Lime'), 1, 'pc'),
(17, (SELECT id FROM ingredients WHERE name = 'Salt'), 1, 'tsp');

INSERT INTO recipe_steps (recipe_id, step_number, instruction) VALUES
(17, 1, 'Blend turmeric, garlic, shallots, ginger and galangal into a paste'),
(17, 2, 'Fry the paste briefly, then add chicken stock, lemongrass and lime leaves'),
(17, 3, 'Add the chicken and poach until cooked through'),
(17, 4, 'Remove chicken, shred the meat, and return bones to the stock to simmer'),
(17, 5, 'Cook rice noodles and blanch bean sprouts'),
(17, 6, 'Boil and halve the eggs'),
(17, 7, 'Assemble bowls with noodles, chicken, egg and sprouts, pour over hot stock'),
(17, 8, 'Top with chopped celery, fried shallots and a squeeze of lime');

-- Recipe 18: Ayam Bakar
INSERT INTO recipes (title, description, instructions, nutrition_json) VALUES
('Ayam Bakar', 'Indonesian grilled chicken marinated in sweet soy and spices',
'Marinate chicken in kecap manis and ground spices. Grill or pan-sear, basting until glazed and charred.',
'{"calories": 480, "protein": 38, "carbs": 14, "fats": 28}');

INSERT INTO recipe_ingredients (recipe_id, ingredient_id, quantity, unit) VALUES
(18, (SELECT id FROM ingredients WHERE name = 'Chicken'), 600, 'g'),
(18, (SELECT id FROM ingredients WHERE name = 'Sweet Soy Sauce'), 4, 'tbsp'),
(18, (SELECT id FROM ingredients WHERE name = 'Garlic'), 5, 'clove'),
(18, (SELECT id FROM ingredients WHERE name = 'Shallots'), 5, 'pc'),
(18, (SELECT id FROM ingredients WHERE name = 'Coriander Seeds'), 1, 'tsp'),
(18, (SELECT id FROM ingredients WHERE name = 'Turmeric'), 10, 'g'),
(18, (SELECT id FROM ingredients WHERE name = 'Tamarind Paste'), 1, 'tbsp'),
(18, (SELECT id FROM ingredients WHERE name = 'Palm Sugar'), 1, 'tbsp'),
(18, (SELECT id FROM ingredients WHERE name = 'Lime'), 1, 'pc'),
(18, (SELECT id FROM ingredients WHERE name = 'Vegetable Oil'), 2, 'tbsp'),
(18, (SELECT id FROM ingredients WHERE name = 'Salt'), 1, 'tsp');

INSERT INTO recipe_steps (recipe_id, step_number, instruction) VALUES
(18, 1, 'Blend garlic, shallots, coriander and turmeric into a paste'),
(18, 2, 'Mix the paste with sweet soy sauce, tamarind, palm sugar and salt'),
(18, 3, 'Coat the chicken and marinate for at least 1 hour'),
(18, 4, 'Grill or pan-sear the chicken, basting with the marinade'),
(18, 5, 'Cook until the skin is charred and glazed'),
(18, 6, 'Rest briefly and serve with lime');

-- Recipe 19: Gado-Gado
INSERT INTO recipes (title, description, instructions, nutrition_json) VALUES
('Gado-Gado', 'Indonesian vegetable salad with peanut sauce',
'Blanch vegetables and boil egg. Crush peanuts with chili, palm sugar and tamarind into a sauce. Dress and serve.',
'{"calories": 380, "protein": 16, "carbs": 26, "fats": 24}');

INSERT INTO recipe_ingredients (recipe_id, ingredient_id, quantity, unit) VALUES
(19, (SELECT id FROM ingredients WHERE name = 'Bean Sprouts'), 100, 'g'),
(19, (SELECT id FROM ingredients WHERE name = 'Green Beans'), 100, 'g'),
(19, (SELECT id FROM ingredients WHERE name = 'Potato'), 2, 'pc'),
(19, (SELECT id FROM ingredients WHERE name = 'Eggs'), 2, 'pc'),
(19, (SELECT id FROM ingredients WHERE name = 'Cucumber'), 1, 'pc'),
(19, (SELECT id FROM ingredients WHERE name = 'Tofu'), 150, 'g'),
(19, (SELECT id FROM ingredients WHERE name = 'Tempeh'), 150, 'g'),
(19, (SELECT id FROM ingredients WHERE name = 'Peanuts'), 150, 'g'),
(19, (SELECT id FROM ingredients WHERE name = 'Garlic'), 3, 'clove'),
(19, (SELECT id FROM ingredients WHERE name = 'Bird''s Eye Chili'), 3, 'pc'),
(19, (SELECT id FROM ingredients WHERE name = 'Palm Sugar'), 2, 'tbsp'),
(19, (SELECT id FROM ingredients WHERE name = 'Tamarind Paste'), 1, 'tbsp'),
(19, (SELECT id FROM ingredients WHERE name = 'Lime'), 1, 'pc'),
(19, (SELECT id FROM ingredients WHERE name = 'Salt'), 1, 'tsp');

INSERT INTO recipe_steps (recipe_id, step_number, instruction) VALUES
(19, 1, 'Boil potato and eggs until done, then peel and slice'),
(19, 2, 'Blanch bean sprouts and green beans briefly, then drain'),
(19, 3, 'Fry tofu and tempeh until golden, cut into pieces'),
(19, 4, 'Pound peanuts with garlic and chili'),
(19, 5, 'Stir in palm sugar, tamarind, warm water, salt and lime to make a thick sauce'),
(19, 6, 'Arrange vegetables, egg, tofu and tempeh on a plate and pour the peanut sauce over');

-- Recipe 20: Opor Ayam
INSERT INTO recipes (title, description, instructions, nutrition_json) VALUES
('Opor Ayam', 'Chicken braised in coconut milk and aromatic spices',
'Fry a spice paste with candlenut, add chicken, simmer in coconut milk until rich and gently spiced.',
'{"calories": 520, "protein": 32, "carbs": 10, "fats": 38}');

INSERT INTO recipe_ingredients (recipe_id, ingredient_id, quantity, unit) VALUES
(20, (SELECT id FROM ingredients WHERE name = 'Chicken'), 600, 'g'),
(20, (SELECT id FROM ingredients WHERE name = 'Coconut Milk'), 500, 'ml'),
(20, (SELECT id FROM ingredients WHERE name = 'Garlic'), 5, 'clove'),
(20, (SELECT id FROM ingredients WHERE name = 'Shallots'), 6, 'pc'),
(20, (SELECT id FROM ingredients WHERE name = 'Galangal'), 30, 'g'),
(20, (SELECT id FROM ingredients WHERE name = 'Ginger'), 20, 'g'),
(20, (SELECT id FROM ingredients WHERE name = 'Turmeric'), 25, 'g'),
(20, (SELECT id FROM ingredients WHERE name = 'Coriander Seeds'), 1, 'tsp'),
(20, (SELECT id FROM ingredients WHERE name = 'Cumin'), 1, 'tsp'),
(20, (SELECT id FROM ingredients WHERE name = 'Candlenut'), 3, 'pc'),
(20, (SELECT id FROM ingredients WHERE name = 'Lemongrass'), 2, 'stalk'),
(20, (SELECT id FROM ingredients WHERE name = 'Kaffir Lime Leaves'), 4, 'leaf'),
(20, (SELECT id FROM ingredients WHERE name = 'Palm Sugar'), 1, 'tbsp'),
(20, (SELECT id FROM ingredients WHERE name = 'Vegetable Oil'), 2, 'tbsp'),
(20, (SELECT id FROM ingredients WHERE name = 'Salt'), 1, 'tsp');

INSERT INTO recipe_steps (recipe_id, step_number, instruction) VALUES
(20, 1, 'Blend shallots, garlic, galangal, ginger, turmeric, coriander, cumin and candlenut'),
(20, 2, 'Heat oil and fry the paste until fragrant'),
(20, 3, 'Add the chicken pieces and stir until they change color'),
(20, 4, 'Pour in coconut milk with lemongrass and lime leaves'),
(20, 5, 'Simmer gently until the chicken is tender and the sauce thickens'),
(20, 6, 'Season with palm sugar and salt, serve with ketupat or rice');

-- Recipe 21: Semur Daging
INSERT INTO recipes (title, description, instructions, nutrition_json) VALUES
('Semur Daging', 'Indonesian sweet-savory beef stew in sweet soy sauce',
'Brown beef with aromatics. Braise in sweet soy sauce and stock with nutmeg and cloves until tender.',
'{"calories": 540, "protein": 36, "carbs": 18, "fats": 30}');

INSERT INTO recipe_ingredients (recipe_id, ingredient_id, quantity, unit) VALUES
(21, (SELECT id FROM ingredients WHERE name = 'Beef'), 500, 'g'),
(21, (SELECT id FROM ingredients WHERE name = 'Sweet Soy Sauce'), 4, 'tbsp'),
(21, (SELECT id FROM ingredients WHERE name = 'Shallots'), 6, 'pc'),
(21, (SELECT id FROM ingredients WHERE name = 'Garlic'), 3, 'clove'),
(21, (SELECT id FROM ingredients WHERE name = 'Nutmeg'), 1, 'tsp'),
(21, (SELECT id FROM ingredients WHERE name = 'Cloves'), 3, 'pc'),
(21, (SELECT id FROM ingredients WHERE name = 'Cinnamon'), 1, 'stick'),
(21, (SELECT id FROM ingredients WHERE name = 'Potato'), 2, 'pc'),
(21, (SELECT id FROM ingredients WHERE name = 'Tomato'), 1, 'pc'),
(21, (SELECT id FROM ingredients WHERE name = 'Fried Shallots'), 2, 'tbsp'),
(21, (SELECT id FROM ingredients WHERE name = 'Palm Sugar'), 1, 'tbsp'),
(21, (SELECT id FROM ingredients WHERE name = 'Vegetable Oil'), 2, 'tbsp'),
(21, (SELECT id FROM ingredients WHERE name = 'Salt'), 1, 'tsp');

INSERT INTO recipe_steps (recipe_id, step_number, instruction) VALUES
(21, 1, 'Cut beef into chunks and brown in oil'),
(21, 2, 'Add bruised shallots and garlic, cook until fragrant'),
(21, 3, 'Pour in sweet soy sauce, a little water, nutmeg, cloves and cinnamon'),
(21, 4, 'Simmer covered until the beef is tender, about 1 hour'),
(21, 5, 'Add potato and tomato, cook until soft'),
(21, 6, 'Sweeten with palm sugar, season with salt, garnish with fried shallots');

-- Recipe 22: Mie Goreng
INSERT INTO recipes (title, description, instructions, nutrition_json) VALUES
('Mie Goreng', 'Indonesian stir-fried noodles with sweet soy sauce',
'Boil noodles. Stir-fry chicken, shrimp and vegetables. Add noodles and sweet soy sauce. Toss and serve.',
'{"calories": 520, "protein": 22, "carbs": 62, "fats": 18}');

INSERT INTO recipe_ingredients (recipe_id, ingredient_id, quantity, unit) VALUES
(22, (SELECT id FROM ingredients WHERE name = 'Egg Noodles'), 200, 'g'),
(22, (SELECT id FROM ingredients WHERE name = 'Eggs'), 2, 'pc'),
(22, (SELECT id FROM ingredients WHERE name = 'Chicken'), 100, 'g'),
(22, (SELECT id FROM ingredients WHERE name = 'Shrimp'), 100, 'g'),
(22, (SELECT id FROM ingredients WHERE name = 'Cabbage'), 150, 'g'),
(22, (SELECT id FROM ingredients WHERE name = 'Shallots'), 4, 'pc'),
(22, (SELECT id FROM ingredients WHERE name = 'Garlic'), 3, 'clove'),
(22, (SELECT id FROM ingredients WHERE name = 'Sweet Soy Sauce'), 3, 'tbsp'),
(22, (SELECT id FROM ingredients WHERE name = 'Soy Sauce'), 1, 'tbsp'),
(22, (SELECT id FROM ingredients WHERE name = 'Vegetable Oil'), 3, 'tbsp'),
(22, (SELECT id FROM ingredients WHERE name = 'Tomato'), 1, 'pc'),
(22, (SELECT id FROM ingredients WHERE name = 'Spring Onions'), 2, 'tbsp');

INSERT INTO recipe_steps (recipe_id, step_number, instruction) VALUES
(22, 1, 'Boil the egg noodles until just tender, then drain'),
(22, 2, 'Heat oil in a wok and fry chopped shallots and garlic'),
(22, 3, 'Add diced chicken and shrimp, stir-fry until cooked'),
(22, 4, 'Push to one side and scramble the eggs'),
(22, 5, 'Add shredded cabbage and tomato, stir-fry briefly'),
(22, 6, 'Add the noodles, sweet soy sauce and soy sauce, toss well'),
(22, 7, 'Serve hot topped with spring onions');

-- ============================================
-- SAMPLE PANTRY DATA
-- ============================================

INSERT INTO pantry (ingredient_id, current_quantity, unit) VALUES
((SELECT id FROM ingredients WHERE name = 'Jasmine Rice'), 2000, 'g'),
((SELECT id FROM ingredients WHERE name = 'Garlic'), 100, 'g'),
((SELECT id FROM ingredients WHERE name = 'Onion'), 500, 'g'),
((SELECT id FROM ingredients WHERE name = 'Eggs'), 12, 'pc'),
((SELECT id FROM ingredients WHERE name = 'Chicken'), 500, 'g'),
((SELECT id FROM ingredients WHERE name = 'Soy Sauce'), 500, 'ml'),
((SELECT id FROM ingredients WHERE name = 'Fish Sauce'), 300, 'ml'),
((SELECT id FROM ingredients WHERE name = 'Olive Oil'), 500, 'ml'),
((SELECT id FROM ingredients WHERE name = 'Salt'), 500, 'g'),
((SELECT id FROM ingredients WHERE name = 'Black Pepper'), 50, 'g'),
((SELECT id FROM ingredients WHERE name = 'Coconut Milk'), 800, 'ml'),
((SELECT id FROM ingredients WHERE name = 'Butter'), 250, 'g'),
((SELECT id FROM ingredients WHERE name = 'Pasta'), 500, 'g'),
((SELECT id FROM ingredients WHERE name = 'Tomato'), 400, 'g');

-- ============================================
-- SAMPLE MEAL PLAN
-- ============================================

INSERT INTO meal_plan (recipe_id, plan_date) VALUES
(1, CURRENT_DATE),
(4, CURRENT_DATE + 1);

COMMIT;
