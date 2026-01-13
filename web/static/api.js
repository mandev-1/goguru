// Centralized API error handling
function handleApiResponse(res) {
  if (res.status === 401 || res.status === 403) {
    // Authentication required - redirect to login
    window.location.href = '/login';
    return Promise.reject(new Error('Authentication required'));
  }
  return res;
}

// Wrapper for fetch that handles auth errors
function apiFetch(url, options = {}) {
  return fetch(url, options)
    .then(handleApiResponse)
    .catch(err => {
      if (err.message === 'Authentication required') {
        return Promise.reject(err);
      }
      throw err;
    });
}

