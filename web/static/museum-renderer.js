import { placementVisible } from './texture-loader.js';

// Rewrite-owned WebGL scene renderer. It consumes only the public layout plan
// and artwork metadata; it has no knowledge of application routes or rules.
const vertexSource = `attribute vec3 position; attribute vec3 color; uniform mat4 projection; uniform mat4 view; varying vec3 shaded; void main() { shaded = color; gl_Position = projection * view * vec4(position, 1.0); }`;
const fragmentSource = `precision mediump float; varying vec3 shaded; void main() { gl_FragColor = vec4(shaded, 1.0); }`;
const artVertexSource = `attribute vec3 position; attribute vec2 uv; uniform mat4 projection; uniform mat4 view; varying vec2 texcoord; void main() { texcoord = uv; gl_Position = projection * view * vec4(position, 1.0); }`;
const artFragmentSource = `precision mediump float; varying vec2 texcoord; uniform sampler2D artwork; uniform vec3 fallback; uniform bool textured; void main() { vec4 image = texture2D(artwork, texcoord); gl_FragColor = textured ? image : vec4(fallback, 1.0); }`;

function compile(gl, type, source) {
  const shader = gl.createShader(type);
  gl.shaderSource(shader, source);
  gl.compileShader(shader);
  if (!gl.getShaderParameter(shader, gl.COMPILE_STATUS)) throw new Error(gl.getShaderInfoLog(shader));
  return shader;
}

function program(gl, vertex, fragment) {
  const value = gl.createProgram();
  gl.attachShader(value, compile(gl, gl.VERTEX_SHADER, vertex));
  gl.attachShader(value, compile(gl, gl.FRAGMENT_SHADER, fragment));
  gl.linkProgram(value);
  if (!gl.getProgramParameter(value, gl.LINK_STATUS)) throw new Error(gl.getProgramInfoLog(value));
  return value;
}

function identity() { return [1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1]; }
function perspective(fovy, aspect, near, far) {
  const f = 1 / Math.tan(fovy / 2), range = 1 / (near - far);
  return [f / aspect, 0, 0, 0, 0, f, 0, 0, 0, 0, (far + near) * range, -1, 0, 0, 2 * far * near * range, 0];
}
export function viewMatrix(camera) {
  const target = camera.target(), eye = camera.position;
  let zx = eye.x - target.x, zy = eye.y - target.y, zz = eye.z - target.z;
  const zLength = Math.hypot(zx, zy, zz) || 1; zx /= zLength; zy /= zLength; zz /= zLength;
  // The right vector is world-up × view-back. Using view-back's y component
  // here collapses at the normal horizontal starting orientation.
  let xx = zz, xy = 0, xz = -zx;
  const xLength = Math.hypot(xx, xy, xz) || 1; xx /= xLength; xy /= xLength; xz /= xLength;
  const yx = xy * zz - xz * zy, yy = xz * zx - xx * zz, yz = xx * zy - xy * zx;
  return [xx, yx, zx, 0, xy, yy, zy, 0, xz, yz, zz, 0, -(xx * eye.x + xy * eye.y + xz * eye.z), -(yx * eye.x + yy * eye.y + yz * eye.z), -(zx * eye.x + zy * eye.y + zz * eye.z), 1];
}

function pushQuad(target, a, b, c, d, color) {
  for (const point of [a, b, c, a, c, d]) target.push(...point, ...color);
}
function pushBox(target, center, size, color) {
  const [x, y, z] = [center.x, center.y, center.z], [w, h, d] = [size.x / 2, size.y / 2, size.z / 2];
  const p = (dx, dy, dz) => [x + dx, y + dy, z + dz];
  pushQuad(target, p(-w,-h,d), p(w,-h,d), p(w,h,d), p(-w,h,d), color);
  pushQuad(target, p(w,-h,-d), p(-w,-h,-d), p(-w,h,-d), p(w,h,-d), color);
  pushQuad(target, p(-w,-h,-d), p(-w,-h,d), p(-w,h,d), p(-w,h,-d), color);
  pushQuad(target, p(w,-h,d), p(w,-h,-d), p(w,h,-d), p(w,h,d), color);
  pushQuad(target, p(-w,h,d), p(w,h,d), p(w,h,-d), p(-w,h,-d), color);
  pushQuad(target, p(-w,-h,-d), p(w,-h,-d), p(w,-h,d), p(-w,-h,d), color);
}

