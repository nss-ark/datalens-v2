import { test, expect } from '@playwright/test';
import fs from 'fs';

test('Auth Flow Trace Headers', async ({ page }) => {
    const log: any = {};

    await page.goto('http://localhost:3000/login');

    await page.getByPlaceholder('e.g. acme-corp').fill('acme.local');
    await page.getByPlaceholder('you@company.com').fill('admin@acme.local');
    await page.getByPlaceholder('••••••••').fill('password123');

    // Trace requests
    const loginPromise = page.waitForResponse(r => r.url().includes('/auth/login'));

    page.on('console', msg => console.log(`[BROWSER] ${msg.text()}`));

    // Intercept the /users/me request to check headers
    let usersMeHeaders = null;
    page.on('request', request => {
        if (request.url().includes('/users/me')) {
            usersMeHeaders = request.headers();
        }
    });

    await page.getByRole('button', { name: 'Sign in' }).click();

    const loginRes = await loginPromise;
    log.loginStatus = loginRes.status();

    // Wait for /users/me to happen (or fail)
    try {
        await page.waitForResponse(r => r.url().includes('/users/me'), { timeout: 3000 });
    } catch (e) {
        log.timeout = true;
    }

    log.usersMeHeaders = usersMeHeaders;
    log.finalUrl = page.url();

    fs.writeFileSync('trace_headers.json', JSON.stringify(log, null, 2));
});
