// Shared header component - ensures consistent header across all pages
function renderHeader(showAuthLinks = true) {
  const header = document.querySelector('header.navbar');
  if (!header) return;

  const navLinksEl = header.querySelector('.nav-links');
  
  // Only process pages that have nav-links
  if (!navLinksEl) return;

  // Ensure user-display exists and is properly positioned
  let userDisplay = header.querySelector('#user-display');
  if (!userDisplay) {
    userDisplay = document.createElement('li');
    userDisplay.id = 'user-display';
    userDisplay.style.cssText = 'color: white; margin-left: 15px;';
    navLinksEl.appendChild(userDisplay);
  }

  // Update user display and auth links based on login status
  if (showAuthLinks) {
    fetch('/api/current-user')
      .then(response => response.json())
      .then(data => {
        if (data.success && data.data && data.data.username) {
          // User is logged in
          userDisplay.textContent = `Hi, ${data.data.username}!`;
          
          // Update Profile link to show username on mobile
          const profileLink = navLinksEl.querySelector('a[href="/user"]');
          if (profileLink) {
            const isMobile = window.innerWidth <= 768;
            if (isMobile) {
              profileLink.textContent = '👤 ' + data.data.username.toUpperCase();
            } else {
              profileLink.textContent = 'Profile';
            }
          }

          // Update Gallery and Editor links text on mobile
          const galleryLink = navLinksEl.querySelector('a[href="/gallery"]');
          const editorLink = navLinksEl.querySelector('a[href="/editor"]');
          const isMobile = window.innerWidth <= 768;
          if (galleryLink) {
            if (isMobile) {
              galleryLink.textContent = 'Community';
            } else {
              galleryLink.textContent = 'Gallery';
            }
          }
          if (editorLink) {
            if (isMobile) {
              editorLink.textContent = 'Create';
            } else {
              editorLink.textContent = 'Editor';
            }
          }
          
          // Remove login/register links if they exist
          const loginLinks = navLinksEl.querySelectorAll('a[href="/login"], a[href="/register"]');
          loginLinks.forEach(link => {
            const li = link.closest('li');
            if (li) li.remove();
          });

          // Add logout link if it doesn't exist
          if (!navLinksEl.querySelector('a[href="/logout"]')) {
            const logoutLi = document.createElement('li');
            const logoutLink = document.createElement('a');
            logoutLink.href = '/logout';
            logoutLink.className = 'a-logout';
            logoutLink.textContent = 'Logout';
            logoutLi.appendChild(logoutLink);
            navLinksEl.appendChild(logoutLi);
          }
        } else {
          // User is not logged in
          userDisplay.textContent = '';
          
          // Remove logout link if it exists
          const logoutLinks = navLinksEl.querySelectorAll('a[href="/logout"]');
          logoutLinks.forEach(link => {
            const li = link.closest('li');
            if (li) li.remove();
          });

          // Remove authenticated-only links (Home, Editor, Profile)
          const authLinks = navLinksEl.querySelectorAll('a[href="/"], a[href="/editor"], a[href="/user"]');
          authLinks.forEach(link => {
            const li = link.closest('li');
            if (li) li.remove();
          });

          // Ensure Gallery link exists and is visible
          let galleryLink = navLinksEl.querySelector('a[href="/gallery"]');
          if (!galleryLink) {
            const galleryLi = document.createElement('li');
            galleryLink = document.createElement('a');
            galleryLink.href = '/gallery';
            galleryLink.textContent = 'Gallery';
            galleryLi.appendChild(galleryLink);
            navLinksEl.appendChild(galleryLi);
          }

          // Add login link if it doesn't exist
          if (!navLinksEl.querySelector('a[href="/login"]')) {
            const loginLi = document.createElement('li');
            const loginLink = document.createElement('a');
            loginLink.href = '/login';
            loginLink.className = 'nav-btn nav-btn-login';
            loginLink.textContent = 'Login';
            loginLi.appendChild(loginLink);
            navLinksEl.appendChild(loginLi);
          }

          // Add register link if it doesn't exist
          if (!navLinksEl.querySelector('a[href="/register"]')) {
            const registerLi = document.createElement('li');
            const registerLink = document.createElement('a');
            registerLink.href = '/register';
            registerLink.className = 'nav-btn nav-btn-register';
            registerLink.textContent = 'Register';
            registerLi.appendChild(registerLink);
            navLinksEl.appendChild(registerLi);
          }
        }
      })
      .catch(() => {
        userDisplay.textContent = '';
      });
  }
}

// Auto-initialize on DOM load
document.addEventListener('DOMContentLoaded', () => {
  renderHeader(true);
  
  // Update link text on window resize (for mobile/desktop switching)
  window.addEventListener('resize', () => {
    const profileLink = document.querySelector('.nav-links a[href="/user"]');
    const galleryLink = document.querySelector('.nav-links a[href="/gallery"]');
    const editorLink = document.querySelector('.nav-links a[href="/editor"]');
    const isMobile = window.innerWidth <= 768;
    
    if (profileLink) {
      fetch('/api/current-user')
        .then(response => response.json())
        .then(data => {
          if (data.success && data.data && data.data.username) {
            if (isMobile) {
              profileLink.textContent = '👤 ' + data.data.username.toUpperCase();
            } else {
              profileLink.textContent = 'Profile';
            }
          }
        })
        .catch(() => {});
    }
    
    if (galleryLink) {
      galleryLink.textContent = isMobile ? 'Community' : 'Gallery';
    }
    
    if (editorLink) {
      editorLink.textContent = isMobile ? 'Create' : 'Editor';
    }
  });
});

