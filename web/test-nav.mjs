import { chromium } from 'playwright';

(async () => {
    const browser = await chromium.launch();
    const page = await browser.newPage();

    page.on('console', msg => console.log('BROWSER LOG:', msg.text()));
    page.on('pageerror', error => console.error('BROWSER ERROR:', error.message));

    console.log("Navigating to / ...");
    // Using 5173 since the first dev server handles the port
    await page.goto('http://localhost:5173/');
    await page.waitForTimeout(1000);

    console.log("Clicking '通知設定' ...");
    await page.click('text=通知設定');
    await page.waitForTimeout(2000);

    await browser.close();
    process.exit(0);
})();
