import { writableBrains, buildAdd, statusText, settingsOf } from "./lib.js";

const $ = (id) => document.getElementById(id);
const say = (text, state) => {
  $("status").textContent = text;
  $("status").dataset.status = state;
};

// The page to propose: the active tab of the window, never an extension
// page (a test opens this popup as a tab of its own).
async function pageTab() {
  const tabs = await chrome.tabs.query({ currentWindow: true });
  const web = tabs.filter((t) => /^https?:/.test(t.url || ""));
  return web.find((t) => t.active) || web[web.length - 1] || null;
}

const settings = settingsOf(await chrome.storage.local.get(["gateway", "key"]));
if (!settings) {
  say("Set the gateway address and your key in the options first.", "no-settings");
  const a = document.createElement("a");
  a.href = chrome.runtime.getURL("options.html");
  a.target = "_blank";
  a.textContent = "Open the options";
  $("status").append(" ", a);
} else {
  const tab = await pageTab();
  if (!tab) {
    say("No web page to propose in this window.", "no-page");
  } else {
    $("title").value = tab.title || "";
    try {
      const [{ result }] = await chrome.scripting.executeScript({
        target: { tabId: tab.id },
        func: () => ({ selection: String(getSelection()), text: document.body ? document.body.innerText : "" }),
      });
      $("content").value = (result.selection || "").trim() || result.text || "";
      $("content").dataset.kind = (result.selection || "").trim() ? "selection" : "page";
    } catch (e) {
      say(`Cannot read this page: ${e.message}`, "no-text");
    }
    let brains = [];
    try {
      const res = await fetch(settings.gateway + "/v1/brains", { headers: { Authorization: "Bearer " + settings.key } });
      const body = await res.json().catch(() => null);
      if (!res.ok) {
        say(statusText(res.status, body), "brains-refused");
      } else {
        brains = writableBrains(body);
      }
    } catch (e) {
      say(statusText(0, null), "no-gateway");
    }
    for (const b of brains) {
      const o = document.createElement("option");
      o.value = b.brain;
      o.textContent = `${b.brain} (${b.access})`;
      o.dataset.access = b.access;
      $("brain").append(o);
    }
    if (brains.length === 0 && !$("status").dataset.status) {
      say("You may not write to any brain.", "no-brain");
    }
    $("propose").disabled = brains.length === 0;
    $("propose").addEventListener("click", async () => {
      $("propose").disabled = true;
      let body;
      try {
        body = buildAdd({
          brain: $("brain").value,
          title: $("title").value,
          content: $("content").value,
          url: tab.url,
          selection: $("content").dataset.kind === "selection" ? $("content").value : "",
        });
      } catch (e) {
        say(e.message, "invalid");
        $("propose").disabled = false;
        return;
      }
      const access = $("brain").selectedOptions[0]?.dataset.access;
      try {
        const res = await fetch(settings.gateway + "/v1/add", {
          method: "POST",
          headers: { Authorization: "Bearer " + settings.key, "Content-Type": "application/json" },
          body: JSON.stringify(body),
        });
        const answer = await res.json().catch(() => null);
        say(statusText(res.status, answer, body.brain, access), res.ok ? "sent" : "refused");
      } catch (e) {
        say(statusText(0, null), "no-gateway");
      }
      $("propose").disabled = false;
    });
  }
}
