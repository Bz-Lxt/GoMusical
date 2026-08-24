import { test, expect } from "@playwright/test";

const BASE = process.env.E2E_BASE || "http://localhost:29471";

test.describe("GoMusical critical path", () => {
  test("login, preview shelf, sponsor unlock copy", async ({ page }) => {
    await page.goto(BASE + "/login");
    await page.getByRole("button", { name: "登录" }).click();
    await expect(page.getByText("阿北")).toBeVisible();
    await page.goto(BASE + "/");
    await expect(page.getByText("河对岸的灯")).toBeVisible();
    await page.getByText("河对岸的灯").click();
    await expect(page.getByText("赞助解锁")).toBeVisible();
    await page.getByText(/赞助解锁/).click();
    await expect(page.getByText("确认赞助金额")).toBeVisible();
    await expect(page.getByText("¥9.00")).toBeVisible();
  });
});
