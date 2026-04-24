// Stockii - Sell Page JavaScript (offline-capable POS)
var conventionId = 0;
var pendingQueue = [];
var qtyModalCpId = 0;
var qtyModalConvId = 0;

// Double-tap tracking
var lastTapCpId = 0;
var lastTapTime = 0;
var tapResetTimer = null;
var DOUBLE_TAP_MS = 400;

// Snackbar
var activeSnackbar = null;
var snackbarTimer = null;

function initSellPage(convId) {
    conventionId = convId;
    loadPendingQueue();
    updateStatusBadge();
    updatePendingDisplay();
    initCategoryStates();

    // Event delegation for product cards
    var productList = document.getElementById('product-list');
    if (productList) {
        productList.addEventListener('click', function(e) {
            var card = e.target.closest('.product-card');
            if (!card) return;
            var cpId = parseInt(card.dataset.cpId);
            var convId = parseInt(card.dataset.convId);
            var name = card.dataset.productName || 'Item';
            handleCardTap(cpId, name, convId);
        });
        productList.addEventListener('contextmenu', function(e) {
            var card = e.target.closest('.product-card');
            if (!card) return;
            e.preventDefault();
            var cpId = parseInt(card.dataset.cpId);
            var convId = parseInt(card.dataset.convId);
            var name = card.dataset.productName || 'Item';
            openQtyModal(cpId, name, convId);
        });
    }

    var searchInput = document.getElementById('product-search');
    if (searchInput) {
        searchInput.addEventListener('input', filterProducts);
    }

    // Auto-sync every 10 seconds when online
    setInterval(function() {
        if (navigator.onLine && pendingQueue.length > 0) {
            syncNow();
        }
    }, 10000);

    window.addEventListener('online', function() {
        updateStatusBadge();
        if (pendingQueue.length > 0) syncNow();
    });
    window.addEventListener('offline', updateStatusBadge);

    // Register service worker
    if ('serviceWorker' in navigator) {
        navigator.serviceWorker.register('/static/sw.js').catch(function(err) {
            console.log('SW registration failed:', err);
        });
    }
}

// ---- Category collapse ----

function initCategoryStates() {
    var headers = document.querySelectorAll('.category-header');
    headers.forEach(function(header) {
        var name = header.dataset.category;
        var grid = header.closest('.category-section').querySelector('.product-grid');
        if (getCategoryState(name)) {
            header.classList.add('open');
            grid.style.display = 'grid';
        }
    });
}

function toggleCategory(header) {
    var grid = header.closest('.category-section').querySelector('.product-grid');
    var isOpen = header.classList.contains('open');
    if (isOpen) {
        header.classList.remove('open');
        grid.style.display = 'none';
    } else {
        header.classList.add('open');
        grid.style.display = 'grid';
    }
    saveCategoryState(header.dataset.category, !isOpen);
}

function saveCategoryState(name, isOpen) {
    try {
        var key = 'stockii-cats-' + conventionId;
        var stored = JSON.parse(localStorage.getItem(key) || '{}');
        stored[name] = isOpen;
        localStorage.setItem(key, JSON.stringify(stored));
    } catch(e) {}
}

function getCategoryState(name) {
    try {
        var key = 'stockii-cats-' + conventionId;
        var stored = JSON.parse(localStorage.getItem(key) || '{}');
        return stored[name] === true;
    } catch(e) { return false; }
}

// ---- Search / filter ----

function filterProducts() {
    var query = document.getElementById('product-search').value.toLowerCase().trim();
    var clearBtn = document.getElementById('search-clear');
    var noResults = document.getElementById('no-results');
    clearBtn.style.display = query ? '' : 'none';

    var sections = document.querySelectorAll('.category-section');
    var totalVisible = 0;

    sections.forEach(function(section) {
        var header = section.querySelector('.category-header');
        var grid = section.querySelector('.product-grid');
        var cards = section.querySelectorAll('.product-card');
        var visibleCount = 0;

        if (query) {
            cards.forEach(function(card) {
                var name = (card.dataset.productName || '').toLowerCase();
                var visible = name.includes(query);
                card.style.display = visible ? '' : 'none';
                if (visible) visibleCount++;
            });
            section.style.display = visibleCount > 0 ? '' : 'none';
            if (visibleCount > 0) {
                grid.style.display = 'grid';
            }
            totalVisible += visibleCount;
        } else {
            section.style.display = '';
            cards.forEach(function(card) { card.style.display = ''; });
            var isOpen = getCategoryState(header.dataset.category);
            if (isOpen) {
                header.classList.add('open');
                grid.style.display = 'grid';
            } else {
                header.classList.remove('open');
                grid.style.display = 'none';
            }
            totalVisible++;
        }
    });

    if (noResults) {
        noResults.style.display = (query && totalVisible === 0) ? '' : 'none';
    }
}

