# Test Coverage Verification

The `verify_coverage.sh` script provides automated test coverage verification for the Thousand Worlds repository. It discovers, runs, and reports on both unit and integration test coverage, ensuring code quality standards are met.

## Quick Start

Run all tests and verify coverage:

```bash
./verify_coverage.sh
```

Run only unit tests:

```bash
./verify_coverage.sh --unit-only
```

Generate HTML report:

```bash
./verify_coverage.sh --html
open coverage-reports/coverage_unit.html
```

## Requirements

- Go 1.21 or higher
- Bash environment (Linux, macOS, WSL)

## Command Line Options

| Option | Description |
|--------|-------------|
| `--unit-only` | Run only unit tests |
| `--integration-only` | Run only integration tests |
| `--html` | Generate HTML coverage report in `coverage-reports/` |
| `--json` | Output results in JSON format (useful for CI/CD) |
| `--verbose` | Show detailed test output |
| `--packages <path>` | Test specific package(s) only (e.g., `./internal/auth`) |
| `--fail-under <n>` | Set custom coverage threshold (overrides config) |
| `--no-color` | Disable colored output |
| `--fail-fast` | Stop on first test failure |
| `--dry-run` | Show what would be executed without running tests |
| `--log <file>` | Write output to log file |
| `--help` | Show usage information |

## Configuration

The script looks for a `.coveragerc` file in the repository root. You can configure thresholds, exclusions, and reporting options.

Example `.coveragerc`:

```yaml
thresholds:
  unit_tests: 80
  integration_tests: 100

exclude:
  - "vendor/*"
  - "**/*_generated.go"

reports:
  html: true
  output_dir: "./coverage-reports"
```

## Interpreting Results

The script provides a clear pass/fail status for each package and overall coverage.

- **PASS**: Coverage meets or exceeds the threshold.
- **FAIL**: Coverage is below the threshold.

Example output:

```text
Running Unit Tests...
Package                                    Coverage    Status
github.com/user/thousand-worlds/internal/lobby      87.3%      ✓ PASS
github.com/user/thousand-worlds/internal/world      76.4%      ✗ FAIL (below 80%)
```

## CI/CD Integration

### GitHub Actions

Add this step to your workflow:

```yaml
- name: Verify Test Coverage
  run: ./verify_coverage.sh --json > coverage-report.json

- name: Upload Coverage Report
  uses: actions/upload-artifact@v3
  with:
    name: coverage-report
    path: coverage-report.json
```

### GitLab CI

```yaml
test_coverage:
  script:
    - ./verify_coverage.sh --json
  artifacts:
    reports:
      coverage_report:
        coverage_format: cobertura
        path: coverage.xml
```

## Troubleshooting

**Error: Go is not installed**
Ensure Go is installed and in your PATH.

**Error: Database not initialized**
Integration tests require a running database. Ensure your Docker containers are up:
`docker-compose up -d`

**Tests failing but coverage is high**
The script checks both test success AND coverage. Fix failing tests first.
