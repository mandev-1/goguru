document.addEventListener('DOMContentLoaded', function() {
    const form = document.getElementById('registerForm');
    const password = document.getElementById('password');
    const confirmPassword = document.getElementById('confirmPassword');

    form.addEventListener('submit', function(e) {
        e.preventDefault();
        
        // Clear previous errors
        clearErrors();

        // Validate
        let isValid = true;

        // Username validation
        const username = document.getElementById('username').value;
        if (username.length < 3) {
            showError('username', 'Username must be at least 3 characters');
            isValid = false;
        }

        // Email validation
        const email = document.getElementById('email').value;
        if (!isValidEmail(email)) {
            showError('email', 'Please enter a valid email address');
            isValid = false;
        }

        // Password validation
        if (password.value.length < 8) {
            showError('password', 'Password must be at least 8 characters');
            isValid = false;
        }

        // Password match validation
        if (password.value !== confirmPassword.value) {
            showError('confirmPassword', 'Passwords do not match');
            isValid = false;
        }

        if (isValid) {
            // Submit form
            submitForm();
        }
    });

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
                alert(data.message || 'Registration failed');
            }
        })
        .catch(error => {
            console.error('Error:', error);
            alert('An error occurred. Please try again.');
        });
    }

    function isValidEmail(email) {
        const re = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        return re.test(email);
    }

    function showError(fieldId, message) {
        const field = document.getElementById(fieldId);
        const formGroup = field.closest('.form-group');
        
        const errorDiv = document.createElement('small');
        errorDiv.className = 'error-message';
        errorDiv.style.color = 'var(--danger)';
        errorDiv.textContent = message;
        
        formGroup.appendChild(errorDiv);
        field.style.borderColor = 'var(--danger)';
    }

    function clearErrors() {
        const errors = document.querySelectorAll('.error-message');
        errors.forEach(error => error.remove());
        
        const inputs = form.querySelectorAll('input');
        inputs.forEach(input => {
            input.style.borderColor = '';
        });
    }
});