function clearSearch() {
    var input = document.getElementById('product-search');
    if (input) input.value = '';
    filterProducts();
}

// ---- Double-tap handling ----

function handleCardTap(cpId, productName, convId) {
    var now = Date.now();

    if (lastTapCpId === cpId && (now - lastTapTime) < DOUBLE_TAP_MS) {
        // Double tap confirmed - record sale
        clearTimeout(tapResetTimer);
        lastTapCpId = 0;
        lastTapTime = 0;
        clearPrimedState();
        recordSale(cpId, 1, convId, productName);
    } else {
        // First tap - prime the card
        clearPrimedState();
        lastTapCpId = cpId;
        lastTapTime = now;

        var card = document.getElementById('card-' + cpId);
        if (card) card.classList.add('tap-primed');

        // Reset after timeout
        clearTimeout(tapResetTimer);
        tapResetTimer = setTimeout(function() {
            lastTapCpId = 0;
            lastTapTime = 0;
            clearPrimedState();
        }, DOUBLE_TAP_MS);
    }
}

function clearPrimedState() {
    var primed = document.querySelectorAll('.product-card.tap-primed');
    for (var i = 0; i < primed.length; i++) {
        primed[i].classList.remove('tap-primed');
    }
}

// ---- Sale recording ----

function recordSale(cpId, quantity, convId, productName) {
    // Optimistic UI update
    var card = document.getElementById('card-' + cpId);
    var previousSold = 0;
    if (card) {
        var soldEl = card.querySelector('.sold-count');
        previousSold = parseInt(card.dataset.qtySold) || 0;
        var newSold = previousSold + quantity;
        var brought = parseInt(card.dataset.qtyBrought) || 0;

        card.dataset.qtySold = newSold;
        soldEl.textContent = newSold;
        updateCardStockLevel(card, newSold, brought);

        // Flash animation
        var flash = document.createElement('div');
        flash.className = 'sale-flash';
        card.appendChild(flash);
        setTimeout(function() { flash.remove(); }, 400);
    }

    if (navigator.onLine) {
        var formData = new FormData();
        formData.append('convention_product_id', cpId);
        formData.append('quantity', quantity);

        fetch('/api/conventions/' + convId + '/sales', {
            method: 'POST',
            body: formData
        })
        .then(function(r) {
            if (!r.ok) throw new Error('HTTP ' + r.status);
            return r.json();
        })
        .then(function(data) {
            if (data && data.sale_id) {
                showUndoSnackbar(productName || 'Item', quantity, data.sale_id, cpId, previousSold);
            } else {
                showSnackbar('+' + quantity + ' ' + (productName || 'Item'));
            }
        })
        .catch(function(err) {
            console.error('Sale failed:', err);
            queueSale(cpId, quantity);
            updatePendingDisplay();
            showSnackbar('Offline - sale queued');
        });
    } else {
        queueSale(cpId, quantity);
        showSnackbar('Offline - sale queued');
    }
}

// ---- Snackbar / Undo ----

function showUndoSnackbar(productName, quantity, saleId, cpId, previousSold) {
    dismissSnackbar();

    var container = document.getElementById('snackbar-container');
    if (!container) return;

    var label = '+' + quantity + ' ' + productName;
    var el = document.createElement('div');
    el.className = 'snackbar';
    el.innerHTML = '<span>' + escHtml(label) + '</span>' +
        '<button class="undo-btn" data-sale-id="' + saleId + '" data-cp-id="' + cpId + '" data-qty="' + quantity + '" data-prev="' + previousSold + '">UNDO</button>';
    el.querySelector('.undo-btn').addEventListener('click', function() {
        undoSale(saleId, cpId, quantity, previousSold);
    });
    container.appendChild(el);
    activeSnackbar = el;

    snackbarTimer = setTimeout(dismissSnackbar, 5000);
}