function doorwayFor(plan, roomID, side) {
  for (const connection of plan.connections || []) {
    if (connection.from === roomID && side === 'east') return connection.from_doorway;
    if (connection.to === roomID && side === 'west') return connection.to_doorway;
  }
  return null;
}
export function sceneGeometry(plan) {
  const data = [];
  for (const room of plan.rooms || []) {
    const x = room.position.x, z = room.position.z || 0, halfWidth = room.width / 2, halfDepth = room.depth / 2;
    pushBox(data, { x, y: -0.12, z }, { x: room.width, y: .24, z: room.depth }, [.30, .27, .22]);
    pushBox(data, { x, y: room.height / 2, z: z - halfDepth }, { x: room.width, y: room.height, z: .18 }, [.60, .55, .47]);
    pushBox(data, { x, y: room.height / 2, z: z + halfDepth }, { x: room.width, y: room.height, z: .18 }, [.64, .59, .50]);
    for (const side of ['west', 'east']) {
      const sign = side === 'east' ? 1 : -1, wallX = x + sign * halfWidth, door = doorwayFor(plan, room.id, side);
      if (!door) { pushBox(data, { x: wallX, y: room.height / 2, z }, { x: .18, y: room.height, z: room.depth }, [.56, .51, .43]); continue; }
      const opening = Math.min(door.width, room.depth - .4), segment = (room.depth - opening) / 2;
      pushBox(data, { x: wallX, y: room.height / 2, z: z - (opening + segment) / 2 }, { x: .18, y: room.height, z: segment }, [.56, .51, .43]);
      pushBox(data, { x: wallX, y: room.height / 2, z: z + (opening + segment) / 2 }, { x: .18, y: room.height, z: segment }, [.56, .51, .43]);
      pushBox(data, { x: wallX, y: (room.height + door.height) / 2, z }, { x: .18, y: room.height - door.height, z: opening }, [.56, .51, .43]);
    }
  }
  return new Float32Array(data);
}
function artVertices(placement) {
  const position = placement.transform?.position || placement.position, normal = placement.transform?.normal || placement.normal;
  const width = placement.width / 2, height = placement.height / 2;
  const horizontal = Math.abs(normal.z) > .5 ? { x: 1, z: 0 } : { x: 0, z: 1 };
  const offset = { x: normal.x * .03, z: normal.z * .03 };
  const point = (sx, sy) => [position.x + horizontal.x * sx + offset.x, position.y + sy, position.z + horizontal.z * sx + offset.z];
  return new Float32Array([...point(-width,-height), 0,1, ...point(width,-height), 1,1, ...point(width,height), 1,0, ...point(-width,-height), 0,1, ...point(width,height), 1,0, ...point(-width,height), 0,0]);
}

