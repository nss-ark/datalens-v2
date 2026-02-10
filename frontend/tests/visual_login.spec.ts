import { test, expect } from '@playwright/test';

test('Visual Login Demo', async ({ page }) => {
    // 1. Go to login
    await page.goto('http://localhost:3000/login');

    // 2. Clear and Fill Form (slowly so user can see)
    await page.getByPlaceholder('e.g. acme-corp').fill('acme.local');
    await page.waitForTimeout(500);

    await page.getByPlaceholder('you@company.com').fill('admin@acme.local');
    await page.waitForTimeout(500);

    await page.getByPlaceholder('••••••••').fill('password123');
    await page.waitForTimeout(500);

    // 3. Click Login
    await page.getByRole('button', { name: 'Sign in' }).click();

    // 4. Wait for dashboard
    await expect(page).toHaveURL(/dashboard/);

    // 5. Keep browser open for user to inspect
    console.log('Login successful! Browser will stay open for 30 seconds...');
    await page.waitForTimeout(30000);
});
