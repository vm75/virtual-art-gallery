// Pure spatial queries for the generated room/corridor plan. Input handling
// stays in museum.js and the camera delegates movement here.
const cameraRadius = .35;

function inRange(value, lower, upper) { return value >= lower && value <= upper; }

function doorwayFor(plan, room, side) {
  for (const connection of plan.connections || []) {
    if (side === 'east' && connection.from === room.id) return connection.from_doorway;
    if (side === 'west' && connection.to === room.id) return connection.to_doorway;
  }
  return null;
}

function insideRoom(plan, room, point, radius) {
  const z = room.position.z || 0, west = room.position.x - room.width / 2, east = room.position.x + room.width / 2;
  if (!inRange(point.z, z - room.depth / 2 + radius, z + room.depth / 2 - radius)) return false;
  if (inRange(point.x, west + radius, east - radius)) return true;
  for (const [side, edge] of [['west', west], ['east', east]]) {
    const door = doorwayFor(plan, room, side);
    if (door && inRange(point.x, side === 'west' ? edge : edge - radius, side === 'west' ? edge + radius : edge) && Math.abs(point.z - door.position.z) <= door.width / 2 - radius) return true;
  }
  return false;
}

function insideCorridor(corridor, point, radius) {
  const z = corridor.position.z || 0, start = corridor.position.x - corridor.length / 2, end = corridor.position.x + corridor.length / 2;
  return inRange(point.x, start, end) && inRange(point.z, z - corridor.width / 2 + radius, z + corridor.width / 2 - radius);
}

export function isTraversable(plan, point, radius = cameraRadius) {
  return (plan.rooms || []).some((room) => insideRoom(plan, room, point, radius)) || (plan.connections || []).some((connection) => connection.corridor && insideCorridor(connection.corridor, point, radius));
}

export function movePoint(plan, point, delta, distance = .65, radius = cameraRadius) {
  const next = { x: point.x + delta.x * distance, y: point.y, z: point.z + delta.z * distance };
  return isTraversable(plan, next, radius) ? next : { ...point };
}
