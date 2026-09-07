function normalize(vector) {
  const length = Math.hypot(vector.x, vector.y, vector.z) || 1;
  return { x: vector.x / length, y: vector.y / length, z: vector.z / length };
}

function dot(a, b) { return a.x * b.x + a.y * b.y + a.z * b.z; }

function cameraBasis(camera) {
  const target = camera.target();
  const forward = normalize({ x: target.x - camera.position.x, y: target.y - camera.position.y, z: target.z - camera.position.z });
  const horizontal = Math.hypot(forward.x, forward.z) || 1;
  const right = { x: -forward.z / horizontal, y: 0, z: forward.x / horizontal };
  const up = normalize({
    x: right.y * forward.z - right.z * forward.y,
    y: right.z * forward.x - right.x * forward.z,
    z: right.x * forward.y - right.y * forward.x,
  });
  return { forward, right, up };
}

function rayForPoint(camera, point, viewport) {
  if (!viewport || viewport.width <= 0 || viewport.height <= 0) return null;
  if (point.x < 0 || point.y < 0 || point.x > viewport.width || point.y > viewport.height) return null;
  const { forward, right, up } = cameraBasis(camera);
  const ndcX = point.x / viewport.width * 2 - 1;
  const ndcY = 1 - point.y / viewport.height * 2;
  const scale = Math.tan(Math.PI / 6);
  const aspect = viewport.width / viewport.height;
  return normalize({
    x: forward.x + right.x * ndcX * scale * aspect + up.x * ndcY * scale,
    y: forward.y + right.y * ndcX * scale * aspect + up.y * ndcY * scale,
    z: forward.z + right.z * ndcX * scale * aspect + up.z * ndcY * scale,
  });
}

function positionOf(placement) { return placement.transform?.position || placement.position; }
function normalOf(placement) { return placement.transform?.normal || placement.normal; }

export function pickArtwork(placements, camera, point, viewport) {
  const ray = rayForPoint(camera, point, viewport);
  if (!ray) return undefined;
  let best;
  for (const placement of placements || []) {
    const position = positionOf(placement), normal = normalOf(placement);
    if (!position || !normal || !placement.width || !placement.height) continue;
    const denominator = dot(ray, normal);
    if (denominator >= -1e-6) continue;
    const toPlane = { x: position.x - camera.position.x, y: position.y - camera.position.y, z: position.z - camera.position.z };
    const distance = dot(toPlane, normal) / denominator;
    if (distance <= 0) continue;
    const hit = {
      x: camera.position.x + ray.x * distance,
      y: camera.position.y + ray.y * distance,
      z: camera.position.z + ray.z * distance,
    };
    const horizontal = Math.abs(normal.z) > .5 ? hit.x - position.x : hit.z - position.z;
    const vertical = hit.y - position.y;
    if (Math.abs(horizontal) > placement.width / 2 || Math.abs(vertical) > placement.height / 2) continue;
    if (!best || distance < best.distance) best = { placement, distance };
  }
  return best?.placement;
}
