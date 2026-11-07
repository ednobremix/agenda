let currentUser = null;
let currentFilter = 'todos';
let productsCache = [];

document.addEventListener('DOMContentLoaded', () => {
    const userData = localStorage.getItem('estoque_user') || 
                     localStorage.getItem('portalnobre-session') ||
                     localStorage.getItem('financeiro_user');
    
    if (!userData) {
        window.location.href = 'login.html';
        return;
    }
    
    currentUser = JSON.parse(userData);

    // Verifica se tem permissão (considerando que o admin tem acesso a tudo)
    const isAdmin = currentUser.role === 'admin';
    const hasEstoque = isAdmin || currentUser.has_estoque;

    if (!hasEstoque) {
        alert("Acesso negado. Esta conta não possui permissão para o App de Estoque.");
        // Limpa apenas se não for admin para evitar deslogar admin por engano
        if (!isAdmin) {
            localStorage.removeItem('estoque_user');
            localStorage.removeItem('portalnobre-session');
            localStorage.removeItem('financeiro_user');
            localStorage.removeItem('role');
            localStorage.removeItem('has_agenda');
            localStorage.removeItem('has_financeiro');
            localStorage.removeItem('has_estoque');
            window.location.href = '/';
            return;
        }
    }

    document.getElementById('user-email-display').innerText = currentUser.email;
    
    const adminLink = document.getElementById('btn-admin');
    if (adminLink && currentUser.role === 'admin') {
        adminLink.classList.remove('hidden');
    }
    
    initApp();
});

