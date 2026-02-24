const { chromium } = require('playwright');

(async () => {
  const browser = await chromium.launch();
  const page = await browser.newPage();
  
  page.on('console', msg => console.log('BROWSER LOG:', msg.text()));
  page.on('pageerror', error => console.error('BROWSER ERROR:', error.message));

  console.log("Navigating to / ...");
  await page.goto('http://localhost:5173/');
  await page.waitForTimeout(1000);

  console.log("Navigating to /settings ...");
  await page.click('text=システム設定');
  await page.waitForTimeout(1000);

  console.log("Navigating to /notifications ...");
  await page.click('text=通知設定');
  await page.waitForTimeout(1000);

  await browser.close();
})();
