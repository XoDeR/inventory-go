# CI package (GitHub Actions + Taskfile)

A self-contained CI setup for a multi-module Go repository. Copy it into another project as described below.

## Contents

```
.github/
  workflows/
    ci.yml                      # entry point: extract-vars -> lint -> test -> coverage
    lint-reusable.yml           # golangci-lint via `task lint`
    test-reusable.yml           # unit tests via `task test`
    test-coverage-reusable.yml  # coverage via `task test-coverage` + Gist badge
  scripts/
    extract-versions.sh         # reads GO_VERSION, GOLANGCI_LINT_VERSION, MODULES from the Taskfile
Taskfile.yml                    # minimal Taskfile with only the tasks CI needs
```

## How to install in another project

1. Copy `.github/` into the repository root.
2. Copy `Taskfile.yml` into the root, or merge its `vars:` block and the `lint`, `test`, `test-coverage` and `install-golangci-lint` tasks into your existing Taskfile.
3. Edit `vars:` in the Taskfile: set `GO_VERSION`, `GOLANGCI_LINT_VERSION` and `MODULES` (space-separated module directories).
4. Add a `.golangci.yml` to the repo root (used by `task lint`).
5. Make sure the repo has a `go.work` file if it contains multiple modules.

## Triggers

Automatic runs are disabled in `ci.yml`. CI runs only manually:

- GitHub UI: Actions tab -> CI -> Run workflow
- CLI: `gh workflow run CI`

To enable automatic runs, uncomment the `push` / `pull_request` blocks under `on:` in `ci.yml`. Keep at least one trigger active, otherwise the workflow file is invalid.

To switch CI off completely, use Actions tab -> CI -> `...` -> Disable workflow.

## Coverage badge (optional)

The `coverage` job (runs only on `main`) writes the total coverage to a Gist, and a README badge reads it.

1. Create a public Gist containing a file named `coverage.json`.
2. Create a personal access token with the `gist` scope and save it as the repository secret `CODE_COVERAGE_GIST`.
3. Save the Gist ID as the repository variable `COVERAGE_GIST_ID` (Settings -> Secrets and variables -> Actions -> Variables).
4. Add the badge to your README (replace `<user>` and `<gist-id>`):

```markdown
![Coverage](https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/<user>/<gist-id>/raw/coverage.json)
```

If you don't want a badge, delete the `coverage` job from `ci.yml` and remove `test-coverage-reusable.yml`.


