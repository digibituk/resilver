const { test, expect } = require("@playwright/test");

test.describe("app bootstrapping", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto("/");
    await page.waitForSelector("#grid");
  });

  test("grid has 9 cells for 3x3 layout", async ({ page }) => {
    const cells = page.locator(".grid-cell");
    await expect(cells).toHaveCount(9);
  });

  test("grid uses correct CSS grid template", async ({ page }) => {
    const grid = page.locator("#grid");
    const columns = await grid.evaluate(
      (el) => getComputedStyle(el).gridTemplateColumns
    );
    // 3 column values
    const colCount = columns.split(" ").length;
    expect(colCount).toBe(3);

    const rows = await grid.evaluate(
      (el) => getComputedStyle(el).gridTemplateRows
    );
    const rowCount = rows.split(" ").length;
    expect(rowCount).toBe(3);
  });

  test("each cell has correct data-position attribute", async ({ page }) => {
    const expected = [
      "top-left", "top-center", "top-right",
      "middle-left", "middle-center", "middle-right",
      "bottom-left", "bottom-center", "bottom-right",
    ];

    for (const pos of expected) {
      const cell = page.locator(`.grid-cell[data-position="${pos}"]`);
      await expect(cell).toHaveCount(1);
    }
  });

  test("empty cells contain no widgets", async ({ page }) => {
    const emptyCell = page.locator('.grid-cell[data-position="top-left"]');
    const children = emptyCell.locator(":scope > *");
    await expect(children).toHaveCount(0);
  });

  test("config endpoint is reachable from the page", async ({ page }) => {
    const config = await page.evaluate(async () => {
      const resp = await fetch("/api/config");
      return resp.json();
    });

    expect(config.server.port).toBe(8080);
    expect(config.layout.columns).toBe(3);
    expect(config.layout.rows).toBe(3);
    expect(config.modules.clock.enabled).toBe(true);
  });

  test("tailwind is loaded and functional", async ({ page }) => {
    // The body has Tailwind class bg-black which should compute to rgb(0,0,0)
    const body = page.locator("body");
    await expect(body).toHaveCSS("background-color", "rgb(0, 0, 0)");

    // The grid has Tailwind class p-5 which should compute to 20px padding
    const grid = page.locator("#grid");
    await expect(grid).toHaveCSS("padding", "20px");
  });
});
