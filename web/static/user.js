document.addEventListener('DOMContentLoaded', () => {
  const profileForm = document.getElementById('profileForm');
  const logoutForm = document.getElementById('logoutForm');
  const profileMsg = document.getElementById('profileMsg');
  const logoutMsg = document.getElementById('logoutMsg');
  const saveChangesBtn = document.getElementById('saveChangesBtn');
  
  let originalUsername = '';
  let originalEmail = '';
  let originalNotify = false;

  // Function to check if form has changes
  function checkForChanges() {
    if (!saveChangesBtn) return;
    
    const username = document.getElementById('username')?.value.trim() || '';
    const email = document.getElementById('email')?.value.trim() || '';
    const password = document.getElementById('password')?.value || '';
    const notify = document.getElementById('notify')?.checked || false;
    
    const hasChanges = 
      username !== originalUsername ||
      email !== originalEmail ||
      password !== '' ||
      notify !== originalNotify;
    
    saveChangesBtn.disabled = !hasChanges;
  }

  // Function to setup event listeners
  function setupListeners() {
    const usernameInput = document.getElementById('username');
    const emailInput = document.getElementById('email');
    const passwordInput = document.getElementById('password');
    const notifyCheckbox = document.getElementById('notify');
    
    if (usernameInput) {
      usernameInput.addEventListener('input', checkForChanges);
      usernameInput.addEventListener('change', checkForChanges);
    }
    if (emailInput) {
      emailInput.addEventListener('input', checkForChanges);
      emailInput.addEventListener('change', checkForChanges);
    }
    if (passwordInput) {
      passwordInput.addEventListener('input', checkForChanges);
      passwordInput.addEventListener('change', checkForChanges);
    }
    if (notifyCheckbox) {
      notifyCheckbox.addEventListener('change', checkForChanges);
      notifyCheckbox.addEventListener('click', checkForChanges);
    }
  }

  // Load current user data
  fetch('/api/current-user')
    .then(res => res.json())
    .then(data => {
      if (data.success && data.data) {
        originalUsername = data.data.username || '';
        originalEmail = data.data.email || '';
        const usernameInput = document.getElementById('username');
        const emailInput = document.getElementById('email');
        if (usernameInput) usernameInput.value = originalUsername;
        if (emailInput) emailInput.value = originalEmail;
        checkForChanges();
      }
    })
    .catch(() => {});

  // Load preferences
  fetch('/api/user/preferences')
    .then(res => res.json())
    .then(data => {
      if (data.success && data.data) {
        originalNotify = data.data.comment_notifications || false;
        const notifyCheckbox = document.getElementById('notify');
        if (notifyCheckbox) notifyCheckbox.checked = originalNotify;
        checkForChanges();
      }
      // Setup listeners after data is loaded
      setupListeners();
    })
    .catch(() => {
      // Setup listeners even if fetch fails
      setupListeners();
    });

  if (profileForm) {
    profileForm.addEventListener('submit', (e) => {
      e.preventDefault();
      profileMsg.textContent = '';
      profileMsg.classList.remove('show');

      const username = document.getElementById('username').value.trim();
      const email = document.getElementById('email').value.trim();
      const password = document.getElementById('password')?.value || '';
      const notify = document.getElementById('notify').checked;

      // Check if profile fields changed
      const profileChanged = 
        username !== originalUsername ||
        email !== originalEmail ||
        password !== '';

      // Check if preferences changed
      const preferencesChanged = notify !== originalNotify;

      // If nothing changed, don't submit
      if (!profileChanged && !preferencesChanged) {
        profileMsg.textContent = 'No changes provided';
        profileMsg.style.color = 'var(--danger)';
        profileMsg.classList.add('show');
        return;
      }

      let profilePromise = Promise.resolve({ success: true });
      let preferencesPromise = Promise.resolve({ success: true });

      // Update profile only if it changed
      if (profileChanged) {
        const formData = new URLSearchParams();
        if (username) formData.set('username', username);
        if (email) formData.set('email', email);
        if (password) formData.set('password', password);

        profilePromise = fetch('/api/user/update', {
          method: 'POST',
          headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
          body: formData.toString()
        })
        .then(res => res.json());
      }

      // Update preferences only if it changed
      if (preferencesChanged) {
        const prefData = new URLSearchParams();
        prefData.set('comment_notifications', notify.toString());
        preferencesPromise = fetch('/api/user/preferences', {
          method: 'POST',
          headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
          body: prefData.toString()
        })
        .then(res => res.json());
      }

      // Wait for both requests to complete
      Promise.all([profilePromise, preferencesPromise])
        .then(([profileData, prefData]) => {
          if (profileData.success && prefData.success) {
            profileMsg.textContent = 'Changes saved successfully';
            profileMsg.style.color = 'green';
            profileMsg.classList.add('show');
            
            // Update original values and disable button
            if (profileChanged) {
              originalUsername = username;
              originalEmail = email;
              if (password) {
                document.getElementById('password').value = '';
              }
            }
            if (preferencesChanged) {
              originalNotify = notify;
            }
            checkForChanges();
          } else {
            const errorMsg = profileData.message || prefData.message || 'Update failed';
            profileMsg.textContent = errorMsg;
            profileMsg.style.color = 'var(--danger)';
            profileMsg.classList.add('show');
          }
        })
        .catch(err => {
          profileMsg.textContent = 'Network error';
          profileMsg.style.color = 'var(--danger)';
          profileMsg.classList.add('show');
        });
    });
  }

  if (logoutForm) {
    logoutForm.addEventListener('submit', (e) => {
      e.preventDefault();
      fetch('/logout', { method: 'POST' })
        .then(() => {
          window.location.href = '/';
        })
        .catch(err => {
          logoutMsg.textContent = 'Logout failed';
          logoutMsg.classList.add('show');
        });
    });
  }
});

