const { chromium } = require('playwright');

(async () => {
    console.log('Launching browser...');
    const browser = await chromium.launch();
    const page = await browser.newPage();

    console.log('Navigating to http://192.168.12.132:5173...');
    try {
        await page.goto('http://192.168.12.132:5173', { timeout: 30000, waitUntil: 'commit' });
        console.log('Page title:', await page.title());
        console.log('SUCCESS: Page loaded');
    } catch (e) {
        console.error('ERROR: Failed to load page:', e.message);
    }

    await browser.close();
})();
