document.addEventListener('DOMContentLoaded', () => {
  const grid = document.getElementById('gallery-grid');
  const template = document.getElementById('gallery-card-template');
  const emptyState = document.getElementById('gallery-empty');
  const statusBar = document.getElementById('gallery-status');
  const paginationContainer = document.getElementById('pagination');

  let page = 1;
  let loading = false;
  let totalPages = 1;
  let total = 0;

  const setStatus = (msg) => {
    statusBar.textContent = msg;
    statusBar.classList.toggle('hidden', !msg);
  };

  const escapeHtml = (value) => {
    return (value || '').replace(/[&<>"']/g, (ch) => ({
      '&': '&amp;',
      '<': '&lt;',
      '>': '&gt;',
      '"': '&quot;',
      "'": '&#39;',
    })[ch]);
  };

  const renderItems = (items, append = false) => {
    if (!append) grid.innerHTML = '';
    items.forEach((item) => {
      const node = template.content.firstElementChild.cloneNode(true);
      node.dataset.id = item.id;

      const img = node.querySelector('.gallery-card-img');
      img.src = item.path;
      img.alt = `Image ${item.id}`;

      node.querySelector('.gallery-card-author').textContent = `By ${item.author}`;
      node.querySelector('.gallery-card-date').textContent = new Date(item.createdAt).toLocaleString();

      const likeBtn = node.querySelector('.like-btn');
      const likeCount = node.querySelector('.like-count');
      likeCount.textContent = item.likes || 0;
      if (item.liked) {
        likeBtn.classList.add('liked');
        likeBtn.disabled = true;
      }
      likeBtn.addEventListener('click', () => handleLike(item.id, likeBtn, likeCount));

      const commentsList = node.querySelector('.comment-list');
      if (Array.isArray(item.comments)) {
        item.comments.forEach((c) => {
          const li = document.createElement('li');
          li.innerHTML = `<strong>${escapeHtml(c.author)}</strong>: ${escapeHtml(c.body)}`;
          commentsList.appendChild(li);
        });
      }

      const commentForm = node.querySelector('.comment-form');
      const commentInput = node.querySelector('.comment-input');
      commentForm.addEventListener('submit', (e) => {
        e.preventDefault();
        handleComment(item.id, commentInput, commentsList);
      });

      grid.appendChild(node);
    });
  };

  const handleLike = async (imageId, btn, countEl) => {
    if (btn.disabled) return;
    btn.disabled = true;
    try {
      const data = new URLSearchParams();
      data.set('image_id', imageId);
      const res = await fetch('/api/gallery/like', {
        method: 'POST',
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
        body: data.toString(),
      });
      
      // Check for redirect before parsing JSON
      if (res.status === 401 || res.status === 403) {
        window.location.href = '/register';
        return;
      }
      
      const json = await res.json();
      if (!res.ok || !json.success) {
        // Check if we need to redirect to register
        if (json.message && json.message.startsWith('redirect:')) {
          window.location.href = json.message.replace('redirect:', '');
          return;
        }
        throw new Error(json.message || 'Unable to like image');
      }
      const current = parseInt(countEl.textContent || '0', 10);
      countEl.textContent = current + 1;
      btn.classList.add('liked');
    } catch (err) {
      // Only show alert if it's not a redirect
      if (!err.message || !err.message.includes('redirect')) {
        alert(err.message || 'Unable to like image');
      }
      btn.disabled = false;
    }
  };

  const handleComment = async (imageId, input, list) => {
    const body = input.value.trim();
    if (!body) return;
    input.disabled = true;
    try {
      const data = new URLSearchParams();
      data.set('image_id', imageId);
      data.set('body', body);
      const res = await fetch('/api/gallery/comment', {
        method: 'POST',
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
        body: data.toString(),
      });
      
      // Check for redirect before parsing JSON
      if (res.status === 401 || res.status === 403) {
        window.location.href = '/register';
        return;
      }
      
      const json = await res.json();
      if (!res.ok || !json.success) {
        // Check if we need to redirect to register
        if (json.message && json.message.startsWith('redirect:')) {
          window.location.href = json.message.replace('redirect:', '');
          return;
        }
        throw new Error(json.message || 'Unable to add comment');
      }
      const li = document.createElement('li');
      li.innerHTML = `<strong>You</strong>: ${escapeHtml(body)}`;
      list.prepend(li);
      input.value = '';
    } catch (err) {
      // Only show alert if it's not a redirect
      if (!err.message || !err.message.includes('redirect')) {
        alert(err.message || 'Unable to add comment');
      }
    } finally {
      input.disabled = false;
    }
  };

  const loadPage = async (targetPage = 1) => {
    if (loading || targetPage < 1) return;
    loading = true;
    setStatus('Loading...');
    let succeeded = false;
    try {
      const res = await fetch(`/api/gallery?page=${targetPage}`);
      const json = await res.json();
      if (!res.ok || !json.success) throw new Error(json.message || 'Failed to fetch gallery');
      
      // Handle response structure: {success: true, data: {items: [...], hasMore: ...}}
      const data = json.data || {};
      const items = data.items || json.items || [];
      totalPages = data.totalPages || 1;
      total = data.total || 0;
      
      renderItems(items, false);
      emptyState.classList.toggle('hidden', items.length > 0);
      page = targetPage;
      succeeded = true;
      updatePagination();
    } catch {
      setStatus('Unable to load gallery right now.');
    } finally {
      loading = false;
      if (succeeded) setStatus('');
    }
  };

  const updatePagination = () => {
    if (!paginationContainer || totalPages <= 1) {
      paginationContainer.innerHTML = '';
      return;
    }

    let html = '';
    
    // Previous button
    if (page > 1) {
      html += `<button class="pagination-btn" data-page="${page - 1}">Previous</button>`;
    } else {
      html += `<button class="pagination-btn" disabled>Previous</button>`;
    }

    // Page numbers
    const maxVisible = 7;
    let startPage = Math.max(1, page - Math.floor(maxVisible / 2));
    let endPage = Math.min(totalPages, startPage + maxVisible - 1);
    
    if (endPage - startPage < maxVisible - 1) {
      startPage = Math.max(1, endPage - maxVisible + 1);
    }

    // First page
    if (startPage > 1) {
      html += `<button class="pagination-btn" data-page="1">1</button>`;
      if (startPage > 2) {
        html += `<span class="pagination-ellipsis">...</span>`;
      }
    }

    // Page range
    for (let i = startPage; i <= endPage; i++) {
      if (i === page) {
        html += `<button class="pagination-btn pagination-btn-active" disabled>${i}</button>`;
      } else {
        html += `<button class="pagination-btn" data-page="${i}">${i}</button>`;
      }
    }

    // Last page
    if (endPage < totalPages) {
      if (endPage < totalPages - 1) {
        html += `<span class="pagination-ellipsis">...</span>`;
      }
      html += `<button class="pagination-btn" data-page="${totalPages}">${totalPages}</button>`;
    }

    // Next button
    if (page < totalPages) {
      html += `<button class="pagination-btn" data-page="${page + 1}">Next</button>`;
    } else {
      html += `<button class="pagination-btn" disabled>Next</button>`;
    }

    paginationContainer.innerHTML = html;

    // Add event listeners
    paginationContainer.querySelectorAll('.pagination-btn[data-page]').forEach(btn => {
      btn.addEventListener('click', () => {
        const targetPage = parseInt(btn.dataset.page, 10);
        if (targetPage !== page) {
          loadPage(targetPage);
          window.scrollTo({ top: 0, behavior: 'smooth' });
        }
      });
    });
  };

  loadPage(1);

  const mockUploadBtn = document.getElementById('mock-upload-btn');
  if (mockUploadBtn) {
    mockUploadBtn.addEventListener('click', () => {
      window.location.href = '/editor';
    });
  }
});
