document.addEventListener('DOMContentLoaded', () => {
  const form = document.getElementById('resetForm');
  const formContent = document.getElementById('resetFormContent');
  const resetTitle = document.getElementById('resetTitle');

  // Check if token is in URL (password reset with token)
  const urlParams = new URLSearchParams(window.location.search);
  const token = urlParams.get('token');

  if (token) {
    // Show password reset form
    resetTitle.textContent = 'Enter your new password';
    formContent.innerHTML = `
      <div class="form-group">
        <label for="password">New Password</label>
        <input type="password" id="password" name="password" placeholder="Enter new password" minlength="8" required />
        <div class="error-message" id="passwordError"></div>
      </div>
      <div class="form-group">
        <label for="confirmPassword">Confirm Password</label>
        <input type="password" id="confirmPassword" name="confirmPassword" placeholder="Confirm new password" minlength="8" required />
        <div class="error-message" id="confirmPasswordError"></div>
      </div>
      <input type="hidden" id="token" name="token" value="${token}" />
      <div class="form-login-button"><button type="submit">Reset Password</button></div>
      <div class="form-links">
        <a href="/login">Back to login</a>
      </div>
    `;

    const password = document.getElementById('password');
    const confirmPassword = document.getElementById('confirmPassword');
    const passwordError = document.getElementById('passwordError');
    const confirmPasswordError = document.getElementById('confirmPasswordError');

    function clearErrors() {
      passwordError.textContent = '';
      passwordError.classList.remove('show');
      confirmPasswordError.textContent = '';
      confirmPasswordError.classList.remove('show');
      password.style.borderColor = '';
      confirmPassword.style.borderColor = '';
    }

    // Real-time validation on blur
    password.addEventListener('blur', () => {
      clearErrors();
      validatePassword();
    });

    confirmPassword.addEventListener('blur', () => {
      clearErrors();
      validateConfirmPassword();
    });

    // Clear errors on input
    password.addEventListener('input', () => {
      clearError('password');
      if (confirmPassword.value) {
        validateConfirmPassword();
      }
    });

    confirmPassword.addEventListener('input', () => {
      clearError('confirmPassword');
      if (password.value) {
        validateConfirmPassword();
      }
    });

    form.addEventListener('submit', (e) => {
      e.preventDefault();
      clearErrors();

      let isValid = true;
      isValid = validatePassword() && isValid;
      isValid = validateConfirmPassword() && isValid;

      if (!isValid) {
        return;
      }

      const formData = new URLSearchParams();
      formData.set('token', token);
      formData.set('password', password.value);
      formData.set('confirmPassword', confirmPassword.value);

      fetch('/reset-password', {
        method: 'POST',
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
        body: formData.toString()
      })
      .then(async (res) => {
        const text = await res.text();
        let json = { success: false, message: 'Invalid response' };
        try { json = JSON.parse(text); } catch {}
        
        if (res.ok && json.success) {
          passwordError.textContent = 'Password reset successful! Redirecting to login...';
          passwordError.style.color = 'green';
          passwordError.classList.add('show');
          setTimeout(() => {
            window.location.href = '/login';
          }, 1500);
        } else {
          passwordError.textContent = json.message || 'Password reset failed';
          passwordError.style.color = 'var(--danger)';
          passwordError.classList.add('show');
          password.style.borderColor = 'var(--danger)';
        }
      })
      .catch(() => {
        passwordError.textContent = 'Network error. Please try again.';
        passwordError.style.color = 'var(--danger)';
        passwordError.classList.add('show');
        password.style.borderColor = 'var(--danger)';
      });
    });

    function validatePassword() {
      const value = password.value;
      clearError('password');

      if (!value) {
        showError('password', 'passwordError', 'Password is required');
        return false;
      }

      if (value.length < 8) {
        showError('password', 'passwordError', 'Password must be at least 8 characters');
        return false;
      }

      return true;
    }

    function validateConfirmPassword() {
      const value = confirmPassword.value;
      clearError('confirmPassword');

      if (!value) {
        showError('confirmPassword', 'confirmPasswordError', 'Please confirm your password');
        return false;
      }

      if (value !== password.value) {
        showError('confirmPassword', 'confirmPasswordError', 'Passwords do not match');
        return false;
      }

      return true;
    }

    function showError(fieldId, errorId, message) {
      const field = document.getElementById(fieldId);
      const errorDiv = document.getElementById(errorId);
      
      if (errorDiv) {
        errorDiv.textContent = message;
        errorDiv.classList.add('show');
      }
      
      field.style.borderColor = 'var(--danger, #dc3545)';
      field.classList.add('error');
    }

    function clearError(fieldId) {
      const field = document.getElementById(fieldId);
      const errorId = fieldId + 'Error';
      const errorDiv = document.getElementById(errorId);
      
      if (errorDiv) {
        errorDiv.textContent = '';
        errorDiv.classList.remove('show');
      }
      
      field.style.borderColor = '';
      field.classList.remove('error');
    }
  } else {
    // No token - redirect to forgot-password page
    window.location.href = '/forgot-password';
  }
});

