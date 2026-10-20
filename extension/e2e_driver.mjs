// Drives the Thalamus extension in Chromium against a real gateway. Called
// by tests/e2e_extension.sh (it lives here so that node finds playwright) with: EXT (extension dir), SITE (local page URL),
// GATEWAY, KEY. Prints PASS/FAIL lines; exits 1 on a failure.
import { chromium } from "playwright";

const { EXT, SITE, GATEWAY, KEY } = process.env;
let fail = 0;
const ok = (s) => console.log("PASS  " + s);
const ko = (s, why) => { fail++; console.log("FAIL  " + s + ": " + String(why).slice(0, 400)); };

const context = await chromium.launchPersistentContext("", {
  channel: "chromium",
  args: [`--disable-extensions-except=${EXT}`, `--load-extension=${EXT}`],
});
try {
  let [sw] = context.serviceWorkers();
  if (!sw) sw = await context.waitForEvent("serviceworker");
  const id = sw.url().split("/")[2];
  const ext = (p) => `chrome-extension://${id}/${p}`;

  // The popup without settings says so.
  const page = await context.newPage();
  await page.goto(SITE);
  let popup = await context.newPage();
  await popup.goto(ext("popup.html"));
  await popup.waitForSelector('#status[data-status="no-settings"]');
  ok("popup without settings asks for the options");
  await popup.close();

  // The options, then a wrong key: the gateway refuses the listing.
  const options = await context.newPage();
  await options.goto(ext("options.html"));
  await options.fill("#gateway", GATEWAY + "/");
  await options.fill("#key", "not-a-key");
  await options.click("#save");
  await options.waitForFunction(() => document.querySelector("#saved").textContent === "Saved.");
  popup = await context.newPage();
  await popup.goto(ext("popup.html"));
  await popup.waitForSelector('#status[data-status="brains-refused"]');
  (await popup.textContent("#status")).includes("unknown to the gateway") ? ok("a wrong key is told") : ko("wrong key", await popup.textContent("#status"));
  await popup.close();

  // The right key: the page, whole, proposed into team/acme.
  await options.fill("#key", KEY);
  await options.click("#save");
  await options.waitForFunction(() => document.querySelector("#saved").textContent === "Saved.");
  await options.close();
  await page.bringToFront();
  popup = await context.newPage();
  await popup.goto(ext("popup.html"));
  await popup.waitForSelector("#brain option", { state: "attached" });
  const opts = await popup.$$eval("#brain option", (os) => os.map((o) => o.textContent));
  JSON.stringify(opts) === '["team/acme (write)"]' ? ok("the popup lists the one brain lfoletti may write to") : ko("brains", JSON.stringify(opts));
  (await popup.inputValue("#title")) === "Cutover plan" ? ok("the title comes from the page") : ko("title", await popup.inputValue("#title"));
  (await popup.inputValue("#content")).includes("MARMOT") ? ok("the content is the page text") : ko("content", await popup.inputValue("#content"));
  await popup.click("#propose");
  await popup.waitForSelector('#status[data-status="sent"]');
  const s1 = await popup.textContent("#status");
  s1 === "Proposed to team/acme: it waits for a curator." ? ok("the page is proposed: " + s1) : ko("status", s1);
  await popup.close();

  // A selection: only the chosen words go.
  await page.bringToFront();
  await page.evaluate(() => {
    const r = document.createRange();
    r.selectNodeContents(document.querySelector("#sel"));
    getSelection().removeAllRanges();
    getSelection().addRange(r);
  });
  popup = await context.newPage();
  await popup.goto(ext("popup.html"));
  await popup.waitForSelector("#brain option", { state: "attached" });
  const sel = await popup.inputValue("#content");
  sel === "The code word for the rehearsal is MARMOT." ? ok("the selection replaces the page text") : ko("selection", sel);
  await popup.click("#propose");
  await popup.waitForSelector('#status[data-status="sent"]');
  ok("the selection is proposed");
  await popup.close();
} catch (e) {
  ko("driver", e.stack || e);
} finally {
  await context.close();
}
process.exit(fail ? 1 : 0);
