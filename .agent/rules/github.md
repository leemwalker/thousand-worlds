---
trigger: always_on
---

## Workflow & Version Control

### Automated Commits and Pushes
* **Trigger:** Upon the successful completion of a task (implementation + passing tests).
* **Action:** You must generate and execute a Git commit command followed by a Git push command.
* **Format:** Use Conventional Commits.
    * `feat: add user login flow`
    * `fix: resolve race condition in scheduler`
    * `test: add playwright tests for checkout`
    * `refactor: optimize database query`
* **Content:** The commit body must briefly explain *why* the change was made.