function showSnackbar(message) {
    dismissSnackbar();

    var container = document.getElementById('snackbar-container');
    if (!container) return;

    var el = document.createElement('div');
    el.className = 'snackbar';
    el.innerHTML = '<span>' + escHtml(message) + '</span>';
    container.appendChild(el);
    activeSnackbar = el;

    snackbarTimer = setTimeout(dismissSnackbar, 3000);
}

function dismissSnackbar() {
    clearTimeout(snackbarTimer);
    if (activeSnackbar) {
        activeSnackbar.remove();
        activeSnackbar = null;
    }
}

function undoSale(saleId, cpId, quantity, previousSold) {
    dismissSnackbar();

    // Revert UI
    var card = document.getElementById('card-' + cpId);
    if (card) {
        var soldEl = card.querySelector('.sold-count');
        var brought = parseInt(card.dataset.qtyBrought) || 0;
        card.dataset.qtySold = previousSold;
        soldEl.textContent = previousSold;
        updateCardStockLevel(card, previousSold, brought);
    }

    if (navigator.onLine) {
        fetch('/api/sales/' + saleId, { method: 'DELETE' }).catch(function() {});
    }
}

function escHtml(s) {
    return String(s).replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;');
}

// ---- Offline queue ----

function queueSale(cpId, quantity) {
    pendingQueue.push({
        convention_product_id: cpId,
        quantity: quantity,
        timestamp: new Date().toISOString()
    });
    savePendingQueue();
    updatePendingDisplay();
}

function syncNow() {
    if (pendingQueue.length === 0) return;

    var toSync = pendingQueue.slice();
    var badge = document.getElementById('sync-badge');
    if (badge) badge.textContent = 'Syncing...';

    fetch('/api/conventions/' + conventionId + '/sales/bulk', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ sales: toSync })
    })
    .then(function(r) {
        if (r.ok) {
            pendingQueue = [];
            savePendingQueue();
            updatePendingDisplay();
        }
    })
    .catch(function() {})
    .finally(function() {
        updatePendingDisplay();
    });
}

// ---- Quantity modal ----

function openQtyModal(cpId, productName, convId) {
    qtyModalCpId = cpId;
    qtyModalConvId = convId;
    document.getElementById('qty-modal-title').textContent = productName;
    document.getElementById('qty-input').value = 1;
    document.getElementById('qty-modal').style.display = 'flex';
    document.getElementById('qty-input').focus();
}

function closeQtyModal() {
    document.getElementById('qty-modal').style.display = 'none';
}

function confirmQtySale() {
    var qty = parseInt(document.getElementById('qty-input').value) || 1;
    var card = document.getElementById('card-' + qtyModalCpId);
    var name = card ? (card.dataset.productName || 'Item') : 'Item';
    recordSale(qtyModalCpId, qty, qtyModalConvId, name);
    closeQtyModal();
}

// ---- UI helpers ----

function updateCardStockLevel(card, sold, brought) {
    card.classList.remove('stock-plenty', 'stock-low', 'stock-critical', 'stock-sold_out');
    if (brought <= 0) {
        card.classList.add('stock-plenty');
        return;
    }
    var pct = sold / brought;
    if (pct >= 1.0) {
        card.classList.add('stock-sold_out');
    } else if (pct >= 0.9) {
        card.classList.add('stock-critical');
    } else if (pct >= 0.7) {
        card.classList.add('stock-low');
    } else {
        card.classList.add('stock-plenty');
    }
}

function updateStatusBadge() {
    var badge = document.getElementById('status-badge');
    if (!badge) return;
    if (navigator.onLine) {
        badge.className = 'status-badge status-online';
        badge.textContent = 'Online';
    } else {
        badge.className = 'status-badge status-offline';
        badge.textContent = 'Offline';
    }
}

function updatePendingDisplay() {
    var badge = document.getElementById('sync-badge');
    if (!badge) return;
    if (pendingQueue.length > 0) {
        badge.textContent = pendingQueue.length + ' pending';
        badge.style.background = '#d29922';
        badge.style.cursor = 'pointer';
    } else {
        badge.textContent = 'Synced';
        badge.style.background = '#3fb950';
        badge.style.cursor = 'default';
    }
    badge.style.display = '';
}

function loadPendingQueue() {
    try {
        var stored = localStorage.getItem('stockii-pending-' + conventionId);
        if (stored) pendingQueue = JSON.parse(stored);
    } catch(e) { pendingQueue = []; }
}

function savePendingQueue() {
    try {
        localStorage.setItem('stockii-pending-' + conventionId, JSON.stringify(pendingQueue));
    } catch(e) {}
}
