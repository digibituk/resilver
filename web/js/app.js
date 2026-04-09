(async function () {
  const resp = await fetch("/api/config");
  const config = await resp.json();

  const grid = document.getElementById("grid");
  const { columns, rows, positions } = config.layout;

  grid.style.gridTemplateColumns = `repeat(${columns}, 1fr)`;
  grid.style.gridTemplateRows = `repeat(${rows}, 1fr)`;

  const positionNames = [
    "top-left", "top-center", "top-right",
    "middle-left", "middle-center", "middle-right",
    "bottom-left", "bottom-center", "bottom-right",
  ];

  const loadedWidgets = new Set();

  for (const name of positionNames) {
    const cell = document.createElement("div");
    cell.className = "grid-cell";
    cell.dataset.position = name;

    const modules = positions[name] || [];
    for (const mod of modules) {
      const moduleConfig = config.modules[mod];
      if (!moduleConfig || !moduleConfig.enabled) continue;

      if (!loadedWidgets.has(mod)) {
        await loadWidget(mod);
        loadedWidgets.add(mod);
      }

      const el = document.createElement(`resilver-${mod}`);
      el.dataset.config = JSON.stringify(moduleConfig.config || {});
      cell.appendChild(el);
    }

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
