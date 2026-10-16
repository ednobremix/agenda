(function() {
  // Configurações e Estado
  const USERS_KEY = 'agenda-users';
  const SESSION_KEY = 'agenda-session';
  
  let currentUser = null;
  let appointments = [];
  let gisInited = false;

  const CLIENT_ID = "665567298487-am8coo2krrd2msdno64p3ng5om1pfr75.apps.googleusercontent.com"; // Substitua pelo seu Client ID
  let config = {
    businessType: '',
    professionals: [],
    services: [
      { }
    ]
  };

  function getStorageKey() {
    return `agenda-data-${currentUser.email}`;
  }

  function getConfigKey() {
    return `agenda-config-${currentUser.email}`;
  }

  let currentView = 'list';
  let calendarMode = 'month'; // 'month', 'week', 'day'
  let currentMonth = new Date();
  let selectedDate = null;
  let formState = {
    payment: false,
    repeat: 'none'
  };

  // Elementos DOM
  const elements = {
    successBanner: document.getElementById('success-banner'),
    successMessage: document.getElementById('success-message'),
    totalReceived: document.getElementById('total-received'),
    totalPending: document.getElementById('total-pending'),
    btnNewAppointment: document.getElementById('btn-new-appointment'),
    viewListBtn: document.getElementById('view-list-btn'),
    viewCalendarBtn: document.getElementById('view-calendar-btn'),
    viewList: document.getElementById('view-list'),
    viewCalendar: document.getElementById('view-calendar'),
    modalOverlay: document.getElementById('modal-overlay'),
    closeModal: document.getElementById('close-modal'),
    appointmentForm: document.getElementById('appointment-form'),
    btnConfirm: document.getElementById('btn-confirm'),
    formError: document.getElementById('form-error'),
    repeatOptions: document.getElementById('repeat-options'),
    repeatCount: document.getElementById('repeat-count'),
    repeatInfo: document.getElementById('repeat-info'),
    conflictBox: document.getElementById('conflict-box'),
    conflictList: document.getElementById('conflict-list'),
    conflictTitle: document.getElementById('conflict-title'),
    forceSchedule: document.getElementById('force-schedule'),
    listContainer: document.getElementById('list-container'),
    filterWhen: document.getElementById('filter-when'),
    filterPayment: document.getElementById('filter-payment'),
    filterProfessional: document.getElementById('filter-professional'),
    filterCount: document.getElementById('filter-count'),
    btnClearFilters: document.getElementById('btn-clear-filters'),
    calendarTitle: document.getElementById('calendar-title'),
    calendarViewContainer: document.getElementById('calendar-view-container'),
    prevPeriod: document.getElementById('prev-period'),
    nextPeriod: document.getElementById('next-period'),
    calViewMonth: document.getElementById('cal-view-month'),
    calViewWeek: document.getElementById('cal-view-week'),
    calViewDay: document.getElementById('cal-view-day'),
    dayDetail: document.getElementById('day-detail'),
    detailDate: document.getElementById('detail-date'),
    detailSessions: document.getElementById('detail-sessions'),
    inputPatient: document.getElementById('patient'),
    inputTherapist: document.getElementById('therapist'),
    inputType: document.getElementById('type'),
    inputValue: document.getElementById('value'),
    inputDuration: document.getElementById('duration'),
    inputDateTime: document.getElementById('datetime'),
    btnSettings: document.getElementById('btn-settings'),
    settingsOverlay: document.getElementById('settings-overlay'),
    closeSettings: document.getElementById('close-settings'),
    settingsForm: document.getElementById('settings-form'),
    servicesList: document.getElementById('services-list'),
    btnAddService: document.getElementById('btn-add-service'),
    professionalsList: document.getElementById('professionals-list'),
    btnAddProfessional: document.getElementById('btn-add-professional'),
    multipleProfessionalsWarning: document.getElementById('multiple-professionals-warning'),
    configBusinessType: document.getElementById('config-business-type'),
    headerTitle: document.getElementById('header-business-title'),
    headerDesc: document.getElementById('header-business-desc'),
    modalTitle: document.querySelector('#modal-overlay h2'),
    therapistLabel: document.querySelector('label[for="therapist"]'),
    therapistInput: document.getElementById('therapist'),
    typeLabel: document.querySelector('label[for="type"]'),
    
    // Auth elements
    authContainer: document.getElementById('auth-container'),
    appContent: document.getElementById('app-content'),
    loginForm: document.getElementById('login-form'),
    registerForm: document.getElementById('register-form'),
    loginCard: document.getElementById('login-card'),
    registerCard: document.getElementById('register-card'),
    goToRegister: document.getElementById('go-to-register'),
    goToLogin: document.getElementById('go-to-login'),
    loginError: document.getElementById('login-error'),
    regError: document.getElementById('reg-error'),
    btnLogout: document.getElementById('btn-logout'),
    userInfo: document.getElementById('user-info'),
    debugUserList: document.getElementById('debug-user-list'),
    
    // Admin elements
    btnAdmin: document.getElementById('btn-admin'),
    adminOverlay: document.getElementById('admin-overlay'),
    closeAdmin: document.getElementById('close-admin'),
    adminUserList: document.getElementById('admin-user-list')
  };

  // --- Inicialização ---

  function init() {
    checkSession();
    setupEventListeners();
    initGoogleAuth();
  }

  function initGoogleAuth() {
    if (typeof google === 'undefined') {
      setTimeout(initGoogleAuth, 100);
      return;
    }

    google.accounts.id.initialize({
      client_id: CLIENT_ID,
      callback: handleCredentialResponse,
      auto_select: true
    });

    gisInited = true;

    google.accounts.id.renderButton(
      document.getElementById("google-login-btn"),
      { theme: "outline", size: "large", width: 300 }
    );

    google.accounts.id.renderButton(
      document.getElementById("google-register-btn"),
      { theme: "outline", size: "large", width: 300 }
    );
  }

  async function handleCredentialResponse(response) {
    try {
      const resp = await fetch('/api/auth/google', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ credential: response.credential })
      });

      if (!resp.ok) {
        const errorText = await resp.text();
        alert('Erro no login com Google: ' + errorText);
        return;
      }

      const user = await resp.json();
      currentUser = user;
      localStorage.setItem(SESSION_KEY, JSON.stringify(currentUser));
      showApp();
    } catch (err) {
      console.error('Erro na autenticação Google:', err);
      alert('Ocorreu um erro ao tentar entrar com o Google.');
    }
  }

  async function checkSession() {
    const session = localStorage.getItem(SESSION_KEY);
    if (session) {
      currentUser = JSON.parse(session);
      // Revalida perfil para atualizar roles/permissões
      try {
        const resp = await fetch(`/api/me?email=${encodeURIComponent(currentUser.email)}`);
        if (resp.ok) {
          const freshUser = await resp.json();
          currentUser = freshUser;
          localStorage.setItem(SESSION_KEY, JSON.stringify(currentUser));
        }
      } catch (err) {
        console.error("Erro ao revalidar sessão:", err);
      }
      await showApp();
    } else {
      showAuth();
    }
  }

  function showAuth() {
    elements.authContainer.classList.remove('hidden');
    elements.appContent.classList.add('hidden');
    showLogin();
  }

  async function showApp() {
    elements.authContainer.classList.add('hidden');
    elements.appContent.classList.remove('hidden');
    
    // Reset state before loading new user data
    appointments = [];
    config = {
      businessType: currentUser.business || 'Massoterapia',
      professionals: [currentUser.name],
      services: [
      ]
    };

    await loadConfig();

    // Se carregar config e não tiver profissionais, usa o nome do usuário
    if (!config.professionals || config.professionals.length === 0) {
      config.professionals = [currentUser.name];
      await saveConfig();
    }
    await loadData();
    updateFinanceSummary();
    applyConfig();
    
    if (elements.userInfo) {
      elements.userInfo.textContent = `Logado como: ${currentUser.name} (${currentUser.email})`;
      if (currentUser.role === 'admin') {
          elements.userInfo.textContent += " (Administrador)";
      }
    }

    // Toggle Admin Button based on role
    if (elements.btnAdmin) {
      const isAdmin = currentUser && currentUser.role === 'admin';
      elements.btnAdmin.classList.toggle('hidden', !isAdmin);
    }
    
    render();
  }

  function showLogin() {
    elements.loginCard.classList.remove('hidden');
    elements.registerCard.classList.add('hidden');
    elements.loginError.classList.add('hidden');
    updateDebugUserList();
  }

  async function updateDebugUserList() {
    if (!elements.debugUserList) return;
    try {
      const resp = await fetch('/api/debug/users');
      const users = await resp.json();
      if (users.length === 0) {
        elements.debugUserList.innerHTML = '<li>Nenhum usuário cadastrado</li>';
        return;
      }
      
      elements.debugUserList.innerHTML = '';
      users.forEach(user => {
        const li = document.createElement('li');
        li.textContent = user.email;
        li.style.cursor = 'pointer';
        li.style.textDecoration = 'underline';
        li.style.marginBottom = '0.25rem';
        li.addEventListener('click', () => {
          document.getElementById('login-email').value = user.email;
          document.getElementById('login-password').value = user.password;
        });
        elements.debugUserList.appendChild(li);
      });
    } catch (err) {
      elements.debugUserList.innerHTML = '<li>Erro ao carregar lista</li>';
    }
  }

  function showRegister() {
    elements.loginCard.classList.add('hidden');
    elements.registerCard.classList.remove('hidden');
    elements.regError.classList.add('hidden');
  }

  async function handleLogin(e) {
    e.preventDefault();
    const email = document.getElementById('login-email').value;
    const password = document.getElementById('login-password').value;

    try {
      const resp = await fetch('/api/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, password })
      });

      if (resp.ok) {
        const user = await resp.json();
        currentUser = user;
        localStorage.setItem(SESSION_KEY, JSON.stringify(user));
        await showApp();
        elements.loginForm.reset();
      } else {
        elements.loginError.classList.remove('hidden');
      }
    } catch (err) {
      elements.loginError.textContent = "Erro de conexão.";
      elements.loginError.classList.remove('hidden');
    }
  }

  async function handleRegister(e) {
    e.preventDefault();
    const name = document.getElementById('reg-name').value;
    const business = document.getElementById('reg-business').value;
    const email = document.getElementById('reg-email').value;
    const password = document.getElementById('reg-password').value;

    try {
      const resp = await fetch('/api/register', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name, business, email, password })
      });

      if (resp.ok) {
        const newUser = await resp.json();
        // Auto login after register
        currentUser = newUser;
        localStorage.setItem(SESSION_KEY, JSON.stringify(newUser));
        await showApp();
        elements.registerForm.reset();
      } else if (resp.status === 409) {
        elements.regError.textContent = "E-mail já cadastrado.";
        elements.regError.classList.remove('hidden');
      } else {
        elements.regError.textContent = "Erro ao cadastrar.";
        elements.regError.classList.remove('hidden');
      }
    } catch (err) {
      elements.regError.textContent = "Erro de conexão.";
      elements.regError.classList.remove('hidden');
    }
  }

  function handleLogout() {
    localStorage.removeItem(SESSION_KEY);
    currentUser = null;
    if (typeof google !== 'undefined' && google.accounts.id) {
      google.accounts.id.disableAutoSelect();
    }
    if (elements.btnAdmin) {
      elements.btnAdmin.classList.add('hidden');
    }
    showAuth();
  }

  async function loadConfig() {
    try {
      const resp = await fetch(`/api/config?email=${encodeURIComponent(currentUser.email)}`);
      if (resp.ok) {
        config = await resp.json();
        // Garantir que arrays existam
        config.professionals = config.professionals || [];
        config.services = config.services || [];
      }
    } catch (err) {
      console.error("Erro ao carregar config:", err);
    }
  }

  async function saveConfig() {
    try {
      await fetch(`/api/config?email=${encodeURIComponent(currentUser.email)}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(config)
      });
      applyConfig();
    } catch (err) {
      console.error("Erro ao salvar config:", err);
    }
  }

  function applyConfig() {
    elements.headerTitle.textContent = `Agenda de ${config.businessType}`;
    
    // Ajustar labels e ícones com base no ramo
    const bizLower = config.businessType.toLowerCase();
    const isPsychology = bizLower.includes('psic') || bizLower.includes('terapia');
    const isBarber = bizLower.includes('barber') || bizLower.includes('cabel');
    const isMedical = bizLower.includes('medic') || bizLower.includes('dent');
    
    if (isPsychology) {
      elements.headerDesc.textContent = 'Organize suas sessões com tranquilidade.';
      elements.therapistLabel.textContent = 'Nome do Psicólogo(a)';
      elements.typeLabel.textContent = 'Tipo de Sessão';
    } else if (isBarber) {
      elements.headerDesc.textContent = 'Gestão ágil para seu negócio.';
      elements.therapistLabel.textContent = 'Nome do Profissional';
      elements.typeLabel.textContent = 'Serviço';
    } else if (isMedical) {
      elements.headerDesc.textContent = 'Controle seus atendimentos clínicos.';
      elements.therapistLabel.textContent = 'Nome do Médico(a)';
      elements.typeLabel.textContent = 'Procedimento';
    } else {
      elements.headerDesc.textContent = 'Seus atendimentos organizados, com leveza.';
      elements.therapistLabel.textContent = 'Nome do Profissional';
      elements.typeLabel.textContent = 'Tipo de Serviço';
    }

    // Atualizar select de profissionais no filtro
    const currentFilterProfessional = elements.filterProfessional.value;
    elements.filterProfessional.innerHTML = '<option value="all">Todos</option>' + (config.professionals || []).map(p => 
      `<option value="${p}">${p}</option>`
    ).join('');
    
    if (config.professionals && config.professionals.includes(currentFilterProfessional)) {
      elements.filterProfessional.value = currentFilterProfessional;
    } else {
      elements.filterProfessional.value = 'all';
    }

    // Atualizar select de profissionais no modal de agendamento
    const currentTherapist = elements.inputTherapist.value;
    elements.inputTherapist.innerHTML = (config.professionals || []).map(p => 
      `<option value="${p}">${p}</option>`
    ).join('');
    
    if (config.professionals && config.professionals.includes(currentTherapist)) {
      elements.inputTherapist.value = currentTherapist;
    } else if (config.professionals && config.professionals.length > 0) {
      elements.inputTherapist.value = config.professionals[0];
    }

    // Atualizar select de tipos de serviço no modal de agendamento
    const currentType = elements.inputType.value;
    elements.inputType.innerHTML = (config.services || []).map(s => 
      `<option value="${s.name}" data-value="${s.value}" data-duration="${s.duration}">${s.name} - ${formatBRL(s.value)} (${s.duration} min)</option>`
    ).join('');
    
    // Tentar manter o valor selecionado ou pegar o primeiro
    if (config.services && Array.isArray(config.services) && config.services.some(s => s.name === currentType)) {
      elements.inputType.value = currentType;
    } else if (config.services && Array.isArray(config.services) && config.services.length > 0) {
      elements.inputType.value = config.services[0].name;
      elements.inputValue.value = config.services[0].value.toFixed(2).replace('.', ',');
      elements.inputDuration.value = config.services[0].duration;
    }
  }

  async function loadData() {
    try {
      const resp = await fetch(`/api/appointments?email=${encodeURIComponent(currentUser.email)}`);
      if (resp.ok) {
        const data = await resp.json();
        appointments = data || [];
      }
    } catch (err) {
      console.error("Erro ao carregar dados:", err);
    }
  }

  async function saveData() {
    try {
      await fetch(`/api/appointments?email=${encodeURIComponent(currentUser.email)}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(appointments)
      });
      updateFinanceSummary();
    } catch (err) {
      console.error("Erro ao salvar dados:", err);
    }
  }

  // --- Lógica de Negócio ---

  function updateFinanceSummary() {
    const apps = appointments || [];
    const received = apps.filter(a => a.pago).reduce((sum, a) => sum + a.valor, 0);
    const pending = apps.filter(a => !a.pago).reduce((sum, a) => sum + a.valor, 0);
    elements.totalReceived.textContent = formatBRL(received);
    elements.totalPending.textContent = formatBRL(pending);
  }

  function generateAppointments(baseApp, repeat, count) {
    const apps = [];
    const baseDate = new Date(baseApp.dataHora);
    
    for (let i = 0; i < count; i++) {
      const newDate = new Date(baseDate);
      if (repeat === 'weekly') {
        newDate.setDate(baseDate.getDate() + (i * 7));
      } else if (repeat === 'biweekly') {
        newDate.setDate(baseDate.getDate() + (i * 14));
      } else if (repeat === 'monthly') {
        const day = baseDate.getDate();
        newDate.setMonth(baseDate.getMonth() + i);
        if (newDate.getDate() < day) {
            newDate.setDate(0); // Último dia do mês anterior
        }
      } else {
        apps.push({ ...baseApp, id: crypto.randomUUID() });
        break;
      }
      apps.push({ 
        ...baseApp, 
        id: crypto.randomUUID(), 
        dataHora: newDate.toISOString() 
      });
    }
    return apps;
  }

  function checkConflicts(newApps) {
    if (!elements.inputTherapist.value.trim()) return [];
    
    const therapist = elements.inputTherapist.value.trim().toLowerCase();
    const conflicts = [];
    const apps = appointments || [];

    newApps.forEach(newApp => {
      const startTime = new Date(newApp.dataHora).getTime();
      const endTime = startTime + (newApp.duracao * 60000);

      const match = apps.find(oldApp => {
        if (!oldApp.massoterapeuta) return false;
        if (oldApp.massoterapeuta.trim().toLowerCase() !== therapist) return false;
        
        const oldStartTime = new Date(oldApp.dataHora).getTime();
        const oldDuration = oldApp.duracao || 60; // fallback para antigos
        const oldEndTime = oldStartTime + (oldDuration * 60000);
        
        // Sobreposição: (StartA < EndB) && (EndA > StartB)
        return (startTime < oldEndTime) && (endTime > oldStartTime);
      });
      if (match) conflicts.push({ newApp, match });
    });

    return conflicts;
  }

  // --- UI Helpers ---

  function formatBRL(val) {
    return new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(val);
  }

  function formatDateShort(iso) {
    const d = new Date(iso);
    return d.toLocaleDateString('pt-BR', { day: '2-digit', month: 'short' }).replace('.', '');
  }

  function formatTime(iso) {
    const d = new Date(iso);
    return d.toLocaleTimeString('pt-BR', { hour: '2-digit', minute: '2-digit' });
  }

  function formatDateFull(date) {
    return date.toLocaleDateString('pt-BR', { day: 'numeric', month: 'long', year: 'numeric' });
  }

  function showBanner(message) {
    elements.successMessage.textContent = message;
    elements.successBanner.classList.remove('hidden');
    setTimeout(() => {
      elements.successBanner.classList.add('hidden');
    }, 5000);
  }

  // --- Handlers de Eventos ---

  function setupEventListeners() {
    // Auth listeners
    elements.goToRegister.onclick = (e) => { e.preventDefault(); showRegister(); };
    elements.goToLogin.onclick = (e) => { e.preventDefault(); showLogin(); };
    elements.loginForm.onsubmit = handleLogin;
    elements.registerForm.onsubmit = handleRegister;
    elements.btnLogout.onclick = handleLogout;

    elements.btnNewAppointment.onclick = openModal;
    elements.closeModal.onclick = closeModal;
    elements.modalOverlay.onclick = (e) => { if (e.target === elements.modalOverlay) closeModal(); };

    // Toggle Buttons (Pagamento e Repetir)
    document.querySelectorAll('.toggle-btn').forEach(btn => {
      btn.onclick = () => {
        const group = btn.parentElement;
        group.querySelectorAll('.toggle-btn').forEach(b => b.classList.remove('active'));
        btn.classList.add('active');

        if (btn.hasAttribute('data-value')) {
          formState.payment = btn.getAttribute('data-value') === 'true';
        } else if (btn.hasAttribute('data-repeat')) {
          formState.repeat = btn.getAttribute('data-repeat');
          elements.repeatOptions.classList.toggle('hidden', formState.repeat === 'none');
          updateRepeatInfo();
        }
        validateForm();
      };
    });

    elements.repeatCount.oninput = updateRepeatInfo;
    elements.appointmentForm.oninput = validateForm;
    elements.forceSchedule.onchange = validateForm;

    elements.appointmentForm.onsubmit = handleFormSubmit;

    // View Switcher
    elements.viewListBtn.onclick = () => switchView('list');
    elements.viewCalendarBtn.onclick = () => switchView('calendar');

    // Configurações
    elements.btnSettings.onclick = openSettings;
    elements.closeSettings.onclick = closeSettings;
    elements.settingsOverlay.onclick = (e) => { if (e.target === elements.settingsOverlay) closeSettings(); };
    elements.btnAddService.onclick = () => addServiceRow();
    elements.btnAddProfessional.onclick = () => addProfessionalRow();
    elements.settingsForm.onsubmit = handleSettingsSubmit;

    // Admin
    elements.btnAdmin.onclick = openAdmin;
    elements.closeAdmin.onclick = closeAdmin;
    elements.adminOverlay.onclick = (e) => { if (e.target === elements.adminOverlay) closeAdmin(); };

    // Atualizar valor e duração ao trocar o tipo de serviço no agendamento
    elements.inputType.onchange = () => {
      const selectedOption = elements.inputType.options[elements.inputType.selectedIndex];
      const val = selectedOption.getAttribute('data-value');
      const dur = selectedOption.getAttribute('data-duration');
      if (val) {
        elements.inputValue.value = parseFloat(val).toFixed(2).replace('.', ',');
      }
      if (dur) {
        elements.inputDuration.value = dur;
      }
      validateForm();
    };

    // Filtros
    elements.filterWhen.onchange = render;
    elements.filterPayment.onchange = render;
    elements.filterProfessional.onchange = render;
    elements.btnClearFilters.onclick = () => {
      elements.filterWhen.value = 'all';
      elements.filterPayment.value = 'all';
      elements.filterProfessional.value = 'all';
      render();
    };

    // Calendário
    elements.prevPeriod.onclick = () => { 
      if (calendarMode === 'month') currentMonth.setMonth(currentMonth.getMonth() - 1);
      else if (calendarMode === 'week') currentMonth.setDate(currentMonth.getDate() - 7);
      else currentMonth.setDate(currentMonth.getDate() - 1);
      render(); 
    };
    elements.nextPeriod.onclick = () => { 
      if (calendarMode === 'month') currentMonth.setMonth(currentMonth.getMonth() + 1);
      else if (calendarMode === 'week') currentMonth.setDate(currentMonth.getDate() + 7);
      else currentMonth.setDate(currentMonth.getDate() + 1);
      render(); 
    };

    elements.calViewMonth.onclick = () => switchCalendarMode('month');
    elements.calViewWeek.onclick = () => switchCalendarMode('week');
    elements.calViewDay.onclick = () => switchCalendarMode('day');
  }

  function switchCalendarMode(mode) {
    calendarMode = mode;
    elements.calViewMonth.classList.toggle('active', mode === 'month');
    elements.calViewWeek.classList.toggle('active', mode === 'week');
    elements.calViewDay.classList.toggle('active', mode === 'day');
    render();
  }

  function openModal(editingApp = null) {
    elements.appointmentForm.reset();
    formState.payment = false;
    formState.repeat = 'none';
    elements.repeatOptions.classList.add('hidden');
    elements.conflictBox.classList.add('hidden');
    elements.formError.classList.add('hidden');
    document.querySelectorAll('.toggle-btn').forEach(btn => btn.classList.remove('active'));
    document.getElementById('pay-false').classList.add('active');
    document.querySelector('[data-repeat="none"]').classList.add('active');
    
    if (editingApp) {
        // Lógica de edição se existir
    } else {
        // Se for novo, garante que o primeiro serviço da config seja selecionado e preencha valor/duração
        if (config.services && config.services.length > 0) {
            const firstService = config.services[0];
            elements.inputType.value = firstService.name;
            elements.inputValue.value = firstService.value.toFixed(2).replace('.', ',');
            elements.inputDuration.value = firstService.duration;
        } else {
            // Caso não tenha serviços cadastrados
            elements.inputValue.value = '';
            elements.inputDuration.value = '';
        }
    }

    elements.modalOverlay.classList.remove('hidden');
    validateForm();
  }

  function closeModal() {
    elements.modalOverlay.classList.add('hidden');
  }

  function openSettings() {
    elements.configBusinessType.value = config.businessType;
    
    elements.professionalsList.innerHTML = '';
    (config.professionals || []).forEach(p => addProfessionalRow(p));
    if (!config.professionals || config.professionals.length === 0) addProfessionalRow(currentUser.name);
    
    elements.servicesList.innerHTML = '';
    config.services.forEach(s => addServiceRow(s.name, s.value, s.duration));
    if (config.services.length === 0) addServiceRow();
    
    updateProfessionalButtons();
    elements.settingsOverlay.classList.remove('hidden');
  }

  function updateProfessionalButtons() {
    const rows = elements.professionalsList.querySelectorAll('.service-row');
    const count = rows.length;
    
    if (currentUser.max_professionals && count >= currentUser.max_professionals) {
      elements.btnAddProfessional.disabled = true;
      elements.btnAddProfessional.style.opacity = '0.5';
      elements.multipleProfessionalsWarning.textContent = `O limite de ${currentUser.max_professionals} profissional(is) foi atingido.`;
      elements.multipleProfessionalsWarning.classList.remove('hidden');
    } else {
      elements.btnAddProfessional.disabled = false;
      elements.btnAddProfessional.style.opacity = '1';
      elements.multipleProfessionalsWarning.classList.add('hidden');
    }
    
    // Impedir remover se for o único
    rows.forEach(row => {
        const btn = row.querySelector('.btn-remove-service');
        if (count <= 1) {
            btn.style.display = 'none';
        } else {
            btn.style.display = 'block';
        }
    });
  }

  function closeSettings() {
    elements.settingsOverlay.classList.add('hidden');
  }

  async function openAdmin() {
    if (currentUser.role !== 'admin') {
      alert("Acesso negado: Apenas administradores podem acessar esta área.");
      return;
    }

    elements.adminOverlay.classList.remove('hidden');
    elements.adminUserList.innerHTML = '<tr><td colspan="5" style="text-align:center;">Carregando...</td></tr>';
    
    try {
      const resp = await fetch('/api/admin/users');
      if (resp.ok) {
        const users = await resp.json();
        renderAdminUsers(users);
      } else {
        elements.adminUserList.innerHTML = '<tr><td colspan="5" style="text-align:center; color:red;">Erro ao carregar usuários.</td></tr>';
      }
    } catch (err) {
      elements.adminUserList.innerHTML = '<tr><td colspan="5" style="text-align:center; color:red;">Erro de conexão.</td></tr>';
    }
  }

  function closeAdmin() {
    elements.adminOverlay.classList.add('hidden');
  }

  function renderAdminUsers(users) {
    if (users.length === 0) {
      elements.adminUserList.innerHTML = '<tr><td colspan="5" style="text-align:center;">Nenhum usuário encontrado.</td></tr>';
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
          <td>${u.business}</td>
          <td>
            <input type="number" value="${u.max_professionals || 1}" min="1" max="50" style="width: 50px; padding: 4px; border: 1px solid #ccc; border-radius: 4px;" onchange="window.app.updateUserPermission('${u.email}', this.value)">
          </td>
          <td>${createdAt.toLocaleDateString('pt-BR')} ${createdAt.toLocaleTimeString('pt-BR', {hour:'2-digit', minute:'2-digit'})}</td>
          <td>${timeText}</td>
        </tr>
      `;
    }).join('');
  }

  async function updateUserPermission(email, maxProfs) {
    try {
      const resp = await fetch('/api/admin/update-permission', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, max_professionals: parseInt(maxProfs) })
      });
      if (!resp.ok) {
        alert('Erro ao atualizar permissão');
      } else {
          // Se for o próprio usuário logado, atualiza o estado local
          if (currentUser.email === email) {
              currentUser.max_professionals = parseInt(maxProfs);
              localStorage.setItem(SESSION_KEY, JSON.stringify(currentUser));
          }
      }
    } catch (err) {
      console.error(err);
      alert('Erro de conexão');
    }
  }

  function addServiceRow(name = '', value = '', duration = '') {
    const row = document.createElement('div');
    row.className = 'service-row';
    row.innerHTML = `
      <input type="text" placeholder="Serviço" value="${name}" required style="flex: 2;">
      <input type="number" placeholder="R$" value="${value}" step="0.01" required style="flex: 1;">
      <input type="number" placeholder="Min" value="${duration}" required style="flex: 0.8;">
      <button type="button" class="btn-remove-service">✕</button>
    `;
    row.querySelector('.btn-remove-service').onclick = () => {
        row.remove();
    };
    elements.servicesList.appendChild(row);
  }

  function addProfessionalRow(name = '') {
    const row = document.createElement('div');
    row.className = 'service-row';
    row.innerHTML = `
      <input type="text" placeholder="Nome do Profissional" value="${name}" required style="flex: 1;">
      <button type="button" class="btn-remove-service">✕</button>
    `;
    row.querySelector('.btn-remove-service').onclick = () => {
        row.remove();
        updateProfessionalButtons();
    };
    elements.professionalsList.appendChild(row);
    updateProfessionalButtons();
  }

  function handleSettingsSubmit(e) {
    e.preventDefault();
    
    const profRows = elements.professionalsList.querySelectorAll('.service-row');
    const newProfessionals = [];
    profRows.forEach(row => {
      newProfessionals.push(row.querySelector('input').value);
    });

    const rows = elements.servicesList.querySelectorAll('.service-row');
    const newServices = [];
    rows.forEach(row => {
      const inputs = row.querySelectorAll('input');
      newServices.push({
        name: inputs[0].value,
        value: parseFloat(inputs[1].value),
        duration: parseInt(inputs[2].value) || 60
      });
    });

    config.businessType = elements.configBusinessType.value;
    config.professionals = newProfessionals;
    config.services = newServices;
    
    saveConfig();
    closeSettings();
    showBanner('Configurações salvas com sucesso!');
    render();
  }

  function updateRepeatInfo() {
    if (formState.repeat === 'none' || !elements.inputDateTime.value) return;
    const count = parseInt(elements.repeatCount.value) || 1;
    const baseApp = { dataHora: elements.inputDateTime.value };
    const apps = generateAppointments(baseApp, formState.repeat, count);
    const lastDate = new Date(apps[apps.length - 1].dataHora);
    elements.repeatInfo.textContent = `${apps.length} datas · última em ${lastDate.toLocaleDateString('pt-BR')}`;
    elements.btnConfirm.textContent = `Confirmar ${apps.length} Agendamentos`;
    validateForm();
  }

  function validateForm() {
    const isBasicValid = elements.inputPatient.value && 
                         elements.inputTherapist.value && 
                         elements.inputType.value && 
                         elements.inputValue.value && 
                         elements.inputDuration.value && 
                         elements.inputDateTime.value;

    let hasConflict = false;
    if (isBasicValid) {
      const count = formState.repeat === 'none' ? 1 : (parseInt(elements.repeatCount.value) || 1);
      const valStr = elements.inputValue.value.replace(',', '.');
      const baseApp = {
        paciente: elements.inputPatient.value,
        massoterapeuta: elements.inputTherapist.value,
        tipo: elements.inputType.value,
        valor: parseFloat(valStr),
        duracao: parseInt(elements.inputDuration.value),
        dataHora: new Date(elements.inputDateTime.value).toISOString(),
        pago: formState.payment
      };
      
      const newApps = generateAppointments(baseApp, formState.repeat, count);
      const conflicts = checkConflicts(newApps);
      
      if (conflicts.length > 0) {
        hasConflict = true;
        elements.conflictBox.classList.remove('hidden');
        elements.conflictTitle.textContent = `⚠️ Conflito de horário em ${conflicts.length} data(s)`;
        elements.conflictList.innerHTML = conflicts.slice(0, 4).map(c => {
          const d = new Date(c.match.dataHora);
          const day = d.toLocaleDateString('pt-BR', { day: 'numeric', month: 'short' });
          const time = d.toLocaleTimeString('pt-BR', { hour: '2-digit', minute: '2-digit' });
          return `<div>${day} às ${time} — ${c.match.paciente} (${c.match.tipo})</div>`;
        }).join('');
        if (conflicts.length > 4) {
          elements.conflictList.innerHTML += `<div>... e mais ${conflicts.length - 4} coincidência(s)</div>`;
        }
      } else {
        elements.conflictBox.classList.add('hidden');
      }
    } else {
      elements.conflictBox.classList.add('hidden');
    }

    const canSubmit = isBasicValid && (!hasConflict || elements.forceSchedule.checked);
    elements.btnConfirm.disabled = !canSubmit;
    elements.btnConfirm.style.opacity = canSubmit ? '1' : '0.5';
    elements.btnConfirm.style.cursor = canSubmit ? 'pointer' : 'not-allowed';

    if (formState.repeat === 'none') {
        elements.btnConfirm.textContent = "Confirmar Agendamento";
    }

    return isBasicValid;
  }

  function handleFormSubmit(e) {
    e.preventDefault();
    if (!validateForm()) {
      elements.formError.classList.remove('hidden');
      return;
    }

    const valStr = elements.inputValue.value.replace(',', '.');
    const baseApp = {
      paciente: elements.inputPatient.value,
      massoterapeuta: elements.inputTherapist.value,
      tipo: elements.inputType.value,
      valor: parseFloat(valStr),
      duracao: parseInt(elements.inputDuration.value),
      dataHora: new Date(elements.inputDateTime.value).toISOString(),
      pago: formState.payment
    };

    const count = formState.repeat === 'none' ? 1 : parseInt(elements.repeatCount.value);
    const newApps = generateAppointments(baseApp, formState.repeat, count);
    
    appointments.push(...newApps);
    saveData();
    closeModal();
    
    let msg = `${newApps.length} sessão(ões) criada(s) para ${baseApp.paciente}`;
    if (formState.repeat !== 'none') {
        const period = { weekly: 'a cada semana', biweekly: 'a cada quinzena', monthly: 'a cada mês' }[formState.repeat];
        msg += ` (${period}).`;
    }
    showBanner(msg);
    render();
  }

  function togglePayment(id) {
    const app = appointments.find(a => a.id === id);
    if (app) {
      app.pago = !app.pago;
      saveData();
      render();
    }
  }

  function confirmCancel(id) {
    const card = document.querySelector(`[data-id="${id}"]`);
    if (!card) return;
    
    const originalContent = card.innerHTML;
    card.innerHTML = `
      <div style="text-align: center; padding: 1rem;">
        <p style="margin-bottom: 1rem; font-weight: 600;">Cancelar esta sessão?</p>
        <div style="display: flex; gap: 0.5rem; justify-content: center;">
          <button class="bg-destructive" style="border: none; padding: 0.5rem 1rem; border-radius: 0.5rem; color: white; cursor: pointer;" onclick="window.app.deleteSession('${id}')">Sim, cancelar</button>
          <button style="border: 1px solid hsl(var(--border)); background: white; padding: 0.5rem 1rem; border-radius: 0.5rem; cursor: pointer;" onclick="window.app.render()">Voltar</button>
        </div>
      </div>
    `;
  }

  async function deleteSession(id) {
    appointments = appointments.filter(a => a.id !== id);
    await saveData();
    render();
  }

  // --- Renderização ---

  function switchView(view) {
    currentView = view;
    elements.viewListBtn.classList.toggle('active', view === 'list');
    elements.viewCalendarBtn.classList.toggle('active', view === 'calendar');
    elements.viewList.classList.toggle('hidden', view !== 'list');
    elements.viewCalendar.classList.toggle('hidden', view !== 'calendar');
    render();
  }

  function render() {
    if (currentView === 'list') {
      renderList();
    } else {
      if (calendarMode === 'month') renderCalendar();
      else if (calendarMode === 'week') renderWeek();
      else renderDay();
    }
  }

  function renderList() {
    const when = elements.filterWhen.value;
    const payment = elements.filterPayment.value;
    const professional = elements.filterProfessional.value;
    const now = new Date().getTime();
    const apps = appointments || [];

    // Atualizar labels na lista se necessário
    const therapistIcon = config.businessType.toLowerCase().includes('psic') ? '👤' : '🤲';

    let filtered = apps.filter(app => {
      let matchWhen = true;
      if (when === 'upcoming') matchWhen = new Date(app.dataHora).getTime() >= now;
      
      let matchPayment = true;
      if (payment === 'paid') matchPayment = app.pago;
      if (payment === 'unpaid') matchPayment = !app.pago;
      
      let matchProfessional = true;
      if (professional !== 'all') matchProfessional = app.massoterapeuta === professional;
      
      return matchWhen && matchPayment && matchProfessional;
    });

    filtered.sort((a, b) => new Date(a.dataHora) - new Date(b.dataHora));

    elements.filterCount.textContent = `${filtered.length} sessões encontradas`;
    elements.btnClearFilters.classList.toggle('hidden', when === 'all' && payment === 'all' && professional === 'all');

    if (apps.length === 0) {
      elements.listContainer.innerHTML = `
        <div style="text-align: center; padding: 3rem 1rem; color: hsl(var(--muted-foreground));">
          <div style="font-size: 3rem; margin-bottom: 1rem;">🌿</div>
          <p style="font-weight: 600; color: hsl(var(--foreground));">Nenhum agendamento ainda</p>
          <p>Toque em "Novo Agendamento" para começar.</p>
        </div>
      `;
    } else if (filtered.length === 0) {
      elements.listContainer.innerHTML = `
        <div style="text-align: center; padding: 3rem 1rem; color: hsl(var(--muted-foreground));">
          <p style="font-weight: 600; color: hsl(var(--foreground));">Nenhuma sessão com esses filtros</p>
          <p>Ajuste os filtros acima para ver outras sessões.</p>
        </div>
      `;
    } else {
      elements.listContainer.innerHTML = filtered.map(app => `
        <div class="appointment-card shadow-sm" data-id="${app.id}">
          <div style="display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 0.75rem;">
            <div>
              <h3 style="font-size: 1rem; font-weight: 700; margin-bottom: 0.25rem;">${app.paciente}</h3>
              <span class="chip">${app.tipo}</span>
            </div>
            <div style="display: flex; flex-direction: column; align-items: flex-end; gap: 0.5rem;">
              <span class="price-tag">${formatBRL(app.valor)}</span>
              <button class="pay-status-btn ${app.pago ? 'paid' : 'unpaid'}" onclick="window.app.togglePayment('${app.id}')">
                ${app.pago ? '✓ Pago' : '● Não pago'}
              </button>
            </div>
          </div>
          <div style="font-size: 0.875rem; color: hsl(var(--muted-foreground)); display: flex; flex-direction: column; gap: 0.25rem;">
            <div style="display: flex; align-items: center; gap: 0.4rem;">🗓 ${formatDateShort(app.dataHora)} às ${formatTime(app.dataHora)} (${app.duracao || 60} min)</div>
            <div style="display: flex; align-items: center; gap: 0.4rem;">${therapistIcon} ${app.massoterapeuta}</div>
          </div>
          <button style="border: none; background: none; color: hsl(var(--destructive)); font-size: 0.75rem; margin-top: 0.75rem; cursor: pointer; padding: 0;" onclick="window.app.confirmCancel('${app.id}')">Cancelar sessão</button>
        </div>
      `).join('');
    }
  }

  function renderCalendar() {
    const year = currentMonth.getFullYear();
    const month = currentMonth.getMonth();
    
    elements.calendarTitle.textContent = currentMonth.toLocaleDateString('pt-BR', { month: 'long', year: 'numeric' }).replace(/^\w/, c => c.toUpperCase());
    
    const firstDay = new Date(year, month, 1).getDay();
    const daysInMonth = new Date(year, month + 1, 0).getDate();
    
    elements.calendarViewContainer.innerHTML = `
      <div style="display: grid; grid-template-columns: repeat(7, 1fr); text-align: center; font-size: 0.75rem; font-weight: 600; color: hsl(var(--muted-foreground)); margin-bottom: 0.5rem;">
        <div>Dom</div><div>Seg</div><div>Ter</div><div>Qua</div><div>Qui</div><div>Sex</div><div>Sáb</div>
      </div>
      <div id="calendar-grid" class="calendar-grid"></div>
    `;
    
    const grid = document.getElementById('calendar-grid');
    let html = '';
    
    for (let i = 0; i < firstDay; i++) {
      html += '<div class="cal-day" style="background: hsl(var(--muted)); cursor: default;"></div>';
    }
    
    const today = new Date();
    const todayStr = `${today.getFullYear()}-${today.getMonth()}-${today.getDate()}`;

    for (let day = 1; day <= daysInMonth; day++) {
      const date = new Date(year, month, day);
      const dateKey = `${year}-${month}-${day}`;
      const isToday = dateKey === todayStr;
      const isSelected = selectedDate && selectedDate.getTime() === date.getTime();
      
      const apps = appointments || [];
      const daySessions = apps.filter(a => {
        const d = new Date(a.dataHora);
        return d.getFullYear() === year && d.getMonth() === month && d.getDate() === day;
      });

      const hasSessions = daySessions.length > 0;
      const paidCount = daySessions.filter(a => a.pago).length;
      const unpaidCount = daySessions.length - paidCount;

      let dotsHtml = '<div class="dots">';
      for (let j = 0; j < Math.min(daySessions.length, 3); j++) {
        // Ordenar pagos primeiro para os pontinhos
        const session = daySessions.sort((a,b) => (a.pago === b.pago) ? 0 : a.pago ? -1 : 1)[j];
        dotsHtml += `<span class="dot ${session.pago ? 'paid' : 'unpaid'}"></span>`;
      }
      dotsHtml += '</div>';

      // Adicionar nomes curtos
      let briefHtml = '';
      daySessions.slice(0, 2).forEach(s => {
        briefHtml += `<div class="cal-session-brief ${s.pago ? 'paid' : 'unpaid'}">${s.paciente.split(' ')[0]}</div>`;
      });

      html += `
        <button class="cal-day ${hasSessions ? 'has-sessions' : ''} ${isToday ? 'today' : ''} ${isSelected ? 'selected' : ''}" 
                onclick="window.app.selectDate(${year}, ${month}, ${day})">
          ${day}
          ${dotsHtml}
          ${briefHtml}
        </button>
      `;
    }
    
    grid.innerHTML = html;
    updateDayDetail();
  }

  function renderWeek() {
    const startOfWeek = new Date(currentMonth);
    startOfWeek.setDate(currentMonth.getDate() - currentMonth.getDay());
    
    const endOfWeek = new Date(startOfWeek);
    endOfWeek.setDate(startOfWeek.getDate() + 6);

    const title = `${startOfWeek.toLocaleDateString('pt-BR', {day:'2-digit', month:'short'})} - ${endOfWeek.toLocaleDateString('pt-BR', {day:'2-digit', month:'short'})} (${currentMonth.getFullYear()})`.replace(/\./g, '');
    elements.calendarTitle.textContent = title;

    elements.calendarViewContainer.innerHTML = `<div class="week-grid"></div>`;
    const grid = elements.calendarViewContainer.querySelector('.week-grid');

    const today = new Date();
    const todayStr = `${today.getFullYear()}-${today.getMonth()}-${today.getDate()}`;

    for (let i = 0; i < 7; i++) {
      const date = new Date(startOfWeek);
      date.setDate(startOfWeek.getDate() + i);
      const dateKey = `${date.getFullYear()}-${date.getMonth()}-${date.getDate()}`;
      
      const apps = appointments || [];
      const sessions = apps.filter(a => {
        const d = new Date(a.dataHora);
        return d.getFullYear() === date.getFullYear() && d.getMonth() === date.getMonth() && d.getDate() === date.getDate();
      }).sort((a, b) => new Date(a.dataHora) - new Date(b.dataHora));
      
      const col = document.createElement('div');
      col.className = 'week-day-col';
      
      const isToday = dateKey === todayStr;
      const dayName = date.toLocaleDateString('pt-BR', { weekday: 'short' }).replace('.', '');
      const dayNum = date.getDate();

      col.innerHTML = `
        <div class="week-day-header ${isToday ? 'today' : ''}">
          <div style="text-transform: uppercase;">${dayName}</div>
          <div style="font-size: 1.1rem;">${dayNum}</div>
        </div>
        <div class="week-sessions-list">
          ${sessions.map(s => `
            <div class="appointment-card" style="padding: 0.5rem; margin-bottom: 0.5rem; font-size: 0.75rem; border-left: 4px solid ${s.pago ? '#48bb78' : '#f56565'}; cursor: pointer;" onclick="window.app.selectDate(${date.getFullYear()}, ${date.getMonth()}, ${date.getDate()}, true)">
              <div style="font-weight: 600;">${formatTime(s.dataHora)}</div>
              <div style="white-space: nowrap; overflow: hidden; text-overflow: ellipsis;">${s.paciente}</div>
            </div>
          `).join('')}
          ${sessions.length === 0 ? '<div style="text-align: center; color: #ccc; font-size: 0.7rem; margin-top: 1rem;">—</div>' : ''}
        </div>
      `;
      grid.appendChild(col);
    }
  }

  function renderDay() {
    elements.calendarTitle.textContent = currentMonth.toLocaleDateString('pt-BR', { day: 'numeric', month: 'long', year: 'numeric' });
    
    const apps = appointments || [];
    const sessions = apps.filter(a => {
        const d = new Date(a.dataHora);
        return d.getFullYear() === currentMonth.getFullYear() && d.getMonth() === currentMonth.getMonth() && d.getDate() === currentMonth.getDate();
    }).sort((a, b) => new Date(a.dataHora) - new Date(b.dataHora));

    elements.calendarViewContainer.innerHTML = `
      <div class="day-view-container">
        <div id="day-view-sessions">
          ${sessions.length > 0 ? '' : '<p style="text-align: center; color: hsl(var(--muted-foreground)); padding: 2rem;">Nenhum agendamento para este dia.</p>'}
        </div>
        <button class="btn-wide bg-secondary" style="margin-top: 1rem; padding: 0.75rem;" onclick="window.app.newAppointmentToday()">+ Novo para este dia</button>
      </div>
    `;

    const list = document.getElementById('day-view-sessions');
    const therapistIcon = config.businessType.toLowerCase().includes('psic') ? '👤' : '🤲';
    
    sessions.forEach(app => {
      const card = document.createElement('div');
      card.className = 'appointment-card';
      card.dataset.id = app.id;
      
      card.innerHTML = `
        <div style="display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 0.75rem;">
          <div>
            <div style="display: flex; align-items: center; gap: 0.5rem; margin-bottom: 0.25rem;">
              <span class="chip">${formatTime(app.dataHora)}</span>
              <span class="chip">${app.duracao || 60} min</span>
            </div>
            <h3 style="font-size: 1.125rem; font-weight: 700;">${app.paciente}</h3>
          </div>
          <div class="price-tag">${formatBRL(app.valor)}</div>
        </div>
        <div style="font-size: 0.875rem; color: hsl(var(--muted-foreground)); margin-bottom: 1rem;">
          <div>${therapistIcon} ${app.massoterapeuta}</div>
          <div>✨ ${app.tipo}</div>
        </div>
        <div style="display: flex; justify-content: space-between; align-items: center; gap: 0.5rem;">
          <button class="pay-status-btn ${app.pago ? 'paid' : 'unpaid'}" onclick="window.app.togglePayment('${app.id}')">
            ${app.pago ? '✓ Pago' : '● Marcar como pago'}
          </button>
          <button style="background: none; border: none; color: hsl(var(--destructive)); cursor: pointer; font-size: 0.875rem;" onclick="window.app.confirmCancel('${app.id}')">Cancelar</button>
        </div>
      `;
      list.appendChild(card);
    });
  }

  function selectDate(y, m, d, goToDay = false) {
    const newDate = new Date(y, m, d);
    if (goToDay) {
      calendarMode = 'day';
      currentMonth = newDate;
      elements.calViewMonth.classList.remove('active');
      elements.calViewWeek.classList.remove('active');
      elements.calViewDay.classList.add('active');
    } else {
      if (selectedDate && selectedDate.getTime() === newDate.getTime()) {
        selectedDate = null;
      } else {
        selectedDate = newDate;
      }
    }
    render();
  }

  function updateDayDetail() {
    if (!selectedDate || calendarMode !== 'month') {
      elements.dayDetail.classList.add('hidden');
      return;
    }

    elements.dayDetail.classList.remove('hidden');
    elements.detailDate.textContent = formatDateFull(selectedDate);
    
    const apps = appointments || [];
    const daySessions = apps.filter(a => {
      const d = new Date(a.dataHora);
      return d.getFullYear() === selectedDate.getFullYear() && 
             d.getMonth() === selectedDate.getMonth() && 
             d.getDate() === selectedDate.getDate();
    });

    daySessions.sort((a, b) => new Date(a.dataHora) - new Date(b.dataHora));

    if (daySessions.length === 0) {
      elements.detailSessions.innerHTML = '<p style="color: hsl(var(--muted-foreground)); text-align: center; padding: 1rem;">Nenhuma sessão neste dia. 🌿</p>';
    } else {
      elements.detailSessions.innerHTML = daySessions.map(app => `
        <div style="display: flex; justify-content: space-between; align-items: center; padding: 0.75rem; border: 1px solid hsl(var(--border)); border-radius: 0.75rem; margin-bottom: 0.5rem;">
          <div>
            <div style="font-weight: 700;">${app.paciente}</div>
            <div style="font-size: 0.75rem; color: hsl(var(--muted-foreground));">${formatTime(app.dataHora)} · ${app.tipo} (${app.duracao || 60} min)</div>
          </div>
          <div style="display: flex; align-items: center; gap: 0.75rem;">
            <div style="font-weight: 600; font-size: 0.875rem;">${formatBRL(app.valor)}</div>
            <button class="pay-status-btn ${app.pago ? 'paid' : 'unpaid'}" style="padding: 0.25rem 0.5rem;" onclick="window.app.togglePayment('${app.id}')">
              ${app.pago ? '✓' : '●'}
            </button>
          </div>
        </div>
      `).join('');
    }
  }

  function newAppointmentToday() {
    openModal();
    const d = new Date(currentMonth);
    const now = new Date();
    d.setHours(now.getHours(), now.getMinutes());
    
    const year = d.getFullYear();
    const month = String(d.getMonth() + 1).padStart(2, '0');
    const day = String(d.getDate()).padStart(2, '0');
    const hours = String(d.getHours()).padStart(2, '0');
    const minutes = String(d.getMinutes()).padStart(2, '0');
    
    elements.inputDateTime.value = `${year}-${month}-${day}T${hours}:${minutes}`;
    validateForm();
  }

  // Exportar para uso global
  window.app = {
    init,
    togglePayment,
    confirmCancel,
    deleteSession,
    selectDate,
    render,
    newAppointmentToday,
    updateUserPermission
  };

  init();
})();
