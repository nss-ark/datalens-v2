import { test, expect } from '@playwright/test';

test('Login flow', async ({ page }) => {
    console.log('Navigating to login page...');
    await page.goto('http://localhost:3000/login');

    // Verify inputs exist
    const domainInput = page.getByPlaceholder('e.g. acme-corp');
    const emailInput = page.getByPlaceholder('you@company.com');
    const passwordInput = page.getByPlaceholder('••••••••');

    await expect(domainInput).toBeVisible();
    console.log('Domain input visible');

    await expect(emailInput).toBeVisible();
    console.log('Email input visible');

    // Fill credentials
    console.log('Filling credentials...');
    await domainInput.fill('acme.local');
    await emailInput.fill('admin@acme.local');
    await passwordInput.fill('password123');

    // Submit
    console.log('Submitting form...');
    await page.getByRole('button', { name: 'Sign in' }).click();

    // Verify navigation
    console.log('Waiting for dashboard...');
    await expect(page).toHaveURL(/\/dashboard/);
    console.log('Dashboard loaded!');

    // Take success screenshot
    await page.screenshot({ path: 'login-success.png' });
});
