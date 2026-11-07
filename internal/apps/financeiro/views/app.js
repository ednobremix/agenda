let currentUser = null;
let expenses = [];
let incomes = [];
let banks = [];
let paymentTypes = [];

let currentFilter = {
    month: new Date().getMonth(), // 0-11 or 'all'
    year: new Date().getFullYear(),
    view: 'list' // 'list' or 'calendar'
};

document.addEventListener('DOMContentLoaded', () => {
    const userData = localStorage.getItem('financeiro_user');
    if (!userData) {
        window.location.href = 'login.html';
        return;
    }
    currentUser = JSON.parse(userData);

    const isAdmin = currentUser.role === 'admin';
    const hasFinanceiro = isAdmin || currentUser.has_financeiro;

    if (!hasFinanceiro) {
        alert("Acesso negado. Esta conta não possui permissão para o App Financeiro.");
        localStorage.removeItem('financeiro_user');
        localStorage.removeItem('agenda-session');
        localStorage.removeItem('role');
        localStorage.removeItem('has_agenda');
        localStorage.removeItem('has_financeiro');
        window.location.href = '/';
        return;
    }

    document.getElementById('user-email-display').textContent = currentUser.email;

    const adminLink = document.getElementById('admin-link');
    if (adminLink && currentUser.role === 'admin') {
      adminLink.classList.remove('hidden');
    }

    initApp();
});

function initApp() {
    // Top Actions
    document.getElementById('btn-config').addEventListener('click', () => showView('view-config'));
    
    // Logout
    document.getElementById('btn-logout').addEventListener('click', () => {
        localStorage.removeItem('financeiro_user');
        localStorage.removeItem('agenda-session');
        localStorage.removeItem('role');
        localStorage.removeItem('has_agenda');
        localStorage.removeItem('has_financeiro');
        window.location.href = 'login.html';
    });

    // Tabs
    const tabs = document.querySelectorAll('.nav-tab');
    tabs.forEach(tab => {
        tab.addEventListener('click', () => {
            tabs.forEach(t => t.classList.remove('active'));
            tab.classList.add('active');
            
            const target = tab.getAttribute('data-target');
            showView(target);
        });
    });

    // Config Forms
    document.getElementById('form-add-bank').addEventListener('submit', addBank);
    document.getElementById('form-add-payment-type').addEventListener('submit', addPaymentType);
    document.getElementById('btn-show-change-password').addEventListener('click', () => {
        document.getElementById('modal-password').classList.remove('hidden');
    });
    document.getElementById('form-change-password').addEventListener('submit', changePassword);

    // Modal Gastos
    document.getElementById('btn-new-expense').addEventListener('click', () => {
        const now = new Date();
        const months = ["Janeiro", "Fevereiro", "Março", "Abril", "Maio", "Junho", "Julho", "Agosto", "Setembro", "Outubro", "Novembro", "Dezembro"];
        document.getElementById('exp-month').value = months[now.getMonth()];
        document.getElementById('exp-date').value = now.toISOString().split('T')[0];
        document.getElementById('modal-expense').classList.remove('hidden');
    });

    // Modal Recebimentos
    document.getElementById('btn-new-income').addEventListener('click', () => {
        const now = new Date();
        const months = ["Janeiro", "Fevereiro", "Março", "Abril", "Maio", "Junho", "Julho", "Agosto", "Setembro", "Outubro", "Novembro", "Dezembro"];
        document.getElementById('inc-month').value = months[now.getMonth()];
        document.getElementById('inc-date').value = now.toISOString().split('T')[0];
        document.getElementById('inc-expense-id').value = '';
        document.getElementById('inc-source').value = '';
        document.getElementById('inc-category').value = 'Recebimento';
        updateLoansDatalist();
        document.getElementById('modal-income').classList.remove('hidden');
    });

    document.getElementById('inc-category').addEventListener('change', (e) => {
        const isLoan = e.target.value === 'Empréstimo';
        const sourceInput = document.getElementById('inc-source');
        if (isLoan) {
            sourceInput.setAttribute('list', 'open-loans');
        } else {
            sourceInput.removeAttribute('list');
            sourceInput.value = '';
            document.getElementById('inc-expense-id').value = '';
        }
    });

    document.getElementById('inc-source').addEventListener('input', (e) => {
        const val = e.target.value;
        const datalist = document.getElementById('open-loans');
        const option = Array.from(datalist.options).find(opt => opt.value === val);
        if (option) {
            document.getElementById('inc-expense-id').value = option.getAttribute('data-id');
            // Opcional: auto-preencher valor se for vínculo
            const loan = expenses.find(exp => exp.id == option.getAttribute('data-id'));
            if (loan) {
                document.getElementById('inc-amount').value = loan.amount.toLocaleString('pt-BR', { minimumFractionDigits: 2 });
            }
        } else {
            document.getElementById('inc-expense-id').value = '';
        }
    });

    // Forms
    document.getElementById('form-expense').addEventListener('submit', saveExpense);
    document.getElementById('form-income').addEventListener('submit', saveIncome);

    // Extrato
    document.getElementById('btn-filter-statement').addEventListener('click', renderStatement);

    // Filters
    initFilters();

    loadData();
    loadConfig();
    checkGoogleLink();
}

