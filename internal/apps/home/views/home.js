(function() {
  const SESSION_KEY_AGENDA = 'agenda-session';
  const USER_ID_FINANCEIRO = 'user_id';
  const USERNAME_FINANCEIRO = 'username';

  function checkAuth() {
    const agendaUser = JSON.parse(localStorage.getItem(SESSION_KEY_AGENDA));
    const financeiroUserId = localStorage.getItem(USER_ID_FINANCEIRO);
    const financeiroUsername = localStorage.getItem(USERNAME_FINANCEIRO);

    const userBar = document.getElementById('user-bar');
    const userNameElem = document.getElementById('user-name');
    
    let loggedIn = false;
    let name = '';

    if (agendaUser) {
      loggedIn = true;
      name = agendaUser.name || agendaUser.email;
    } else if (financeiroUserId && financeiroUsername) {
      loggedIn = true;
      name = financeiroUsername;
    }

    const agendaAppCard = document.getElementById('app-agenda');
    const financeiroAppCard = document.getElementById('app-financeiro');

    if (loggedIn) {
      userBar.classList.remove('hidden');
      userNameElem.textContent = `Olá, ${name}`;
      updateAppStatus(true);
      
      // Se estiver logado, garante que os links funcionem
      agendaAppCard.classList.remove('disabled');
      financeiroAppCard.classList.remove('disabled');
    } else {
      userBar.classList.add('hidden');
      updateAppStatus(false);
    }
  }

  function updateAppStatus(isLoggedIn) {
    const agendaStatus = document.querySelector('#app-agenda .access-status');
    const financeiroStatus = document.querySelector('#app-financeiro .access-status');

    if (isLoggedIn) {
      agendaStatus.textContent = '✓ Acesso Liberado';
      agendaStatus.style.color = 'hsl(var(--primary))';
      financeiroStatus.textContent = '✓ Acesso Liberado';
      financeiroStatus.style.color = 'hsl(var(--primary))';
    } else {
      agendaStatus.textContent = '🔒 Requer Login';
      financeiroStatus.textContent = '🔒 Requer Login';
    }
  }

  window.handleHomeLogout = function() {
    localStorage.removeItem(SESSION_KEY_AGENDA);
    localStorage.removeItem(USER_ID_FINANCEIRO);
    localStorage.removeItem(USERNAME_FINANCEIRO);
    localStorage.removeItem('role');
    location.reload();
  };

  checkAuth();
})();
