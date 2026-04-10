(async function () {
  const resp = await fetch("/api/config");
  const config = await resp.json();

  const grid = document.getElementById("grid");
  const { direction, widgets } = config.layout;
  const count = widgets.length;

  if (count === 0) return;

  const cols = Math.ceil(Math.sqrt(count));
  const rows = Math.ceil(count / cols);

  grid.style.gridTemplateColumns = `repeat(${cols}, 1fr)`;
  grid.style.gridTemplateRows = `repeat(${rows}, 1fr)`;
  grid.style.gridAutoFlow = direction;

  const loadedWidgets = new Set();
  const isOdd = count % 2 !== 0;

  for (let i = 0; i < count; i++) {
    const widget = widgets[i];
    const mod = widget.module;
    const moduleConfig = config.modules[mod];

    if (!moduleConfig) continue;

    const cell = document.createElement("div");
    cell.className = "grid-cell";
    cell.dataset.index = i;

    // Last widget in an odd count spans the remainder
    if (isOdd && i === count - 1) {
      if (direction === "row") {
        cell.style.gridColumn = `span ${cols * rows - count + 1}`;
      } else {
        cell.style.gridRow = `span ${cols * rows - count + 1}`;
      }
    }

    if (!loadedWidgets.has(mod)) {
      await loadWidget(mod);
      loadedWidgets.add(mod);
    }

    const el = document.createElement(`resilver-${mod}`);
    el.dataset.config = JSON.stringify(moduleConfig.config || {});
    cell.appendChild(el);

    grid.appendChild(cell);
  }
})();

function loadWidget(name) {
  return new Promise((resolve, reject) => {
    const script = document.createElement("script");
    script.src = `/js/widgets/${name}/${name}.js`;
    script.onload = resolve;
    script.onerror = () => reject(new Error(`Failed to load widget: ${name}`));
    document.head.appendChild(script);
  });
}
