# Package Manager

CColon includes a package manager for installing and managing third-party libraries.

## Usage

```bash
# Initialize a new project
ccolon pkg init

# Install a package
ccolon pkg install mypackage
ccolon pkg install mypackage@1.0.0

# Remove a package
ccolon pkg remove mypackage

# List installed packages
ccolon pkg list
```

## Project manifest (ccolon.json)

Run `ccolon pkg init` to create a `ccolon.json` in your project directory:

```json
{
    "name": "my-project",
    "version": "0.1.0",
    "description": "",
    "dependencies": {},
    "registry": "https://github.com/TRC-Loop/ccolon-registry"
}
```

### Fields

| Field | Description |
|---|---|
| `name` | Project name |
| `version` | Project version (semver) |
| `description` | Short description |
| `dependencies` | Map of package name to version |
| `registry` | Package registry URL (optional, defaults to the official registry) |

## Package Registry

The default package registry is a GitHub repository at [github.com/TRC-Loop/ccolon-registry](https://github.com/TRC-Loop/ccolon-registry).

### Custom registry

You can use a custom registry by setting the `registry` field in `ccolon.json` or the `CCOLON_REGISTRY` environment variable:

```bash
export CCOLON_REGISTRY=https://github.com/my-org/my-registry
```

## Installation directory

Packages are installed to `~/.ccolon/packages/<name>@<version>/`.

## Publishing packages

To publish a package to the registry, see the [registry documentation](https://github.com/TRC-Loop/ccolon-registry) for instructions on submitting packages.
