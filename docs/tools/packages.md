# Package Manager

CColon includes a package manager for installing third-party libraries from GitHub repositories.

## Usage

```bash
# Initialize a new project
ccolon pkg init

# Install a package from a GitHub repo (latest from main branch)
ccolon pkg install https://github.com/someone/their-package

# Install a specific version (git tag)
ccolon pkg install https://github.com/someone/their-package@1.2.0

# Remove a package
ccolon pkg remove mypackage

# List installed packages
ccolon pkg list
```

## How it works

Each CColon package is a GitHub repository. When you run `ccolon pkg install`, the tool:

1. Fetches the `ccolon.json` from the repo to get the package name and metadata
2. Downloads the repository as a tarball (from the specified tag or main branch)
3. Extracts it to `~/.ccolon/packages/<name>@<version>/`

Versions correspond to git tags on the repository.

## Project manifest (ccolon.json)

Run `ccolon pkg init` to create a `ccolon.json` in your project directory:

```json
{
  "name": "my-project",
  "version": "0.1.0",
  "description": "A short description of your project",
  "dependencies": {},
  "type": "ccl"
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

## Creating a package

### CColon packages (type: "ccl")

A CColon package is a GitHub repo with this structure:

```
your-package/
  ccolon.json        # required: package manifest
  lib.ccl            # entry point (or whatever "entry" specifies)
  utils.ccl          # additional source files
  README.md          # optional: documentation
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

For performance-critical code or system-level functionality, packages can be written in Go. These work similar to how Python allows C extensions.

A Go native package is a GitHub repo with:

```
your-go-package/
  ccolon.json        # required: type must be "go"
  plugin.go          # Go source that registers native functions
  README.md          # optional
```

The `ccolon.json`:

```json
{
  "name": "your-go-package",
  "version": "1.0.0",
  "description": "A native Go package for CColon",
  "type": "go",
  "entry": "plugin.go"
}
```

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

## Installation directory

Packages are installed to `~/.ccolon/packages/<name>@<version>/`.
