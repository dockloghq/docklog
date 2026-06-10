/** Shared stroke icons (Lucide-style) used across DockLog UI */
export const icons = {
  dashboard: [
    { tag: "rect", attrs: { x: 3, y: 3, width: 7, height: 7 } },
    { tag: "rect", attrs: { x: 14, y: 3, width: 7, height: 7 } },
    { tag: "rect", attrs: { x: 14, y: 14, width: 7, height: 7 } },
    { tag: "rect", attrs: { x: 3, y: 14, width: 7, height: 7 } },
  ],
  containers: [
    {
      tag: "path",
      attrs: {
        d: "M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z",
      },
    },
  ],
  logs: [
    { tag: "line", attrs: { x1: 8, y1: 6, x2: 21, y2: 6 } },
    { tag: "line", attrs: { x1: 8, y1: 12, x2: 21, y2: 12 } },
    { tag: "line", attrs: { x1: 8, y1: 18, x2: 21, y2: 18 } },
    { tag: "line", attrs: { x1: 3, y1: 6, x2: 3.01, y2: 6 } },
    { tag: "line", attrs: { x1: 3, y1: 12, x2: 3.01, y2: 12 } },
    { tag: "line", attrs: { x1: 3, y1: 18, x2: 3.01, y2: 18 } },
  ],
  activity: [{ tag: "path", attrs: { d: "M22 12h-4l-3 9L9 3l-3 9H2" } }],
  users: [
    { tag: "path", attrs: { d: "M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" } },
    { tag: "circle", attrs: { cx: 9, cy: 7, r: 4 } },
    { tag: "path", attrs: { d: "M23 21v-2a4 4 0 0 0-3-3.87" } },
    { tag: "path", attrs: { d: "M16 3.13a4 4 0 0 1 0 7.75" } },
  ],
  shield: [
    { tag: "path", attrs: { d: "M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" } },
  ],
  settings: [
    { tag: "circle", attrs: { cx: 12, cy: 12, r: 3 } },
    {
      tag: "path",
      attrs: {
        d: "M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z",
      },
    },
  ],
  play: [{ tag: "polygon", attrs: { points: "5 3 19 12 5 21 5 3" } }],
  stop: [
    {
      tag: "rect",
      attrs: { x: 6, y: 6, width: 12, height: 12, fill: "currentColor", stroke: "currentColor" },
    },
  ],
  stopOutline: [{ tag: "rect", attrs: { x: 7, y: 7, width: 10, height: 10, rx: 1 } }],
  refresh: [
    { tag: "polyline", attrs: { points: "23 4 23 10 17 10" } },
    { tag: "path", attrs: { d: "M20.49 15a9 9 0 1 1-2.12-9.36L23 10" } },
  ],
  search: [
    { tag: "circle", attrs: { cx: 11, cy: 11, r: 8 } },
    { tag: "line", attrs: { x1: 21, y1: 21, x2: 16.65, y2: 16.65 } },
  ],
  server: [
    { tag: "rect", attrs: { x: 2, y: 2, width: 20, height: 8, rx: 2, ry: 2 } },
    { tag: "rect", attrs: { x: 2, y: 14, width: 20, height: 8, rx: 2, ry: 2 } },
    { tag: "line", attrs: { x1: 6, y1: 6, x2: 6.01, y2: 6 } },
    { tag: "line", attrs: { x1: 6, y1: 18, x2: 6.01, y2: 18 } },
  ],
  bell: [
    { tag: "path", attrs: { d: "M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9" } },
    { tag: "path", attrs: { d: "M13.73 21a2 2 0 0 1-3.46 0" } },
  ],
  checkCircle: [
    { tag: "path", attrs: { d: "M22 11.08V12a10 10 0 1 1-5.93-9.14" } },
    { tag: "polyline", attrs: { points: "22 4 12 14.01 9 11.01" } },
  ],
  info: [
    { tag: "circle", attrs: { cx: 12, cy: 12, r: 10 } },
    { tag: "line", attrs: { x1: 12, y1: 16, x2: 12, y2: 12 } },
    { tag: "line", attrs: { x1: 12, y1: 8, x2: 12.01, y2: 8 } },
  ],
  alert: [
    { tag: "path", attrs: { d: "M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z" } },
    { tag: "line", attrs: { x1: 12, y1: 9, x2: 12, y2: 13 } },
    { tag: "line", attrs: { x1: 12, y1: 17, x2: 12.01, y2: 17 } },
  ],
  box: [
    { tag: "path", attrs: { d: "M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z" } },
  ],
  user: [
    { tag: "path", attrs: { d: "M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2" } },
    { tag: "circle", attrs: { cx: 12, cy: 7, r: 4 } },
  ],
  plus: [
    { tag: "line", attrs: { x1: 12, y1: 5, x2: 12, y2: 19 } },
    { tag: "line", attrs: { x1: 5, y1: 12, x2: 19, y2: 12 } },
  ],
  lock: [
    { tag: "rect", attrs: { x: 3, y: 11, width: 18, height: 11, rx: 2, ry: 2 } },
    { tag: "path", attrs: { d: "M7 11V7a5 5 0 0 1 10 0v4" } },
  ],
  terminal: [
    { tag: "path", attrs: { d: "M12 19h8" } },
    { tag: "path", attrs: { d: "m4 17 6-6-6-6" } },
    { tag: "path", attrs: { d: "M3 7v10a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2V7a2 2 0 0 0-2-2H5a2 2 0 0 0-2 2z" } },
  ],
  logsBubble: [
    { tag: "path", attrs: { d: "M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z" } },
    { tag: "path", attrs: { d: "M8 9h8M8 13h5" } },
  ],
  copy: [
    { tag: "rect", attrs: { x: 9, y: 9, width: 13, height: 13, rx: 2 } },
    { tag: "path", attrs: { d: "M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" } },
  ],
  chevronLeft: [{ tag: "polyline", attrs: { points: "15 18 9 12 15 6" } }],
  trash: [
    { tag: "polyline", attrs: { points: "3 6 5 6 21 6" } },
    { tag: "path", attrs: { d: "M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" } },
  ],
  external: [
    { tag: "path", attrs: { d: "M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6" } },
    { tag: "polyline", attrs: { points: "15 3 21 3 21 9" } },
    { tag: "line", attrs: { x1: 10, y1: 14, x2: 21, y2: 3 } },
  ],
};
