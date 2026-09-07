import assert from 'node:assert/strict';
import { MuseumCamera } from './museum-camera.js';
import { pickArtwork } from './museum-picking.js';

const camera = new MuseumCamera({ x: 0, y: 1.6, z: 0 });
const placements = [
  { artwork_slug: 'left', position: { x: -2, y: 1.6, z: -5 }, normal: { x: 0, y: 0, z: 1 }, width: 1, height: 1.5 },
  { artwork_slug: 'right', position: { x: 2, y: 1.6, z: -5 }, normal: { x: 0, y: 0, z: 1 }, width: 1, height: 1.5 },
];
const viewport = { width: 800, height: 600 };

assert.equal(pickArtwork(placements, camera, { x: 200, y: 300 }, viewport)?.artwork_slug, 'left', 'left-side tap picks left artwork');
assert.equal(pickArtwork(placements, camera, { x: 600, y: 300 }, viewport)?.artwork_slug, 'right', 'right-side tap picks right artwork');
assert.equal(pickArtwork(placements, camera, { x: 400, y: 300 }, viewport), undefined, 'empty-space tap does not pick nearest artwork');
assert.equal(pickArtwork(placements, camera, { x: -1, y: 300 }, viewport), undefined, 'tap outside viewport is ignored');

const overlapping = [
  { artwork_slug: 'near', position: { x: 0, y: 1.6, z: -4 }, normal: { x: 0, y: 0, z: 1 }, width: 2, height: 2 },
  { artwork_slug: 'far', position: { x: 0, y: 1.6, z: -8 }, normal: { x: 0, y: 0, z: 1 }, width: 3, height: 3 },
];
assert.equal(pickArtwork(overlapping, camera, { x: 400, y: 300 }, viewport)?.artwork_slug, 'near', 'nearest intersected artwork wins');

console.log('museum artwork picking tests passed');
