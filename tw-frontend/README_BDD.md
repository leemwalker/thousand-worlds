# Frontend BDD Testing

This directory contains the Gherkin `.feature` files for behavior-driven development testing of the frontend.

## Setup (Pending `npm` availability)

To run these tests, we recommend using `playwright-bdd`.

1. **Install Dependencies**:
   ```bash
   npm install -D playwright-bdd @playwright/test
   ```

2. **Configure Playwright**:
   Update `playwright.config.ts` to include the BDD configuration.

3. **Generate Tests**:
   ```bash
   npx bddgen
   ```

4. **Run Tests**:
   ```bash
   npx playwright test
   ```

## Feature Structure

- `features/`: Contains all `.feature` files.
- `tests/bdd/`: (To be created) Will contain the generated spec files and step definitions.

## Writing Steps

Create step definition files in `tests/bdd/steps/*.ts` using the Cucumber syntax compatible with Playwright.

Example:
```typescript
import { createBdd } from 'playwright-bdd';
const { Given, When, Then } = createBdd();

Given('I am on the main menu', async ({ page }) => {
  await page.goto('/');
});
```
