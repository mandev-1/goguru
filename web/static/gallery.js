document.addEventListener('DOMContentLoaded', () => {
  const grid = document.getElementById('gallery-grid');
  const template = document.getElementById('gallery-card-template');
  const emptyState = document.getElementById('gallery-empty');
  const statusBar = document.getElementById('gallery-status');
  const loadMoreBtn = document.getElementById('gallery-load-more');

  let page = 1;
  let loading = false;
  let hasMore = true;

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
      const json = await res.json();
      if (!res.ok || !json.success) {
        throw new Error(json.message || 'Unable to like image');
      }
      const current = parseInt(countEl.textContent || '0', 10);
      countEl.textContent = current + 1;
      btn.classList.add('liked');
    } catch (err) {
      alert(err.message);
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
      const json = await res.json();
      if (!res.ok || !json.success) {
        throw new Error(json.message || 'Unable to add comment');
      }
      const li = document.createElement('li');
      li.innerHTML = `<strong>You</strong>: ${escapeHtml(body)}`;
      list.prepend(li);
      input.value = '';
    } catch (err) {
      alert(err.message);
    } finally {
      input.disabled = false;
    }
  };

  const loadPage = async (targetPage = 1, append = false) => {
    if (loading || (!hasMore && append)) return;
    loading = true;
    setStatus('Loading...');
    let succeeded = false;
    try {
      const res = await fetch(`/api/gallery?page=${targetPage}`);
      const json = await res.json();
      if (!res.ok || !json.success) throw new Error(json.message || 'Failed to fetch gallery');
      const items = json.items || [];
      renderItems(items, append);
      emptyState.classList.toggle('hidden', (items.length > 0) || append);
      hasMore = !!json.hasMore;
      page = targetPage;
      succeeded = true;
      if (!hasMore) loadMoreBtn?.setAttribute('disabled', 'true');
    } catch (err) {
      console.error(err);
      setStatus('Unable to load gallery right now.');
    } finally {
      loading = false;
      if (succeeded) setStatus('');
    }
  };

  loadPage(1, false);

  if (loadMoreBtn) {
    loadMoreBtn.addEventListener('click', () => {
      if (hasMore) loadPage(page + 1, true);
    });
  }

  const mockUploadBtn = document.getElementById('mock-upload-btn');
  if (mockUploadBtn) {
    mockUploadBtn.addEventListener('click', async () => {
      mockUploadBtn.disabled = true;
      mockUploadBtn.textContent = 'Uploading...';
      try {
        const res = await fetch('/api/gallery/mock-upload', { method: 'POST' });
        const json = await res.json();
        if (!res.ok || !json.success) throw new Error(json.message || 'Upload failed');
        await loadPage(1, false);
        setStatus('Uploaded mock image successfully.');
      } catch (err) {
        setStatus(err.message);
      } finally {
        mockUploadBtn.disabled = false;
        mockUploadBtn.textContent = 'Upload / Create';
      }
    });
  }
});
