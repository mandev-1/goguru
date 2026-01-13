document.addEventListener('DOMContentLoaded', () => {
  const form = document.getElementById('forgotPasswordForm');
  const email = document.getElementById('email');
  const emailError = document.getElementById('emailError');

  function clearErrors() {
    emailError.textContent = '';
    emailError.classList.remove('show');
    email.style.borderColor = '';
  }

  // Real-time validation on blur
  email.addEventListener('blur', () => {
    clearErrors();
    validateEmail();
  });

  // Clear errors on input
  email.addEventListener('input', () => {
    clearError('email');
  });

  form.addEventListener('submit', (e) => {
    e.preventDefault();
    clearErrors();

    if (!validateEmail()) {
      return;
    }

    const formData = new URLSearchParams();
    formData.set('email', email.value.trim());

    fetch('/forgot-password', {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: formData.toString()
    })
    .then(async (res) => {
      const text = await res.text();
      let json = { success: false, message: 'Invalid response' };
      try { json = JSON.parse(text); } catch {}
      
      if (res.ok && json.success) {
        emailError.textContent = json.message || 'If the email exists, a reset link has been sent.';
        emailError.style.color = 'green';
        emailError.classList.add('show');
        email.value = ''; // Clear the field
      } else {
        emailError.textContent = json.message || 'Failed to send reset link';
        emailError.style.color = 'var(--danger)';
        emailError.classList.add('show');
        email.style.borderColor = 'var(--danger)';
      }
    })
    .catch(() => {
      emailError.textContent = 'Network error. Please try again.';
      emailError.style.color = 'var(--danger)';
      emailError.classList.add('show');
      email.style.borderColor = 'var(--danger)';
    });
  });

  function validateEmail() {
    const value = email.value.trim();
    clearError('email');

    if (!value) {
      showError('email', 'emailError', 'Email is required');
      return false;
    }

    if (!isValidEmail(value)) {
      showError('email', 'emailError', 'Please enter a valid email address');
      return false;
    }

    return true;
  }

  function isValidEmail(email) {
    const re = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return re.test(email);
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
});

