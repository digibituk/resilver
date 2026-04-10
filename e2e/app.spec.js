const { test, expect } = require("@playwright/test");

test.describe("app bootstrapping", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto("/");
    await page.waitForSelector("#grid");
  });

  test("grid has correct number of cells for widget count", async ({
    page,
  }) => {
    // Default config has 2 widgets (clock + weather) → 2 cells
    const cells = page.locator(".grid-cell");
    await expect(cells).toHaveCount(2);
  });

  test("grid uses auto-calculated columns and rows", async ({ page }) => {
    // 2 widgets, direction row → 2 columns, 1 row
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
    expect(rowCount).toBe(1);
  });

  test("widgets are rendered in config order", async ({ page }) => {
    const cells = page.locator(".grid-cell");

    // First cell should contain the clock widget
    const firstWidget = cells.nth(0).locator("resilver-clock");
    await expect(firstWidget).toBeVisible();

    // Second cell should contain the weather widget
    const secondWidget = cells.nth(1).locator("resilver-weather");
    await expect(secondWidget).toBeVisible();
  });

  test("each cell has data-index attribute", async ({ page }) => {
    const cells = page.locator(".grid-cell");
    await expect(cells.nth(0)).toHaveAttribute("data-index", "0");
    await expect(cells.nth(1)).toHaveAttribute("data-index", "1");
  });

  test("config endpoint is reachable from the page", async ({ page }) => {
    const config = await page.evaluate(async () => {
      const resp = await fetch("/api/config");
      return resp.json();
    });

    expect(config.server.port).toBe(8080);
    expect(config.layout.maxWidgets).toBe(8);
    expect(config.layout.direction).toBe("row");
    expect(config.layout.widgets).toHaveLength(2);
    expect(config.modules.clock).toBeDefined();
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
