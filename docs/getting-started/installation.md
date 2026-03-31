# Installation

## Quick install

### Linux / macOS

```sh
curl -fsSL https://raw.githubusercontent.com/TRC-Loop/CColon/main/install.sh | sh
```

This downloads the latest release binary and installs it to `/usr/local/bin`. You may be prompted for your password if sudo is required.

To install a specific version:

```sh
curl -fsSL https://raw.githubusercontent.com/TRC-Loop/CColon/main/install.sh | sh -s v0.1.0
```

### Windows (PowerShell)

```powershell
irm https://raw.githubusercontent.com/TRC-Loop/CColon/main/install.ps1 | iex
```

This installs CColon to `%LOCALAPPDATA%\CColon` and adds it to your user PATH. You will need to restart your terminal after installation.

To install a specific version:

```powershell
& ([scriptblock]::Create((irm https://raw.githubusercontent.com/TRC-Loop/CColon/main/install.ps1))) v0.1.0
```

## Build from source

Building from source requires [Go](https://go.dev/dl/) 1.22 or later.

```sh
git clone https://github.com/TRC-Loop/CColon.git
cd CColon/ccolon
go build -o ccolon .
```

Then move the binary to a directory on your PATH:

=== "Linux / macOS"

    ```sh
    sudo mv ccolon /usr/local/bin/
    ```

=== "Windows"

    Move `ccolon.exe` to a folder that is on your PATH, or add the build directory to your PATH.

## Verify the installation

```sh
ccolon --version
```

You should see output like:

```
CColon v0.1.0
```

## Next steps

Continue with the [Hello World](hello-world.md) guide to write your first CColon program.
