// Rewrite-owned spatial camera. It deliberately knows only world coordinates,
// leaving plan interpretation and WebGL drawing to their respective modules.
export class MuseumCamera {
  constructor(position = { x: 0, y: 1.6, z: 0 }) {
    this.position = { ...position };
    this.yaw = 0;
    this.pitch = 0;
  }

  move(direction, distance = 0.65) {
    const forward = { x: Math.sin(this.yaw), z: -Math.cos(this.yaw) };
    const sideways = { x: Math.cos(this.yaw), z: Math.sin(this.yaw) };
    if (direction === 'forward') { this.position.x += forward.x * distance; this.position.z += forward.z * distance; }
    if (direction === 'back') { this.position.x -= forward.x * distance; this.position.z -= forward.z * distance; }
    if (direction === 'left') { this.position.x -= sideways.x * distance; this.position.z -= sideways.z * distance; }
    if (direction === 'right') { this.position.x += sideways.x * distance; this.position.z += sideways.z * distance; }
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
