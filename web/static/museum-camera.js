import { movePoint } from './museum-navigation.js';

// Rewrite-owned spatial camera. It delegates collision queries to the pure plan
// module and leaves input handling and WebGL drawing to their own modules.
export class MuseumCamera {
  constructor(position = { x: 0, y: 1.6, z: 0 }) {
    this.position = { ...position };
    this.yaw = 0;
    this.pitch = 0;
  }

  move(direction, distance = 0.65, plan) {
    const forward = { x: Math.sin(this.yaw), z: -Math.cos(this.yaw) };
    const sideways = { x: Math.cos(this.yaw), z: Math.sin(this.yaw) };
    let delta = { x: 0, z: 0 };
    if (direction === 'forward') delta = forward;
    if (direction === 'back') delta = { x: -forward.x, z: -forward.z };
    if (direction === 'left') delta = { x: -sideways.x, z: -sideways.z };
    if (direction === 'right') delta = sideways;
    const next = plan ? movePoint(plan, this.position, delta, distance) : { x: this.position.x + delta.x * distance, y: this.position.y, z: this.position.z + delta.z * distance };
    const moved = next.x !== this.position.x || next.z !== this.position.z;
    this.position = next;
    return moved;
  }

  look(deltaX, deltaY) {
    this.yaw += deltaX * 0.008;
    this.pitch = Math.max(-1.2, Math.min(1.2, this.pitch + deltaY * 0.008));
  }

  target() {
    const horizontal = Math.cos(this.pitch);
    return {
      x: this.position.x + Math.sin(this.yaw) * horizontal,
      y: this.position.y + Math.sin(this.pitch),
      z: this.position.z - Math.cos(this.yaw) * horizontal,
    };
  }
}
