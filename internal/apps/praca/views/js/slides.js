(function() {
  let slides = [];
  let currentIndex = 0;
  let isPlaying = false;
  let isShuffle = false;
  let isAdmin = false;
  let intervalMs = 5000;
  let timerId = null;

  const slidesToolbar = document.getElementById('slides-toolbar');
  const stageEl = document.getElementById('slideshow-stage');
  const stageFrame = document.getElementById('stage-frame');
  const emptyStateEl = document.getElementById('empty-state');
  const emptyStateMsg = document.getElementById('empty-state-msg');
  const thumbsSection = document.getElementById('thumbnails-section');
  const thumbsGrid = document.getElementById('thumbnails-grid');
  const totalCountEl = document.getElementById('total-count');

  const slideImg = document.getElementById('slide-image');
  const slideCounter = document.getElementById('slide-counter');
  const slideTitle = document.getElementById('slide-title');
  const slideCaption = document.getElementById('slide-caption');
  const slideOverlay = document.getElementById('slide-overlay-info');
  const progressBar = document.getElementById('slide-progress-bar');

  const btnPrev = document.getElementById('btn-prev');
  const btnNext = document.getElementById('btn-next');
  const btnPlayPause = document.getElementById('btn-play-pause');
  const intervalSelect = document.getElementById('interval-select');
  const btnShuffle = document.getElementById('btn-shuffle');
  const btnFullscreen = document.getElementById('btn-fullscreen');

  const uploadModal = document.getElementById('upload-modal');
  const btnOpenUpload = document.getElementById('btn-open-upload');
  const btnCloseModal = document.getElementById('btn-close-modal');
  const uploadForm = document.getElementById('upload-form');
  const photoInput = document.getElementById('photo-input');
  const previewWrapper = document.getElementById('preview-wrapper');
  const previewCount = document.getElementById('preview-count');
  const previewGrid = document.getElementById('preview-grid');
  const uploadError = document.getElementById('upload-error');
  const btnSubmitUpload = document.getElementById('btn-submit-upload');

  async function checkAdmin() {
    try {
      const resp = await fetch('/praca/api/me', { credentials: 'same-origin' });
      if (resp.ok) {
        const data = await resp.json();
        isAdmin = Boolean(data.is_admin);
      }
    } catch (err) {
      isAdmin = false;
    }

    if (isAdmin) {
      slidesToolbar.classList.remove('hidden');
      if (emptyStateMsg) {
        emptyStateMsg.textContent = 'Clique em "+ Adicionar Fotos" para enviar imagens para o expositor.';
      }
    } else {
      slidesToolbar.classList.add('hidden');
    }
  }

  async function loadSlides() {
    try {
      const resp = await fetch('/praca/api/slides');
      if (!resp.ok) throw new Error('Falha ao carregar fotos');
      slides = await resp.json();
      if (!Array.isArray(slides)) slides = [];
      if (currentIndex >= slides.length) currentIndex = 0;
      renderAll();
    } catch (err) {
      console.error(err);
    }
  }

  function renderAll() {
    if (slides.length === 0) {
      stopAutoPlay();
      stageEl.classList.add('hidden');
      thumbsSection.classList.add('hidden');
      emptyStateEl.classList.remove('hidden');
      return;
    }

    emptyStateEl.classList.add('hidden');
    stageEl.classList.remove('hidden');
    thumbsSection.classList.remove('hidden');
    totalCountEl.textContent = slides.length;

    renderCurrentSlide();
    renderThumbnails();
  }

  function renderCurrentSlide() {
    const slide = slides[currentIndex];
    if (!slide) return;

    slideImg.src = slide.image_url;
    slideImg.alt = slide.title || 'Foto do expositor';
    slideCounter.textContent = `${currentIndex + 1} / ${slides.length}`;

    slideTitle.textContent = slide.title || '';
    slideCaption.textContent = slide.caption || '';
    if (!slide.title && !slide.caption) {
      slideOverlay.classList.add('is-empty');
    } else {
      slideOverlay.classList.remove('is-empty');
    }

    document.querySelectorAll('.thumb-card').forEach((el, idx) => {
      el.classList.toggle('active', idx === currentIndex);
    });

    resetProgressBar();
  }

  function renderThumbnails() {
    thumbsGrid.innerHTML = '';
    slides.forEach((slide, idx) => {
      const card = document.createElement('div');
      card.className = `thumb-card ${idx === currentIndex ? 'active' : ''}`;
      card.addEventListener('click', () => {
        currentIndex = idx;
        renderCurrentSlide();
        if (isPlaying) restartTimer();
      });

      const img = document.createElement('img');
      img.src = slide.image_url;
      img.alt = slide.title || `Foto ${idx + 1}`;
      img.loading = 'lazy';

      card.appendChild(img);

      if (isAdmin) {
        const delBtn = document.createElement('button');
        delBtn.className = 'thumb-delete-btn';
        delBtn.title = 'Remover foto';
        delBtn.textContent = '🗑️';
        delBtn.addEventListener('click', async (e) => {
          e.stopPropagation();
          await deleteSlide(slide.id);
        });
        card.appendChild(delBtn);
      }

      thumbsGrid.appendChild(card);
    });
  }

  function nextSlide() {
    if (slides.length <= 1) return;
    if (isShuffle) {
      let nextIdx = currentIndex;
      while (nextIdx === currentIndex) {
        nextIdx = Math.floor(Math.random() * slides.length);
      }
      currentIndex = nextIdx;
    } else {
      currentIndex = (currentIndex + 1) % slides.length;
    }
    renderCurrentSlide();
  }

  function prevSlide() {
    if (slides.length <= 1) return;
    currentIndex = (currentIndex - 1 + slides.length) % slides.length;
    renderCurrentSlide();
  }

  function resetProgressBar() {
    progressBar.style.transition = 'none';
    progressBar.style.width = '0%';
    if (isPlaying && slides.length > 1) {
      void progressBar.offsetWidth;
      progressBar.style.transition = `width ${intervalMs}ms linear`;
      progressBar.style.width = '100%';
    }
  }

  function startAutoPlay() {
    if (slides.length <= 1) return;
    isPlaying = true;
    btnPlayPause.textContent = '⏸ Pausar';
    btnPlayPause.classList.add('active');
    restartTimer();
  }

  function stopAutoPlay() {
    isPlaying = false;
    btnPlayPause.textContent = '▶ Reproduzir';
    btnPlayPause.classList.remove('active');
    if (timerId) {
      clearInterval(timerId);
      timerId = null;
    }
    resetProgressBar();
  }

  function restartTimer() {
    if (timerId) clearInterval(timerId);
    resetProgressBar();
    timerId = setInterval(nextSlide, intervalMs);
  }

  async function deleteSlide(id) {
    if (!confirm('Remover esta foto do expositor?')) return;
    try {
      const resp = await fetch(`/praca/api/slides?id=${id}`, {
        method: 'DELETE',
        credentials: 'same-origin'
      });
      if (!resp.ok) throw new Error('Erro ao remover foto');
      await loadSlides();
    } catch (err) {
      alert(err.message);
    }
  }

  btnPrev.addEventListener('click', () => {
    prevSlide();
    if (isPlaying) restartTimer();
  });

  btnNext.addEventListener('click', () => {
    nextSlide();
    if (isPlaying) restartTimer();
  });

  btnPlayPause.addEventListener('click', () => {
    if (isPlaying) {
      stopAutoPlay();
    } else {
      startAutoPlay();
    }
  });

  intervalSelect.addEventListener('change', (e) => {
    intervalMs = parseInt(e.target.value, 10) || 5000;
    if (isPlaying) restartTimer();
  });

  btnShuffle.addEventListener('click', () => {
    isShuffle = !isShuffle;
    btnShuffle.classList.toggle('active', isShuffle);
  });

  btnFullscreen.addEventListener('click', () => {
    if (!document.fullscreenElement) {
      stageFrame.requestFullscreen().catch(() => {});
    } else {
      document.exitFullscreen().catch(() => {});
    }
  });

  // Navegação por teclado
  document.addEventListener('keydown', (e) => {
    if (!uploadModal.classList.contains('hidden')) return;
    if (slides.length === 0) return;

    if (e.key === 'ArrowRight') {
      nextSlide();
      if (isPlaying) restartTimer();
    } else if (e.key === 'ArrowLeft') {
      prevSlide();
      if (isPlaying) restartTimer();
    } else if (e.key === ' ') {
      e.preventDefault();
      if (isPlaying) stopAutoPlay();
      else startAutoPlay();
    } else if (e.key === 'f' || e.key === 'F') {
      btnFullscreen.click();
    }
  });

  // Suporte a swipe em telas touch
  let touchStartX = 0;
  stageFrame.addEventListener('touchstart', (e) => {
    touchStartX = e.changedTouches[0].screenX;
  }, { passive: true });
  stageFrame.addEventListener('touchend', (e) => {
    const diff = e.changedTouches[0].screenX - touchStartX;
    if (Math.abs(diff) > 40) {
      if (diff < 0) nextSlide();
      else prevSlide();
      if (isPlaying) restartTimer();
    }
  }, { passive: true });

  // Modal de upload
  btnOpenUpload.addEventListener('click', () => {
    uploadForm.reset();
    previewWrapper.classList.add('hidden');
    uploadError.classList.add('hidden');
    uploadModal.classList.remove('hidden');
  });

  btnCloseModal.addEventListener('click', () => {
    uploadModal.classList.add('hidden');
  });

  uploadModal.addEventListener('click', (e) => {
    if (e.target === uploadModal) {
      uploadModal.classList.add('hidden');
    }
  });

  photoInput.addEventListener('change', () => {
    const files = Array.from(photoInput.files || []);
    previewGrid.innerHTML = '';
    if (files.length === 0) {
      previewWrapper.classList.add('hidden');
      return;
    }

    previewCount.textContent = `${files.length} foto(s) selecionada(s)`;
    previewWrapper.classList.remove('hidden');

    files.forEach((file) => {
      const reader = new FileReader();
      reader.onload = (ev) => {
        const img = document.createElement('img');
        img.src = ev.target.result;
        img.alt = file.name;
        previewGrid.appendChild(img);
      };
      reader.readAsDataURL(file);
    });
  });

  uploadForm.addEventListener('submit', async (e) => {
    e.preventDefault();
    uploadError.classList.add('hidden');

    const files = Array.from(photoInput.files || []);
    if (files.length === 0) {
      uploadError.textContent = 'Selecione pelo menos uma imagem.';
      uploadError.classList.remove('hidden');
      return;
    }

    for (const file of files) {
      if (file.size > 5 * 1024 * 1024) {
        uploadError.textContent = `A imagem "${file.name}" ultrapassa o limite de 5 MB.`;
        uploadError.classList.remove('hidden');
        return;
      }
    }

    btnSubmitUpload.disabled = true;
    btnSubmitUpload.textContent = `Enviando ${files.length} foto(s)...`;

    try {
      const formData = new FormData();
      files.forEach((file) => formData.append('images', file));
      formData.append('title', document.getElementById('photo-title').value);
      formData.append('caption', document.getElementById('photo-caption').value);

      const resp = await fetch('/praca/api/slides', {
        method: 'POST',
        credentials: 'same-origin',
        body: formData
      });

      if (!resp.ok) {
        const msg = await resp.text();
        throw new Error(msg || 'Erro ao enviar fotos');
      }

      uploadModal.classList.add('hidden');
      currentIndex = 0;
      await loadSlides();
    } catch (err) {
      uploadError.textContent = err.message;
      uploadError.classList.remove('hidden');
    } finally {
      btnSubmitUpload.disabled = false;
      btnSubmitUpload.textContent = 'Enviar Fotos';
    }
  });

  (async () => {
    await checkAdmin();
    await loadSlides();
  })();
})();
