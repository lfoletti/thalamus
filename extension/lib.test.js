import test from "node:test";
import assert from "node:assert/strict";
import { writableBrains, documentID, buildAdd, statusText, settingsOf } from "./lib.js";

// Only the brains the employee may write to, with a key, sorted by name.
test("writableBrains keeps the brains one may write to, with a key, sorted", () => {
  const got = writableBrains([
    { brain: "team/zeta", access: "write", key: true },
    { brain: "org/secrets", access: "read", key: true },
    { brain: "team/acme", access: "curate", key: true },
    { brain: "team/nokey", access: "write", key: false },
  ]);
  assert.deepEqual(got.map((b) => b.brain), ["team/acme", "team/zeta"]);
  assert.deepEqual(writableBrains([]), []);
});

// The document id follows the page: stable, without the fragment, and
// distinct for two pages.
test("documentID is stable per page and ignores the fragment", () => {
  const a = documentID("https://intranet/spec?v=2#top");
  assert.equal(a, documentID("https://intranet/spec?v=2"));
  assert.notEqual(a, documentID("https://intranet/spec?v=3"));
  assert.match(a, /^web-[0-9a-f]{8}$/);
});

// The body of POST /v1/add: the selection wins over the page text, the
// whitespace is collapsed, the title and the content are bounded, and the
// metadata says where it comes from. Nothing to propose is an error.
test("buildAdd prefers the selection and bounds the fields", () => {
  const page = { brain: "team/acme", title: "  Spec  ", content: "The  page\n\n text", url: "https://intranet/spec#x", selection: "" };
  let got = buildAdd(page);
  assert.equal(got.brain, "team/acme");
  assert.equal(got.title, "Spec");
  assert.equal(got.content, "The page text");
  assert.equal(got.source_url, "https://intranet/spec#x");
  assert.equal(got.document_id, documentID(page.url));
  assert.deepEqual(got.metadata, { via: "chrome-extension", kind: "page" });
  got = buildAdd({ ...page, selection: " the  chosen words " });
  assert.equal(got.content, "the chosen words");
  assert.equal(got.metadata.kind, "selection");
  assert.equal(buildAdd({ ...page, title: "t".repeat(300) }).title.length, 200);
  assert.equal(buildAdd({ ...page, content: "c".repeat(200000) }).content.length, 100000);
  assert.throws(() => buildAdd({ ...page, content: "  ", selection: "" }), /nothing to propose/);
  assert.throws(() => buildAdd({ ...page, brain: "" }), /choose a brain/);
});

// One sentence per outcome, by status, with the brain and the access.
test("statusText says what happened, by status", () => {
  assert.equal(statusText(200, {}, "team/acme", "write"), "Proposed to team/acme: it waits for a curator.");
  assert.equal(statusText(200, {}, "team/acme", "curate"), "Added to team/acme: you curate it, so it is served at once.");
  assert.equal(statusText(401, { error: "refused: unknown key" }), "Your key is unknown to the gateway: check the options.");
  assert.equal(statusText(403, { error: "th_add: x may not write to team/acme" }), "You may not write to this brain.");
  assert.equal(statusText(404, { error: "th_add: unknown brain" }), "Unknown brain: th_add: unknown brain");
  assert.equal(statusText(400, { error: "th_add: content is required" }), "Refused: th_add: content is required");
  assert.equal(statusText(502, { error: "th_add: brain says no" }), "The brain refused: th_add: brain says no");
  assert.equal(statusText(0, null), "The gateway did not answer: check the address in the options.");
  assert.equal(statusText(500, { error: "key file of team/acme: broken" }), "Error 500: key file of team/acme: broken");
  assert.equal(statusText(500, "not json"), "Error 500: not json");
});

// The settings: both fields trimmed and required, the gateway without a
// trailing slash.
test("settingsOf trims, requires both fields and drops the trailing slash", () => {
  assert.deepEqual(settingsOf({ gateway: " http://gw:8787/ ", key: " k " }), { gateway: "http://gw:8787", key: "k" });
  assert.equal(settingsOf({ gateway: "", key: "k" }), null);
  assert.equal(settingsOf({ gateway: "http://gw", key: " " }), null);
  assert.equal(settingsOf({}), null);
});
