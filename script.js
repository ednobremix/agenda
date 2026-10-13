(function() {
  // Configurações e Estado
  const STORAGE_KEY = 'agenda-massoterapia';
  const CONFIG_KEY = 'app-config';
  const OLD_STORAGE_KEY = 'massage_appointments';
  let appointments = [];
  let config = {
    businessType: 'Massoterapia',
    services: [
      { name: 'Massagem Relaxante', value: 150, duration: 60 },
      { name: 'Drenagem Linfática', value: 160, duration: 60 },
      { name: 'Massagem Modeladora', value: 160, duration: 60 },
      { name: 'Massagem Terapêutica', value: 170, duration: 60 },
      { name: 'Reflexologia', value: 100, duration: 30 },
      { name: 'Shiatsu', value: 180, duration: 60 }
    ]
  };
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
    configBusinessType: document.getElementById('config-business-type'),
    headerTitle: document.getElementById('header-business-title'),
    headerDesc: document.getElementById('header-business-desc'),
    modalTitle: document.querySelector('#modal-overlay h2'),
    therapistLabel: document.querySelector('label[for="therapist"]'),
    therapistInput: document.getElementById('therapist'),
    typeLabel: document.querySelector('label[for="type"]')
  };

  // --- Inicialização ---

  function init() {
    loadConfig();
    loadData();
    setupEventListeners();
    updateFinanceSummary();
    applyConfig();
    render();
  }

  function loadConfig() {
    const data = localStorage.getItem(CONFIG_KEY);
    if (data) {
      config = JSON.parse(data);
    }
  }

  function saveConfig() {
    localStorage.setItem(CONFIG_KEY, JSON.stringify(config));
    applyConfig();
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
      elements.therapistInput.placeholder = 'Ex.: Dr. Roberto';
      elements.typeLabel.textContent = 'Tipo de Sessão';
    } else if (isBarber) {
      elements.headerDesc.textContent = 'Gestão ágil para seu negócio.';
      elements.therapistLabel.textContent = 'Nome do Profissional';
      elements.therapistInput.placeholder = 'Ex.: Mestre Cuca';
      elements.typeLabel.textContent = 'Serviço';
    } else if (isMedical) {
      elements.headerDesc.textContent = 'Controle seus atendimentos clínicos.';
      elements.therapistLabel.textContent = 'Nome do Médico(a)';
      elements.therapistInput.placeholder = 'Ex.: Dra. Helena';
      elements.typeLabel.textContent = 'Procedimento';
    } else {
      elements.headerDesc.textContent = 'Seus atendimentos organizados, com leveza.';
      elements.therapistLabel.textContent = 'Nome do Profissional';
      elements.therapistInput.placeholder = 'Ex.: Ana Souza';
      elements.typeLabel.textContent = 'Tipo de Serviço';
    }

    // Atualizar select de tipos de serviço no modal
    const currentType = elements.inputType.value;
    elements.inputType.innerHTML = config.services.map(s => 
      `<option value="${s.name}" data-value="${s.value}" data-duration="${s.duration}">${s.name} - ${formatBRL(s.value)} (${s.duration} min)</option>`
    ).join('');
    
    // Tentar manter o valor selecionado ou pegar o primeiro
    if (config.services.some(s => s.name === currentType)) {
      elements.inputType.value = currentType;
    } else if (config.services.length > 0) {
      elements.inputType.value = config.services[0].name;
      elements.inputValue.value = config.services[0].value.toFixed(2).replace('.', ',');
      elements.inputDuration.value = config.services[0].duration;
    }
  }

  function loadData() {
    const data = localStorage.getItem(STORAGE_KEY);
    if (data) {
      appointments = JSON.parse(data).map(app => ({
        ...app,
        pago: app.pago ?? false // Migração para registros antigos
      }));
    } else {
      // Migração da chave antiga se existir
      const oldData = localStorage.getItem(OLD_STORAGE_KEY);
      if (oldData) {
        appointments = JSON.parse(oldData).map(app => ({
          id: app.id.toString(),
          paciente: app.patientName,
          massoterapeuta: app.therapistName,
          tipo: app.massageType,
          valor: app.sessionValue,
          dataHora: app.dateTime,
          pago: false
        }));
        saveData();
      }
    }
  }

  function saveData() {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(appointments));
    updateFinanceSummary();
  }

  // --- Lógica de Negócio ---

  function updateFinanceSummary() {
    const received = appointments.filter(a => a.pago).reduce((sum, a) => sum + a.valor, 0);
    const pending = appointments.filter(a => !a.pago).reduce((sum, a) => sum + a.valor, 0);
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

    newApps.forEach(newApp => {
      const startTime = new Date(newApp.dataHora).getTime();
      const endTime = startTime + (newApp.duracao * 60000);

      const match = appointments.find(oldApp => {
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
    elements.settingsForm.onsubmit = handleSettingsSubmit;

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
    elements.btnClearFilters.onclick = () => {
      elements.filterWhen.value = 'all';
      elements.filterPayment.value = 'all';
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

  function openModal() {
    elements.appointmentForm.reset();
    formState.payment = false;
    formState.repeat = 'none';
    elements.repeatOptions.classList.add('hidden');
    elements.conflictBox.classList.add('hidden');
    elements.formError.classList.add('hidden');
    document.querySelectorAll('.toggle-btn').forEach(btn => btn.classList.remove('active'));
    document.getElementById('pay-false').classList.add('active');
    document.querySelector('[data-repeat="none"]').classList.add('active');
    elements.modalOverlay.classList.remove('hidden');
    validateForm();
  }

  function closeModal() {
    elements.modalOverlay.classList.add('hidden');
  }

  function openSettings() {
    elements.configBusinessType.value = config.businessType;
    elements.servicesList.innerHTML = '';
    config.services.forEach(s => addServiceRow(s.name, s.value, s.duration));
    if (config.services.length === 0) addServiceRow();
    elements.settingsOverlay.classList.remove('hidden');
  }

  function closeSettings() {
    elements.settingsOverlay.classList.add('hidden');
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
    row.querySelector('.btn-remove-service').onclick = () => row.remove();
    elements.servicesList.appendChild(row);
  }

  function handleSettingsSubmit(e) {
    e.preventDefault();
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

  function deleteSession(id) {
    appointments = appointments.filter(a => a.id !== id);
    saveData();
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
    const now = new Date().getTime();

    // Atualizar labels na lista se necessário
    const therapistIcon = config.businessType.toLowerCase().includes('psic') ? '👤' : '🤲';

    let filtered = appointments.filter(app => {
      let matchWhen = true;
      if (when === 'upcoming') matchWhen = new Date(app.dataHora).getTime() >= now;
      
      let matchPayment = true;
      if (payment === 'paid') matchPayment = app.pago;
      if (payment === 'unpaid') matchPayment = !app.pago;
      
      return matchWhen && matchPayment;
    });

    filtered.sort((a, b) => new Date(a.dataHora) - new Date(b.dataHora));

    elements.filterCount.textContent = `${filtered.length} sessões encontradas`;
    elements.btnClearFilters.classList.toggle('hidden', when === 'all' && payment === 'all');

    if (appointments.length === 0) {
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
      
      const daySessions = appointments.filter(a => {
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
      
      const sessions = appointments.filter(a => {
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
    
    const sessions = appointments.filter(a => {
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
    
    const daySessions = appointments.filter(a => {
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
    newAppointmentToday
  };

  init();
})();
