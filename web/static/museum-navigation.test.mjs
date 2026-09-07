import assert from 'node:assert/strict';
import { MuseumCamera } from './museum-camera.js';
import { isTraversable, movePoint } from './museum-navigation.js';

const plan = {
  spawn_position: { x: 5, y: 1.6, z: 0 },
  rooms: [
    { id: 'first', position: { x: 5, z: 0 }, width: 10, depth: 10 },
    { id: 'second', position: { x: 19, z: 0 }, width: 10, depth: 10 },
  ],
  connections: [{
    from: 'first', to: 'second',
    from_doorway: { position: { x: 10, y: 1.5, z: 0 }, width: 2.4, height: 3 },
    to_doorway: { position: { x: 14, y: 1.5, z: 0 }, width: 2.4, height: 3 },
    corridor: { position: { x: 12, z: 0 }, length: 4, width: 2.4, height: 3 },
  }],
};

assert.ok(isTraversable(plan, plan.spawn_position), 'spawn is traversable');
assert.equal(isTraversable(plan, { x: 10.2, y: 1.6, z: 3 }), false, 'wall outside doorway is blocked');
assert.equal(isTraversable(plan, { x: 12, y: 1.6, z: 0 }), true, 'corridor is traversable');
assert.equal(isTraversable(plan, { x: 12, y: 1.6, z: 2 }), false, 'corridor side wall is blocked');

const blocked = movePoint(plan, { x: 9.65, y: 1.6, z: 3 }, { x: 1, z: 0 }, .65);
assert.deepEqual(blocked, { x: 9.65, y: 1.6, z: 3 }, 'movement cannot pass a wall');
let point = { x: 9.35, y: 1.6, z: 0 };
for (let step = 0; step < 10; step++) point = movePoint(plan, point, { x: 1, z: 0 }, .65);
assert.ok(point.x > 14.5 && isTraversable(plan, point), 'movement crosses doorway, corridor, and into next room');

const camera = new MuseumCamera(plan.spawn_position);
camera.yaw = Math.PI / 2;
assert.equal(camera.move('forward', .65, plan), true, 'camera uses shared collision primitive');
camera.position = { x: 9.65, y: 1.6, z: 3 };
assert.equal(camera.move('forward', .65, plan), false, 'camera reports a blocked wall');

console.log('museum navigation collision tests passed');
