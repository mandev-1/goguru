document.addEventListener('DOMContentLoaded', () => {
    const userDisplay = document.getElementById('user-display');
    if (!userDisplay) return;

    fetch('/api/current-user')
        .then(response => response.json())
        .then(data => {
            if (data.success && data.data && data.data.username) {
                userDisplay.textContent = `Hi, ${data.data.username}!`;
            } else {
                // User not logged in - hide or show login link
                userDisplay.textContent = '';
            }
        })
        .catch(() => {
            userDisplay.textContent = '';
        });
});