function initApp() {
    // Configura datas padrão para movimentações (dia atual)
    const today = new Date().toISOString().split('T')[0];
    const startDateInput = document.getElementById('mov-start-date');
    const endDateInput = document.getElementById('mov-end-date');
    if (startDateInput) startDateInput.value = today;
    if (endDateInput) endDateInput.value = today;

    loadProducts();
    loadMovements();
    updateStats();

    // Tabs
    document.querySelectorAll('.nav-tab').forEach(tab => {
        tab.addEventListener('click', () => {
            const targetId = tab.dataset.target;
            if (!targetId) return;

            document.querySelectorAll('.nav-tab').forEach(t => t.classList.remove('active'));
            document.querySelectorAll('.tab-content').forEach(c => c.classList.add('hidden'));
            
            tab.classList.add('active');
            const target = document.getElementById(targetId);
            if (target) target.classList.remove('hidden');
            
            if (targetId === 'view-products') loadProducts();
            if (targetId === 'view-movements') {
                // Ao entrar na aba de movimentações, garante que as datas estejam certas se vazias
                const start = document.getElementById('mov-start-date');
                const end = document.getElementById('mov-end-date');
                const today = new Date().toISOString().split('T')[0];
                if (start && !start.value) start.value = today;
                if (end && !end.value) end.value = today;
                loadMovements();
            }
        });
    });

    const btnFilterMov = document.getElementById('btn-filter-movements');
    if (btnFilterMov) {
        btnFilterMov.addEventListener('click', () => {
            loadMovements();
        });
    }

    // Filtros de Produtos
    const filterProdName = document.getElementById('filter-product-name');
    if (filterProdName) {
        filterProdName.addEventListener('input', () => {
            renderProducts(productsCache);
        });
    }

    const filterProdDate = document.getElementById('filter-product-date');
    if (filterProdDate) {
        filterProdDate.addEventListener('change', () => {
            renderProducts(productsCache);
        });
    }

    // Botões de tipo de produto
    document.querySelectorAll('.filter-btn[data-type]').forEach(btn => {
        // Ignora os botões que não são da aba de produtos (se houver ambiguidade, mas aqui são específicos)
        if (btn.closest('#view-products')) {
            btn.addEventListener('click', () => {
                document.querySelectorAll('#view-products .filter-btn').forEach(b => b.classList.remove('active'));
                btn.classList.add('active');
                currentFilter = btn.dataset.type;
                renderProducts(productsCache);
            });
        }
    });

    // Modals
    const btnNewBought = document.getElementById('btn-new-bought');
    if (btnNewBought) {
        btnNewBought.addEventListener('click', () => {
            const formBought = document.getElementById('form-bought');
            if (formBought) formBought.reset();
            const boughtId = document.getElementById('bought-id');
            if (boughtId) boughtId.value = '';
            openModal('modal-bought');
        });
    }

    const btnNewCreated = document.getElementById('btn-new-created');
    if (btnNewCreated) {
        btnNewCreated.addEventListener('click', () => {
            const formCreated = document.getElementById('form-created');
            if (formCreated) formCreated.reset();
            const createdId = document.getElementById('created-id');
            if (createdId) createdId.value = '';
            openModal('modal-created');
        });
    }

    // Forms
    const formBought = document.getElementById('form-bought');
    if (formBought) {
        formBought.addEventListener('submit', async (e) => {
            e.preventDefault();
            const id = document.getElementById('bought-id').value;
            const product = {
                name: document.getElementById('bought-name').value,
                quantity: parseFloat(document.getElementById('bought-qty').value),
                min_quantity: parseFloat(document.getElementById('bought-min').value),
                value: parseFloat(document.getElementById('bought-value').value),
                location: document.getElementById('bought-location').value,
                type: 'comprado',
                user_email: currentUser.email
            };

            const method = id ? 'PUT' : 'POST';
            if (id) product.id = parseInt(id);

            const res = await fetch(`/api/estoque/products?email=${currentUser.email}`, {
                method,
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(product)
            });

            if (res.ok) {
                closeModal('modal-bought');
                loadProducts();
            }
        });
    }

    const formCreated = document.getElementById('form-created');
    if (formCreated) {
        formCreated.addEventListener('submit', async (e) => {
            e.preventDefault();
            const id = document.getElementById('created-id').value;
            const product = {
                name: document.getElementById('created-name').value,
                quantity: parseFloat(document.getElementById('created-qty').value),
                min_quantity: parseFloat(document.getElementById('created-min').value),
                production_cost: parseFloat(document.getElementById('created-cost').value || 0),
                type: 'criado',
                user_email: currentUser.email
            };

            const method = id ? 'PUT' : 'POST';
            if (id) product.id = parseInt(id);

            const res = await fetch(`/api/estoque/products?email=${currentUser.email}`, {
                method,
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(product)
            });

            if (res.ok) {
                closeModal('modal-created');
                loadProducts();
            }
        });
    }

    const formMovement = document.getElementById('form-movement');
    if (formMovement) {
        formMovement.addEventListener('submit', async (e) => {
            e.preventDefault();
            let type = document.getElementById('mov-type').value;
            if (type === 'saida') {
                const checkedOut = document.querySelector('input[name="out-type"]:checked');
                type = checkedOut ? checkedOut.value : 'venda';
            }
            const movement = {
                product_id: parseInt(document.getElementById('mov-product-id').value),
                type: type,
                batch_id: parseInt(document.getElementById('mov-batch-id').value) || null,
                quantity: parseFloat(document.getElementById('mov-qty').value),
                sale_value: type === 'venda' ? parseFloat(document.getElementById('mov-sale-value').value) : 0,
                batch_value: (type === 'entrada') ? parseFloat(document.getElementById('mov-entry-value').value) : 0,
                reason: document.getElementById('mov-reason').value
            };

            const res = await fetch(`/api/estoque/movements?email=${currentUser.email}`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(movement)
            });

            if (res.ok) {
                closeModal('modal-movement');
                loadProducts();
            }
        });
    }
}

// Modals
function openModal(id) { 
    const modal = document.getElementById(id);
    if (modal) modal.classList.remove('hidden'); 
}
function closeModal(id) { 
    const modal = document.getElementById(id);
    if (modal) modal.classList.add('hidden'); 
}

async function loadProducts() {
    updateStats();
    const res = await fetch(`/api/estoque/products?email=${currentUser.email}`);
    const allProducts = await res.json();
    productsCache = allProducts;
    updateDatalists();
    renderProducts(allProducts);
}

