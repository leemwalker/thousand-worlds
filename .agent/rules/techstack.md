---
trigger: always_on
---

## Technology Stack Guidelines

### Go (Golang)
* **Style:** Follow "Effective Go" and standard formatting (`gofmt`).
* **Error Handling:** Handle errors explicitly. Do not ignore errors. Use wrapping for context (`fmt.Errorf("...: %w", err)`).
* **Concurrency:** Use channels and goroutines judiciously. Prevent race conditions.
* **Structure:** Prefer domain-driven package structure over flat structures for complex apps.

### Svelte, HTML, JavaScript, TypeScript
* **Svelte:** Use functional components and reactive statements (`$:`) appropriately. Keep logic in `script lang="ts"`.
* **TypeScript:** strict mode must be enabled. Avoid `any`; define interfaces/types for all data structures and props.
* **HTML:** Use semantic HTML5 elements (article, section, nav, aside) for accessibility.
* **Styling:** Scope styles within Svelte components unless global styles are strictly necessary.

### Shell Scripting
* **Safety:** Always use `set -euo pipefail` at the start of scripts to ensure strict error handling.
* **Portability:** Write POSIX-compliant scripts where possible, or specify `#!/bin/bash` explicitly.

### Deployment Style
* **Remote Server:** We deploy to the babylon server at 10.0.0.17 where everything runs in docker containers. We use the update_build.sh script to pull down the most recent changes from Git and deploy to Docker.