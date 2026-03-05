import { contextBridge as d, ipcRenderer as e } from "electron";
d.exposeInMainWorld("electronAPI", {
  // Window controls
  minimizeWindow: () => e.send("window-minimize"),
  maximizeWindow: () => e.send("window-maximize"),
  closeWindow: () => e.send("window-close"),
  isMaximized: () => e.invoke("window-is-maximized"),
  // Platform info
  platform: process.platform,
  // IPC communication
  on: (i, n) => {
    e.on(i, (m, ...o) => n(...o));
  },
  removeListener: (i, n) => {
    e.removeListener(i, n);
  },
  send: (i, ...n) => {
    e.send(i, ...n);
  },
  invoke: (i, ...n) => e.invoke(i, ...n)
});
