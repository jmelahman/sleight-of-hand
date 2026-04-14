# sleight-of-hand - Agent Guide

## Architecture

This is a **multicall binary** (like busybox). A single compiled binary
dispatches based on `os.Args[0]`:

- Invoked as `sleight-of-hand` -> management CLI (install/uninstall)
- Invoked as `gh` (via symlink) -> gh tool handler with custom subcommands

### Key packages

- `internal/dispatch` - Tool registry. Maps tool names to handler functions.
  Add new tools here.
- `internal/passthrough` - Transparent exec of the real binary. Forwards
  stdin/stdout/stderr, signals (SIGINT, SIGTERM, SIGHUP), and exit codes.
- `internal/lookup` - Finds the real binary in PATH, skipping any entry that
  resolves to the sleight-of-hand binary itself (via `filepath.EvalSymlinks`).
- `tools/<name>/` - Per-tool packages. Each exposes `Run(args []string) int`.
- `cmd/` - Management CLI commands (install, uninstall).

### Command flow

```
os.Args[0] == "gh"
  -> dispatch.Run("gh", args)
    -> gh.Run(args)
      -> cobra matches "pr retry" -> custom handler
      -> cobra doesn't match -> passthrough.Exec("gh", args) -> real gh binary
```

## Adding a new tool

1. Create `tools/<name>/<name>.go` with `func Run(args []string) int`
2. Build a cobra command tree with `DisableFlagParsing: true` on the root
3. Set the root's `RunE` to call `passthrough.Exec("<name>", args)` as fallback
4. Register in `internal/dispatch/dispatch.go`: add to the `registry` map
5. Rebuild and run `sleight-of-hand install <name>`

## Adding a subcommand to an existing tool

1. Create a new `.go` file in the tool's package (or a sub-package for
   subcommand groups)
2. Return a `*cobra.Command` from a constructor function
3. Register it with `cmd.AddCommand()` in the parent command
4. Subcommand groups that also need passthrough should set
   `DisableFlagParsing: true` and prepend their name in the fallback args

## Invariants

- **Passthrough must be transparent**: exit codes, signals, stdin/stdout/stderr,
  and environment variables must all forward correctly to the real binary.
- **Lookup must skip self**: `internal/lookup` uses `filepath.EvalSymlinks` to
  compare candidates against `os.Executable()`. This prevents infinite
  recursion when the symlink and binary resolve to the same file.
- **Install is safe**: only creates/overwrites symlinks, never overwrites
  regular files.
- **Uninstall is safe**: only removes symlinks that point back to the
  sleight-of-hand binary.

## Testing

- Build: `go build -o sleight-of-hand .`
- Test passthrough: `./sleight-of-hand install gh && gh --version`
- Test custom command: `gh pr retry <PR#>`
- Test uninstall: `./sleight-of-hand uninstall gh`
