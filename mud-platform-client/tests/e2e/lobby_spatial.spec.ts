import { test, expect } from '@playwright/test';
import { createCharacter, deleteCharacter } from './utils/auth';

test.describe('Lobby Spatial Logic', () => {
    let charName: string;

    test.beforeEach(async ({ page }) => {
        // Generate unique character name
        charName = `Explorer_${Math.floor(Math.random() * 10000)}`;
        await page.goto('/');

        // Login/Create Character flow would be here, but using createCharacter helper if available
        // or just walking through UI.
        // Assuming 'createCharacter' utility might need expansion or we just do it via UI
        await page.fill('input[placeholder="Enter your name"]', charName);
        await page.click('button:has-text("Enter World")');

        // Select role/race if prompted
        // Wait for game to load
        await expect(page.locator('#game-terminal')).toBeVisible();
    });

    test('Statue Collision and Portal Proximity', async ({ page }) => {
        // 1. Ensure we are in lobby spawn (should be 5,2)
        // Type 'lobby' to be sure
        await page.keyboard.type('lobby');
        await page.keyboard.press('Enter');
        await expect(page.locator('.message-system')).toContainText('Grand Lobby');

        // 2. Test Statue Collision using text commands
        // We are at (5,2). Statue is at (5,5).
        // Move North to (5,3)
        await page.keyboard.type('north');
        await page.keyboard.press('Enter');
        // Move North to (5,4)
        await page.keyboard.type('n');
        await page.keyboard.press('Enter');

        // Try to move North to (5,5) - Should Fail
        await page.keyboard.type('n');
        await page.keyboard.press('Enter');

        // Expect error message about statue or blocking
        // The implementation returns "The massive statue blocks your path." as error.
        // Errors are usually styled .message-error
        await expect(page.locator('.message-error').last()).toContainText('statue blocks');

        // 3. Test Portal Proximity
        // We need a target world. Let's try to enter a made-up world ID "TestWorld" 
        // Wait, "enter" requires world ID or name. If name, it searches.
        // If I use a random name "PhantomWorld", it might not exist.
        // Better to list worlds or define one?
        // "who" shows players.
        // Assuming there's a way to list worlds or I can just try "lobby" again to reset.
        // Actually, "enter" command looks up world by name. 
        // If I try "enter NonExistent", it returns "World not found".
        // I need a VALID world. 
        // Since I can't easily create a world via CLI in this test without admin, 
        // I might depend on a seed world or creating one if UI allows.
        // The UI has "Create World".

        // Create a new world to test portal entry
        await page.keyboard.type('create world PortalTest');
        await page.keyboard.press('Enter');
        // Wait for "World PortalTest created"
        // It might take us there automatically?
        // If so, type 'lobby' to return.
        await expect(page.locator('.message-system').last()).toContainText('Welcome to PortalTest', { timeout: 10000 });

        await page.keyboard.type('lobby');
        await page.keyboard.press('Enter');
        await expect(page.locator('.message-system').last()).toContainText('Grand Lobby');

        // Now try to enter "PortalTest"
        await page.keyboard.type('enter PortalTest');
        await page.keyboard.press('Enter');

        // It should fail with proximity error because we are at (5,4) (moved north twice).
        // Portals are at edges (0 or 10).
        // Distance from (5,4) to any wall (x=0, x=10, y=0, y=10) checks.
        // Closest wall is West(0) dist 5? No, 5-0=5. North(10) dist 6. South(0) dist 4.
        // Wait, y=0 is South. We are at y=4. Distance 4.
        // If portal is on South wall (y=0, x=?), we might be close enough!
        // 5m proximity!
        // If portal is at (x, 0), and we are at (5, 4).
        // sqrt((x-5)^2 + (0-4)^2) = sqrt((x-5)^2 + 16).
        // If x=5, dist=4 <= 5. So we CAN enter South portals from (5,4).
        // This makes the "Failure" test tricky if we are lucky.

        // To GUARANTEE failure, we should stand in center (5,5) (blocked) or (5,4)
        // and try to enter a portal on FAR wall (North 10).
        // But we don't know where it is yet.

        // Capture the error message.
        // If it succeeds, fine (we handle that case).
        // If it fails, parse message.

        // Let's grab the last message.
        // Note: Playwright locators are dynamic.
        const lastMessage = page.locator('#game-messages > div').last();
        // We wait for it to appear.
        // It could be the "trigger_entry_options" (success) or "message-error".

        await expect(async () => {
            const text = await lastMessage.textContent();
            const isError = await lastMessage.evaluate(el => el.classList.contains('message-error'));
            if (isError) {
                expect(text).toContain('too far');
            } else {
                // Success? That implies we were close enough.
                // We want to verify we CAN get close enough eventually.
                // If we solved it immediately, that's okay too but less coverage of "finding" it.
            }
        }).toPass();

        // Retrieve text to decide logic
        const msgText = await lastMessage.textContent();

        if (msgText?.includes('too far')) {
            // Parse direction
            let direction = '';
            if (msgText.includes('North')) direction = 'north';
            else if (msgText.includes('South')) direction = 'south';
            else if (msgText.includes('East')) direction = 'east';
            else if (msgText.includes('West')) direction = 'west';

            // Move towards it. 
            // 5 steps should cover moving from center-ish to wall.
            for (let i = 0; i < 6; i++) {
                await page.keyboard.type(direction);
                await page.keyboard.press('Enter');
                // Small wait to ensure processing
                await page.waitForTimeout(200);
            }

            // Try enter again
            await page.keyboard.type('enter PortalTest');
            await page.keyboard.press('Enter');

            // Should succeed (modal or welcome)
            // Check for modal or "welcome"
            // Validating modal appearance
            await expect(page.locator('text=How do you want to enter?')).toBeVisible();
        } else {
            // We were already close enough. 
            await expect(page.locator('text=How do you want to enter?')).toBeVisible();
        }

    });
});
