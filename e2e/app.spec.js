const { test, expect } = require("@playwright/test");

test.describe("app bootstrapping", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto("/");
    await page.waitForSelector("#grid");
  });

  test("grid has correct number of cells for widget count", async ({
    page,
  }) => {
    // Default config has 3 widgets (clock + weather + news) → 3 cells
    const cells = page.locator(".grid-cell");
    await expect(cells).toHaveCount(3);
  });

  test("grid uses auto-calculated columns and rows", async ({ page }) => {
    // 3 widgets, direction row → 2 columns, 2 rows
    const grid = page.locator("#grid");
    const columns = await grid.evaluate(
      (el) => getComputedStyle(el).gridTemplateColumns
    );
    const colCount = columns.split(" ").length;
    expect(colCount).toBe(2);

    const rows = await grid.evaluate(
      (el) => getComputedStyle(el).gridTemplateRows
    );
    const rowCount = rows.split(" ").length;
    expect(rowCount).toBe(2);
  });

  test("widgets are rendered in config order", async ({ page }) => {
    const cells = page.locator(".grid-cell");

    const firstWidget = cells.nth(0).locator("resilver-clock");
    await expect(firstWidget).toBeVisible();

    const secondWidget = cells.nth(1).locator("resilver-weather");
    await expect(secondWidget).toBeVisible();

    const thirdWidget = cells.nth(2).locator("resilver-news");
    await expect(thirdWidget).toBeVisible();
  });

  test("each cell has data-index attribute", async ({ page }) => {
    const cells = page.locator(".grid-cell");
    await expect(cells.nth(0)).toHaveAttribute("data-index", "0");
    await expect(cells.nth(1)).toHaveAttribute("data-index", "1");
    await expect(cells.nth(2)).toHaveAttribute("data-index", "2");
  });

  test("last widget in odd count spans remainder", async ({ page }) => {
    const lastCell = page.locator('.grid-cell[data-index="2"]');
    const gridColumn = await lastCell.evaluate(
      (el) => el.style.gridColumn
    );
    expect(gridColumn).toBe("span 2");
  });

  test("config endpoint is reachable from the page", async ({ page }) => {
    const config = await page.evaluate(async () => {
      const resp = await fetch("/api/config");
      return resp.json();
    });

    expect(config.server.port).toBe(8080);
    expect(config.layout.maxWidgets).toBe(8);
    expect(config.layout.direction).toBe("row");
    expect(config.layout.widgets).toHaveLength(3);
    expect(config.modules.clock).toBeDefined();
    expect(config.modules.news).toBeDefined();
  });

  test("tailwind is loaded and functional", async ({ page }) => {
    const body = page.locator("body");
    await expect(body).toHaveCSS("background-color", "rgb(0, 0, 0)");

    const grid = page.locator("#grid");
    await expect(grid).toHaveCSS("padding", "20px");
  });

  test("grid uses auto-flow based on direction config", async ({ page }) => {
    const grid = page.locator("#grid");
    const autoFlow = await grid.evaluate(
      (el) => getComputedStyle(el).gridAutoFlow
    );
    // Default direction is "row"
    expect(autoFlow).toBe("row");
  });
});
