document.addEventListener('DOMContentLoaded', () => {
    const userDisplay = document.getElementById('user-display');

    fetch('/api/current-user')
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                userDisplay.textContent = `Hi, ${data.username}!`;
            }
        })
        .catch(error => console.error('Error fetching user:', error));
});
