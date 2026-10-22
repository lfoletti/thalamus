import test from "node:test";
import assert from "node:assert/strict";
import { readFileSync } from "node:fs";

const dir = new URL("./", import.meta.url);
const manifest = JSON.parse(readFileSync(new URL("manifest.json", dir)));

// The size of a PNG, read from its IHDR chunk; an error if the file is not
// a PNG.
function pngSize(path) {
  const b = readFileSync(new URL(path, dir));
  assert.equal(b.subarray(0, 8).toString("hex"), "89504e470d0a1a0a", `${path} is not a PNG`);
  assert.equal(b.subarray(12, 16).toString(), "IHDR", `${path} has no IHDR first`);
  return [b.readUInt32BE(16), b.readUInt32BE(20)];
}

// Every icon the manifest names, for the extension and for its button,
// exists, is a PNG and has the size its key says. Chrome shows a grey
// letter instead of an icon that is missing or wrong.
test("the manifest names an icon of each size, and each file is that size", () => {
  for (const icons of [manifest.icons, manifest.action.default_icon]) {
    assert.deepEqual(Object.keys(icons || {}).sort(), ["128", "16", "32", "48"]);
    for (const [size, path] of Object.entries(icons)) {
      assert.deepEqual(pngSize(path), [Number(size), Number(size)], path);
    }
  }
});
