const { test, expect } = require("@playwright/test");

test.describe("weather widget", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto("/");
    await page.waitForSelector("resilver-weather");
  });

  test("renders the weather element", async ({ page }) => {
    const weather = page.locator("resilver-weather");
    await expect(weather).toBeVisible();
  });

  test("displays temperature", async ({ page }) => {
    const temp = page.locator(".resilver-weather__temp");
    await expect(temp).toBeVisible();
    await expect(temp).not.toHaveText("--", { timeout: 5000 });
    const text = await temp.textContent();
    expect(text).toMatch(/\d+°[CF]/);
  });

  test("displays weather icon", async ({ page }) => {
    const icon = page.locator(".resilver-weather__icon");
    await expect(icon).toBeVisible();
    const iconEl = icon.locator("i.wi");
    await expect(iconEl).toBeVisible({ timeout: 5000 });
  });

  test("displays weather description", async ({ page }) => {
    const desc = page.locator(".resilver-weather__desc");
    await expect(desc).toBeVisible();
    await expect(desc).not.toHaveText("Unable to load weather", {
      timeout: 5000,
    });
    const text = await desc.textContent();
    expect(text.length).toBeGreaterThan(0);
  });

  test("displays feels-like details", async ({ page }) => {
    const details = page.locator(".resilver-weather__details");
    await expect(details).toBeVisible();
    await expect(details).not.toHaveText("", { timeout: 5000 });
    const text = await details.textContent();
    expect(text).toContain("Feels");
  });

  test("displays location name", async ({ page }) => {
    const location = page.locator(".resilver-weather__location");
    await expect(location).toBeVisible();
    await expect(location).toHaveText("Grays");
  });

  test("is placed in the second grid cell", async ({ page }) => {
    const cell = page.locator('.grid-cell[data-index="1"]');
    const weather = cell.locator("resilver-weather");
    await expect(weather).toBeVisible();
  });
});
