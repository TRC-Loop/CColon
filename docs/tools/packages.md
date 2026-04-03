# Package Manager

CColon includes a package manager for installing third-party libraries from GitHub repositories.

## Quick start

```bash
// Install a package
ccolon pkg install https://github.com/TRC-Loop/ccl-testpkg

// Use it in your code
```

```
import ccl_testpkg

function main() {
    ccl_testpkg.test()  // prints "If you see this, it works!"
}
```

## Commands

```bash
// Initialize a new project
ccolon pkg init

// Install a package (latest from main branch)
ccolon pkg install https://github.com/TRC-Loop/ccl-testpkg

// Install a specific version (git tag)
ccolon pkg install https://github.com/TRC-Loop/ccl-testpkg@0.1.0

// Install into the current project only
ccolon pkg install --local https://github.com/TRC-Loop/ccl-testpkg

// Upgrade a package to latest
ccolon pkg upgrade ccl-testpkg

// Remove a package
ccolon pkg remove ccl-testpkg

// List installed packages (shows name, version, and source repo)
ccolon pkg list
```

### Flags

| Flag | Applies to | Description |
|------|-----------|-------------|
| `--local` | install, upgrade | Install into `./ccolon_packages/` instead of global `~/.ccolon/packages/` |

## How it works

Each CColon package is a GitHub repository. When you run `ccolon pkg install`, the tool:

1. Fetches the `ccolon.json` from the repo to get the package name and metadata
2. Downloads the repository as a tarball (from the specified tag or main branch)
3. Extracts it to `~/.ccolon/packages/<name>@<version>/` (or `./ccolon_packages/` with `--local`)
4. Registers the package as an importable module

## Using packages

Installed packages are available as modules. Import them like built-in modules:

```
import ccl_testpkg

function main() {
    ccl_testpkg.test()
}
```

You can also use selective imports:

```
from ccl_testpkg import test

function main() {
    test()
}
```

## Example package

[ccl-testpkg](https://github.com/TRC-Loop/ccl-testpkg) is a test package you can use to verify your setup:

```bash
ccolon pkg install https://github.com/TRC-Loop/ccl-testpkg
```

Its `ccolon.json`:

```json
{
  "name": "ccl-testpkg",
  "version": "0.1.0",
  "description": "Pakage to test CColon's builtin pm",
  "dependencies": {},
  "type": "ccl"
}
```

## Project manifest (ccolon.json)

Run `ccolon pkg init` to create a `ccolon.json` in your project directory:

```json
{
  "name": "my-project",
  "version": "0.1.0",
  "dependencies": {}
}
```

### Fields

| Field | Description |
|---|---|
| `name` | Package name (used for the install directory) |
| `version` | Package version (semver recommended) |
| `description` | Short description |
| `dependencies` | Map of dependency name to GitHub URL with version |
| `type` | Package type: `"ccl"` for CColon source, `"go"` for Go native plugins |
| `entry` | Entry point file (default: `lib.ccl` for ccl packages) |
| `repository` | Source GitHub URL (set automatically during install) |

## Creating a package

### CColon packages (type: "ccl")

A CColon package is a GitHub repo with this structure:

```
your-package/
  ccolon.json        // required: package manifest
  lib.ccl            // entry point (or whatever "entry" specifies)
  utils.ccl          // additional source files
  README.md          // optional: documentation
```

The `ccolon.json` should look like:

```json
{
  "name": "your-package",
  "version": "1.0.0",
  "description": "What this package does",
  "type": "ccl",
  "entry": "lib.ccl"
}
```

### Go native packages (type: "go")

For performance-critical code or system-level functionality, packages can be written in Go.

Go packages must implement a `Register` function that takes a `*vm.VM` and registers modules/functions. See the CColon stdlib source code for examples of how to create native modules.

## Versioning

Versions are git tags on the repository. To release a new version:

```bash
git tag 1.0.0
git push origin 1.0.0
```

Users can then install that specific version:

```bash
ccolon pkg install https://github.com/you/your-package@1.0.0
```

Without a version, the latest code from the `main` branch is used.

## Installation directories

- **Global**: `~/.ccolon/packages/<name>@<version>/` (default)
- **Local**: `./ccolon_packages/<name>@<version>/` (with `--local`)

Both directories are scanned when loading packages. Local packages take effect only for programs run from that project directory.
