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
    // Wait for data to load
    await expect(temp).not.toHaveText("--", { timeout: 5000 });
    const text = await temp.textContent();
    // Should contain a number and degree symbol
    expect(text).toMatch(/\d+°[CF]/);
  });

  test("displays weather icon", async ({ page }) => {
    const icon = page.locator(".resilver-weather__icon");
    await expect(icon).toBeVisible();
    // Wait for data to load — icon should not be the error icon
    await expect(icon).not.toHaveText("⚠️", { timeout: 5000 });
    const text = await icon.textContent();
    expect(text.length).toBeGreaterThan(0);
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

  test("displays feels-like and humidity details", async ({ page }) => {
    const details = page.locator(".resilver-weather__details");
    await expect(details).toBeVisible();
    await expect(details).not.toHaveText("", { timeout: 5000 });
    const text = await details.textContent();
    expect(text).toContain("Feels");
    expect(text).toContain("Humidity");
  });

  test("displays location name", async ({ page }) => {
    const location = page.locator(".resilver-weather__location");
    await expect(location).toBeVisible();
    await expect(location).toHaveText("Grays");
  });

  test("is placed in the correct grid position", async ({ page }) => {
    const cell = page.locator('.grid-cell[data-position="top-right"]');
    const weather = cell.locator("resilver-weather");
    await expect(weather).toBeVisible();
  });
});