function showView(viewId) {
    document.querySelectorAll('.tab-content').forEach(c => c.classList.add('hidden'));
    document.getElementById(viewId).classList.remove('hidden');

    // Update active tab if needed
    if (viewId === 'view-expenses' || viewId === 'view-incomes' || viewId === 'view-statement') {
        const tabs = document.querySelectorAll('.nav-tab');
        tabs.forEach(t => {
            if (t.getAttribute('data-target') === viewId) t.classList.add('active');
            else t.classList.remove('active');
        });
    }

    if (viewId === 'view-statement') {
        initStatementDates();
        renderStatement();
    }
}

function initStatementDates() {
    const startInput = document.getElementById('statement-start-date');
    const endInput = document.getElementById('statement-end-date');
    
    if (!startInput.value || !endInput.value) {
        const now = new Date();
        const firstDay = new Date(now.getFullYear(), now.getMonth(), 1);
        const lastDay = new Date(now.getFullYear(), now.getMonth() + 1, 0);
        
        startInput.value = firstDay.toISOString().split('T')[0];
        endInput.value = lastDay.toISOString().split('T')[0];
    }
}

function renderStatement() {
    const typeFilter = document.getElementById('statement-type').value;
    const startDate = document.getElementById('statement-start-date').value;
    const endDate = document.getElementById('statement-end-date').value;
    const listContainer = document.getElementById('statement-list');
    
    let filtered = expenses.filter(exp => {
        const pDate = new Date(exp.purchase_date).toISOString().split('T')[0];
        const typeMatch = typeFilter === 'all' || exp.payment_type === typeFilter;
        const dateMatch = (!startDate || pDate >= startDate) && (!endDate || pDate <= endDate);
        return typeMatch && dateMatch;
    });

    // Ordenar por data
    filtered.sort((a, b) => new Date(b.purchase_date) - new Date(a.purchase_date));

    listContainer.innerHTML = '';
    let total = 0;

    filtered.forEach(exp => {
        total += exp.amount;
        const isLoan = exp.category === 'Empréstimo';
        const valueColor = isLoan ? '#854d0e' : '#dc2626';
        const bgColor = isLoan ? '#fef9c3' : 'white';
        
        const item = document.createElement('div');
        item.className = 'data-item';
        if (isLoan) item.style.backgroundColor = bgColor;
        
        item.innerHTML = `
            <div class="data-info">
                <h4>${exp.description} ${isLoan ? '<span class="tag-loan" style="font-size:0.7rem; background:#eab308; color:white; padding:1px 4px; border-radius:4px; margin-left:5px;">Empréstimo</span>' : ''}</h4>
                <p>${exp.payment_type} | ${new Date(exp.purchase_date).toLocaleDateString()}</p>
            </div>
            <div style="display: flex; align-items: center;">
                <span class="data-value" style="color: ${valueColor};">- ${formatCurrency(exp.amount)}</span>
            </div>
        `;
        listContainer.appendChild(item);
    });

    document.getElementById('statement-total').textContent = formatCurrency(total);
}

