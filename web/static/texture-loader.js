// Image cache and room policy are kept independent of WebGL resource ownership.
export class TextureLoader {
  constructor(limit = 6) { this.limit = Math.max(1, limit); this.cache = new Map(); }
  async load(key, url, retries = 1) {
    const cached = this.cache.get(key); if (cached) { cached.used = performance.now(); return cached.image; }
    const image = await new Promise((resolve) => {
      const attempt = (remaining) => { const value = new Image(); value.onload = () => resolve(value); value.onerror = () => remaining ? attempt(remaining - 1) : resolve(this.placeholder()); value.src = url; };
      attempt(retries);
    });
    this.cache.set(key, { image, used: performance.now() }); this.evict(); return image;
  }
  unloadExcept(keys) { const removed = []; for (const [key] of this.cache) if (!keys.has(key)) { this.cache.delete(key); removed.push(key); } return removed; }
  clear() { const removed = [...this.cache.keys()]; this.cache.clear(); return removed; }
  placeholder() { const value = new Image(); value.src = 'data:image/svg+xml,<svg xmlns="http://www.w3.org/2000/svg" width="2" height="2"><rect width="2" height="2" fill="%233b3028"/></svg>'; return value; }
  evict() { while (this.cache.size > this.limit) { let oldest; for (const [key, value] of this.cache) if (!oldest || value.used < oldest.value.used) oldest = { key, value }; this.cache.delete(oldest.key); } }
}

export function textureCapability(connection = navigator.connection, memory = navigator.deviceMemory) { return { saveData: connection?.saveData === true, constrained: memory !== undefined && memory <= 2 }; }
export function museumSource(artwork, capability = textureCapability()) { return capability.saveData || capability.constrained ? (artwork.image.medium || artwork.image.museum) : (artwork.image.museum || artwork.image.medium); }

function positionOf(placement) { return placement.transform?.position || placement.position; }
function normalOf(placement) { return placement.transform?.normal || placement.normal; }
function forwardOf(camera) { const target = camera.target(); const x = target.x - camera.position.x, z = target.z - camera.position.z, length = Math.hypot(x, z) || 1; return { x: x / length, z: z / length }; }
export function placementVisible(placement, camera, maxDistance = 24) {
  const position = positionOf(placement), dx = position.x - camera.position.x, dz = position.z - camera.position.z;
  const distance = Math.hypot(dx, dz); if (distance > maxDistance) return false;
  const normal = normalOf(placement); if (normal && normal.x * -dx + normal.z * -dz < 0) return false;
  const forward = forwardOf(camera); return distance === 0 || (forward.x * dx + forward.z * dz) / distance > -.2;
}
function roomFor(plan, point) {
  const containing = (plan.rooms || []).find((room) => Math.abs(point.x - room.position.x) <= room.width / 2 && Math.abs(point.z - (room.position.z || 0)) <= room.depth / 2);
  if (containing) return containing;
  return [...(plan.rooms || [])].sort((a, b) => Math.hypot(a.position.x - point.x, (a.position.z || 0) - point.z) - Math.hypot(b.position.x - point.x, (b.position.z || 0) - point.z))[0];
}
function adjacentRooms(plan, roomID) { const rooms = new Set([roomID]); for (const link of plan.connections || []) { if (link.from === roomID) rooms.add(link.to); if (link.to === roomID) rooms.add(link.from); } return rooms; }
export function textureTargets(plan, camera, maxEntries = 6) {
  const current = roomFor(plan, camera.position); if (!current) return [];
  const rooms = adjacentRooms(plan, current.id), ranked = (plan.placements || []).filter((placement) => rooms.has(placement.room_id)).map((placement) => ({ placement, position: positionOf(placement) })).sort((a, b) => Math.hypot(a.position.x - camera.position.x, a.position.z - camera.position.z) - Math.hypot(b.position.x - camera.position.x, b.position.z - camera.position.z));
  return ranked.filter(({ placement }) => placement.room_id === current.id || placementVisible(placement, camera, 32)).slice(0, maxEntries).map(({ placement }) => placement.artwork_slug);
}

export class MuseumTextureLifecycle {
  constructor({ plan, artworks, renderer, redraw, limit = 6, capability = textureCapability() }) { this.plan = plan; this.artworks = artworks; this.renderer = renderer; this.redraw = redraw; this.loader = new TextureLoader(limit); this.limit = limit; this.capability = capability; this.active = new Set(); this.pending = new Set(); this.resident = new Set(); }
  update(camera) {
    const next = new Set(textureTargets(this.plan, camera, this.limit));
    this.loader.unloadExcept(next);
    for (const key of this.resident) if (!next.has(key)) { this.renderer.removeArtworkImage(key); this.resident.delete(key); }
    this.active = next;
    for (const key of next) {
      const artwork = this.artworks.get(key), source = artwork && museumSource(artwork, this.capability); if (!source) continue;
      if (this.resident.has(key) || this.pending.has(key)) continue;
      this.pending.add(key); this.loader.load(key, source).then((image) => { if (this.active.has(key)) { this.renderer.setArtworkImage(key, image); this.resident.add(key); this.redraw?.(); } }).catch(() => {}).finally(() => this.pending.delete(key));
    }
  }
  dispose() { this.loader.clear(); for (const key of this.resident) this.renderer.removeArtworkImage(key); this.resident.clear(); this.active.clear(); this.pending.clear(); }
}
