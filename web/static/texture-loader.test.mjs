import assert from 'node:assert/strict';
import { MuseumCamera } from './museum-camera.js';
import { MuseumTextureLifecycle } from './texture-loader.js';

globalThis.Image = class {
  set src(value) { this.value = value; queueMicrotask(() => this.onload?.()); }
};

const rooms = [
  { id: 'one', position: { x: 5, z: 0 }, width: 10, depth: 10 },
  { id: 'two', position: { x: 19, z: 0 }, width: 10, depth: 10 },
  { id: 'three', position: { x: 33, z: 0 }, width: 10, depth: 10 },
];
const plan = {
  rooms,
  connections: [
    { from: 'one', to: 'two' },
    { from: 'two', to: 'three' },
  ],
  placements: rooms.map((room, index) => ({ artwork_slug: `work-${index + 1}`, room_id: room.id, transform: { position: { x: room.position.x, y: 2, z: 1 }, normal: { x: 0, z: -1 } } })),
};
const artworks = new Map(plan.placements.map((placement) => [placement.artwork_slug, { image: { museum: `/${placement.artwork_slug}.jpg` } }]));
const calls = { set: [], remove: [] };
const lifecycle = new MuseumTextureLifecycle({ plan, artworks, renderer: { setArtworkImage: (key) => calls.set.push(key), removeArtworkImage: (key) => calls.remove.push(key) }, limit: 2, capability: {} });
const firstCamera = new MuseumCamera({ x: 5, y: 1.6, z: 0 });
firstCamera.yaw = Math.PI;
lifecycle.update(firstCamera);
await new Promise((resolve) => setTimeout(resolve, 0));
const initialUploads = calls.set.length;
assert.ok(initialUploads > 0 && initialUploads <= 2, 'initial GPU residency is bounded');
lifecycle.update(firstCamera);
await new Promise((resolve) => setTimeout(resolve, 0));
assert.equal(calls.set.length, initialUploads, 'same-room navigation does not re-upload resident textures');

const thirdCamera = new MuseumCamera({ x: 33, y: 1.6, z: 0 });
thirdCamera.yaw = Math.PI;
lifecycle.update(thirdCamera);
await new Promise((resolve) => setTimeout(resolve, 0));
assert.ok(calls.remove.length > 0, 'textures outside the room policy are released');
assert.ok(calls.set.length <= initialUploads + 2, 'new-room GPU residency remains bounded');
lifecycle.dispose();
assert.equal(lifecycle.resident.size, 0, 'dispose releases GPU residency');

console.log('museum texture lifecycle tests passed');
