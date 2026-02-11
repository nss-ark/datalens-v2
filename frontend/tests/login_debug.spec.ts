import { test } from '@playwright/test';

test('Login Network Debug', async ({ page }) => {
    console.log('Navigating to login page...');
    await page.goto('http://localhost:3000/login');

    // Fill credentials
    await page.getByPlaceholder('e.g. acme-corp').fill('acme.local');
    await page.getByPlaceholder('you@company.com').fill('admin@acme.local');
    await page.getByPlaceholder('••••••••').fill('password123');

    // Setup network listener
    const responsePromise = page.waitForResponse(response => response.url().includes('/auth/login'));

    // Submit
    console.log('Submitting form...');
    await page.getByRole('button', { name: 'Sign in' }).click();

    // Wait for response
    console.log('Waiting for API response...');
    const response = await responsePromise;
    console.log('Login API Status:', response.status());

    try {
        const body = await response.json();
        console.log('Login API Body:', JSON.stringify(body));
    } catch {
        console.log('Could not parse response body JSON');
    }

    // Check URL after a moment
    await page.waitForTimeout(1000);
    console.log('Current URL:', page.url());
});