function renderProducts(allProducts) {
    const list = document.getElementById('products-list');
    if (!list) return;
    list.innerHTML = '';

    const searchTerm = (document.getElementById('filter-product-name')?.value || '').toLowerCase();
    const searchDate = document.getElementById('filter-product-date')?.value;

    // Calcula quantidades para os filtros (sempre sobre o total ou sobre a busca?)
    // O pedido original era "ter filtro pela letra e por data".
    // Geralmente os badges de contagem refletem o estado sem o filtro de texto/data, ou com ele?
    // Vou manter as contagens baseadas apenas no tipo para consistência com o layout anterior, 
    // mas filtrar a exibição.

    const countTodos = allProducts.length;
    const countComprado = allProducts.filter(p => p.type === 'comprado').length;
    const countCriado = allProducts.filter(p => p.type === 'criado').length;
    const countMinimo = allProducts.filter(p => p.quantity <= p.min_quantity).length;

    const btnTodos = document.querySelector('#view-products .filter-btn[data-type="todos"]');
    const btnComprado = document.querySelector('#view-products .filter-btn[data-type="comprado"]');
    const btnCriado = document.querySelector('#view-products .filter-btn[data-type="criado"]');
    const btnMinimo = document.querySelector('#view-products .filter-btn[data-type="minimo"]');

    if (btnTodos) btnTodos.innerText = `Todos (${countTodos})`;
    if (btnComprado) btnComprado.innerText = `Comprados (${countComprado})`;
    if (btnCriado) btnCriado.innerText = `Criados (${countCriado})`;
    if (btnMinimo) btnMinimo.innerText = `Estoque Mínimo (${countMinimo})`;

    let products = allProducts;
    
    // 1. Filtro por Tipo
    if (currentFilter === 'comprado') {
        products = products.filter(p => p.type === 'comprado');
    } else if (currentFilter === 'criado') {
        products = products.filter(p => p.type === 'criado');
    } else if (currentFilter === 'minimo') {
        products = products.filter(p => p.quantity <= p.min_quantity);
    }

    // 2. Filtro por Letra/Nome
    if (searchTerm) {
        products = products.filter(p => p.name.toLowerCase().includes(searchTerm));
    }

    // 3. Filtro por Data
    if (searchDate) {
        products = products.filter(p => {
            // Verifica se o produto foi criado na data ou se algum lote foi criado na data
            const prodDate = p.created_at ? p.created_at.split('T')[0] : '';
            if (prodDate === searchDate) return true;
            
            if (p.batches) {
                return p.batches.some(b => b.created_at && b.created_at.split('T')[0] === searchDate);
            }
            return false;
        });
    }

    // Ordena: produtos com estoque > 0 primeiro, depois os zerados
    products.sort((a, b) => {
        if (a.quantity > 0 && b.quantity === 0) return -1;
        if (a.quantity === 0 && b.quantity > 0) return 1;
        return 0;
    });

    products.forEach(p => {
        const isLow = p.quantity <= p.min_quantity;
        const typeClass = p.type === 'comprado' ? 'type-bought' : 'type-created';
        
        // Exibe lotes individuais com data para maior controle (FIFO)
        let batchHtml = '';
        if (p.batches) {
            p.batches.forEach((b, index) => {
                if (b.quantity <= 0) return;
                const batchNum = index; // Começa de 0 conforme pedido ("FIFO 0")
                const val = p.type === 'comprado' ? b.value : b.production_cost;
                const date = new Date(b.created_at).toLocaleDateString('pt-BR');
                batchHtml += `<span class="batch-info" style="border: 1px solid #4a7c59; font-weight: bold;" title="Lote ${batchNum} de ${date}">
                    Lote ${batchNum}: ${b.quantity}x R$ ${parseFloat(val).toFixed(2)} (${date})
                </span>`;
            });
        }

        const card = document.createElement('div');
        card.className = `product-card ${isLow ? 'low' : ''}`;
        card.innerHTML = `
            <div style="display: flex; justify-content: space-between; align-items: flex-start;">
                <div>
                    <h3 style="margin: 0;">${p.name} <span class="type-badge ${typeClass}">${p.type}</span></h3>
                    <p style="margin: 0.5rem 0;">
                        Qtd Total: <span class="${isLow ? 'low-stock' : ''}">${p.quantity}</span> 
                        <small>(Mín: ${p.min_quantity})</small>
                    </p>
                    <div style="margin-bottom: 0.5rem;">${batchHtml}</div>
                    ${p.type === 'comprado' ? `<p style="margin: 0; font-size: 0.85rem; color: #666;">Local: ${p.location}</p>` : ''}
                </div>
                <div style="display: flex; flex-direction: column; gap: 0.5rem;">
                    <div style="display: flex; gap: 0.25rem;">
                        <button onclick="openMovement(${p.id}, 'entrada', ${p.quantity}, ${p.type === 'comprado' ? p.value : p.production_cost})" class="btn-small bg-success" title="Entrada">➕</button>
                        <button onclick="openMovement(${p.id}, 'saida', ${p.quantity}, ${p.type === 'comprado' ? p.value : p.production_cost})" class="btn-small bg-destructive" title="Saída">➖</button>
                    </div>
                    ${p.quantity === 0 ? `<button onclick="deleteProduct(${p.id}, ${p.quantity})" class="btn-small" style="background: #ccc;">Excluir</button>` : ''}
                </div>
            </div>
        `;
        list.appendChild(card);
    });
}

