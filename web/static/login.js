document.addEventListener('DOMContentLoaded', () => {
  const form = document.getElementById('loginForm');
  const username = document.getElementById('username');
  const password = document.getElementById('password');
  const usernameError = document.getElementById('usernameError');
  const passwordError = document.getElementById('passwordError');

  function clearErrors() {
    usernameError.textContent = '';
    usernameError.classList.remove('show');
    passwordError.textContent = '';
    passwordError.classList.remove('show');
    username.style.borderColor = '';
    password.style.borderColor = '';
  }

  form.addEventListener('submit', (e) => {
    e.preventDefault();
    clearErrors();

    let ok = true;
    if (!username.value.trim()) {
      usernameError.textContent = 'Username is required';
      usernameError.classList.add('show');
      username.style.borderColor = 'var(--danger)';
      ok = false;
    }
    if (!password.value) {
      passwordError.textContent = 'Password is required';
      passwordError.classList.add('show');
      password.style.borderColor = 'var(--danger)';
      ok = false;
    }
    if (!ok) return;

    const data = new URLSearchParams();
    data.set('username', username.value.trim());
    data.set('password', password.value);

    fetch('/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: data.toString()
    })
    .then(async (res) => {
      const text = await res.text();
      let json = { success: false, message: 'Invalid response' };
      try { json = JSON.parse(text); } catch {}
      if (res.ok && json.success) {
        window.location.href = '/';
        return;
      }
      // Handle 403 (not verified): offer resend
      if (res.status === 403) {
        const msg = 'Not verified yet. Want me to resend verification?';
        passwordError.textContent = msg;
        passwordError.classList.add('show');
        password.style.borderColor = 'var(--warning, #ff9800)';
        const btn = document.createElement('button');
        btn.type = 'button';
        btn.className = 'resend-verification-btn';
        btn.textContent = 'Resend verification email';
        btn.addEventListener('click', () => {
          const d = new URLSearchParams();
          d.set('username', username.value.trim());
          fetch('/resend-verification', {
            method: 'POST',
            headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
            body: d.toString()
          }).then(r => r.text()).then(t => {
            let j = { success: false, message: 'Failed to resend' };
            try { j = JSON.parse(t); } catch {}
            passwordError.textContent = j.message || 'Sent. Please check your email.';
            passwordError.classList.add('show');
          }).catch(() => {
            passwordError.textContent = 'Network error. Please try again.';
            passwordError.classList.add('show');
          });
        });
        // Append button once
        if (!passwordError.querySelector('button')) passwordError.appendChild(btn);
        return;
      }

      // Handle 401 (invalid credentials)
      if (res.status === 401) {
        const msg = 'Invalid username/password combination, sorry! (🇨🇦)';
        passwordError.textContent = msg;
        passwordError.classList.add('show');
        password.style.borderColor = 'var(--danger)';
        return;
      }

      // Generic error fallback
      const msg = (json && json.message) ? json.message : 'Login failed';
      passwordError.textContent = msg;
      passwordError.classList.add('show');
      password.style.borderColor = 'var(--danger)';
    })
    .catch(() => {
      passwordError.textContent = 'Network error. Please try again.';
      password.style.borderColor = 'var(--danger)';
    });
  });
});
