document.addEventListener('DOMContentLoaded', () => {
  const form = document.getElementById('resetForm');
  const emailInput = document.getElementById('email');

  // Check if token is in URL (password reset with token)
  const urlParams = new URLSearchParams(window.location.search);
  const token = urlParams.get('token');

  if (token) {
    // Show password reset form
    form.innerHTML = `
      <div class="form-group">
        <label for="password">New Password</label>
        <input type="password" id="password" name="password" placeholder="Enter new password" minlength="8" required />
      </div>
      <div class="form-group">
        <label for="confirmPassword">Confirm Password</label>
        <input type="password" id="confirmPassword" name="confirmPassword" placeholder="Confirm new password" minlength="8" required />
      </div>
      <input type="hidden" id="token" name="token" value="${token}" />
      <button type="submit">Reset Password</button>
    `;

    form.addEventListener('submit', (e) => {
      e.preventDefault();
      const password = document.getElementById('password').value;
      const confirmPassword = document.getElementById('confirmPassword').value;

      if (password.length < 8) {
        alert('Password must be at least 8 characters');
        return;
      }

      if (password !== confirmPassword) {
        alert('Passwords do not match');
        return;
      }

      const formData = new URLSearchParams();
      formData.set('token', token);
      formData.set('password', password);
      formData.set('confirmPassword', confirmPassword);

      fetch('/reset-password', {
        method: 'POST',
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
        body: formData.toString()
      })
      .then(res => res.json())
      .then(data => {
        if (data.success) {
          alert('Password reset successful! Redirecting to login...');
          window.location.href = '/login';
        } else {
          alert(data.message || 'Password reset failed');
        }
      })
      .catch(err => {
        alert('Network error. Please try again.');
      });
    });
  } else {
    // Show forgot password form
    form.addEventListener('submit', (e) => {
      e.preventDefault();
      const email = emailInput.value.trim();

      if (!email) {
        alert('Please enter your email address');
        return;
      }

      const formData = new URLSearchParams();
      formData.set('email', email);

      fetch('/forgot-password', {
        method: 'POST',
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
        body: formData.toString()
      })
      .then(res => res.json())
      .then(data => {
        alert(data.message || 'If the email exists, a reset link has been sent.');
      })
      .catch(err => {
        alert('Network error. Please try again.');
      });
    });
  }
});