function closeModal(modalId) {
    document.getElementById(modalId).classList.add('hidden');
}

function initFilters() {
    const monthSel = document.getElementById('filter-month');
    const yearSel = document.getElementById('filter-year');
    
    monthSel.value = currentFilter.month;
    
    const currentYear = new Date().getFullYear();
    for (let y = currentYear - 2; y <= currentYear + 2; y++) {
        const opt = document.createElement('option');
        opt.value = y;
        opt.textContent = y;
        if (y === currentFilter.year) opt.selected = true;
        yearSel.appendChild(opt);
    }

    monthSel.addEventListener('change', (e) => {
        currentFilter.month = e.target.value === 'all' ? 'all' : parseInt(e.target.value);
        renderAll();
    });
    yearSel.addEventListener('change', (e) => {
        currentFilter.year = parseInt(e.target.value);
        renderAll();
    });

    document.getElementById('btn-list-view').addEventListener('click', () => {
        currentFilter.view = 'list';
        document.getElementById('btn-list-view').classList.add('active');
        document.getElementById('btn-calendar-view').classList.remove('active');
        renderAll();
    });

    document.getElementById('btn-calendar-view').addEventListener('click', () => {
        currentFilter.view = 'calendar';
        document.getElementById('btn-calendar-view').classList.add('active');
        document.getElementById('btn-list-view').classList.remove('active');
        renderAll();
    });
}

function renderAll() {
    renderExpenses();
    renderIncomes();
    updateSummary();
}

async function loadData() {
    const email = currentUser.email;
    try {
        const [expRes, incRes] = await Promise.all([
            fetch(`/api/financeiro/expenses?email=${email}`),
            fetch(`/api/financeiro/incomes?email=${email}`)
        ]);

        if (expRes.ok) expenses = await expRes.json();
        if (incRes.ok) incomes = await incRes.json();

        renderAll();
    } catch (err) {
        console.error('Erro ao carregar dados:', err);
    }
}

async function loadConfig() {
    const email = currentUser.email;
    try {
        const [bankRes, typeRes] = await Promise.all([
            fetch(`/api/financeiro/banks?email=${email}`),
            fetch(`/api/financeiro/payment-types?email=${email}`)
        ]);

        if (bankRes.ok) banks = await bankRes.json();
        if (typeRes.ok) paymentTypes = await typeRes.json();

        renderBanks();
        renderPaymentTypes();
        updateDatalists();
    } catch (err) {
        console.error('Erro ao carregar configurações:', err);
    }
}

function renderBanks() {
    const container = document.getElementById('banks-list');
    container.innerHTML = '';
    banks.forEach(b => {
        const item = document.createElement('div');
        item.className = 'tag-item';
        item.innerHTML = `
            <span>${b.name}</span>
            <button onclick="deleteConfig('banks', ${b.id})">✕</button>
        `;
        container.appendChild(item);
    });
}

function renderPaymentTypes() {
    const container = document.getElementById('payment-types-list');
    container.innerHTML = '';
    paymentTypes.forEach(t => {
        const item = document.createElement('div');
        item.className = 'tag-item';
        item.innerHTML = `
            <span>${t.name}</span>
            <button onclick="deleteConfig('payment-types', ${t.id})">✕</button>
        `;
        container.appendChild(item);
    });
}

function updateDatalists() {
    updateLoansDatalist();
    
    // Update Statement Type select
    const statementType = document.getElementById('statement-type');
    if (statementType) {
        const currentVal = statementType.value;
        statementType.innerHTML = '<option value="all">Todos</option>';
        paymentTypes.forEach(t => {
            const opt = document.createElement('option');
            opt.value = t.name;
            opt.textContent = t.name;
            statementType.appendChild(opt);
        });
        statementType.value = currentVal || 'all';
    }

    const bankDatalist = document.getElementById('banks');
    bankDatalist.innerHTML = '';
    banks.forEach(b => {
        const opt = document.createElement('option');
        opt.value = b.name;
        bankDatalist.appendChild(opt);
    });

    const typeDatalist = document.getElementById('payment-types');
    typeDatalist.innerHTML = '';
    paymentTypes.forEach(t => {
        const opt = document.createElement('option');
        opt.value = t.name;
        typeDatalist.appendChild(opt);
    });
}

