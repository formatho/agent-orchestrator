import { nativeImage as h, Tray as f, Menu as u, app as i, BrowserWindow as r, ipcMain as s } from "electron";
import n from "path";
import { fileURLToPath as l } from "url";
const g = n.dirname(l(import.meta.url));
let t = null;
function b(e) {
  const d = n.join(g, "../public/icon.png"), m = h.createFromPath(d).resize({ width: 16, height: 16 });
  t = new f(m);
  const p = u.buildFromTemplate([
    {
      label: "Open Agent Orchestrator",
      click: () => {
        e && (e.show(), e.focus());
      }
    },
    {
      label: "Dashboard",
      click: () => {
        e && (e.show(), e.webContents.send("navigate", "/"));
      }
    },
    { type: "separator" },
    {
      label: "Agents",
      click: () => {
        e && (e.show(), e.webContents.send("navigate", "/agents"));
      }
    },
    {
      label: "TODOs",
      click: () => {
        e && (e.show(), e.webContents.send("navigate", "/todos"));
      }
    },
    { type: "separator" },
    {
      label: "Quit",
      click: () => {
        i.quit();
      }
    }
  ]);
  return t.setToolTip("Agent Orchestrator"), t.setContextMenu(p), t.on("click", () => {
    e && (e.isVisible() ? e.hide() : (e.show(), e.focus()));
  }), t;
}
const a = n.dirname(l(import.meta.url));
process.env.DIST_ELECTRON = n.join(a, "../");
process.env.DIST = n.join(a, "../dist");
process.env.VITE_PUBLIC = process.env.VITE_PUBLIC || n.join(a, "../public");
let o = null;
const w = n.join(a, "./preload.js");
function c() {
  return o = new r({
    width: 1400,
    height: 900,
    minWidth: 1e3,
    minHeight: 700,
    title: "Agent Orchestrator",
    icon: n.join(process.env.VITE_PUBLIC || "", "icon.png"),
    backgroundColor: "#0f0f0f",
    frame: !1,
    titleBarStyle: "hiddenInset",
    webPreferences: {
      preload: w,
      nodeIntegration: !1,
      contextIsolation: !0
    }
  }), o.webContents.on("did-finish-load", () => {
    o?.webContents.send("main-process-message", (/* @__PURE__ */ new Date()).toLocaleString());
  }), process.env.VITE_DEV_SERVER_URL ? (o.loadURL(process.env.VITE_DEV_SERVER_URL), o.webContents.openDevTools()) : o.loadFile(n.join(process.env.DIST || "", "index.html")), o;
}
i.on("window-all-closed", () => {
  process.platform !== "darwin" && (i.quit(), o = null);
});
i.on("activate", () => {
  r.getAllWindows().length === 0 && c();
});
s.on("window-minimize", () => {
  o?.minimize();
});
s.on("window-maximize", () => {
  o?.isMaximized() ? o.unmaximize() : o?.maximize();
});
s.on("window-close", () => {
  o?.close();
});
s.handle("window-is-maximized", () => o?.isMaximized() || !1);
i.whenReady().then(() => {
  c(), b(o);
});
