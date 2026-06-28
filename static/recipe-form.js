// Shared front-end logic for Wareg. Included by the dashboard (base.html), the
// recipes page (recipes.html), and the server-rendered detail page (pageShell).

window.WaregUnits = ["g","kg","ml","l","pc","clove","tbsp","tsp","cup","lb","oz","stalk","leaf","sprig"];
window.__ingredients = []; // cached for comboboxes + autocomplete

function escapeHtml(s){
    return String(s == null ? "" : s)
        .replace(/&/g,"&amp;").replace(/</g,"&lt;").replace(/>/g,"&gt;")
        .replace(/"/g,"&quot;").replace(/'/g,"&#39;");
}

function unitOptionsHTML(name, required, selected){
    var sel = selected || '';
    var h = '<select name="'+name+'"'+(required?' required':'')+' class="unit-select w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-emerald-500">' +
        '<option value="">Unit</option>';
    WaregUnits.forEach(function(u){
        h += '<option value="'+u+'"'+(u===sel?' selected':'')+'>'+u+'</option>';
    });
    h += '</select>';
    return h;
}

function fillUnitSelects(){
    document.querySelectorAll('[data-unit-select]').forEach(function(sel){
        if(sel.dataset.filled) return;
        var placeholder = sel.innerHTML;
        sel.innerHTML = placeholder + WaregUnits.map(function(u){
            return '<option value="'+u+'">'+u+'</option>';
        }).join('');
        sel.dataset.filled = '1';
    });
}

function showToast(message, type){
    type = type || 'success';
    var c = document.getElementById('toast-container');
    if(!c) return;
    var t = document.createElement('div');
    t.className = (type === 'error' ? 'bg-red-500' : 'bg-green-500') + ' text-white px-6 py-4 rounded-lg shadow-lg flex items-center gap-3 transform transition-all duration-300 translate-x-full';
    t.textContent = message;
    c.appendChild(t);
    setTimeout(function(){ t.classList.remove('translate-x-full'); }, 10);
    setTimeout(function(){ t.classList.add('translate-x-full'); setTimeout(function(){ t.remove(); }, 300); }, 3000);
}

// Show plain-text error responses as a toast and keep them from clobbering the
// DOM. Server errors are returned as plain text (no SQL/stack leaked).
document.body.addEventListener('htmx:beforeSwap', function(evt){
    var x = evt.detail && evt.detail.xhr;
    if(!x || typeof x.status === 'undefined') return;
    if(x.status >= 400){
        evt.detail.shouldSwap = false;
        showToast((x.responseText || 'Operation failed.').trim(), 'error');
    }
});

// Global toast for any successful HTMX create/delete (covers pages that don't
// wire their own after-request handlers).
document.body.addEventListener('htmx:afterRequest', function(evt){
    var x = evt.detail && evt.detail.xhr;
    if(!x || typeof x.status === 'undefined') return;
    if(x.status >= 200 && x.status < 300){
        var m = (x.method || '').toUpperCase();
        var u = x.url || '';
        // Only show for DELETE (POST results are handled by form-specific handlers
        // to avoid duplicate toasts). This covers delete buttons everywhere.
        if(m === 'DELETE'){
            if(u.indexOf('/api/recipes') >= 0) showToast('Recipe deleted.', 'success');
            else if(u.indexOf('/api/ingredients') >= 0) showToast('Ingredient deleted.', 'success');
            else if(u.indexOf('/api/pantry') >= 0) showToast('Pantry item removed.', 'success');
            else if(u.indexOf('/api/meal-plan') >= 0) showToast('Meal removed.', 'success');
        }
    } else if(x.status === 0){
        showToast('Network error — please try again.', 'error');
    }
});

// Close any open combobox dropdown when clicking elsewhere.
document.addEventListener('click', function(e){
    document.querySelectorAll('.ingredient-combo').forEach(function(box){
        if(!box.contains(e.target)){
            var l = box.querySelector('.ic-list');
            if(l) l.classList.add('hidden');
        }
    });
});

// ----- Ingredient data (clean JSON endpoint) -----
async function loadIngredientsJSON(){
    try { var r = await fetch('/api/ingredients/options'); return await r.json(); } catch(e){ return []; }
}
async function loadRecipesJSON(){
    try { var r = await fetch('/api/recipes/options'); return await r.json(); } catch(e){ return []; }
}

// Refresh the cached ingredient list + the add-ingredient <datalist>. Call this
// after creating an ingredient so pickers see it immediately.
function refreshIngredientCache(){
    return loadIngredientsJSON().then(function(items){
        window.__ingredients = items || [];
        var dl = document.getElementById('existing-ingredients');
        if(dl){
            dl.innerHTML = window.__ingredients.map(function(i){
                return '<option value="'+escapeHtml(i.name)+'"></option>';
            }).join('');
        }
    });
}

function fillRecipeSelects(){
    return loadRecipesJSON().then(function(items){
        document.querySelectorAll('[data-recipe-select]').forEach(function(sel){
            var cur = sel.value;
            sel.innerHTML = '<option value="">Select recipe...</option>' +
                items.map(function(i){ return '<option value="'+i.id+'">'+escapeHtml(i.title)+'</option>'; }).join('');
            if(cur) sel.value = cur;
        });
    });
}

// ----- Searchable / typeable ingredient picker (combobox) -----
// Renders a text input (type to filter) + a hidden id field (the submitted
// value) + a suggestion dropdown. Lets users type OR pick from the list.
// When an ingredient is selected, the unit dropdown in the same row is
// auto-set to the ingredient's default unit.
function ingredientComboHTML(name){
    return '<div class="ingredient-combo relative flex-1 min-w-[180px]">' +
        '<input type="text" class="ic-input w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-emerald-500" placeholder="Type or search ingredient..." autocomplete="off">' +
        '<input type="hidden" name="'+name+'" class="ic-id">' +
        '<div class="ic-list absolute z-20 left-0 right-0 mt-1 bg-white border rounded-md shadow-lg max-h-60 overflow-auto hidden"></div>' +
        '</div>';
}

function wireIngredientCombos(){
    document.querySelectorAll('.ingredient-combo').forEach(function(box){
        if(box.dataset.wired) return;
        box.dataset.wired = '1';
        var input = box.querySelector('.ic-input');
        var idField = box.querySelector('.ic-id');
        var list = box.querySelector('.ic-list');

        function render(q){
            var items = window.__ingredients || [];
            q = (q || '').trim().toLowerCase();
            var matches = items.filter(function(i){
                return !q || String(i.name).toLowerCase().indexOf(q) >= 0;
            }).slice(0, 60);
            if(matches.length === 0){ list.classList.add('hidden'); return; }
            list.innerHTML = matches.map(function(i){
                var unitTxt = i.unit ? ' <span class="text-gray-400">('+escapeHtml(i.unit)+')</span>' : '';
                return '<div class="ic-item px-3 py-2 hover:bg-emerald-50 cursor-pointer text-sm" data-id="'+i.id+'" data-name="'+escapeHtml(i.name)+'" data-unit="'+escapeHtml(i.unit||'')+'">'+escapeHtml(i.name)+unitTxt+'</div>';
            }).join('');
            list.classList.remove('hidden');
        }
        input.addEventListener('focus', function(){ render(input.value); });
        input.addEventListener('input', function(){ idField.value = ''; render(input.value); });
        list.addEventListener('mousedown', function(e){
            var item = e.target.closest('.ic-item');
            if(!item) return;
            e.preventDefault();
            idField.value = item.dataset.id;
            input.value = item.dataset.name;
            list.classList.add('hidden');
            // Auto-select the ingredient's default unit in the nearest unit dropdown.
            var defaultUnit = item.dataset.unit || '';
            if(defaultUnit){
                // Recipe form: unit is inside .ingredient-row > .unit-select
                var row = box.closest('.ingredient-row');
                var unitSel = row ? row.querySelector('.unit-select') : null;
                // Pantry/dashboard form: unit is a sibling select[name="unit"] in the same <form>
                if(!unitSel){
                    var form = box.closest('form');
                    if(form){
                        unitSel = form.querySelector('select[name="unit"]');
                    }
                }
                if(unitSel) unitSel.value = defaultUnit;
            }
        });
    });
}

function fillIngredientSelects(){
    return refreshIngredientCache().then(function(){
        wireIngredientCombos();
    });
}

// ----- Add-ingredient form: autocomplete + live "already exists" hint -----
function wireNewIngredientCheck(){
    var input = document.getElementById('new-ingredient-name');
    if(!input || input.dataset.wired) return;
    input.dataset.wired = '1';
    var hint = document.getElementById('ingredient-dup-hint');
    var submit = document.querySelector('#ingredient-add-form button[type=submit]');
    if(!window.__ingredients.length){ refreshIngredientCache(); }
    input.addEventListener('input', function(){
        var v = input.value.trim().toLowerCase();
        var dup = (window.__ingredients || []).some(function(i){ return String(i.name).toLowerCase() === v; });
        if(hint){ hint.classList.toggle('hidden', !dup); }
        if(submit){ submit.disabled = dup; submit.classList.toggle('opacity-50', dup); submit.classList.toggle('cursor-not-allowed', dup); }
    });
}

// ----- Recipe add-form -----
function ingredientRowHTML(idx){
    return '<div class="ingredient-row flex flex-wrap gap-2 items-end">' +
        ingredientComboHTML('ingredients.'+idx+'.ingredient_id') +
        '<div class="w-24"><label class="block text-xs font-medium mb-1">Qty</label><input type="number" step="0.01" min="0.01" name="ingredients.'+idx+'.quantity" required class="w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-emerald-500"></div>' +
        '<div class="w-24"><label class="block text-xs font-medium mb-1">Unit</label>' + unitOptionsHTML('ingredients.'+idx+'.unit', true) + '</div>' +
        '<div class="w-24 flex flex-col"><label class="block text-xs font-medium mb-1">Required</label><input type="checkbox" name="ingredients.'+idx+'.mandatory" checked class="mt-2 h-5 w-5" title="Uncheck to mark optional"></div>' +
        '<button type="button" onclick="removeIngredientRow(this)" class="bg-red-600 text-white px-3 py-2 rounded-md hover:bg-red-700">\u00d7</button>' +
        '</div>';
}
function stepRowHTML(idx, num){
    return '<div class="step-row flex gap-2 items-start">' +
        '<div class="w-16"><label class="block text-xs font-medium mb-1">Step</label><input type="number" name="steps.'+idx+'.step_number" required readonly value="'+num+'" class="w-full px-3 py-2 border rounded-md bg-gray-100 focus:outline-none"></div>' +
        '<div class="flex-1"><label class="block text-xs font-medium mb-1">Instruction</label><input type="text" name="steps.'+idx+'.instruction" required class="w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-emerald-500" placeholder="e.g. Chop and saut\u00e9 the vegetables"></div>' +
        '<button type="button" onclick="removeStepRow(this)" class="bg-red-600 text-white px-3 py-2 rounded-md hover:bg-red-700 mt-5">\u00d7</button>' +
        '</div>';
}

function removeIngredientRow(btn){ btn.closest('.ingredient-row').remove(); reindexRows('recipe-ingredients-list','ingredient'); }
function removeStepRow(btn){ btn.closest('.step-row').remove(); reindexRows('recipe-steps-list','step'); }

function reindexRows(listId, kind){
    var list = document.getElementById(listId);
    if(!list) return;
    var rows = list.children;
    for(var i=0;i<rows.length;i++){
        rows[i].querySelectorAll('[name]').forEach(function(el){
            el.name = el.name.replace(/\.\d+\./, '.'+i+'.');
        });
        var stepInput = rows[i].querySelector('input[type=number][readonly]');
        if(stepInput){ stepInput.value = String(i+1); }
    }
}

function addIngredientRow(){
    var list = document.getElementById('recipe-ingredients-list');
    var div = document.createElement('div'); div.innerHTML = ingredientRowHTML(list.children.length);
    list.appendChild(div.firstElementChild); wireIngredientCombos();
}
function addStepRow(){
    var list = document.getElementById('recipe-steps-list');
    var idx = list.children.length;
    var div = document.createElement('div'); div.innerHTML = stepRowHTML(idx, idx+1);
    list.appendChild(div.firstElementChild);
}

// ----- Quick-add ingredient (inline, inside recipe form) -----
// Lets the user create a new ingredient without leaving the recipe form.
function quickAddIngredient(){
    var nameInput = document.getElementById('quick-ingredient-name');
    var unitInput = document.getElementById('quick-ingredient-unit');
    if(!nameInput || !unitInput) return;
    var name = nameInput.value.trim();
    var unit = unitInput.value.trim();
    if(!name){ showToast('Enter an ingredient name first.', 'error'); return; }

    fetch('/api/ingredients', {
        method: 'POST',
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
        body: 'name=' + encodeURIComponent(name) + '&unit=' + encodeURIComponent(unit)
    }).then(function(resp){
        if(!resp.ok) return resp.text().then(function(t){ throw new Error(t || 'Failed'); });
        return refreshIngredientCache();
    }).then(function(){
        showToast('Ingredient "' + name + '" added.', 'success');
        nameInput.value = '';
        unitInput.value = '';
        wireIngredientCombos();
    }).catch(function(err){
        showToast((err.message || 'Could not add ingredient.').trim(), 'error');
    });
}

function recipeFormHTML(){
    return '' +
    '<h3 class="text-xl font-semibold mb-4">Add New Recipe</h3>' +
    '<form action="/api/recipes" method="post" hx-post="/api/recipes" hx-target="#recipes-list" hx-swap="innerHTML" ' +
    'hx-on::after-request="if(event.detail.xhr.status<300){document.getElementById(\'recipe-form-container\').classList.add(\'hidden\'); document.getElementById(\'recipe-form-container\').innerHTML=\'\'; showToast(\'Recipe created.\',\'success\');}" ' +
    'class="space-y-4">' +
      '<div><label class="block text-sm font-medium mb-1">Title *</label><input type="text" name="title" required class="w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-emerald-500"></div>' +
      '<div><label class="block text-sm font-medium mb-1">Description</label><textarea name="description" rows="2" class="w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-emerald-500"></textarea></div>' +
      '<div class="grid grid-cols-1 sm:grid-cols-2 gap-4">' +
        '<div><label class="block text-sm font-medium mb-1">Image URL</label><input type="text" name="image_url" class="w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-emerald-500" placeholder="https://..."></div>' +
        '<div><label class="block text-sm font-medium mb-1">Video URL (YouTube/Instagram)</label><input type="text" name="video_url" class="w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-emerald-500" placeholder="https://..."></div>' +
      '</div>' +
      '<div class="grid grid-cols-2 sm:grid-cols-4 gap-4">' +
        '<div><label class="block text-sm font-medium mb-1">Calories</label><input type="number" step="0.1" name="nutrition.calories" class="w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-emerald-500"></div>' +
        '<div><label class="block text-sm font-medium mb-1">Protein (g)</label><input type="number" step="0.1" name="nutrition.protein" class="w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-emerald-500"></div>' +
        '<div><label class="block text-sm font-medium mb-1">Carbs (g)</label><input type="number" step="0.1" name="nutrition.carbs" class="w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-emerald-500"></div>' +
        '<div><label class="block text-sm font-medium mb-1">Fats (g)</label><input type="number" step="0.1" name="nutrition.fats" class="w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-emerald-500"></div>' +
      '</div>' +

      // ----- Quick-add ingredient box -----
      '<div class="bg-gray-50 border border-gray-200 rounded-lg p-3">' +
        '<div class="text-sm font-semibold text-gray-700 mb-2">Quick-add an ingredient (if not in list)</div>' +
        '<div class="flex gap-2 items-end flex-wrap">' +
          '<div class="flex-1 min-w-[160px]"><label class="block text-xs font-medium mb-1">Name</label><input type="text" id="quick-ingredient-name" class="w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-emerald-500" placeholder="e.g. Egg yolk"></div>' +
          '<div class="w-32"><label class="block text-xs font-medium mb-1">Default unit</label>' + unitOptionsHTML('quick-ingredient-unit', false) + '</div>' +
          '<button type="button" onclick="quickAddIngredient()" class="bg-gray-600 text-white px-4 py-2 rounded-md hover:bg-gray-700 text-sm">+ Add ingredient</button>' +
        '</div>' +
      '</div>' +

      '<div class="border-t pt-4"><div class="flex justify-between items-center mb-3"><h4 class="text-lg font-semibold">Ingredients</h4>' +
        '<button type="button" onclick="addIngredientRow()" class="bg-gray-600 text-white px-3 py-1.5 rounded-md hover:bg-gray-700 text-sm">+ Add</button></div>' +
        '<div id="recipe-ingredients-list" class="space-y-2">' + ingredientRowHTML(0) + '</div></div>' +
      '<div class="border-t pt-4"><div class="flex justify-between items-center mb-3"><h4 class="text-lg font-semibold">Cooking Steps</h4>' +
        '<button type="button" onclick="addStepRow()" class="bg-gray-600 text-white px-3 py-1.5 rounded-md hover:bg-gray-700 text-sm">+ Add</button></div>' +
        '<div id="recipe-steps-list" class="space-y-2">' + stepRowHTML(0,1) + '</div></div>' +
      '<div class="flex gap-3 pt-2">' +
        '<button type="submit" class="bg-emerald-600 text-white px-6 py-2 rounded-md hover:bg-emerald-700 transition">Save Recipe</button>' +
        '<button type="button" onclick="toggleRecipeForm()" class="bg-gray-400 text-white px-6 py-2 rounded-md hover:bg-gray-500 transition">Cancel</button>' +
      '</div>' +
    '</form>';
}

function toggleRecipeForm(){
    var c = document.getElementById('recipe-form-container');
    if(!c) return;
    c.classList.toggle('hidden');
    if(!c.classList.contains('hidden')){
        c.innerHTML = recipeFormHTML();
        fillIngredientSelects();
        // CRITICAL: tell HTMX to scan the newly-injected form for hx-* attributes.
        // Without this, the browser falls back to a native GET submission.
        if(window.htmx && typeof htmx.process === 'function'){
            htmx.process(c);
        }
    } else {
        c.innerHTML = '';
    }
}

// ----- Recipe search/filter -----
function filterRecipes(query){
    var q = (query || '').trim().toLowerCase();
    var cards = document.querySelectorAll('#recipes-list [id^="recipe-"]');
    cards.forEach(function(card){
        if(!q){
            card.style.display = '';
            return;
        }
        var titleEl = card.querySelector('h3, h2');
        var descEl = card.querySelector('p');
        var title = titleEl ? titleEl.textContent.toLowerCase() : '';
        var desc = descEl ? descEl.textContent.toLowerCase() : '';
        if(title.indexOf(q) >= 0 || desc.indexOf(q) >= 0){
            card.style.display = '';
        } else {
            card.style.display = 'none';
        }
    });
}

// ----- Matches (What Can I Cook?) search/filter -----
function filterMatches(query){
    var q = (query || '').trim().toLowerCase();
    var cards = document.querySelectorAll('#matches-list .match-card');
    cards.forEach(function(card){
        if(!q){
            card.style.display = '';
            return;
        }
        var title = card.getAttribute('data-title') || '';
        var missing = card.getAttribute('data-missing') || '';
        if(title.indexOf(q) >= 0 || missing.indexOf(q) >= 0){
            card.style.display = '';
        } else {
            card.style.display = 'none';
        }
    });
}
