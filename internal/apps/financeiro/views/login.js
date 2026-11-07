document.addEventListener('DOMContentLoaded', () => {
    const loginForm = document.getElementById('login-form');
    const loginError = document.getElementById('login-error');
    const CLIENT_ID = "665567298487-am8coo2krrd2msdno64p3ng5om1pfr75.apps.googleusercontent.com";

    // Submissão de Login
    if (loginForm) {
        loginForm.addEventListener('submit', async (e) => {
            e.preventDefault();
            const email = document.getElementById('login-email').value;
            const password = document.getElementById('login-password').value;

            try {
                const res = await fetch('/api/login', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ email, password })
                });

                if (res.ok) {
                    const user = await res.json();
                    localStorage.setItem('financeiro_user', JSON.stringify(user));
                    localStorage.setItem('agenda-session', JSON.stringify(user)); // Sincroniza com Agenda
                    localStorage.setItem('role', user.role); // Salva a role para a Home
                    localStorage.setItem('has_agenda', user.has_agenda);
                    localStorage.setItem('has_financeiro', user.has_financeiro);
                    window.location.href = 'index.html';
                } else {
                    loginError.textContent = 'Credenciais inválidas ou erro no servidor.';
                    loginError.classList.remove('hidden');
                }
            } catch (err) {
                loginError.textContent = 'Erro ao conectar ao servidor.';
                loginError.classList.remove('hidden');
            }
        });
    }

    function initGoogleAuth() {
        const btn = document.getElementById("google-login-btn");
        if (!btn) return;
        
        if (typeof google === 'undefined') {
            setTimeout(initGoogleAuth, 100);
            return;
        }

        google.accounts.id.initialize({
            client_id: CLIENT_ID,
            callback: handleCredentialResponse
        });
        google.accounts.id.renderButton(
            btn,
            { theme: "outline", size: "large", width: 280 }
        );
    }

    initGoogleAuth();
});

async function handleCredentialResponse(response) {
    try {
        const res = await fetch('/api/auth/google', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ credential: response.credential })
        });
        if (res.ok) {
            const user = await res.json();
            localStorage.setItem('financeiro_user', JSON.stringify(user));
            localStorage.setItem('agenda-session', JSON.stringify(user)); // Sincroniza com Agenda
            localStorage.setItem('role', user.role); // Salva a role para a Home
            localStorage.setItem('has_agenda', user.has_agenda);
            localStorage.setItem('has_financeiro', user.has_financeiro);
            window.location.href = 'index.html';
        } else {
            console.error('Erro no Google Auth:', await res.text());
        }
    } catch (err) {
        console.error('Google Auth Error:', err);
    }
}
