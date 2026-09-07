export class TextureLoader {
  constructor(limit = 6) { this.limit = Math.max(1, limit); this.cache = new Map(); }
  async load(key, url, retries = 1) {
    const cached = this.cache.get(key); if (cached) { cached.used = performance.now(); return cached.image; }
    const image = await new Promise((resolve, reject) => { const value = new Image(); value.onload = () => resolve(value); value.onerror = () => retries > 0 ? this.load(key, url, retries - 1).then(resolve, reject) : resolve(this.placeholder()); value.src = url; });
    this.cache.set(key, {image, used: performance.now()}); this.evict(); return image;
  }
  async preload(entries, count = 2) { for (const entry of entries.slice(0, count)) await this.load(entry.key, entry.url); }
  unloadExcept(keys) { for (const [key] of this.cache) if (!keys.has(key)) this.cache.delete(key); }
  clear() { this.cache.clear(); }
  placeholder() { const value = new Image(); value.src = 'data:image/svg+xml,<svg xmlns="http://www.w3.org/2000/svg" width="2" height="2"><rect width="2" height="2" fill="%233b3028"/></svg>'; return value; }
  evict() { while (this.cache.size > this.limit) { let oldest; for (const [key, value] of this.cache) if (!oldest || value.used < oldest.value.used) oldest = {key, value}; this.cache.delete(oldest.key); } }
}

export function museumSource(artwork, saveData = false) { return saveData ? (artwork.image.medium || artwork.image.museum) : (artwork.image.museum || artwork.image.medium); }
export function placementVisible(placement, camera, maxDistance = 24) { const dx = placement.x - camera.x; const dz = placement.z - camera.z; if (Math.hypot(dx, dz) > maxDistance) return false; if (!placement.normal || !camera.forward) return true; return placement.normal.x * camera.forward.x + placement.normal.z * camera.forward.z < 0; }