async function addBank(e) {
    e.preventDefault();
    const name = document.getElementById('bank-name').value;
    const res = await fetch('/api/financeiro/banks', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ user_email: currentUser.email, name })
    });
    if (res.ok) {
        document.getElementById('bank-name').value = '';
        loadConfig();
    }
}

async function addPaymentType(e) {
    e.preventDefault();
    const name = document.getElementById('payment-type-name').value;
    const res = await fetch('/api/financeiro/payment-types', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ user_email: currentUser.email, name })
    });
    if (res.ok) {
        document.getElementById('payment-type-name').value = '';
        loadConfig();
    }
}

async function deleteConfig(type, id) {
    if (!confirm('Deseja excluir este item?')) return;
    const res = await fetch(`/api/financeiro/${type}?id=${id}`, { method: 'DELETE' });
    if (res.ok) loadConfig();
}

async function changePassword(e) {
    e.preventDefault();
    const current_password = document.getElementById('current-pwd').value;
    const new_password = document.getElementById('new-pwd').value;

    const res = await fetch(`/api/user/update-password?email=${currentUser.email}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ current_password, new_password })
    });

    if (res.ok) {
        alert('Senha atualizada com sucesso!');
        closeModal('modal-password');
        document.getElementById('form-change-password').reset();
    } else {
        const msg = await res.text();
        alert('Erro: ' + msg);
    }
}

async function checkGoogleLink() {
    const res = await fetch(`/api/me?email=${currentUser.email}`);
    if (res.ok) {
        const user = await res.json();
        const status = document.getElementById('google-status');
        const container = document.getElementById('google-link-button');

        if (user.google_id) {
            status.textContent = 'Conta vinculada ao Google ✅';
            container.innerHTML = '';
        } else {
            status.textContent = 'Conta não vinculada ao Google ⚠️';
            renderGoogleButton();
        }
    }
}

function renderGoogleButton() {
    const container = document.getElementById('google-link-button');
    google.accounts.id.initialize({
        client_id: "1037474087695-h03skslnhv1q4074r32l1p67pks3p06u.apps.googleusercontent.com",
        callback: handleGoogleLink
    });
    google.accounts.id.renderButton(container, { theme: "outline", size: "large" });
}

async function handleGoogleLink(response) {
    const res = await fetch(`/api/user/link-google?email=${currentUser.email}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ credential: response.credential })
    });

    if (res.ok) {
        alert('Conta vinculada com sucesso!');
        checkGoogleLink();
    } else {
        const msg = await res.text();
        alert('Erro ao vincular: ' + msg);
    }
}

function getFilteredData(data, type) {
    const months = ["Janeiro", "Fevereiro", "Março", "Abril", "Maio", "Junho", "Julho", "Agosto", "Setembro", "Outubro", "Novembro", "Dezembro"];
    
    return data.filter(item => {
        const date = new Date(type === 'expense' ? item.purchase_date : item.received_date);
        const yearMatch = date.getFullYear() === currentFilter.year;
        
        let monthMatch = true;
        if (currentFilter.month !== 'all') {
            const itemMonth = type === 'expense' ? item.payment_month : item.received_month;
            monthMatch = itemMonth === months[currentFilter.month];
        }
        
        return yearMatch && monthMatch;
    });
}

function updateLoansDatalist() {
    const datalist = document.getElementById('open-loans');
    datalist.innerHTML = '';
    
    // Pegar IDs de gastos já vinculados a recebimentos
    const linkedExpenseIds = incomes
        .filter(inc => inc.expense_id)
        .map(inc => inc.expense_id);

    expenses.filter(exp => exp.category === 'Empréstimo' && !linkedExpenseIds.includes(exp.id))
        .forEach(exp => {
            const opt = document.createElement('option');
            opt.value = `${exp.description} - ${exp.payment_month} - ${formatCurrency(exp.amount)}`;
            opt.setAttribute('data-id', exp.id);
            datalist.appendChild(opt);
        });
}

