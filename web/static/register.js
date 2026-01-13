document.addEventListener('DOMContentLoaded', function() {
    const form = document.getElementById('registerForm');
    const username = document.getElementById('username');
    const email = document.getElementById('email');
    const password = document.getElementById('password');
    const confirmPassword = document.getElementById('confirmPassword');

    // Real-time validation on blur
    username.addEventListener('blur', () => validateUsername());
    email.addEventListener('blur', () => validateEmail());
    password.addEventListener('blur', () => validatePassword());
    confirmPassword.addEventListener('blur', () => validateConfirmPassword());

    // Clear errors on input
    [username, email, password, confirmPassword].forEach(input => {
        input.addEventListener('input', function() {
            clearError(this.id);
        });
    });

    form.addEventListener('submit', function(e) {
        e.preventDefault();
        
        // Clear previous errors
        clearAllErrors();

        // Validate all fields
        let isValid = true;
        isValid = validateUsername() && isValid;
        isValid = validateEmail() && isValid;
        isValid = validatePassword() && isValid;
        isValid = validateConfirmPassword() && isValid;

        if (isValid) {
            submitForm();
        }
    });

    function validateUsername() {
        const value = username.value.trim();
        const errorId = 'usernameError';
        clearError('username');

        if (!value) {
            showError('username', errorId, 'Username is required');
            return false;
        }

        if (value.length < 3) {
            showError('username', errorId, 'Username must be at least 3 characters');
            return false;
        }

        if (value.length > 20) {
            showError('username', errorId, 'Username must be at most 20 characters');
            return false;
        }

        if (!/^[a-zA-Z0-9_]+$/.test(value)) {
            showError('username', errorId, 'Username can only contain letters, numbers, and underscores');
            return false;
        }

        return true;
    }

    function validateEmail() {
        const value = email.value.trim();
        const errorId = 'emailError';
        clearError('email');

        if (!value) {
            showError('email', errorId, 'Email is required');
            return false;
        }

        if (!isValidEmail(value)) {
            showError('email', errorId, 'Please enter a valid email address');
            return false;
        }

        return true;
    }

    function validatePassword() {
        const value = password.value;
        const errorId = 'passwordError';
        clearError('password');

        if (!value) {
            showError('password', errorId, 'Password is required');
            return false;
        }

        if (value.length < 8) {
            showError('password', errorId, 'Password must be at least 8 characters');
            return false;
        }

        if (value.length > 128) {
            showError('password', errorId, 'Password must be at most 128 characters');
            return false;
        }

        return true;
    }

    function validateConfirmPassword() {
        const value = confirmPassword.value;
        const errorId = 'confirmPasswordError';
        clearError('confirmPassword');

        if (!value) {
            showError('confirmPassword', errorId, 'Please confirm your password');
            return false;
        }

        if (value !== password.value) {
            showError('confirmPassword', errorId, 'Passwords do not match');
            return false;
        }

        return true;
    }

    function submitForm() {
        const formData = new FormData(form);
        
        fetch('/register', {
            method: 'POST',
            body: formData
        })
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                alert('Registration successful! Please check your email to verify your account.');
                window.location.href = '/login';
            } else {
                // Display server-side validation errors
                if (data.message) {
                    // Try to match error to specific field
                    if (data.message.toLowerCase().includes('username')) {
                        showError('username', 'usernameError', data.message);
                    } else if (data.message.toLowerCase().includes('email')) {
                        showError('email', 'emailError', data.message);
                    } else if (data.message.toLowerCase().includes('password')) {
                        if (data.message.toLowerCase().includes('match')) {
                            showError('confirmPassword', 'confirmPasswordError', data.message);
                        } else {
                            showError('password', 'passwordError', data.message);
                        }
                    } else {
                        // Generic error at the top
                        showGenericError(data.message);
                    }
                } else {
                    showGenericError('Registration failed. Please try again.');
                }
            }
        })
        .catch(() => {
            showGenericError('An error occurred. Please try again.');
        });
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
            errorDiv.style.display = 'block';
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
            errorDiv.style.display = 'none';
        }
        
        field.style.borderColor = '';
        field.classList.remove('error');
    }

    function clearAllErrors() {
        ['username', 'email', 'password', 'confirmPassword'].forEach(id => {
            clearError(id);
        });
        
        // Clear generic error if exists
        const genericError = document.getElementById('genericError');
        if (genericError) {
            genericError.remove();
        }
    }

    function showGenericError(message) {
        // Remove existing generic error
        const existing = document.getElementById('genericError');
        if (existing) {
            existing.remove();
        }

        // Create generic error element
        const errorDiv = document.createElement('div');
        errorDiv.id = 'genericError';
        errorDiv.className = 'error-message';
        errorDiv.style.color = 'var(--danger, #dc3545)';
        errorDiv.style.marginBottom = '1rem';
        errorDiv.style.padding = '0.5rem';
        errorDiv.style.backgroundColor = 'rgba(220, 53, 69, 0.1)';
        errorDiv.style.borderRadius = '4px';
        errorDiv.textContent = message;

        // Insert before form
        form.parentNode.insertBefore(errorDiv, form);
    }
});