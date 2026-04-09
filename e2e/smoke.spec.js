const { test, expect } = require("@playwright/test");

test("page loads with correct title", async ({ page }) => {
  await page.goto("/");
  await expect(page).toHaveTitle("Resilver");
});

test("page has dark background", async ({ page }) => {
  await page.goto("/");
  const body = page.locator("body");
  await expect(body).toHaveCSS("background-color", "rgb(0, 0, 0)");
});
