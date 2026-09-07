// Page wiring stays separate from the renderer, camera, and texture modules.
import { MuseumCamera } from './museum-camera.js';
import { MuseumRenderer } from './museum-renderer.js';
import { MuseumTextureLifecycle } from './texture-loader.js';

const canvas = document.querySelector('#museum-canvas');
const fallback = document.querySelector('#museum-fallback');

function detailsDialog() {
  const info = document.querySelector('#museum-info');
  info.querySelector('.info-close').addEventListener('click', () => info.close());
  return (item) => {
    info.querySelector('h2').textContent = item.name;
    info.querySelector('[data-field="date"]').textContent = item.date || '—';
    info.querySelector('[data-field="surface"]').textContent = item.surface || '—';
    info.querySelector('[data-field="medium"]').textContent = item.medium || '—';
    info.querySelector('[data-field="tags"]').textContent = (item.tags || []).join(', ');
    const link = info.querySelector('[data-field="link"]'); link.href = `/artwork/${encodeURIComponent(item.slug)}`;
    info.showModal();
  };
}

function installControls(camera, plan, redraw, inspect) {
  const status = document.createElement('p'); status.className = 'museum-status'; status.setAttribute('aria-live', 'polite'); canvas.after(status);
  const move = (direction) => { const moved = camera.move(direction, .65, plan); status.textContent = moved ? `Position ${camera.position.x.toFixed(1)}, ${camera.position.z.toFixed(1)}` : 'A wall blocks that direction.'; redraw(); };
  const keys = { ArrowUp: 'forward', w: 'forward', ArrowDown: 'back', s: 'back', ArrowLeft: 'left', a: 'left', ArrowRight: 'right', d: 'right' };
  canvas.addEventListener('keydown', (event) => { if (event.key === 'Escape') { canvas.blur(); return; } if (keys[event.key]) { event.preventDefault(); move(keys[event.key]); } });
  document.querySelectorAll('[data-move]').forEach((button) => button.addEventListener('click', () => { canvas.focus(); move(button.dataset.move); }));
  let last;
  canvas.addEventListener('pointerdown', (event) => { last = { x: event.clientX, y: event.clientY, moved: false }; canvas.setPointerCapture(event.pointerId); });
  canvas.addEventListener('pointermove', (event) => { if (!last || !canvas.hasPointerCapture(event.pointerId)) return; const dx = event.clientX - last.x, dy = event.clientY - last.y; last.moved ||= Math.hypot(dx, dy) > 6; camera.look(dx, dy); last.x = event.clientX; last.y = event.clientY; redraw(); });
  canvas.addEventListener('pointerup', (event) => { const tap = last && !last.moved; last = null; if (canvas.hasPointerCapture(event.pointerId)) canvas.releasePointerCapture(event.pointerId); if (tap) inspect(); });
  window.addEventListener('resize', redraw); window.addEventListener('orientationchange', redraw);
}

function artworkList(placements, bySlug, showInfo) {
  const list = document.createElement('ul'); list.className = 'museum-artworks';
  for (const placement of placements) {
    const item = bySlug.get(placement.artwork_slug); if (!item) continue;
    const entry = document.createElement('li'), button = document.createElement('button'); button.type = 'button'; button.textContent = item.name; button.addEventListener('click', () => showInfo(item)); entry.append(button); list.append(entry);
  }
  document.querySelector('.museum-page').append(list);
}

async function start() {
  let renderer;
  try { renderer = new MuseumRenderer(canvas); } catch { fallback.hidden = false; return; }
  let camera;
  let textures;
  const redraw = () => { renderer.render(camera); textures?.update(camera); };
  try {
    const sceneResponse = await fetch('/api/museum'); if (!sceneResponse.ok) throw new Error('museum is not published');
    const scene = await sceneResponse.json(), plan = scene.plan;
    const response = await fetch('/api/artworks'); if (!response.ok) throw new Error('artworks are unavailable');
    const artworks = await response.json(), bySlug = new Map(artworks.map((item) => [item.slug, item]));
    camera = new MuseumCamera(plan.spawn_position);
    // Generated placements begin on the room's north wall, facing inward.
    // Start looking toward that wall so the first artwork is visible on entry.
    camera.yaw = Math.PI;
    renderer.setScene(plan); fallback.hidden = true; redraw();
    const showInfo = detailsDialog(); installControls(camera, plan, redraw, () => { const placement = renderer.nearestArtwork(camera), item = placement && bySlug.get(placement.artwork_slug); if (item) showInfo(item); }); artworkList(plan.placements || [], bySlug, showInfo);
    textures = new MuseumTextureLifecycle({ plan, artworks: bySlug, renderer, redraw }); textures.update(camera);
    window.addEventListener('pagehide', () => { textures.dispose(); renderer.dispose(); }, { once: true });
  } catch {
    renderer.dispose(); fallback.hidden = false;
  }
}

start();
