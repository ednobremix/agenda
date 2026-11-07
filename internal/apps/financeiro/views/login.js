document.addEventListener('DOMContentLoaded', () => {
    const CLIENT_ID = "665567298487-am8coo2krrd2msdno64p3ng5om1pfr75.apps.googleusercontent.com";

    function checkExistingAuth() {
        const userData = localStorage.getItem('financeiro_user') || 
                         localStorage.getItem('portalnobre-session') ||
                         localStorage.getItem('estoque_user');
        
        if (userData) {
            try {
                const user = JSON.parse(userData);
                const role = user.role || localStorage.getItem('role');
                const hasFinanceiro = user.has_financeiro || localStorage.getItem('has_financeiro') === 'true';

                if (role === 'admin' || hasFinanceiro) {
                    window.location.href = 'index.html';
                }
            } catch (e) {
                console.error("Erro ao processar sessão existente", e);
            }
        }
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

    checkExistingAuth();
    initGoogleAuth();
});

window.clearCache = function() {
    if (confirm('Deseja limpar todo o cache e dados locais? Isso fará você sair de todas as sessões.')) {
        localStorage.clear();
        if ('caches' in window) {
            caches.keys().then(names => {
                for (let name of names) caches.delete(name);
            });
        }
        if ('serviceWorker' in navigator) {
            navigator.serviceWorker.getRegistrations().then(registrations => {
                for (let registration of registrations) {
                    registration.unregister();
                }
            });
        }
        setTimeout(() => {
            window.location.reload(true);
        }, 500);
    }
};

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
            localStorage.setItem('portalnobre-session', JSON.stringify(user));
            localStorage.setItem('estoque_user', JSON.stringify(user));
            localStorage.setItem('role', user.role);
            localStorage.setItem('has_agenda', user.has_agenda);
            localStorage.setItem('has_financeiro', user.has_financeiro);
            localStorage.setItem('has_estoque', user.has_estoque);
            window.location.href = 'index.html';
        } else {
            console.error('Erro no Google Auth:', await res.text());
        }
    } catch (err) {
        console.error('Google Auth Error:', err);
    }
}
