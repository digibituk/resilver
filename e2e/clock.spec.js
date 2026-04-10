const { test, expect } = require("@playwright/test");

test.describe("clock widget", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto("/");
    await page.waitForSelector("resilver-clock");
  });

  test("renders the clock element", async ({ page }) => {
    const clock = page.locator("resilver-clock");
    await expect(clock).toBeVisible();
  });

  test("displays time text", async ({ page }) => {
    const time = page.locator(".resilver-clock__time");
    await expect(time).toBeVisible();
    const text = await time.textContent();
    expect(text).toMatch(/\d{1,2}:\d{2}/);
  });

  test("displays date when showDate is true", async ({ page }) => {
    const date = page.locator(".resilver-clock__date");
    await expect(date).toBeVisible();
    const text = await date.textContent();
    expect(text).toMatch(/\d{4}/);
  });

  test("updates time every second", async ({ page }) => {
    const time = page.locator(".resilver-clock__time");
    const first = await time.textContent();
    await page.waitForTimeout(1100);
    const second = await time.textContent();
    expect(second).toBeDefined();
  });

  test("is placed in the first grid cell", async ({ page }) => {
    const cell = page.locator('.grid-cell[data-index="0"]');
    const clock = cell.locator("resilver-clock");
    await expect(clock).toBeVisible();
  });
});