function openMovement(productId, type, currentQty, purchaseValue) {
    document.getElementById('mov-product-id').value = productId;
    document.getElementById('mov-type').value = type;
    document.getElementById('mov-batch-id').value = ''; // Limpa seleção anterior
    const titleEl = document.getElementById('mov-title');
    const stockInfoEl = document.getElementById('mov-stock-info');
    const batchesInfoEl = document.getElementById('mov-batches-info');
    const btnConfirm = document.getElementById('btn-confirm-movement');
    const groupOut = document.getElementById('group-out-type');
    const groupSale = document.getElementById('group-sale-value');
    const groupEntry = document.getElementById('group-entry-value');
    const labelEntry = document.getElementById('label-entry-value');
    const inputQty = document.getElementById('mov-qty');
    const inputSaleValue = document.getElementById('mov-sale-value');
    const inputEntryValue = document.getElementById('mov-entry-value');
    const totalDisplay = document.getElementById('mov-total-display');
    
    stockInfoEl.innerText = `Estoque Atual: ${currentQty}`;
    totalDisplay.style.display = 'none';
    
    const p = productsCache.find(prod => prod.id === productId);

    // Exibe detalhamento de lotes individuais com data
    batchesInfoEl.innerHTML = '';
    if (p && p.batches) {
        let label = document.createElement('div');
        label.style = 'font-size: 0.8rem; color: #4a7c59; font-weight: bold; margin-bottom: 4px;';
        label.innerText = type === 'saida' ? 'Selecione um Lote (ou deixe nenhum para FIFO):' : 'Lotes Disponíveis:';
        batchesInfoEl.appendChild(label);
        
        p.batches.forEach((b, index) => {
            if (b.quantity <= 0) return;
            const batchNum = index;
            const val = p.type === 'comprado' ? b.value : b.production_cost;
            const date = new Date(b.created_at).toLocaleDateString('pt-BR');
            const span = document.createElement('span');
            span.className = 'batch-info' + (type === 'saida' ? ' batch-select' : '');
            span.style = 'border: 1px solid #4a7c59; font-weight: bold;';
            span.innerText = `Lote ${batchNum}: ${b.quantity}x R$ ${parseFloat(val).toFixed(2)} (${date})`;
            
            if (type === 'saida') {
                span.onclick = () => {
                    // Toggle seleção
                    const isSelected = span.classList.contains('selected');
                    document.querySelectorAll('.batch-select').forEach(s => s.classList.remove('selected'));
                    if (!isSelected) {
                        span.classList.add('selected');
                        document.getElementById('mov-batch-id').value = b.id;
                        // Ao selecionar lote, podemos sugerir o valor de venda baseado no custo do lote se for venda
                        const outType = document.querySelector('input[name="out-type"]:checked').value;
                        if (outType === 'venda') {
                            document.getElementById('mov-sale-value').value = val;
                        }
                    } else {
                        document.getElementById('mov-batch-id').value = '';
                    }
                    updateMovementTotal();
                };
            }
            
            batchesInfoEl.appendChild(span);
        });
    }
    
    // Armazena o valor de compra para uso dinâmico
    window.currentPurchaseValue = purchaseValue;

    if (type === 'entrada') {
        titleEl.innerText = 'Entrada de Estoque';
        btnConfirm.className = 'btn-wide bg-success';
        groupOut.classList.add('hidden');
        groupSale.classList.add('hidden');
        groupEntry.classList.remove('hidden');
        labelEntry.innerText = p.type === 'comprado' ? 'Valor Pago (R$)' : 'Custo de Produção (R$)';
        inputQty.removeAttribute('max');
        inputSaleValue.value = 0;
        inputEntryValue.value = purchaseValue;
    } else {
        titleEl.innerText = 'Saída de Estoque';
        groupOut.classList.remove('hidden');
        groupSale.classList.remove('hidden'); // Venda é o padrão
        groupEntry.classList.add('hidden');
        inputQty.setAttribute('max', currentQty);
        // Reset sub-type to venda and updates color
        document.querySelector('input[name="out-type"][value="venda"]').checked = true;
        btnConfirm.className = 'btn-wide bg-success';
        inputSaleValue.value = purchaseValue;
        inputEntryValue.value = 0;
    }
    
    document.getElementById('mov-qty').value = 1;
    document.getElementById('mov-reason').value = '';
    updateMovementTotal();
    openModal('modal-movement');
}

