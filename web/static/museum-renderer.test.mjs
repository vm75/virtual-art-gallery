import assert from 'node:assert/strict';
import { MuseumCamera } from './museum-camera.js';
import { sceneGeometry, viewMatrix } from './museum-renderer.js';

function column(matrix, offset) {
  return [matrix[offset], matrix[offset + 4], matrix[offset + 8]];
}

function dot(a, b) {
  return a[0] * b[0] + a[1] * b[1] + a[2] * b[2];
}

for (const [yaw, pitch] of [[0, 0], [.7, .3], [-1.2, -.6], [2.4, 1.1]]) {
  const camera = new MuseumCamera({ x: 4, y: 1.6, z: -3 });
  camera.yaw = yaw;
  camera.pitch = pitch;
  const matrix = viewMatrix(camera);
  const basis = [column(matrix, 0), column(matrix, 1), column(matrix, 2)];

  assert.ok(matrix.every(Number.isFinite), `matrix is finite at yaw=${yaw}, pitch=${pitch}`);
  for (const vector of basis) assert.ok(Math.abs(Math.hypot(...vector) - 1) < 1e-9, 'basis vector has unit length');
  assert.ok(Math.abs(dot(basis[0], basis[1])) < 1e-9, 'right and up are orthogonal');
  assert.ok(Math.abs(dot(basis[0], basis[2])) < 1e-9, 'right and back are orthogonal');
  assert.ok(Math.abs(dot(basis[1], basis[2])) < 1e-9, 'up and back are orthogonal');
}

console.log('museum view matrix basis tests passed');

const room = { id: 'a', position: { x: 0, z: 0 }, width: 10, depth: 10, height: 5 };
const doorway = { position: { x: 5, y: 1.5, z: 0 }, width: 2.4, height: 3 };
const withoutConnector = sceneGeometry({ rooms: [room], connections: [] });
const withConnector = sceneGeometry({ rooms: [room], connections: [{ from: 'a', to: 'b', from_doorway: doorway, corridor: { position: { x: 7, z: 0 }, length: 4, width: 2.4, height: 3 } }] });
assert.ok(withConnector.length > withoutConnector.length, 'connector adds continuous floor and side-wall geometry');

console.log('museum connector geometry tests passed');
