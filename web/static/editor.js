// Camagru Editor JavaScript (consolidated)
document.addEventListener('DOMContentLoaded', () => {
  const webcam = document.getElementById('webcam');
  const captureBtn = document.getElementById('capture-btn');
  const thumbnailList = document.querySelector('.thumbnail-list');
  const uploadInput = document.getElementById('upload-btn');
  const placeImageBtn = document.getElementById('place-image-btn');
  const canvasContent = document.querySelector('.canvas-content');
  const zoomInBtn = document.getElementById('zoom-in-btn');
  const zoomOutBtn = document.getElementById('zoom-out-btn');
  const zoomLevelDisplay = document.getElementById('zoom-level');
  const canvasStage = document.getElementById('canvas-stage');

  let selectedFilter = null;
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
        canvasContent.innerHTML = '';
        const stage = document.createElement('div');
        stage.className = 'canvas-stage';
        stage.id = 'canvas-stage';

        const video = document.createElement('video');
        video.id = 'webcam';
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
      })
      .catch(err => {
        console.log('Webcam error:', err);
        canvasContent.innerHTML = '<p style="color: white;">Could not access webcam.</p>';
      });
  }

  startWebcam();

  document.querySelectorAll('.filter-item').forEach(item => {
    item.addEventListener('click', function () {
      document.querySelectorAll('.filter-item').forEach(i => i.classList.remove('active'));
      this.classList.add('active');
      selectedFilter = this.dataset.src;
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
      (window.canvasStage || canvasStage || canvasContent).appendChild(overlayImage);
    });
  });

  if (captureBtn) {
    captureBtn.addEventListener('click', () => {
      const canvas = document.getElementById('photo-canvas');
      const context = canvas.getContext('2d');
      const src = document.getElementById('webcam') || canvasContent.querySelector('.canvas-image');
      if (!src) return;
      const targetW = 1080;
      const targetH = 720;
      canvas.width = targetW;
      canvas.height = targetH;
      let sw, sh;
      if (src.tagName === 'VIDEO') { sw = src.videoWidth; sh = src.videoHeight; }
      else { sw = src.naturalWidth; sh = src.naturalHeight; }
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
      context.drawImage(src, sx, sy, cropW, cropH, 0, 0, targetW, targetH);
      if (overlayImage) {
        const scaleX = targetW / 1080;
        const scaleY = targetH / 720;
        const ox = overlayState.x * scaleX;
        const oy = overlayState.y * scaleY;
        const ow = overlayState.w * scaleX;
        const oh = overlayState.h * scaleY;
        context.drawImage(overlayImage, ox, oy, ow, oh);
      }
      const dataUrl = canvas.toDataURL('image/png');
      const a = document.createElement('a');
      a.href = dataUrl;
      a.download = `camagru_${Date.now()}.png`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      const thumb = document.createElement('img');
      thumb.src = dataUrl;
      thumb.classList.add('thumbnail');
      if (thumbnailList) thumbnailList.prepend(thumb);
    });
  }

  if (uploadInput) {
    uploadInput.addEventListener('change', (e) => {
      if (e.target.files && e.target.files[0]) {
        const reader = new FileReader();
        reader.onload = function (event) {
          uploadedImage = new Image();
          uploadedImage.src = event.target.result;
          uploadedImage.onload = () => { placeImageBtn.disabled = false; };
        };
        reader.readAsDataURL(e.target.files[0]);
      } else {
        placeImageBtn.disabled = true;
        uploadedImage = null;
      }
    });
  }

  if (placeImageBtn) {
    placeImageBtn.addEventListener('click', () => {
      if (!uploadedImage) return;
      const webcamEl = document.getElementById('webcam');
      if (webcamEl && webcamEl.srcObject) { webcamEl.srcObject.getTracks().forEach(t => t.stop()); }
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
      captureBtn.disabled = !selectedFilter;
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
});
