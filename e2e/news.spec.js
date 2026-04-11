const { test, expect } = require("@playwright/test");

test.describe("news widget", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto("/");
    await page.waitForSelector("resilver-news");
  });

  test("renders the news element", async ({ page }) => {
    const news = page.locator("resilver-news");
    await expect(news).toBeVisible();
  });

  test("displays a headline after loading", async ({ page }) => {
    const headline = page.locator(".resilver-news__headline");
    await expect(headline).toBeVisible();
    await expect(headline).not.toHaveText("", { timeout: 5000 });
    const text = await headline.textContent();
    expect(text.length).toBeGreaterThan(0);
  });

  test("displays source name", async ({ page }) => {
    const source = page.locator(".resilver-news__source");
    await expect(source).toBeVisible();
    await expect(source).not.toHaveText("", { timeout: 5000 });
    const text = await source.textContent();
    expect(text.length).toBeGreaterThan(0);
  });

  test("content has fade transition style", async ({ page }) => {
    const content = page.locator(".resilver-news__content");
    await expect(content).toBeVisible();
    const transition = await content.evaluate(
      (el) => getComputedStyle(el).transition
    );
    expect(transition).toContain("opacity");
  });

  test("is placed in the third grid cell", async ({ page }) => {
    const cell = page.locator('.grid-cell[data-index="2"]');
    const news = cell.locator("resilver-news");
    await expect(news).toBeVisible();
  });
});
