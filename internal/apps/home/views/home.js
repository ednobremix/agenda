(function() {
  const SESSION_KEY_AGENDA = 'agenda-session';
  const SESSION_KEY_FINANCEIRO = 'financeiro_user';
  const USER_ID_FINANCEIRO = 'user_id';
  const USERNAME_FINANCEIRO = 'username';
  const CLIENT_ID = "665567298487-am8coo2krrd2msdno64p3ng5om1pfr75.apps.googleusercontent.com";

  function checkAuth() {
    const agendaUser = JSON.parse(localStorage.getItem(SESSION_KEY_AGENDA));
    const financeiroUser = JSON.parse(localStorage.getItem(SESSION_KEY_FINANCEIRO));
    const financeiroUserId = localStorage.getItem(USER_ID_FINANCEIRO);
    const financeiroUsername = localStorage.getItem(USERNAME_FINANCEIRO);

    const userBar = document.getElementById('user-bar');
    const userNameElem = document.getElementById('user-name');
    const authSection = document.getElementById('auth-section');
    const registerText = document.getElementById('register-text');
    
    let loggedIn = false;
    let name = '';
    let isAdmin = false;

    if (agendaUser) {
      loggedIn = true;
      name = agendaUser.name || agendaUser.email;
      isAdmin = agendaUser.role === 'admin';
    } else if (financeiroUser) {
      loggedIn = true;
      name = financeiroUser.name || financeiroUser.email;
      isAdmin = financeiroUser.role === 'admin';
    } else if (financeiroUserId && financeiroUsername) {
      loggedIn = true;
      name = financeiroUsername;
      isAdmin = localStorage.getItem('role') === 'admin';
    }

    const agendaAppCard = document.getElementById('app-agenda');
    const financeiroAppCard = document.getElementById('app-financeiro');

    if (loggedIn) {
      userBar.classList.remove('hidden');
      if (authSection) authSection.classList.add('hidden');
      if (registerText) registerText.classList.add('hidden');
      userNameElem.textContent = isAdmin ? `👑 Administrador: ${name}` : `Olá, ${name}`;
      
      const userObj = agendaUser || financeiroUser || (financeiroUserId ? { role: localStorage.getItem('role'), has_agenda: localStorage.getItem('has_agenda') === 'true', has_financeiro: localStorage.getItem('has_financeiro') === 'true' } : null);

      const hasAgenda = isAdmin || (userObj && userObj.has_agenda);
      const hasFinanceiro = isAdmin || (userObj && userObj.has_financeiro);

      updateAppStatus(true, isAdmin, hasAgenda, hasFinanceiro);
      
      if (hasAgenda) {
        agendaAppCard.classList.remove('disabled');
        agendaAppCard.querySelector('.btn-primary').style.pointerEvents = 'auto';
        agendaAppCard.querySelector('.btn-primary').style.opacity = '1';
      } else {
        agendaAppCard.classList.add('disabled');
        agendaAppCard.querySelector('.btn-primary').style.pointerEvents = 'none';
        agendaAppCard.querySelector('.btn-primary').style.opacity = '0.5';
      }

      if (hasFinanceiro) {
        financeiroAppCard.classList.remove('disabled');
        financeiroAppCard.querySelector('.btn-primary').style.pointerEvents = 'auto';
        financeiroAppCard.querySelector('.btn-primary').style.opacity = '1';
      } else {
        financeiroAppCard.classList.add('disabled');
        financeiroAppCard.querySelector('.btn-primary').style.pointerEvents = 'none';
        financeiroAppCard.querySelector('.btn-primary').style.opacity = '0.5';
      }
    } else {
      userBar.classList.add('hidden');
      if (authSection) authSection.classList.remove('hidden');
      if (registerText) registerText.classList.remove('hidden');
      initGoogleAuth();
      updateAppStatus(false);
    }
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

    google.accounts.id.renderButton(
      document.getElementById("google-login-btn"),
      { theme: "outline", size: "large", width: 280 }
    );
  }

  async function handleCredentialResponse(response) {
    try {
      const resp = await fetch('/api/auth/google', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ credential: response.credential })
      });

      if (resp.ok) {
        const user = await resp.json();
        // Salva em ambas as chaves para compatibilidade
        localStorage.setItem(SESSION_KEY_AGENDA, JSON.stringify(user));
        localStorage.setItem(SESSION_KEY_FINANCEIRO, JSON.stringify(user));
        localStorage.setItem('role', user.role);
        localStorage.setItem('has_agenda', user.has_agenda);
        localStorage.setItem('has_financeiro', user.has_financeiro);
        location.reload();
      } else {
        const errorText = await resp.text();
        console.error('Erro Google Auth:', errorText);
        alert('Erro na autenticação com Google.');
      }
    } catch (err) {
      console.error('Erro na conexão com Google Auth:', err);
    }
  }

  function updateAppStatus(isLoggedIn, isAdmin, hasAgenda, hasFinanceiro) {
    const agendaStatus = document.querySelector('#app-agenda .access-status');
    const financeiroStatus = document.querySelector('#app-financeiro .access-status');

    if (isLoggedIn) {
      if (isAdmin) {
        agendaStatus.textContent = '👑 Admin (Acesso Total)';
        agendaStatus.style.color = 'hsl(var(--primary))';
        financeiroStatus.textContent = '👑 Admin (Acesso Total)';
        financeiroStatus.style.color = 'hsl(var(--primary))';
      } else {
        agendaStatus.textContent = hasAgenda ? '✓ Acesso Liberado' : '🚫 Acesso Restrito';
        agendaStatus.style.color = hasAgenda ? 'hsl(var(--primary))' : '#dc2626';
        financeiroStatus.textContent = hasFinanceiro ? '✓ Acesso Liberado' : '🚫 Acesso Restrito';
        financeiroStatus.style.color = hasFinanceiro ? 'hsl(var(--primary))' : '#dc2626';
      }
    } else {
      agendaStatus.textContent = '🔒 Requer Login';
      financeiroStatus.textContent = '🔒 Requer Login';
    }
  }

  window.handleHomeLogout = function() {
    localStorage.removeItem(SESSION_KEY_AGENDA);
    localStorage.removeItem(SESSION_KEY_FINANCEIRO);
    localStorage.removeItem(USER_ID_FINANCEIRO);
    localStorage.removeItem(USERNAME_FINANCEIRO);
    localStorage.removeItem('role');
    localStorage.removeItem('has_agenda');
    localStorage.removeItem('has_financeiro');
    location.reload();
  };

  checkAuth();
})();