export class MuseumRenderer {
  constructor(canvas) {
    this.canvas = canvas; this.gl = canvas.getContext('webgl', { antialias: true });
    if (!this.gl) throw new Error('WebGL is unavailable');
    const gl = this.gl;
    this.architectureProgram = program(gl, vertexSource, fragmentSource);
    this.artProgram = program(gl, artVertexSource, artFragmentSource);
    this.architectureBuffer = gl.createBuffer(); this.artBuffer = gl.createBuffer(); this.textures = new Map(); this.artworks = [];
    gl.enable(gl.DEPTH_TEST); gl.enable(gl.CULL_FACE); gl.clearColor(.07, .06, .05, 1);
  }
  setScene(plan) {
    const gl = this.gl; this.plan = plan; this.artworks = plan.placements || [];
    gl.bindBuffer(gl.ARRAY_BUFFER, this.architectureBuffer); gl.bufferData(gl.ARRAY_BUFFER, sceneGeometry(plan), gl.STATIC_DRAW);
    this.canvas.dataset.roomsRendered = String((plan.rooms || []).length); this.canvas.dataset.artworksRendered = String(this.artworks.length);
  }
  setArtworkImage(slug, image) {
    const gl = this.gl, old = this.textures.get(slug); if (old) gl.deleteTexture(old);
    const texture = gl.createTexture(); gl.bindTexture(gl.TEXTURE_2D, texture);
    gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE); gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE);
    gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR); gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR);
    gl.texImage2D(gl.TEXTURE_2D, 0, gl.RGBA, gl.RGBA, gl.UNSIGNED_BYTE, image); this.textures.set(slug, texture);
  }
  removeArtworkImage(slug) { const texture = this.textures.get(slug); if (texture) this.gl.deleteTexture(texture); this.textures.delete(slug); }
  nearestArtwork(camera) {
    return this.artworks.filter((artwork) => placementVisible(artwork, camera, 12)).map((artwork) => {
      const position = artwork.transform?.position || artwork.position;
      return { artwork, distance: Math.hypot(position.x - camera.position.x, position.z - camera.position.z) };
    }).sort((a, b) => a.distance - b.distance)[0]?.artwork;
  }
  render(camera) {
    const gl = this.gl, width = this.canvas.clientWidth * devicePixelRatio, height = this.canvas.clientHeight * devicePixelRatio;
    if (this.canvas.width !== width || this.canvas.height !== height) { this.canvas.width = width; this.canvas.height = height; }
    gl.viewport(0, 0, width, height); gl.clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT);
    const projection = perspective(Math.PI / 3, Math.max(1, width) / Math.max(1, height), .1, 250), view = viewMatrix(camera);
    gl.useProgram(this.architectureProgram); gl.bindBuffer(gl.ARRAY_BUFFER, this.architectureBuffer);
    const p = gl.getAttribLocation(this.architectureProgram, 'position'), c = gl.getAttribLocation(this.architectureProgram, 'color');
    gl.enableVertexAttribArray(p); gl.vertexAttribPointer(p, 3, gl.FLOAT, false, 24, 0); gl.enableVertexAttribArray(c); gl.vertexAttribPointer(c, 3, gl.FLOAT, false, 24, 12);
    gl.uniformMatrix4fv(gl.getUniformLocation(this.architectureProgram, 'projection'), false, projection); gl.uniformMatrix4fv(gl.getUniformLocation(this.architectureProgram, 'view'), false, view); gl.drawArrays(gl.TRIANGLES, 0, this.gl.getBufferParameter(gl.ARRAY_BUFFER, gl.BUFFER_SIZE) / 24);
    gl.disable(gl.CULL_FACE); gl.useProgram(this.artProgram); gl.uniformMatrix4fv(gl.getUniformLocation(this.artProgram, 'projection'), false, projection); gl.uniformMatrix4fv(gl.getUniformLocation(this.artProgram, 'view'), false, view);
    const ap = gl.getAttribLocation(this.artProgram, 'position'), uv = gl.getAttribLocation(this.artProgram, 'uv'); gl.uniform1i(gl.getUniformLocation(this.artProgram, 'artwork'), 0);
    for (const artwork of this.artworks) {
      if (!placementVisible(artwork, camera)) continue;
      gl.bindBuffer(gl.ARRAY_BUFFER, this.artBuffer); gl.bufferData(gl.ARRAY_BUFFER, artVertices(artwork), gl.STREAM_DRAW); gl.enableVertexAttribArray(ap); gl.vertexAttribPointer(ap, 3, gl.FLOAT, false, 20, 0); gl.enableVertexAttribArray(uv); gl.vertexAttribPointer(uv, 2, gl.FLOAT, false, 20, 12);
      const texture = this.textures.get(artwork.artwork_slug); gl.uniform1i(gl.getUniformLocation(this.artProgram, 'textured'), texture ? 1 : 0); gl.uniform3f(gl.getUniformLocation(this.artProgram, 'fallback'), .48, .22, .14);
      if (texture) { gl.activeTexture(gl.TEXTURE0); gl.bindTexture(gl.TEXTURE_2D, texture); } gl.drawArrays(gl.TRIANGLES, 0, 6);
    }
    gl.enable(gl.CULL_FACE);
  }
  dispose() { const gl = this.gl; for (const texture of this.textures.values()) gl.deleteTexture(texture); gl.deleteBuffer(this.architectureBuffer); gl.deleteBuffer(this.artBuffer); gl.deleteProgram(this.architectureProgram); gl.deleteProgram(this.artProgram); }
}
