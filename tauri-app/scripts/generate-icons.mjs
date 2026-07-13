// Generates the source app icon with no external deps (raw PNG encoder):
//   - app-icon.png : 1024px rounded-square app icon (indigo, white git-branch)
// Run `cargo tauri icon app-icon.png` afterwards to fan out the platform set.
import { deflateSync } from "node:zlib";
import { writeFileSync, mkdirSync } from "node:fs";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";

const root = resolve(dirname(fileURLToPath(import.meta.url)), "..");

// ---- PNG encoding -----------------------------------------------------------
const CRC = (() => {
  const t = new Uint32Array(256);
  for (let n = 0; n < 256; n++) {
    let c = n;
    for (let k = 0; k < 8; k++) c = c & 1 ? 0xedb88320 ^ (c >>> 1) : c >>> 1;
    t[n] = c >>> 0;
  }
  return t;
})();
function crc32(buf) {
  let c = 0xffffffff;
  for (let i = 0; i < buf.length; i++) c = CRC[(c ^ buf[i]) & 0xff] ^ (c >>> 8);
  return (c ^ 0xffffffff) >>> 0;
}
function chunk(type, data) {
  const len = Buffer.alloc(4);
  len.writeUInt32BE(data.length, 0);
  const t = Buffer.from(type, "ascii");
  const crc = Buffer.alloc(4);
  crc.writeUInt32BE(crc32(Buffer.concat([t, data])), 0);
  return Buffer.concat([len, t, data, crc]);
}
function encodePng(w, h, rgba) {
  const sig = Buffer.from([137, 80, 78, 71, 13, 10, 26, 10]);
  const ihdr = Buffer.alloc(13);
  ihdr.writeUInt32BE(w, 0);
  ihdr.writeUInt32BE(h, 4);
  ihdr[8] = 8; // bit depth
  ihdr[9] = 6; // color type RGBA
  const raw = Buffer.alloc((w * 4 + 1) * h);
  for (let y = 0; y < h; y++) {
    raw[y * (w * 4 + 1)] = 0; // filter: none
    rgba.copy(raw, y * (w * 4 + 1) + 1, y * w * 4, (y + 1) * w * 4);
  }
  return Buffer.concat([
    sig,
    chunk("IHDR", ihdr),
    chunk("IDAT", deflateSync(raw, { level: 9 })),
    chunk("IEND", Buffer.alloc(0)),
  ]);
}

// ---- vector helpers (signed-distance, anti-aliased) -------------------------
const clamp01 = (x) => Math.min(1, Math.max(0, x));
const cover = (d) => clamp01(0.5 - d); // 1px anti-aliased edge
function distSeg(px, py, ax, ay, bx, by) {
  const vx = bx - ax,
    vy = by - ay;
  const wx = px - ax,
    wy = py - ay;
  const t = clamp01((wx * vx + wy * vy) / (vx * vx + vy * vy));
  const cx = ax + t * vx,
    cy = ay + t * vy;
  return Math.hypot(px - cx, py - cy);
}
// git-branch glyph coverage at pixel (x,y) within a size×size box.
function branchAlpha(x, y, size) {
  const s = size,
    cx = (v) => v * s,
    node = 0.11 * s,
    stroke = 0.075 * s;
  // trunk: vertical line on the left with a node top and bottom
  const trunkX = cx(0.3);
  let d = distSeg(x, y, trunkX, cx(0.2), trunkX, cx(0.78)) - stroke;
  d = Math.min(d, Math.hypot(x - trunkX, y - cx(0.82)) - node); // bottom node
  // branch node top-right + curved connector back to the trunk
  const bx = cx(0.72),
    by = cx(0.2);
  d = Math.min(d, Math.hypot(x - bx, y - by) - node);
  const p0 = [bx, by + node],
    p1 = [bx, cx(0.5)],
    p2 = [trunkX, cx(0.5)];
  for (let i = 0; i < 24; i++) {
    const t = i / 23,
      it = 1 - t;
    const qx = it * it * p0[0] + 2 * it * t * p1[0] + t * t * p2[0];
    const qy = it * it * p0[1] + 2 * it * t * p1[1] + t * t * p2[1];
    d = Math.min(d, Math.hypot(x - qx, y - qy) - stroke);
  }
  return cover(d);
}

// ---- app icon: rounded indigo square + white glyph --------------------------
function appIcon(size) {
  const buf = Buffer.alloc(size * size * 4);
  const r = 0.22 * size,
    pad = 0.16 * size,
    inner = size - 2 * pad;
  for (let y = 0; y < size; y++) {
    for (let x = 0; x < size; x++) {
      // rounded-rect mask (distance to inset rect corners)
      const qx = Math.max(r - x, x - (size - r), 0);
      const qy = Math.max(r - y, y - (size - r), 0);
      const bg = cover(Math.hypot(qx, qy) - r);
      // vertical indigo→violet gradient
      const t = y / size;
      const cr = Math.round(79 + (124 - 79) * t);
      const cg = Math.round(70 + (58 - 70) * t);
      const cb = Math.round(229 + (237 - 229) * t);
      const g = branchAlpha((x - pad) , (y - pad), inner);
      const i = (y * size + x) * 4;
      // composite white glyph over gradient, all masked by rounded rect
      buf[i] = Math.round(cr * (1 - g) + 255 * g);
      buf[i + 1] = Math.round(cg * (1 - g) + 255 * g);
      buf[i + 2] = Math.round(cb * (1 - g) + 255 * g);
      buf[i + 3] = Math.round(255 * bg);
    }
  }
  return encodePng(size, size, buf);
}

mkdirSync(resolve(root, "src-tauri/icons"), { recursive: true });
writeFileSync(resolve(root, "app-icon.png"), appIcon(1024));
console.log("wrote app-icon.png");
