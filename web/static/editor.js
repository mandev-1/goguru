// Camagru Editor JavaScript
document.addEventListener('DOMContentLoaded', () => {
  const webcam = document.getElementById('webcam');
  const captureBtn = document.getElementById('capture-btn');
  const snapBtn = document.getElementById('snap-btn');
  const thumbnailList = document.querySelector('.thumbnail-list');
  const uploadInput = document.getElementById('upload-btn');
  const placeImageBtn = document.getElementById('place-image-btn');
  const canvasContent = document.querySelector('.canvas-content');
  const zoomInBtn = document.getElementById('zoom-in-btn');
  const zoomOutBtn = document.getElementById('zoom-out-btn');
  const zoomLevelDisplay = document.getElementById('zoom-level');
  const filterGrid = document.querySelector('.filter-grid');

  let selectedFilter = null;
  let selectedAssetId = null;
  let uploadedImage = null;
  let overlayImage = null;
  let zoomLevel = 1.0;
  const ZOOM_STEP = 0.1;
  const MIN_ZOOM = 0.25;
  let stageOffsetX = 0;
  let stageOffsetY = 0;
  let overlayState = { x: 0, y: 0, w: 200, h: 200 };
  let isDragging = false;
  let isDraggingOverlay = false;
  let initialX = 0, initialY = 0;
  let isVideoFrozen = false;
  let videoStream = null;

  // Load assets from API
  function loadAssets() {
    if (!filterGrid) {
      return;
    }
    
    fetch('/api/assets')
      .then(res => {
        if (!res.ok) {
          throw new Error(`HTTP error! status: ${res.status}`);
        }
        return res.json();
      })
      .then(response => {
        // Handle both response formats: {success: true, data: [...]} or direct array
        let assets = [];
        if (response && response.success && Array.isArray(response.data)) {
          assets = response.data;
        } else if (Array.isArray(response)) {
          assets = response;
        } else if (response && Array.isArray(response.data)) {
          assets = response.data;
        }
        
        if (assets.length === 0) {
          filterGrid.innerHTML = '<p style="color: #9fb2c9; text-align: center; padding: 20px;">No superposable images available</p>';
          return;
        }
        
        filterGrid.innerHTML = '';
        assets.forEach(asset => {
          if (!asset) {
            return;
          }
          
          // Handle both camelCase and PascalCase property names
          const assetId = asset.id || asset.ID;
          const assetPath = asset.path || asset.Path;
          const assetName = asset.name || asset.Name;
          
          if (!assetId || !assetPath) {
            return;
          }
          
          const item = document.createElement('div');
          item.className = 'filter-item';
          item.dataset.src = assetPath;
          item.dataset.id = assetId;
          item.innerHTML = `
            <img src="${assetPath}" alt="${assetName || 'Asset'}" class="filter-preview" onerror="this.style.display='none';" />
            <label>${assetName || 'Unnamed'}</label>
          `;
          item.addEventListener('click', function() {
            document.querySelectorAll('.filter-item').forEach(i => i.classList.remove('active'));
            this.classList.add('active');
            selectedFilter = this.dataset.src;
            selectedAssetId = parseInt(this.dataset.id, 10);
            if (captureBtn) captureBtn.disabled = false;
            if (overlayImage) overlayImage.remove();
            overlayImage = new Image();
            overlayImage.src = selectedFilter;
            overlayImage.classList.add('overlay-image');
            overlayImage.setAttribute('draggable', 'false');
            overlayImage.onload = () => {
              const stageRect = canvasContent.getBoundingClientRect();
              const targetWidth = Math.max(80, stageRect.width * 0.25);
              const aspect = overlayImage.naturalWidth / overlayImage.naturalHeight;
              overlayState.w = targetWidth;
              overlayState.h = targetWidth / aspect;
              overlayState.x = (stageRect.width * 0.5 / zoomLevel) - overlayState.w / 2;
              overlayState.y = (stageRect.height * 0.5 / zoomLevel) - overlayState.h / 2;
              applyOverlayTransform();
            };
            overlayImage.onerror = () => {
              alert('Failed to load overlay image. Please check the image path.');
            };
            (window.canvasStage || document.getElementById('canvas-stage') || canvasContent).appendChild(overlayImage);
          });
          filterGrid.appendChild(item);
        });
      })
      .catch(() => {
        if (filterGrid) {
          filterGrid.innerHTML = '<p style="color: #ff6b6b; text-align: center; padding: 20px;">Failed to load superposable images</p>';
        }
      });
  }

  loadAssets();
  loadPreviousImages();

  function applyStageTransform() {
    const stage = window.canvasStage || document.getElementById('canvas-stage');
    if (!stage) return;
    stage.style.transform = `translate(${stageOffsetX}px, ${stageOffsetY}px) scale(${zoomLevel})`;
  }

  function applyOverlayTransform() {
    if (!overlayImage) return;
    overlayImage.style.width = overlayState.w + 'px';
    overlayImage.style.height = overlayState.h + 'px';
    const screenX = overlayState.x * zoomLevel + stageOffsetX;
    const screenY = overlayState.y * zoomLevel + stageOffsetY;
    overlayImage.style.transform = `translate(${screenX}px, ${screenY}px) scale(${zoomLevel})`;
  }

  function updateAllTransforms() {
    applyStageTransform();
    applyOverlayTransform();
  }

  if (zoomLevelDisplay) zoomLevelDisplay.textContent = `${Math.round(zoomLevel * 100)}%`;

  if (canvasContent) {
    canvasContent.addEventListener('pointerdown', (e) => {
      const isOverlay = e.target.classList.contains('overlay-image');
      isDragging = true;
      isDraggingOverlay = isOverlay;
      initialX = e.clientX;
      initialY = e.clientY;
      e.preventDefault();
    });

    window.addEventListener('pointermove', (e) => {
      if (!isDragging) return;
      const dx = e.clientX - initialX;
      const dy = e.clientY - initialY;
      if (isDraggingOverlay && overlayImage) {
        overlayState.x += dx / zoomLevel;
        overlayState.y += dy / zoomLevel;
        applyOverlayTransform();
      } else {
        stageOffsetX += dx;
        stageOffsetY += dy;
        applyStageTransform();
      }
      initialX = e.clientX;
      initialY = e.clientY;
    });

    window.addEventListener('pointerup', () => {
      isDragging = false;
      isDraggingOverlay = false;
    });
  }

  function startWebcam() {
    if (!navigator.mediaDevices || !canvasContent) return;
    navigator.mediaDevices.getUserMedia({ video: true })
      .then(stream => {
        videoStream = stream;
        canvasContent.innerHTML = '';
        const stage = document.createElement('div');
        stage.className = 'canvas-stage';
        stage.id = 'canvas-stage';

        const video = document.createElement('video');
        video.id = 'webcam';
        video.className = 'canvas-video';
        video.autoplay = true;
        video.playsinline = true;
        video.srcObject = stream;
        video.play();
        stage.appendChild(video);

        const hiddenCanvas = document.createElement('canvas');
        hiddenCanvas.id = 'photo-canvas';
        hiddenCanvas.style.display = 'none';
        stage.appendChild(hiddenCanvas);

        canvasContent.appendChild(stage);
        window.canvasStage = stage;
        updateAllTransforms();
        
        // Enable snap button when video is ready
        if (snapBtn) {
          snapBtn.disabled = false;
        }
        if (captureBtn) {
          captureBtn.disabled = false;
        }
      })
      .catch(() => {
        canvasContent.innerHTML = '<p style="color: white; padding: 20px;">Could not access webcam. Please allow camera access or upload an image instead.</p>';
      });
  }

  startWebcam();

  // Snap button - freeze/unfreeze video
  if (snapBtn) {
    snapBtn.addEventListener('click', () => {
      const video = document.getElementById('webcam');
      if (!video || !videoStream) return;
      
      if (isVideoFrozen) {
        // Unfreeze - resume video
        video.play();
        videoStream.getVideoTracks().forEach(track => {
          track.enabled = true;
        });
        isVideoFrozen = false;
        snapBtn.textContent = 'Snap';
      } else {
        // Freeze - pause video
        video.pause();
        videoStream.getVideoTracks().forEach(track => {
          track.enabled = false;
        });
        isVideoFrozen = true;
        snapBtn.textContent = 'Resume';
      }
    });
  }

  if (captureBtn) {
    captureBtn.addEventListener('click', async () => {
      if (!selectedAssetId || isNaN(selectedAssetId)) {
        alert('Please select a superposable image first');
        return;
      }

      const canvas = document.getElementById('photo-canvas');
      if (!canvas) {
        alert('Canvas not found');
        return;
      }
      const context = canvas.getContext('2d');
      const src = document.getElementById('webcam') || canvasContent.querySelector('.canvas-image');
      if (!src) {
        alert('No image source available');
        return;
      }
      
      // Wait for video to be ready
      if (src.tagName === 'VIDEO' && src.readyState < 2) {
        src.addEventListener('loadeddata', () => {
          // Retry after video is loaded
          setTimeout(() => captureBtn.click(), 100);
        });
        return;
      }

      const targetW = 1080;
      const targetH = 720;
      canvas.width = targetW;
      canvas.height = targetH;
      
      let sw, sh;
      if (src.tagName === 'VIDEO') {
        sw = src.videoWidth;
        sh = src.videoHeight;
      } else {
        sw = src.naturalWidth;
        sh = src.naturalHeight;
      }

      const fitScale = Math.max(targetW / sw, targetH / sh);
      const effectiveScale = fitScale * zoomLevel;
      const cropW = targetW / effectiveScale;
      const cropH = targetH / effectiveScale;
      let sx = (sw - cropW) / 2;
      let sy = (sh - cropH) / 2;
      sx -= stageOffsetX / effectiveScale;
      sy -= stageOffsetY / effectiveScale;
      sx = Math.max(0, Math.min(sw - cropW, sx));
      sy = Math.max(0, Math.min(sh - cropH, sy));

      // Draw base image only (server will add overlay)
      context.drawImage(src, sx, sy, cropW, cropH, 0, 0, targetW, targetH);

      // Convert canvas to blob and upload
      canvas.toBlob(async (blob) => {
        if (!blob) {
          alert('Failed to create image');
          return;
        }

        captureBtn.disabled = true;
        captureBtn.textContent = 'Uploading...';

        // Calculate overlay position in final image coordinates (1080x720)
        // overlayState is in stage coordinates (unzoomed, relative to stage)
        // The stage is inside canvas-content which is 1080x720
        // When we capture, we create a 1080x720 image, so coordinates map 1:1
        
        // overlayState.x/y are in unzoomed stage coordinates
        // The overlay's position on the canvas-content (after zoom/pan) is:
        // screenX = overlayState.x * zoomLevel + stageOffsetX
        // screenY = overlayState.y * zoomLevel + stageOffsetY
        
        // Since canvas-content is 1080x720 and final image is 1080x720, we can use screen coordinates directly
        // But we need to ensure they're within bounds
        const overlayX = Math.max(0, Math.min(1080, (overlayState.x * zoomLevel) + stageOffsetX));
        const overlayY = Math.max(0, Math.min(720, (overlayState.y * zoomLevel) + stageOffsetY));
        const overlayW = Math.max(10, Math.min(1080 - overlayX, overlayState.w * zoomLevel));
        const overlayH = Math.max(10, Math.min(720 - overlayY, overlayState.h * zoomLevel));

        const formData = new FormData();
        formData.append('image', blob, 'photo.png');
        formData.append('asset_id', String(selectedAssetId));
        formData.append('overlay_x', String(Math.round(overlayX)));
        formData.append('overlay_y', String(Math.round(overlayY)));
        formData.append('overlay_w', String(Math.round(overlayW)));
        formData.append('overlay_h', String(Math.round(overlayH)));

        try {
          const res = await fetch('/api/compose', {
            method: 'POST',
            body: formData
          });

          const json = await res.json();
          if (!res.ok || !json.success) {
            throw new Error(json.message || 'Upload failed');
          }

          // Get path from response (could be in data.path or directly in path)
          const imagePath = (json.data && json.data.path) || json.path;
          if (!imagePath) {
            throw new Error('No image path returned from server');
          }

          // Add thumbnail to sidebar
          addThumbnail(imagePath);

          alert('Image saved successfully!');
          captureBtn.textContent = 'Take Picture & Save';
        } catch (err) {
          alert('Failed to upload: ' + err.message);
          captureBtn.textContent = 'Take Picture & Save';
        } finally {
          captureBtn.disabled = false;
        }
      }, 'image/png');
    });
  }

  if (uploadInput) {
    uploadInput.addEventListener('change', (e) => {
      if (e.target.files && e.target.files[0]) {
        const file = e.target.files[0];
        const filename = file.name;
        
        // Validate filename (only allow safe characters: alphanumeric, underscore, hyphen, dot)
        const validFilenamePattern = /^[a-zA-Z0-9_.-]+$/;
        if (!validFilenamePattern.test(filename)) {
          alert('Please rename the file before uploading thanks!');
          e.target.value = ''; // Clear the input
          if (placeImageBtn) placeImageBtn.disabled = true;
          uploadedImage = null;
          return;
        }
        
        // Validate MIME type
        const validMIMETypes = ['image/jpeg', 'image/jpg', 'image/png', 'image/gif', 'image/webp'];
        if (!validMIMETypes.includes(file.type)) {
          alert('Invalid file type. Please upload an image file.');
          e.target.value = ''; // Clear the input
          if (placeImageBtn) placeImageBtn.disabled = true;
          uploadedImage = null;
          return;
        }
        
        const reader = new FileReader();
        reader.onload = function (event) {
          uploadedImage = new Image();
          uploadedImage.src = event.target.result;
          uploadedImage.onload = () => { 
            if (placeImageBtn) placeImageBtn.disabled = false; 
          };
        };
        reader.readAsDataURL(file);
      } else {
        if (placeImageBtn) placeImageBtn.disabled = true;
        uploadedImage = null;
      }
    });
  }

  if (placeImageBtn) {
    placeImageBtn.addEventListener('click', () => {
      if (!uploadedImage) return;
      const webcamEl = document.getElementById('webcam');
      if (webcamEl && webcamEl.srcObject) {
        webcamEl.srcObject.getTracks().forEach(t => t.stop());
      }
      canvasContent.innerHTML = '';
      uploadedImage.classList.add('canvas-image');
      const stage = document.createElement('div');
      stage.className = 'canvas-stage';
      stage.id = 'canvas-stage';
      stage.appendChild(uploadedImage);
      const hiddenCanvas = document.createElement('canvas');
      hiddenCanvas.id = 'photo-canvas';
      hiddenCanvas.style.display = 'none';
      stage.appendChild(hiddenCanvas);
      canvasContent.appendChild(stage);
      window.canvasStage = stage;
      updateAllTransforms();
      if (selectedFilter) {
        overlayImage = new Image();
        overlayImage.src = selectedFilter;
        overlayImage.classList.add('overlay-image');
        overlayImage.setAttribute('draggable', 'false');
        stage.appendChild(overlayImage);
        overlayState.x = 0;
        overlayState.y = 0;
        applyOverlayTransform();
      }
      if (captureBtn) captureBtn.disabled = !selectedFilter;
    });
  }

  if (zoomInBtn) {
    zoomInBtn.addEventListener('click', () => {
      zoomLevel += ZOOM_STEP;
      updateAllTransforms();
      if (zoomLevelDisplay) zoomLevelDisplay.textContent = `${Math.round(zoomLevel * 100)}%`;
    });
  }

  if (zoomOutBtn) {
    zoomOutBtn.addEventListener('click', () => {
      zoomLevel = Math.max(MIN_ZOOM, zoomLevel - ZOOM_STEP);
      updateAllTransforms();
      if (zoomLevelDisplay) zoomLevelDisplay.textContent = `${Math.round(zoomLevel * 100)}%`;
      });
  }

  // Load previous images on page load
  function loadPreviousImages() {
    if (!thumbnailList) return;
    
    fetch('/api/user/images')
      .then(res => res.json())
      .then(response => {
        if (response.success && Array.isArray(response.data)) {
          thumbnailList.innerHTML = '';
          response.data.forEach(img => {
            if (img.path) {
              addThumbnail(img.path);
            }
          });
        }
      })
      .catch(() => {
      });
  }

  // Helper function to add a thumbnail
  function addThumbnail(imagePath) {
    if (!thumbnailList) return;
    
    const thumb = document.createElement('img');
    thumb.src = imagePath;
    thumb.classList.add('thumbnail');
    thumb.style.cursor = 'pointer';
    thumb.onclick = () => window.location.href = '/gallery';
    thumb.title = 'Click to view in gallery';
    thumbnailList.prepend(thumb);
  }
});