function renderExpenses() {
    const listContainer = document.getElementById('expenses-list');
    const calContainer = document.getElementById('expenses-calendar');
    
    const filtered = getFilteredData(expenses, 'expense');

    if (currentFilter.view === 'list') {
        listContainer.classList.remove('hidden');
        calContainer.classList.add('hidden');
        
        listContainer.innerHTML = '';
        filtered.forEach(exp => {
            const isLoan = exp.category === 'Empréstimo';
            const isPaid = incomes.some(inc => inc.expense_id === exp.id);
            
            const valueColor = isLoan ? '#854d0e' : '#dc2626';
            const bgColor = isLoan ? '#fef9c3' : 'white';
            
            const item = document.createElement('div');
            item.className = 'data-item';
            if (isLoan) item.style.backgroundColor = bgColor;
            
            item.innerHTML = `
                <div class="data-info">
                    <h4>
                        ${isPaid ? '<span style="margin-right:5px;">✅</span>' : ''}
                        ${exp.description} 
                        ${isLoan ? '<span class="tag-loan" style="font-size:0.7rem; background:#eab308; color:white; padding:1px 4px; border-radius:4px; margin-left:5px;">Empréstimo</span>' : ''}
                    </h4>
                    <p>${exp.payment_month} | ${exp.payment_type} | ${exp.installments}x</p>
                    <p>${new Date(exp.purchase_date).toLocaleDateString()}</p>
                </div>
                <div style="display: flex; align-items: center;">
                    <span class="data-value" style="color: ${valueColor};">- ${formatCurrency(exp.amount)}</span>
                    ${isPaid ? '' : `<button class="btn-delete" onclick="deleteItem('expenses', ${exp.id})">🗑️</button>`}
                </div>
            `;
            listContainer.appendChild(item);
        });
    } else {
        listContainer.classList.add('hidden');
        calContainer.classList.remove('hidden');
        renderCalendar(calContainer, filtered, 'expense');
    }
}

function renderIncomes() {
    const listContainer = document.getElementById('incomes-list');
    const calContainer = document.getElementById('incomes-calendar');
    
    const filtered = getFilteredData(incomes, 'income');

    if (currentFilter.view === 'list') {
        listContainer.classList.remove('hidden');
        calContainer.classList.add('hidden');
        
        listContainer.innerHTML = '';
        filtered.forEach(inc => {
            const isLoan = inc.category === 'Empréstimo';
            const valueColor = isLoan ? '#854d0e' : '#059669';
            const bgColor = isLoan ? '#fef9c3' : 'white';

            const item = document.createElement('div');
            item.className = 'data-item';
            if (isLoan) item.style.backgroundColor = bgColor;

            item.innerHTML = `
                <div class="data-info">
                    <h4>${inc.source} ${isLoan ? '<span class="tag-loan" style="font-size:0.7rem; background:#eab308; color:white; padding:1px 4px; border-radius:4px; margin-left:5px;">Empréstimo</span>' : ''}</h4>
                    <p>${inc.received_month} | ${inc.bank}</p>
                    <p>${new Date(inc.received_date).toLocaleDateString()}</p>
                </div>
                <div style="display: flex; align-items: center;">
                    <span class="data-value" style="color: ${valueColor};">+ ${formatCurrency(inc.amount)}</span>
                    <button class="btn-delete" onclick="deleteItem('incomes', ${inc.id})">🗑️</button>
                </div>
            `;
            listContainer.appendChild(item);
        });
    } else {
        listContainer.classList.add('hidden');
        calContainer.classList.remove('hidden');
        renderCalendar(calContainer, filtered, 'income');
    }
}

