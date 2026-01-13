document.addEventListener('DOMContentLoaded', () => {
  const assetList = document.getElementById('assetList');
  const assetUploadForm = document.getElementById('assetUploadForm');
  const assetFile = document.getElementById('assetFile');
  const assetName = document.getElementById('assetName');
  const baseImage = document.getElementById('baseImage');
  const canvas = document.getElementById('canvas');
  const ctx = canvas.getContext('2d');
  const btnClear = document.getElementById('btnClear');
  const btnCompose = document.getElementById('btnCompose');
  const composeStatus = document.getElementById('composeStatus');

  // init white canvas
  ctx.fillStyle = '#ffffff';
  ctx.fillRect(0, 0, canvas.width, canvas.height);

  // Load asset list
  function loadAssets() {
    fetch('/api/assets')
      .then(r => r.json())
      .then(list => {
        assetList.innerHTML = '';
        if (!Array.isArray(list) || list.length === 0) {
          assetList.innerHTML = '<p class="muted">No assets yet. Upload to begin.</p>';
          return;
        }
        list.forEach(a => {
          const item = document.createElement('button');
          item.type = 'button';
          item.className = 'asset-item';
          item.style.display = 'block';
          item.style.marginBottom = '8px';
          item.innerHTML = `<img src="${a.path}" alt="${a.name}" style="max-width:100%;">`;
          item.addEventListener('click', () => {
            assetList.dataset.selected = String(a.id);
            [...assetList.querySelectorAll('.asset-item')].forEach(el => el.classList.remove('selected'));
            item.classList.add('selected');
            updateComposeButtonState();
          });
          assetList.appendChild(item);
        });
      }).catch(() => {
        assetList.innerHTML = '<p class="muted">Failed to load assets.</p>';
      });
  }
  loadAssets();

  // Upload asset
  assetUploadForm.addEventListener('change', (e) => {
    if (e.target === assetFile) {
      const fd = new FormData();
      if (assetName.value.trim()) fd.set('name', assetName.value.trim());
      if (assetFile.files[0]) fd.set('file', assetFile.files[0]);
      fetch('/api/assets/upload', { method: 'POST', body: fd })
        .then(r => r.json())
        .then(j => {
          composeStatus.textContent = j.message || 'Uploaded';
          loadAssets();
        })
        .catch(() => { composeStatus.textContent = 'Upload failed'; });
    }
  });

  // Base image preview on canvas
  baseImage.addEventListener('change', () => {
    const file = baseImage.files[0];
    if (!file) return;
    const url = URL.createObjectURL(file);
    const img = new Image();
    img.onload = () => {
      ctx.clearRect(0, 0, canvas.width, canvas.height);
      ctx.fillStyle = '#ffffff';
      ctx.fillRect(0, 0, canvas.width, canvas.height);
      const scale = Math.min(canvas.width / img.width, canvas.height / img.height, 1);
      const w = Math.floor(img.width * scale);
      const h = Math.floor(img.height * scale);
      const x = Math.floor((canvas.width - w) / 2);
      const y = Math.floor((canvas.height - h) / 2);
      ctx.drawImage(img, x, y, w, h);
      updateComposeButtonState();
    };
    img.src = url;
  });

  btnClear.addEventListener('click', () => {
    ctx.clearRect(0, 0, canvas.width, canvas.height);
    ctx.fillStyle = '#ffffff';
    ctx.fillRect(0, 0, canvas.width, canvas.height);
    baseImage.value = '';
    composeStatus.textContent = '';
    updateComposeButtonState();
  });

  function hasSelectedAsset() {
    return !!assetList.dataset.selected;
  }
  function hasBaseImage() {
    return baseImage.files && baseImage.files[0];
  }
  function updateComposeButtonState() {
    btnCompose.disabled = !(hasSelectedAsset() && hasBaseImage());
  }

  btnCompose.addEventListener('click', () => {
    if (btnCompose.disabled) return;
    const fd = new FormData();
    fd.set('asset_id', assetList.dataset.selected);
    fd.set('image', baseImage.files[0]);
    fetch('/api/compose', { method: 'POST', body: fd })
      .then(r => r.text())
      .then(t => {
        let j = { success: false, message: 'Failed' };
        try { j = JSON.parse(t); } catch {}
        if (j.success) {
          composeStatus.innerHTML = `Saved: <a href="${j.path}" target="_blank">${j.path}</a>`;
        } else {
          composeStatus.textContent = j.message || 'Compose failed';
        }
      })
      .catch(() => { composeStatus.textContent = 'Network error'; });
  });
});
