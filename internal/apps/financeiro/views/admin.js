(function() {
  const SESSION_KEY = 'financeiro_user';
  let currentUser = null;

  const elements = {
    adminTabUsers: document.getElementById('admin-tab-users'),
    adminTabPayments: document.getElementById('admin-tab-payments'),
    adminViewUsers: document.getElementById('admin-view-users'),
    adminViewPayments: document.getElementById('admin-view-payments'),
    adminUserList: document.getElementById('admin-user-list'),
    userCount: document.getElementById('user-count'),
    filterApp: document.getElementById('filter-app'),
    adminPaymentList: document.getElementById('admin-payment-list'),
    btnNewPayment: document.getElementById('btn-new-payment'),
    btnPaymentHistory: document.getElementById('btn-payment-history'),
    paymentModalOverlay: document.getElementById('payment-modal-overlay'),
    closePaymentModal: document.getElementById('close-payment-modal'),
    paymentForm: document.getElementById('payment-form'),
    paymentUserId: document.getElementById('payment-user-id'),
    btnLogout: document.getElementById('btn-logout-admin')
  };

  async function init() {
    const session = localStorage.getItem(SESSION_KEY);
    if (!session) {
      window.location.href = 'login.html';
      return;
    }

    currentUser = JSON.parse(session);
    if (currentUser.role !== 'admin') {
      alert("Acesso negado.");
      window.location.href = 'index.html';
      return;
    }

    setupEventListeners();
    loadAdminUsers();
  }

  function setupEventListeners() {
    elements.adminTabUsers.onclick = () => switchAdminTab('users');
    elements.adminTabPayments.onclick = () => switchAdminTab('payments');
    if (elements.filterApp) {
      elements.filterApp.onchange = () => {
        loadAdminUsers();
      };
    }
    elements.btnNewPayment.onclick = () => openPaymentModal();
    elements.btnPaymentHistory.onclick = () => {
        const history = !elements.btnPaymentHistory.getAttribute('data-history');
        elements.btnPaymentHistory.setAttribute('data-history', history ? '1' : '');
        elements.btnPaymentHistory.innerText = history ? 'Ver Pendentes' : 'Ver Histórico';
        loadAdminPayments();
    };
    elements.closePaymentModal.onclick = () => elements.paymentModalOverlay.classList.add('hidden');
    elements.paymentForm.onsubmit = handlePaymentSubmit;
    elements.btnLogout.onclick = handleLogout;
  }

  function handleLogout() {
    localStorage.removeItem(SESSION_KEY);
    localStorage.removeItem('agenda-session');
    localStorage.removeItem('role');
    window.location.href = 'login.html';
  }

  function switchAdminTab(tab) {
    elements.adminTabUsers.classList.toggle('active', tab === 'users');
    elements.adminTabPayments.classList.toggle('active', tab === 'payments');
    elements.adminViewUsers.classList.toggle('hidden', tab !== 'users');
    elements.adminViewPayments.classList.toggle('hidden', tab !== 'payments');

    if (tab === 'users') {
      loadAdminUsers();
    } else {
      loadAdminPayments();
    }
  }

  async function loadAdminUsers() {
    const colspan = 9;
    elements.adminUserList.innerHTML = `<tr><td colspan="${colspan}" style="text-align:center;">Carregando...</td></tr>`;
    try {
      const resp = await fetch('/api/financeiro/admin/users');
      if (resp.ok) {
        let allUsers = await resp.json();
        let filteredUsers = allUsers;
        
        // Aplicar filtro se existir
        if (elements.filterApp) {
          const filter = elements.filterApp.value;
          if (filter === 'agenda') {
            filteredUsers = allUsers.filter(u => u.has_agenda);
          } else if (filter === 'financeiro') {
            filteredUsers = allUsers.filter(u => u.has_financeiro);
          } else if (filter === 'both') {
            filteredUsers = allUsers.filter(u => u.has_agenda && u.has_financeiro);
          }
        }

        if (elements.userCount) {
          elements.userCount.innerText = filteredUsers.length;
        }
        renderAdminUsers(filteredUsers);
      } else {
        elements.adminUserList.innerHTML = `<tr><td colspan="${colspan}" style="text-align:center; color:red;">Erro ao carregar usuários.</td></tr>`;
      }
    } catch (err) {
      elements.adminUserList.innerHTML = `<tr><td colspan="${colspan}" style="text-align:center; color:red;">Erro de conexão.</td></tr>`;
    }
  }

  function renderAdminUsers(users) {
    const colspan = 9;
    if (users.length === 0) {
      elements.adminUserList.innerHTML = `<tr><td colspan="${colspan}" style="text-align:center;">Nenhum usuário encontrado.</td></tr>`;
      return;
    }

    const now = new Date();
    elements.adminUserList.innerHTML = users.map(u => {
      const createdAt = new Date(u.created_at);
      const diffMs = now - createdAt;
      const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));
      const diffHours = Math.floor(diffMs / (1000 * 60 * 60)) % 24;
      const diffMins = Math.floor(diffMs / (1000 * 60)) % 60;

      let timeText = "";
      if (diffDays > 0) timeText = `${diffDays}d ${diffHours}h`;
      else if (diffHours > 0) timeText = `${diffHours}h ${diffMins}m`;
      else timeText = `${diffMins}m`;

      return `
        <tr>
          <td><strong>${u.name}</strong> ${u.role === 'admin' ? '🛡️' : ''}</td>
          <td>${u.email}</td>
          <td>
            <div style="display: flex; gap: 8px; align-items: center;">
              <label title="Acesso Agenda" style="cursor: pointer; display: flex; align-items: center; gap: 2px;">
                <input type="checkbox" ${u.has_agenda ? 'checked' : ''} onchange="toggleUserAccess('${u.email}', this.checked, ${u.has_financeiro})">
                📅
              </label>
              <label title="Acesso Financeiro" style="cursor: pointer; display: flex; align-items: center; gap: 2px;">
                <input type="checkbox" ${u.has_financeiro ? 'checked' : ''} onchange="toggleUserAccess('${u.email}', ${u.has_agenda}, this.checked)">
                💰
              </label>
            </div>
          </td>
          <td>${u.business || ''}</td>
          <td>
            <input type="number" value="${u.max_professionals || 1}" min="1" max="50" style="width: 50px; padding: 4px; border: 1px solid #ccc; border-radius: 4px;" onchange="updateUserPermission('${u.email}', this.value, ${u.free_scheduling})">
          </td>
          <td>
            <input type="checkbox" ${u.free_scheduling !== false ? 'checked' : ''} onchange="updateUserPermission('${u.email}', ${u.max_professionals || 1}, this.checked)">
          </td>
          <td>${createdAt.toLocaleDateString('pt-BR')} ${createdAt.toLocaleTimeString('pt-BR', {hour:'2-digit', minute:'2-digit'})}</td>
          <td>${timeText}</td>
          <td>
            <button class="bg-primary" style="border: none; padding: 4px 8px; border-radius: 4px; color: white; cursor: pointer; font-size: 0.75rem; margin-right: 4px;" onclick="openPaymentForUser('${u.ID || u.id}')">Agendar Pagto</button>
            ${u.email !== currentUser.email && u.email !== 'edynobre@gmail.com' ? `
              <button class="bg-destructive" style="border: none; padding: 4px 8px; border-radius: 4px; color: white; cursor: pointer; font-size: 0.75rem;" onclick="deleteUser('${u.email}')">Excluir</button>
            ` : ''}
          </td>
        </tr>
      `;
    }).join('');
  }

  window.updateUserPermission = async function(email, maxProfs, freeScheduling) {
    try {
      const resp = await fetch('/api/financeiro/admin/update-permission', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ 
          email, 
          max_professionals: parseInt(maxProfs),
          free_scheduling: !!freeScheduling
        })
      });
      if (!resp.ok) alert('Erro ao atualizar permissão');
    } catch (err) {
      console.error(err);
      alert('Erro de conexão');
    }
  };

  window.toggleUserAccess = async function(email, hasAgenda, hasFinanceiro) {
    try {
      const resp = await fetch('/api/financeiro/admin/update-access', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ 
          email, 
          has_agenda: !!hasAgenda,
          has_financeiro: !!hasFinanceiro
        })
      });
      if (resp.ok) {
        loadAdminUsers(); // Recarrega para refletir mudanças se necessário
      } else {
        alert('Erro ao atualizar acesso');
        loadAdminUsers();
      }
    } catch (err) {
      console.error(err);
      alert('Erro de conexão');
      loadAdminUsers();
    }
  };

  window.deleteUser = async function(email) {
    if (!confirm(`Tem certeza que deseja excluir o usuário ${email} e TODOS os seus dados financeiros e de agenda? Esta ação é irreversível!`)) return;
    try {
      const resp = await fetch(`/api/financeiro/admin/users?email=${encodeURIComponent(email)}&admin_email=${encodeURIComponent(currentUser.email)}`, {
        method: 'DELETE'
      });
      if (resp.ok) {
        alert('Usuário e todos os seus dados foram excluídos!');
        loadAdminUsers();
      } else {
        const text = await resp.text();
        alert('Erro: ' + text);
      }
    } catch (err) {
      console.error(err);
      alert('Erro de conexão');
    }
  };

  async function loadAdminPayments() {
    elements.adminPaymentList.innerHTML = '<tr><td colspan="6" style="text-align:center;">Carregando...</td></tr>';
    try {
      const history = !!elements.btnPaymentHistory.getAttribute('data-history');
      const resp = await fetch(`/api/financeiro/admin/payments${history ? '?history=1' : ''}`);
      if (resp.ok) {
        const payments = await resp.json();
        renderAdminPayments(payments);
      } else {
        elements.adminPaymentList.innerHTML = '<tr><td colspan="6" style="text-align:center; color:red;">Erro ao carregar pagamentos.</td></tr>';
      }
    } catch (err) {
      elements.adminPaymentList.innerHTML = '<tr><td colspan="6" style="text-align:center; color:red;">Erro de conexão.</td></tr>';
    }
  }

  function renderAdminPayments(payments) {
    if (payments.length === 0) {
      elements.adminPaymentList.innerHTML = '<tr><td colspan="6" style="text-align:center;">Nenhum pagamento encontrado.</td></tr>';
      return;
    }

    elements.adminPaymentList.innerHTML = payments.map(p => {
      const dueDate = new Date(p.due_date);
      let statusClass = 'unpaid';
      if (p.status === 'paid') statusClass = 'paid';
      else if (new Date() > dueDate) statusClass = 'overdue';

      return `
        <tr>
          <td>${p.user ? p.user.email : 'N/A'}</td>
          <td>${p.reference}</td>
          <td>${formatBRL(p.amount)}</td>
          <td>${dueDate.toLocaleDateString('pt-BR')}</td>
          <td><span class="dot ${statusClass}"></span> ${p.status === 'paid' ? 'Pago' : (statusClass === 'overdue' ? 'Atrasado' : 'Pendente')}</td>
          <td>
            ${p.status !== 'paid' ? `<button onclick="updatePaymentStatus(${p.ID}, 'paid')" style="padding: 4px 8px; border-radius: 4px; cursor: pointer; font-size: 0.75rem;">Marcar Pago</button>` : `<button onclick="updatePaymentStatus(${p.ID}, 'pending')" style="padding: 4px 8px; border-radius: 4px; cursor: pointer; font-size: 0.75rem;">Reverter</button>`}
          </td>
        </tr>
      `;
    }).join('');
  }

  window.updatePaymentStatus = async function(id, status) {
    try {
      const resp = await fetch('/api/financeiro/admin/payments', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ id, status })
      });
      if (resp.ok) loadAdminPayments();
      else alert('Erro ao atualizar status');
    } catch (err) {
      console.error(err);
    }
  };

  window.openPaymentForUser = async function(userId) {
    openPaymentModal(userId);
  };

  async function openPaymentModal(userId = null) {
    elements.paymentModalOverlay.classList.remove('hidden');
    elements.paymentUserId.innerHTML = '<option value="">Carregando usuários...</option>';
    try {
        const resp = await fetch('/api/financeiro/admin/users');
        if (resp.ok) {
            const users = await resp.json();
            elements.paymentUserId.innerHTML = users.map(u => 
                `<option value="${u.ID}" ${userId == u.ID ? 'selected' : ''}>${u.email} (${u.name})</option>`
            ).join('');
        }
    } catch (e) {
        console.error(e);
    }
  }

  async function handlePaymentSubmit(e) {
    e.preventDefault();
    const dueDateVal = document.getElementById('payment-due-date').value;
    const data = {
        user_id: parseInt(document.getElementById('payment-user-id').value),
        reference: document.getElementById('payment-reference').value,
        amount: parseFloat(document.getElementById('payment-amount').value),
        due_date: dueDateVal ? new Date(dueDateVal + 'T12:00:00').toISOString() : null
    };

    try {
        const resp = await fetch('/api/financeiro/admin/payments', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(data)
        });
        if (resp.ok) {
            elements.paymentModalOverlay.classList.add('hidden');
            elements.paymentForm.reset();
            loadAdminPayments();
        } else {
            const text = await resp.text();
            alert('Erro ao criar pagamento: ' + text);
        }
    } catch (err) {
        console.error(err);
    }
  }

  function formatBRL(val) {
    return new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(val);
  }

  init();
})();