function updateMovementTotal() {
    const type = document.getElementById('mov-type').value;
    const qty = parseFloat(document.getElementById('mov-qty').value) || 0;
    const totalDisplay = document.getElementById('mov-total-display');
    
    let unitValue = 0;
    let showTotal = false;

    if (type === 'entrada') {
        unitValue = parseFloat(document.getElementById('mov-entry-value').value) || 0;
        showTotal = true;
    } else if (type === 'saida') {
        const outType = document.querySelector('input[name="out-type"]:checked').value;
        if (outType === 'venda') {
            unitValue = parseFloat(document.getElementById('mov-sale-value').value) || 0;
            showTotal = true;
        }
    }

    if (showTotal && qty > 0) {
        const total = qty * unitValue;
        totalDisplay.innerText = `Total: R$ ${total.toLocaleString('pt-BR', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
        totalDisplay.style.display = 'block';
    } else {
        totalDisplay.style.display = 'none';
    }
}

// Listeners para atualizar total
document.getElementById('mov-qty').addEventListener('input', updateMovementTotal);
document.getElementById('mov-entry-value').addEventListener('input', updateMovementTotal);
document.getElementById('mov-sale-value').addEventListener('input', updateMovementTotal);

// listener para mudar cor do botão de confirmar no modal de movimento
document.querySelectorAll('input[name="out-type"]').forEach(radio => {
    radio.addEventListener('change', (e) => {
        const btnConfirm = document.getElementById('btn-confirm-movement');
        const groupSale = document.getElementById('group-sale-value');
        const inputSaleValue = document.getElementById('mov-sale-value');
        const outType = e.target.value;
        if (outType === 'venda') {
            btnConfirm.className = 'btn-wide bg-success';
            groupSale.classList.remove('hidden');
            inputSaleValue.value = window.currentPurchaseValue || 0;
        } else {
            btnConfirm.className = 'btn-wide bg-destructive';
            groupSale.classList.add('hidden');
            inputSaleValue.value = 0;
        }
        updateMovementTotal();
    });
});

async function deleteProduct(id, quantity) {
    if (quantity !== 0) {
        alert('Não é possível excluir um produto com estoque. Zere o estoque primeiro.');
        return;
    }
    if (confirm('Deseja realmente excluir este produto?')) {
        const res = await fetch(`/api/estoque/products?id=${id}&email=${currentUser.email}`, { method: 'DELETE' });
        if (res.ok) {
            loadProducts();
        } else {
            const error = await res.text();
            alert(error);
        }
    }
}

function updateDatalists() {
    const boughtList = document.getElementById('bought-names');
    const createdList = document.getElementById('created-names');
    boughtList.innerHTML = '';
    createdList.innerHTML = '';
    productsCache.forEach(p => {
        const option = document.createElement('option');
        option.value = p.name;
        if (p.type === 'comprado') boughtList.appendChild(option);
        else createdList.appendChild(option);
    });
}

document.getElementById('bought-name').addEventListener('input', (e) => {
    const name = e.target.value;
    const p = productsCache.find(prod => prod.name === name && prod.type === 'comprado');
    if (p) {
        document.getElementById('bought-min').value = p.min_quantity;
        document.getElementById('bought-location').value = p.location;
        document.getElementById('bought-value').value = p.value;
    }
});

document.getElementById('created-name').addEventListener('input', (e) => {
    const name = e.target.value;
    const p = productsCache.find(prod => prod.name === name && prod.type === 'criado');
    if (p) {
        document.getElementById('created-min').value = p.min_quantity;
        document.getElementById('created-cost').value = p.production_cost;
    }
});

async function loadMovements() {
    updateStats();
    const start = document.getElementById('mov-start-date').value;
    const end = document.getElementById('mov-end-date').value;
    
    let url = `/api/estoque/movements?email=${currentUser.email}`;
    if (start) url += `&start_date=${start}`;
    if (end) url += `&end_date=${end}`;

    const res = await fetch(url);
    const movements = await res.json();
    const list = document.getElementById('movements-list');
    list.innerHTML = '';

    movements.forEach(m => {
        const date = new Date(m.created_at).toLocaleString('pt-BR');
        let badgeClass = 'out';
        if (m.type === 'entrada' || m.type === 'cadastro' || m.type === 'venda') badgeClass = 'in';
        else if (m.type === 'editado') badgeClass = 'edit';
        else if (m.type === 'saida' || m.type === 'baixa') badgeClass = 'out';

        // Cores específicas para Venda (Verde) e Baixa (Vermelho)
        let customBadgeClass = badgeClass;
        if (m.type === 'venda') customBadgeClass = 'venda';
        else if (m.type === 'baixa') customBadgeClass = 'baixa';

        let saleValueText = '';
        if (m.type === 'venda' && m.sale_value > 0) {
            saleValueText = ` | Valor Venda: R$ ${m.sale_value.toFixed(2)}`;
        }

        // Mostra o valor do lote se disponível (para entradas/cadastros)
        let batchValueText = '';
        if (m.batch) {
            const bVal = m.product && m.product.type === 'comprado' ? m.batch.value : m.batch.production_cost;
            if (bVal > 0) {
                batchValueText = ` | Valor Lote: R$ ${bVal.toFixed(2)}`;
            }
        } else if (m.batch_value > 0) {
            batchValueText = ` | Valor Lote: R$ ${m.batch_value.toFixed(2)}`;
        }

        const item = document.createElement('div');
        item.className = 'product-card';
        
        const operationTotal = m.quantity * (m.type === 'venda' ? m.sale_value : m.batch_value);
        let totalColor = '#4a7c59'; // verde para venda
        if (m.type === 'baixa') totalColor = '#d9534f'; // vermelho para baixa

        item.innerHTML = `
            <div style="display: flex; justify-content: space-between;">
                <span><strong>${m.product ? m.product.name : 'Produto Removido'}</strong></span>
                <span class="movement-badge ${customBadgeClass}">${m.type.toUpperCase()}</span>
            </div>
            <p style="margin: 0.5rem 0;">Qtd: ${m.quantity}${saleValueText}${batchValueText} | ${m.reason || 'Sem motivo'}</p>
            ${m.type === 'venda' || m.type === 'baixa' ? `<p style="margin: -0.25rem 0 0.5rem 0; font-size: 0.85rem; color: ${totalColor}; font-weight: bold;">Total: R$ ${operationTotal.toFixed(2)}</p>` : ''}
            <small style="color: #999;">${date}</small>
        `;
        list.appendChild(item);
    });
}

async function updateStats() {
    const start = document.getElementById('mov-start-date')?.value || '';
    const end = document.getElementById('mov-end-date')?.value || '';
    
    let url = `/api/estoque/stats?email=${currentUser.email}`;
    if (start) url += `&start_date=${start}`;
    if (end) url += `&end_date=${end}`;

    const res = await fetch(url);
    if (res.ok) {
        const stats = await res.json();
        document.getElementById('stat-stock-value').innerText = `Estoque: R$ ${stats.stock_value.toLocaleString('pt-BR', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
        document.getElementById('stat-sale-total').innerHTML = `<span style="color: #4a7c59;">Vendas: R$ ${stats.sale_total.toLocaleString('pt-BR', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}</span>`;
        document.getElementById('stat-baixa-total').innerHTML = `<span style="color: #d9534f;">Baixas: R$ ${stats.baixa_total.toLocaleString('pt-BR', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}</span>`;
    }
}

// Initial load
document.querySelectorAll('.filter-btn').forEach(btn => {
    btn.addEventListener('click', () => {
        document.querySelectorAll('.filter-btn').forEach(b => b.classList.remove('active'));
        btn.classList.add('active');
        currentFilter = btn.dataset.type;
        loadProducts();
    });
});
