import { test } from '@playwright/test';
import fs from 'fs';

test('Auth Flow Trace JSON', async ({ page }) => {
    const log: Record<string, unknown> = {};

    await page.goto('http://localhost:3000/login');

    await page.getByPlaceholder('e.g. acme-corp').fill('acme.local');
    await page.getByPlaceholder('you@company.com').fill('admin@acme.local');
    await page.getByPlaceholder('••••••••').fill('password123');

    // Trace requests
    const loginPromise = page.waitForResponse(r => r.url().includes('/auth/login'));
    const mePromise = page.waitForResponse(r => r.url().includes('/users/me')).catch(() => null);

    await page.getByRole('button', { name: 'Sign in' }).click();

    const loginRes = await loginPromise;
    log.loginStatus = loginRes.status();

    // Only wait 3s for getMe
    try {
        const mePromiseWithTimeout = Promise.race([
            mePromise,
            new Promise<null>(r => setTimeout(() => r(null), 3000))
        ]);
        const meRes = await mePromiseWithTimeout;

        if (meRes) {
            log.meStatus = meRes.status();
            log.meBody = await meRes.text();
        } else {
            log.meStatus = 'TIMEOUT';
        }
    } catch (e) {
        log.meError = String(e);
    }

    log.finalUrl = page.url();

    fs.writeFileSync('trace_log.json', JSON.stringify(log, null, 2));
});
