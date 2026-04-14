# sleight-of-hand

A multicall CLI wrapper that transparently extends existing command-line tools
with custom subcommands.

## How it works

sleight-of-hand is a single Go binary that impersonates other CLI tools via
symlinks. When invoked, it checks `os.Args[0]` to determine which tool it's
acting as:

- If it recognizes a custom subcommand, it handles it directly
- Otherwise, it forwards all arguments to the real binary transparently
  (preserving stdin/stdout/stderr, signals, and exit codes)

## Build

```sh
cd sleight-of-hand
go build -o sleight-of-hand .
```

## Install

```sh
# Install symlinks for all registered tools
./sleight-of-hand install

# Install a specific tool
./sleight-of-hand install gh

# Custom bin directory
./sleight-of-hand install --bin-dir /usr/local/bin
```

## Uninstall

```sh
# Remove all symlinks (only removes symlinks pointing to sleight-of-hand)
./sleight-of-hand uninstall

# Remove a specific tool
./sleight-of-hand uninstall gh
```

## Wrapped tools

### gh (GitHub CLI)

Custom subcommands:

- `gh pr retry <PR#>` - Retry all failed CI jobs for a pull request
  - `--repo`, `-R` - Specify repository in OWNER/REPO format

All other `gh` commands are forwarded to the real GitHub CLI.

## Adding a new tool

1. Create a package under `tools/<name>/`
2. Implement `func Run(args []string) int`
3. Register in `internal/dispatch/dispatch.go`
4. Run `sleight-of-hand install <name>`
