import { settingsOf } from "./lib.js";

const $ = (id) => document.getElementById(id);
const stored = await chrome.storage.local.get(["gateway", "key"]);
$("gateway").value = stored.gateway || "";
$("key").value = stored.key || "";
$("save").addEventListener("click", async () => {
  const s = settingsOf({ gateway: $("gateway").value, key: $("key").value });
  if (!s) {
    $("saved").textContent = "Both fields are required.";
    return;
  }
  await chrome.storage.local.set(s);
  $("saved").textContent = "Saved.";
});
