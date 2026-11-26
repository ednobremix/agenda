(function() {
  let isAdmin = false;
  let isEditMode = false;
  let galleryData = { config: { title: 'Museu Virtual da Praça', subtitle: '', template: 'classic' }, rooms: [], artworks: [] };
  let currentRoomId = 0;
  let slidesList = [];
  let selectedArtwork = null;
  let editingRoomId = 0;
  let tourTimer = null;
  let tourIndex = 0;

  // Estado da câmera 3D
  const cam = {
    x: 0,
    z: 180,
    yaw: 0,
    pitch: 0
  };

  const worldEl = document.getElementById('museum-world');
  const viewportEl = document.getElementById('museum-viewport');
  const emptyHintEl = document.getElementById('museum-empty-hint');
  const titleEl = document.getElementById('exhibition-title');
  const subtitleEl = document.getElementById('exhibition-subtitle');

  const btnShare = document.getElementById('btn-share');
  const shareToast = document.getElementById('share-toast');
  const btnAdminRooms = document.getElementById('btn-admin-rooms');
  const btnAdminEditMode = document.getElementById('btn-admin-edit-mode');
  const editModeBanner = document.getElementById('edit-mode-banner');
  const btnAdminConfig = document.getElementById('btn-admin-config');
  const btnAdminAdd = document.getElementById('btn-admin-add');
  const roomSelect = document.getElementById('room-select');
  const btnTour = document.getElementById('btn-tour');

  const roomsModal = document.getElementById('rooms-modal');
  const closeRoomsModal = document.getElementById('close-rooms-modal');
  const roomsListEl = document.getElementById('rooms-list');
  const roomForm = document.getElementById('room-form');
  const btnNewRoom = document.getElementById('btn-new-room');
  const btnEnterRoom = document.getElementById('btn-enter-room');
  const btnDeleteRoom = document.getElementById('btn-delete-room');

  const detailModal = document.getElementById('artwork-detail-modal');
  const closeDetailModal = document.getElementById('close-detail-modal');
  const detailImage = document.getElementById('detail-image');
  const detailTitle = document.getElementById('detail-title');
  const detailArtistYear = document.getElementById('detail-artist-year');
  const detailTechDim = document.getElementById('detail-tech-dim');
  const detailDescription = document.getElementById('detail-description');
  const detailExtLink = document.getElementById('detail-external-link');
  const detailAdminActions = document.getElementById('detail-admin-actions');
  const btnEditArtwork = document.getElementById('btn-edit-artwork');
  const btnDeleteArtwork = document.getElementById('btn-delete-artwork');

  const configModal = document.getElementById('config-modal');
  const closeConfigModal = document.getElementById('close-config-modal');
  const configForm = document.getElementById('config-form');

  const artworkFormModal = document.getElementById('artwork-form-modal');
  const closeArtworkFormModal = document.getElementById('close-artwork-form-modal');
  const artworkForm = document.getElementById('artwork-form');
  const artworkFormHeading = document.getElementById('artwork-form-heading');
  const artworkFormError = document.getElementById('artwork-form-error');
  const artSlideSelect = document.getElementById('art-slide-id');
  const artFileInput = document.getElementById('art-file-input');
  const srcExistingBox = document.getElementById('src-existing-box');
  const srcUploadBox = document.getElementById('src-upload-box');
  let imageSourceMode = 'existing';

  function updateCameraTransform(smooth = false) {
    if (smooth) {
      worldEl.classList.add('smooth-move');
      setTimeout(() => worldEl.classList.remove('smooth-move'), 680);
    }
    const roomH = (typeof getCurrentRoom === 'function' ? getCurrentRoom().height : 600) || 600;
    // Visão mais baixa (~40% da altura da porta / 28% acima do piso)
    const eyeY = roomH * 0.22;
    worldEl.style.transform =
      `translate3d(0, 0, 820px) rotateX(${cam.pitch}deg) rotateY(${cam.yaw}deg) translate3d(${-cam.x}px, ${-eyeY}px, ${-cam.z}px)`;
  }

  function focusWall(wall) {
    const targets = {
      north: { x: 0, z: 180, yaw: 0, pitch: 0 },
      east:  { x: -180, z: 0, yaw: 90, pitch: 0 },
      south: { x: 0, z: -180, yaw: 180, pitch: 0 },
      west:  { x: 180, z: 0, yaw: -90, pitch: 0 }
    };
    const t = targets[wall] || targets.north;
    cam.x = t.x;
    cam.z = t.z;
    cam.yaw = t.yaw;
    cam.pitch = t.pitch;
    updateCameraTransform(true);

    document.querySelectorAll('.wall-btn[data-focus-wall]').forEach((btn) => {
      btn.classList.toggle('active', btn.dataset.focusWall === wall);
    });
  }

  function moveCamera(forwardStep, strafeStep) {
    const rad = (cam.yaw * Math.PI) / 180;
    const dx = Math.sin(rad) * forwardStep + Math.cos(rad) * strafeStep;
    const dz = -Math.cos(rad) * forwardStep + Math.sin(rad) * strafeStep;

    cam.x = Math.max(-440, Math.min(440, cam.x + dx));
    cam.z = Math.max(-440, Math.min(440, cam.z + dz));
    updateCameraTransform(false);
  }

  async function checkAdmin() {
    try {
      const resp = await fetch('/praca/api/me', { credentials: 'same-origin' });
      if (resp.ok) {
        const data = await resp.json();
        isAdmin = Boolean(data.is_admin);
      }
    } catch (e) {
      isAdmin = false;
    }

    btnAdminRooms.classList.toggle('hidden', !isAdmin);
    btnAdminEditMode.classList.toggle('hidden', !isAdmin);
    btnAdminConfig.classList.toggle('hidden', !isAdmin);
    btnAdminAdd.classList.toggle('hidden', !isAdmin);
  }

  async function loadGallery() {
    try {
      const resp = await fetch('/praca/api/gallery');
      if (!resp.ok) throw new Error('Falha ao carregar galeria');
      galleryData = await resp.json();
      renderGallery();
    } catch (err) {
      console.error(err);
    }
  }

  async function loadSlidesForAdmin() {
    if (!isAdmin) return;
    try {
      const resp = await fetch('/praca/api/slides');
      if (resp.ok) {
        slidesList = await resp.json();
        artSlideSelect.innerHTML = '';
        if (!Array.isArray(slidesList) || slidesList.length === 0) {
          const opt = document.createElement('option');
          opt.value = '';
          opt.textContent = 'Nenhuma imagem na galeria (envie uma nova)';
          artSlideSelect.appendChild(opt);
        } else {
          slidesList.forEach((s, idx) => {
            const opt = document.createElement('option');
            opt.value = s.id;
            opt.textContent = s.title ? `${s.title} (#${s.id})` : `Foto #${s.id}`;
            if (idx === 0) opt.selected = true;
            artSlideSelect.appendChild(opt);
          });
        }
      }
    } catch (e) {
      console.error(e);
    }
  }

  function applyTemplate(template) {
    document.body.classList.remove('template-classic', 'template-modern', 'template-dark');
    document.body.classList.add(`template-${template || 'classic'}`);
  }

  async function persistArtworkLayout(art) {
    try {
      const resp = await fetch('/praca/api/gallery/artworks', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'same-origin',
        body: JSON.stringify(art)
      });
      if (!resp.ok) throw new Error('Falha ao salvar posição/tamanho');
    } catch (err) {
      console.error(err);
    }
  }

  function applyImageSize(img, art) {
    if (art.width_px && art.width_px > 0) {
      img.style.width = `${Math.round(art.width_px)}px`;
      img.style.height = 'auto';
      img.style.maxWidth = 'none';
      img.style.maxHeight = '360px';
    } else {
      img.style.width = '';
      img.style.height = '';
      img.style.maxWidth = '';
      img.style.maxHeight = '';
    }
  }

  function getCurrentRoom() {
    const rooms = galleryData.rooms || [];
    if (rooms.length === 0) {
      return { id: 1, name: 'Sala Principal', template: galleryData.config?.template || 'classic', width: 1200, depth: 1200, height: 600 };
    }
    return rooms.find((r) => r.id === currentRoomId) || rooms[0];
  }

  function applyRoomGeometry(room) {
    const w = room.width || 1200;
    const d = room.depth || 1200;
    const h = room.height || 600;
    worldEl.style.setProperty('--room-w', `${w}px`);
    worldEl.style.setProperty('--room-d', `${d}px`);
    worldEl.style.setProperty('--room-h', `${h}px`);
    worldEl.style.setProperty('--half-w', `${w / 2}px`);
    worldEl.style.setProperty('--half-d', `${d / 2}px`);
    worldEl.style.setProperty('--half-h', `${h / 2}px`);
  }

  function enterRoom(roomId) {
    currentRoomId = roomId;
    renderGallery();
    focusWall('north');
  }

  function renderRoomSelect() {
    const rooms = galleryData.rooms || [];
    const active = getCurrentRoom();
    currentRoomId = active.id;
    roomSelect.innerHTML = '';
    rooms.forEach((r) => {
      const opt = document.createElement('option');
      opt.value = r.id;
      opt.textContent = `🏛️ ${r.name}`;
      if (r.id === currentRoomId) opt.selected = true;
      roomSelect.appendChild(opt);
    });
  }

  function renderGallery() {
    const cfg = galleryData.config || {};
    const room = getCurrentRoom();
    titleEl.textContent = cfg.title || 'Museu Virtual da Praça';
    subtitleEl.textContent = room.description ? `${room.name} — ${room.description}` : (cfg.subtitle || room.name);
    applyTemplate(room.template || cfg.template);
    applyRoomGeometry(room);
    renderRoomSelect();

    document.querySelectorAll('.artwork-mount, .museum-doorway').forEach((el) => el.remove());

    // Renderizar portas entre salas
    const rooms = galleryData.rooms || [];
    const doorMap = [
      { wall: 'north', targetId: room.door_north_room_id, posKey: 'door_north_pos_x' },
      { wall: 'east', targetId: room.door_east_room_id, posKey: 'door_east_pos_x' },
      { wall: 'south', targetId: room.door_south_room_id, posKey: 'door_south_pos_x' },
      { wall: 'west', targetId: room.door_west_room_id, posKey: 'door_west_pos_x' }
    ];
    doorMap.forEach(({ wall, targetId, posKey }) => {
      if (!targetId) return;
      const targetRoom = rooms.find((r) => r.id === targetId);
      if (!targetRoom) return;
      const wallEl = document.querySelector(`.museum-wall[data-wall="${wall}"]`);
      if (!wallEl) return;
      const door = document.createElement('div');
      door.className = 'museum-doorway';
      door.style.left = `${room[posKey] || 82}%`;
      door.innerHTML = `<span>🚪</span><strong>${targetRoom.name}</strong>`;

      let draggingDoor = false;
      let doorMoved = false;
      let doorStartX = 0;
      let doorStartPos = 82;
      let doorBadge = null;

      door.addEventListener('pointerdown', (e) => {
        if (!isAdmin || !isEditMode) return;
        e.stopPropagation();
        draggingDoor = true;
        doorMoved = false;
        doorStartX = e.clientX;
        doorStartPos = room[posKey] || 82;
        door.setPointerCapture(e.pointerId);
      });

      door.addEventListener('pointermove', (e) => {
        if (!draggingDoor) return;
        e.stopPropagation();
        const dx = e.clientX - doorStartX;
        if (Math.abs(dx) > 3) doorMoved = true;
        const wallSpan = (wall === 'north' || wall === 'south') ? (room.width || 1200) : (room.depth || 1200);
        const nextPos = Math.max(10, Math.min(90, +(doorStartPos + (dx / (wallSpan / 100))).toFixed(1)));
        room[posKey] = nextPos;
        door.style.left = `${nextPos}%`;
        if (!doorBadge) {
          doorBadge = document.createElement('div');
          doorBadge.className = 'artwork-live-badge';
          door.appendChild(doorBadge);
        }
        doorBadge.textContent = `Porta X:${Math.round(nextPos)}%`;
      });

      door.addEventListener('pointerup', async (e) => {
        if (!draggingDoor) return;
        e.stopPropagation();
        draggingDoor = false;
        if (doorBadge) {
          doorBadge.remove();
          doorBadge = null;
        }
        if (doorMoved) {
          await fetch('/praca/api/gallery/rooms', {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            credentials: 'same-origin',
            body: JSON.stringify(room)
          });
        } else {
          enterRoom(targetRoom.id);
        }
      });

      door.addEventListener('click', (e) => {
        e.stopPropagation();
        if (isEditMode) return;
        enterRoom(targetRoom.id);
      });
      wallEl.appendChild(door);
    });

    const artworks = (galleryData.artworks || []).filter((a) => (a.room_id || rooms[0]?.id || 1) === room.id);
    emptyHintEl.classList.toggle('hidden', artworks.length > 0);

    artworks.forEach((art) => {
      const wallEl = document.querySelector(`.museum-wall[data-wall="${art.wall}"]`);
      if (!wallEl) return;

      const mount = document.createElement('div');
      mount.className = 'artwork-mount';
      mount.style.left = `${art.pos_x}%`;
      mount.style.top = `${art.pos_y}%`;

      const frame = document.createElement('div');
      frame.className = `artwork-frame frame-${art.frame_style || 'classic-gold'} scale-${art.scale || 'medium'}`;

      const img = document.createElement('img');
      img.src = art.image_url;
      img.alt = art.title || 'Obra';
      applyImageSize(img, art);
      frame.appendChild(img);

      const resizeHandle = document.createElement('div');
      resizeHandle.className = 'artwork-resize-handle';
      resizeHandle.title = 'Arraste para redimensionar';
      resizeHandle.textContent = '↘';
      frame.appendChild(resizeHandle);

      const plaque = document.createElement('div');
      plaque.className = 'artwork-plaque';
      const strong = document.createElement('strong');
      strong.textContent = art.title || 'Sem título';
      const span = document.createElement('span');
      span.textContent = [art.artist, art.year].filter(Boolean).join(' • ') || 'Ver ficha técnica';
      plaque.appendChild(strong);
      plaque.appendChild(span);

      mount.appendChild(frame);
      mount.appendChild(plaque);

      // Arrastar obra na parede (Modo Edição)
      let mode = null; // 'move' | 'resize'
      let startX = 0;
      let startY = 0;
      let startPosX = 0;
      let startPosY = 0;
      let startWidth = 0;
      let moved = false;
      let liveBadge = null;

      function showBadge(text) {
        if (!liveBadge) {
          liveBadge = document.createElement('div');
          liveBadge.className = 'artwork-live-badge';
          mount.appendChild(liveBadge);
        }
        liveBadge.textContent = text;
      }

      function removeBadge() {
        if (liveBadge) {
          liveBadge.remove();
          liveBadge = null;
        }
      }

      resizeHandle.addEventListener('pointerdown', (e) => {
        if (!isAdmin || !isEditMode) return;
        e.stopPropagation();
        e.preventDefault();
        mode = 'resize';
        moved = false;
        startX = e.clientX;
        startWidth = art.width_px && art.width_px > 0 ? art.width_px : img.getBoundingClientRect().width || 220;
        resizeHandle.setPointerCapture(e.pointerId);
      });

      resizeHandle.addEventListener('pointermove', (e) => {
        if (mode !== 'resize') return;
        e.stopPropagation();
        const dx = e.clientX - startX;
        if (Math.abs(dx) > 2) moved = true;
        const nextWidth = Math.max(100, Math.min(440, startWidth + dx));
        art.width_px = Math.round(nextWidth);
        applyImageSize(img, art);
        showBadge(`${art.width_px}px`);
      });

      resizeHandle.addEventListener('pointerup', async (e) => {
        if (mode !== 'resize') return;
        e.stopPropagation();
        mode = null;
        removeBadge();
        if (moved) {
          await persistArtworkLayout(art);
        }
      });

      mount.addEventListener('pointerdown', (e) => {
        if (!isAdmin || !isEditMode || e.target === resizeHandle) return;
        e.stopPropagation();
        mode = 'move';
        moved = false;
        startX = e.clientX;
        startY = e.clientY;
        startPosX = art.pos_x;
        startPosY = art.pos_y;
        mount.setPointerCapture(e.pointerId);
      });

      mount.addEventListener('pointermove', (e) => {
        if (mode !== 'move') return;
        e.stopPropagation();
        const dx = e.clientX - startX;
        const dy = e.clientY - startY;
        if (Math.abs(dx) > 3 || Math.abs(dy) > 3) moved = true;

        art.pos_x = Math.max(10, Math.min(90, +(startPosX + dx / 12).toFixed(1)));
        art.pos_y = Math.max(25, Math.min(75, +(startPosY + dy / 6).toFixed(1)));
        mount.style.left = `${art.pos_x}%`;
        mount.style.top = `${art.pos_y}%`;
        showBadge(`X:${Math.round(art.pos_x)}% Y:${Math.round(art.pos_y)}%`);
      });

      mount.addEventListener('pointerup', async (e) => {
        if (mode !== 'move') return;
        e.stopPropagation();
        mode = null;
        removeBadge();
        if (moved) {
          await persistArtworkLayout(art);
        } else {
          openDetailModal(art);
        }
      });

      mount.addEventListener('click', (e) => {
        e.stopPropagation();
        if (isEditMode) return;
        openDetailModal(art);
      });

      wallEl.appendChild(mount);
    });
  }

  function openDetailModal(art) {
    selectedArtwork = art;
    detailImage.src = art.image_url;
    detailTitle.textContent = art.title || 'Sem título';
    detailArtistYear.textContent = [art.artist, art.year].filter(Boolean).join(' — ');
    detailTechDim.textContent = [art.technique, art.dimensions].filter(Boolean).join(' | ');
    detailDescription.textContent = art.description || '';

    if (art.external_link) {
      detailExtLink.href = art.external_link;
      detailExtLink.textContent = art.external_link_label || '🔗 Acessar link externo';
      detailExtLink.classList.remove('hidden');
    } else {
      detailExtLink.classList.add('hidden');
    }

    detailAdminActions.classList.toggle('hidden', !isAdmin);
    detailModal.classList.remove('hidden');
  }

  function stopTour() {
    if (tourTimer) {
      clearInterval(tourTimer);
      tourTimer = null;
    }
    btnTour.classList.remove('active');
  }

  function toggleTour() {
    if (tourTimer) {
      stopTour();
      return;
    }
    const artworks = galleryData.artworks || [];
    if (artworks.length === 0) return;
    btnTour.classList.add('active');
    const step = () => {
      const art = artworks[tourIndex % artworks.length];
      focusWall(art.wall);
      tourIndex++;
    };
    step();
    tourTimer = setInterval(step, 4000);
  }

  // Arrastar para olhar ao redor (mouse & touch)
  let dragging = false;
  let lastX = 0;
  let lastY = 0;

  viewportEl.addEventListener('pointerdown', (e) => {
    if (e.target.closest('.artwork-mount')) return;
    stopTour();
    dragging = true;
    lastX = e.clientX;
    lastY = e.clientY;
  });

  window.addEventListener('pointermove', (e) => {
    if (!dragging) return;
    const dx = e.clientX - lastX;
    const dy = e.clientY - lastY;
    lastX = e.clientX;
    lastY = e.clientY;

    cam.yaw -= dx * 0.28;
    cam.pitch = Math.max(-25, Math.min(25, cam.pitch + dy * 0.18));
    updateCameraTransform(false);
  });

  window.addEventListener('pointerup', () => {
    dragging = false;
  });

  // Teclado (Desktop)
  window.addEventListener('keydown', (e) => {
    if (!detailModal.classList.contains('hidden') ||
        !configModal.classList.contains('hidden') ||
        !artworkFormModal.classList.contains('hidden')) {
      if (e.key === 'Escape') {
        detailModal.classList.add('hidden');
        configModal.classList.add('hidden');
        artworkFormModal.classList.add('hidden');
      }
      return;
    }

    if (e.key === 'w' || e.key === 'W' || e.key === 'ArrowUp') moveCamera(45, 0);
    if (e.key === 's' || e.key === 'S' || e.key === 'ArrowDown') moveCamera(-45, 0);
    if (e.key === 'a' || e.key === 'A') moveCamera(0, -45);
    if (e.key === 'd' || e.key === 'D') moveCamera(0, 45);
    if (e.key === 'ArrowLeft') { cam.yaw -= 12; updateCameraTransform(false); }
    if (e.key === 'ArrowRight') { cam.yaw += 12; updateCameraTransform(false); }
  });

  // Botões de parede & movimento
  document.querySelectorAll('.wall-btn[data-focus-wall]').forEach((btn) => {
    btn.addEventListener('click', () => {
      stopTour();
      focusWall(btn.dataset.focusWall);
    });
  });

  document.getElementById('pad-forward').addEventListener('click', () => moveCamera(60, 0));
  document.getElementById('pad-back').addEventListener('click', () => moveCamera(-60, 0));
  document.getElementById('pad-left').addEventListener('click', () => { cam.yaw -= 25; updateCameraTransform(true); });
  document.getElementById('pad-right').addEventListener('click', () => { cam.yaw += 25; updateCameraTransform(true); });
  btnTour.addEventListener('click', toggleTour);

  // Compartilhar link
  btnShare.addEventListener('click', async () => {
    const shareUrl = window.location.href;
    if (navigator.share) {
      try {
        await navigator.share({
          title: titleEl.textContent,
          text: 'Visite a Galeria 3D na Praça do Portal Nobre!',
          url: shareUrl
        });
        return;
      } catch (e) {}
    }
    try {
      await navigator.clipboard.writeText(shareUrl);
    } catch (e) {}
    shareToast.classList.remove('hidden');
    setTimeout(() => shareToast.classList.add('hidden'), 2500);
  });

  closeDetailModal.addEventListener('click', () => detailModal.classList.add('hidden'));
  detailModal.addEventListener('click', (e) => { if (e.target === detailModal) detailModal.classList.add('hidden'); });

  roomSelect.addEventListener('change', (e) => {
    stopTour();
    enterRoom(parseInt(e.target.value, 10));
  });

  // Admin: Modelador de Estrutura de Salas
  function updateBlueprint() {
    const w = parseInt(document.getElementById('room-width').value, 10) || 1200;
    const d = parseInt(document.getElementById('room-depth').value, 10) || 1200;
    const h = parseInt(document.getElementById('room-height').value, 10) || 600;
    document.getElementById('room-width-val').textContent = w;
    document.getElementById('room-depth-val').textContent = d;
    document.getElementById('room-height-val').textContent = h;
    document.getElementById('bp-dims').textContent = `${w} × ${d}`;

    const bp = document.getElementById('room-blueprint');
    bp.style.width = `${Math.round((w / 2000) * 180)}px`;
    bp.style.height = `${Math.round((d / 2000) * 115)}px`;

    document.getElementById('bp-n').classList.toggle('has-door', Boolean(parseInt(document.getElementById('room-door-n').value, 10)));
    document.getElementById('bp-e').classList.toggle('has-door', Boolean(parseInt(document.getElementById('room-door-e').value, 10)));
    document.getElementById('bp-s').classList.toggle('has-door', Boolean(parseInt(document.getElementById('room-door-s').value, 10)));
    document.getElementById('bp-w').classList.toggle('has-door', Boolean(parseInt(document.getElementById('room-door-w').value, 10)));
  }

  function populateDoorSelects(excludeRoomId) {
    const rooms = galleryData.rooms || [];
    ['room-door-n', 'room-door-e', 'room-door-s', 'room-door-w'].forEach((id) => {
      const sel = document.getElementById(id);
      sel.innerHTML = '<option value="0">Sem porta</option>';
      rooms.forEach((r) => {
        if (r.id === excludeRoomId) return;
        const opt = document.createElement('option');
        opt.value = r.id;
        opt.textContent = r.name;
        sel.appendChild(opt);
      });
    });
  }

  function loadRoomIntoForm(room) {
    editingRoomId = room ? room.id : 0;
    document.getElementById('room-id').value = editingRoomId || '';
    document.getElementById('room-name').value = room ? room.name : '';
    document.getElementById('room-desc').value = room ? (room.description || '') : '';
    document.getElementById('room-template').value = room ? (room.template || 'classic') : 'classic';
    document.getElementById('room-width').value = room ? room.width : 1200;
    document.getElementById('room-depth').value = room ? room.depth : 1200;
    document.getElementById('room-height').value = room ? room.height : 600;

    populateDoorSelects(editingRoomId);
    document.getElementById('room-door-n').value = room ? (room.door_north_room_id || 0) : 0;
    document.getElementById('room-door-e').value = room ? (room.door_east_room_id || 0) : 0;
    document.getElementById('room-door-s').value = room ? (room.door_south_room_id || 0) : 0;
    document.getElementById('room-door-w').value = room ? (room.door_west_room_id || 0) : 0;

    btnEnterRoom.classList.toggle('hidden', !editingRoomId);
    btnDeleteRoom.classList.toggle('hidden', !editingRoomId || (galleryData.rooms || []).length <= 1);
    renderRoomsChips();
    updateBlueprint();
  }

  function renderRoomsChips() {
    roomsListEl.innerHTML = '';
    (galleryData.rooms || []).forEach((r) => {
      const btn = document.createElement('button');
      btn.type = 'button';
      btn.className = `room-chip ${r.id === editingRoomId ? 'active' : ''}`;
      btn.textContent = `${r.name} (${r.width}×${r.depth})`;
      btn.addEventListener('click', () => loadRoomIntoForm(r));
      roomsListEl.appendChild(btn);
    });
  }

  btnAdminRooms.addEventListener('click', () => {
    loadRoomIntoForm(getCurrentRoom());
    roomsModal.classList.remove('hidden');
  });
  closeRoomsModal.addEventListener('click', () => roomsModal.classList.add('hidden'));
  btnNewRoom.addEventListener('click', () => loadRoomIntoForm(null));

  ['room-width', 'room-depth', 'room-height', 'room-door-n', 'room-door-e', 'room-door-s', 'room-door-w'].forEach((id) => {
    document.getElementById(id).addEventListener('input', updateBlueprint);
    document.getElementById(id).addEventListener('change', updateBlueprint);
  });

  btnEnterRoom.addEventListener('click', () => {
    if (!editingRoomId) return;
    roomsModal.classList.add('hidden');
    enterRoom(editingRoomId);
  });

  btnDeleteRoom.addEventListener('click', async () => {
    if (!editingRoomId || !confirm('Excluir esta sala? As obras dela serão movidas para outra sala.')) return;
    const resp = await fetch(`/praca/api/gallery/rooms?id=${editingRoomId}`, {
      method: 'DELETE',
      credentials: 'same-origin'
    });
    if (resp.ok) {
      await loadGallery();
      loadRoomIntoForm(getCurrentRoom());
    } else {
      alert(await resp.text());
    }
  });

  roomForm.addEventListener('submit', async (e) => {
    e.preventDefault();
    const idVal = parseInt(document.getElementById('room-id').value, 10) || 0;
    const payload = {
      id: idVal,
      name: document.getElementById('room-name').value,
      description: document.getElementById('room-desc').value,
      template: document.getElementById('room-template').value,
      width: parseInt(document.getElementById('room-width').value, 10),
      depth: parseInt(document.getElementById('room-depth').value, 10),
      height: parseInt(document.getElementById('room-height').value, 10),
      door_north_room_id: parseInt(document.getElementById('room-door-n').value, 10) || 0,
      door_east_room_id: parseInt(document.getElementById('room-door-e').value, 10) || 0,
      door_south_room_id: parseInt(document.getElementById('room-door-s').value, 10) || 0,
      door_west_room_id: parseInt(document.getElementById('room-door-w').value, 10) || 0
    };
    const resp = await fetch('/praca/api/gallery/rooms', {
      method: idVal ? 'PUT' : 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'same-origin',
      body: JSON.stringify(payload)
    });
    if (resp.ok) {
      const saved = await resp.json();
      await loadGallery();
      loadRoomIntoForm(saved);
      enterRoom(saved.id);
    } else {
      alert('Erro ao salvar sala.');
    }
  });

  // Admin: Modo Edição na Parede
  btnAdminEditMode.addEventListener('click', () => {
    if (!isAdmin) return;
    isEditMode = !isEditMode;
    document.body.classList.toggle('is-editing-walls', isEditMode);
    editModeBanner.classList.toggle('hidden', !isEditMode);
    btnAdminEditMode.classList.toggle('hud-btn-editing', isEditMode);
    btnAdminEditMode.textContent = isEditMode ? '✅ Concluir Edição' : '✋ Editar Paredes';
    if (isEditMode) stopTour();
  });

  // Admin: Configuração do espaço
  btnAdminConfig.addEventListener('click', () => {
    document.getElementById('cfg-title').value = galleryData.config.title || '';
    document.getElementById('cfg-subtitle').value = galleryData.config.subtitle || '';
    document.getElementById('cfg-template').value = galleryData.config.template || 'classic';
    configModal.classList.remove('hidden');
  });

  closeConfigModal.addEventListener('click', () => configModal.classList.add('hidden'));

  configForm.addEventListener('submit', async (e) => {
    e.preventDefault();
    const payload = {
      title: document.getElementById('cfg-title').value,
      subtitle: document.getElementById('cfg-subtitle').value,
      template: document.getElementById('cfg-template').value
    };
    const resp = await fetch('/praca/api/gallery/config', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'same-origin',
      body: JSON.stringify(payload)
    });
    if (resp.ok) {
      configModal.classList.add('hidden');
      await loadGallery();
    } else {
      alert('Erro ao salvar configurações do espaço.');
    }
  });

  // Admin: Adicionar / Editar obra
  document.querySelectorAll('.source-tab').forEach((tab) => {
    tab.addEventListener('click', () => {
      imageSourceMode = tab.dataset.src;
      document.querySelectorAll('.source-tab').forEach((t) => t.classList.toggle('active', t === tab));
      srcExistingBox.classList.toggle('hidden', imageSourceMode !== 'existing');
      srcUploadBox.classList.toggle('hidden', imageSourceMode !== 'upload');
    });
  });

  document.getElementById('art-posx').addEventListener('input', (e) => {
    document.getElementById('posx-val').textContent = e.target.value;
  });
  document.getElementById('art-posy').addEventListener('input', (e) => {
    document.getElementById('posy-val').textContent = e.target.value;
  });

  async function openArtworkForm(art = null) {
    await loadSlidesForAdmin();
    artworkForm.reset();
    artworkFormError.classList.add('hidden');

    const artRoomSelect = document.getElementById('art-room-id');
    artRoomSelect.innerHTML = '';
    (galleryData.rooms || []).forEach((r) => {
      const opt = document.createElement('option');
      opt.value = r.id;
      opt.textContent = r.name;
      artRoomSelect.appendChild(opt);
    });

    if (art) {
      artworkFormHeading.textContent = 'Editar Obra na Parede';
      document.getElementById('art-id').value = art.id;
      artRoomSelect.value = art.room_id || getCurrentRoom().id;
      artSlideSelect.value = art.slide_id;
      document.getElementById('art-wall').value = art.wall;
      document.getElementById('art-frame').value = art.frame_style;
      document.getElementById('art-posx').value = art.pos_x;
      document.getElementById('posx-val').textContent = Math.round(art.pos_x);
      document.getElementById('art-posy').value = art.pos_y;
      document.getElementById('posy-val').textContent = Math.round(art.pos_y);
      document.getElementById('art-scale').value = art.scale;
      document.getElementById('art-title').value = art.title || '';
      document.getElementById('art-artist').value = art.artist || '';
      document.getElementById('art-year').value = art.year || '';
      document.getElementById('art-technique').value = art.technique || '';
      document.getElementById('art-dimensions').value = art.dimensions || '';
      document.getElementById('art-description').value = art.description || '';
      document.getElementById('art-ext-link').value = art.external_link || '';
      document.getElementById('art-ext-label').value = art.external_link_label || '';
    } else {
      artworkFormHeading.textContent = 'Posicionar Obra na Parede';
      document.getElementById('art-id').value = '';
      artRoomSelect.value = getCurrentRoom().id;
      document.getElementById('posx-val').textContent = '50';
      document.getElementById('posy-val').textContent = '48';
    }

    artworkFormModal.classList.remove('hidden');
  }

  btnAdminAdd.addEventListener('click', () => openArtworkForm(null));
  closeArtworkFormModal.addEventListener('click', () => artworkFormModal.classList.add('hidden'));

  btnEditArtwork.addEventListener('click', () => {
    detailModal.classList.add('hidden');
    openArtworkForm(selectedArtwork);
  });

  btnDeleteArtwork.addEventListener('click', async () => {
    if (!selectedArtwork || !confirm('Remover esta obra da parede virtual?')) return;
    const resp = await fetch(`/praca/api/gallery/artworks?id=${selectedArtwork.id}`, {
      method: 'DELETE',
      credentials: 'same-origin'
    });
    if (resp.ok) {
      detailModal.classList.add('hidden');
      await loadGallery();
    } else {
      alert('Erro ao remover obra.');
    }
  });

  artworkForm.addEventListener('submit', async (e) => {
    e.preventDefault();
    artworkFormError.classList.add('hidden');

    let slideId = parseInt(artSlideSelect.value, 10) || 0;

    if (imageSourceMode === 'upload') {
      const file = artFileInput.files[0];
      if (!file) {
        artworkFormError.textContent = 'Selecione uma imagem para enviar.';
        artworkFormError.classList.remove('hidden');
        return;
      }
      const fd = new FormData();
      fd.append('images', file);
      fd.append('title', document.getElementById('art-title').value);
      const uploadResp = await fetch('/praca/api/slides', {
        method: 'POST',
        credentials: 'same-origin',
        body: fd
      });
      if (!uploadResp.ok) {
        artworkFormError.textContent = await uploadResp.text();
        artworkFormError.classList.remove('hidden');
        return;
      }
      const uploaded = await uploadResp.json();
      slideId = uploaded[0].id;
    }

    if (!slideId) {
      artworkFormError.textContent = 'Selecione ou envie uma imagem.';
      artworkFormError.classList.remove('hidden');
      return;
    }

    const idVal = parseInt(document.getElementById('art-id').value, 10) || 0;
    const payload = {
      id: idVal,
      room_id: parseInt(document.getElementById('art-room-id').value, 10) || getCurrentRoom().id,
      slide_id: slideId,
      wall: document.getElementById('art-wall').value,
      pos_x: parseFloat(document.getElementById('art-posx').value),
      pos_y: parseFloat(document.getElementById('art-posy').value),
      scale: document.getElementById('art-scale').value,
      width_px: selectedArtwork && idVal ? (selectedArtwork.width_px || 0) : 0,
      frame_style: document.getElementById('art-frame').value,
      title: document.getElementById('art-title').value,
      artist: document.getElementById('art-artist').value,
      year: document.getElementById('art-year').value,
      technique: document.getElementById('art-technique').value,
      dimensions: document.getElementById('art-dimensions').value,
      description: document.getElementById('art-description').value,
      external_link: document.getElementById('art-ext-link').value,
      external_link_label: document.getElementById('art-ext-label').value
    };

    const resp = await fetch('/praca/api/gallery/artworks', {
      method: idVal ? 'PUT' : 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'same-origin',
      body: JSON.stringify(payload)
    });

    if (!resp.ok) {
      artworkFormError.textContent = await resp.text();
      artworkFormError.classList.remove('hidden');
      return;
    }

    artworkFormModal.classList.add('hidden');
    await loadGallery();
    focusWall(payload.wall);
  });

  (async () => {
    updateCameraTransform(false);
    await checkAdmin();
    await loadGallery();
  })();
})();