function renderCalendar(container, data, type) {
    if (currentFilter.month === 'all') {
        container.innerHTML = '<p style="text-align:center; padding:2rem;">O modo calendário não está disponível para visualização anual. Selecione um mês específico.</p>';
        return;
    }

    const year = currentFilter.year;
    const month = currentFilter.month;
    
    const firstDay = new Date(year, month, 1).getDay();
    const daysInMonth = new Date(year, month + 1, 0).getDate();
    
    let html = `
        <div class="calendar-grid">
            <div class="calendar-day-head">Dom</div>
            <div class="calendar-day-head">Seg</div>
            <div class="calendar-day-head">Ter</div>
            <div class="calendar-day-head">Qua</div>
            <div class="calendar-day-head">Qui</div>
            <div class="calendar-day-head">Sex</div>
            <div class="calendar-day-head">Sáb</div>
    `;
    
    // Empty days
    for (let i = 0; i < firstDay; i++) {
        html += '<div class="calendar-day empty"></div>';
    }
    
    const today = new Date();
    const isCurrentMonth = today.getFullYear() === year && today.getMonth() === month;

    for (let d = 1; d <= daysInMonth; d++) {
        const isToday = isCurrentMonth && today.getDate() === d;
        const dateStr = `${year}-${String(month + 1).padStart(2, '0')}-${String(d).padStart(2, '0')}`;
        
        const dayEvents = data.filter(item => {
            const itemDate = new Date(type === 'expense' ? item.purchase_date : item.received_date);
            return itemDate.getDate() === d;
        });

        html += `
            <div class="calendar-day ${isToday ? 'cal-day-today' : ''}">
                <div class="calendar-day-num">${d}</div>
                ${dayEvents.map(e => {
                    const isLoan = e.category === 'Empréstimo';
                    let isPaid = false;
                    if (type === 'expense' && isLoan) {
                        isPaid = incomes.some(inc => inc.expense_id === e.id);
                    }
                    return `
                        <div class="cal-event ${type} ${isLoan ? 'loan' : ''}">
                            ${isPaid ? '✅ ' : ''}${type === 'expense' ? e.description : e.source} (${formatCurrency(e.amount)})
                        </div>
                    `;
                }).join('')}
            </div>
        `;
    }
    
    html += '</div>';
    container.innerHTML = html;
}

function updateSummary() {
    const filteredExp = getFilteredData(expenses, 'expense');
    const filteredInc = getFilteredData(incomes, 'income');

    const totalExp = filteredExp.reduce((sum, e) => sum + e.amount, 0);
    const totalInc = filteredInc.reduce((sum, i) => sum + i.amount, 0);
    const balance = totalInc - totalExp;

    document.getElementById('total-expense').textContent = formatCurrency(totalExp);
    document.getElementById('total-income').textContent = formatCurrency(totalInc);
    document.getElementById('total-balance').textContent = formatCurrency(balance);
}

async function saveExpense(e) {
    e.preventDefault();
    const data = {
        user_email: currentUser.email,
        description: document.getElementById('exp-desc').value,
        payment_month: document.getElementById('exp-month').value,
        installments: parseInt(document.getElementById('exp-installments').value),
        purchase_date: new Date(document.getElementById('exp-date').value).toISOString(),
        amount: parseFloat(document.getElementById('exp-amount').value.replace(',', '.')),
        payment_type: document.getElementById('exp-type').value,
        category: document.getElementById('exp-category').value
    };

    const res = await fetch('/api/financeiro/expenses', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data)
    });

    if (res.ok) {
        closeModal('modal-expense');
        document.getElementById('form-expense').reset();
        loadData();
    }
}

async function saveIncome(e) {
    e.preventDefault();
    const data = {
        user_email: currentUser.email,
        source: document.getElementById('inc-source').value,
        received_month: document.getElementById('inc-month').value,
        received_date: new Date(document.getElementById('inc-date').value).toISOString(),
        amount: parseFloat(document.getElementById('inc-amount').value.replace(',', '.')),
        bank: document.getElementById('inc-bank').value,
        category: document.getElementById('inc-category').value,
        expense_id: document.getElementById('inc-expense-id').value ? parseInt(document.getElementById('inc-expense-id').value) : null
    };

    const res = await fetch('/api/financeiro/incomes', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data)
    });

    if (res.ok) {
        closeModal('modal-income');
        document.getElementById('form-income').reset();
        loadData();
    }
}

async function deleteItem(type, id) {
    if (!confirm('Deseja excluir este item?')) return;

    const res = await fetch(`/api/financeiro/${type}?id=${id}`, { method: 'DELETE' });
    if (res.ok) loadData();
}

function formatCurrency(value) {
    return value.toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' });
}
