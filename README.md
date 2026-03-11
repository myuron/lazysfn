# lazysfn

A TUI tool for browsing AWS Step Functions from your terminal. Inspired by [lazygit](https://github.com/jesseduffield/lazygit).

[https://github.com/user-attachments/assets/8f76fb05-b507-4118-8d61-6bbaa8145417](https://github.com/user-attachments/assets/39484e9b-fa0a-4abf-abaa-779c115b5f41)

[日本語版 README](docs/README.ja.md)

## Features

- AWS profile selection (loaded from `~/.aws/config`)
- State machine listing (Standard type only, sorted by name)
- Latest execution status shown with colored indicator `●`
- Execution history (execution ID, status, failed state, start/stop time, duration, input params)
- Color-coded statuses (SUCCEEDED: green, FAILED: red, RUNNING: blue, TIMED_OUT: yellow, ABORTED: gray)
- Incremental search for state machine names
- Keybinding help modal
- Manual refresh
- Error modal with recovery to profile selection
- Vim-style keybindings

## Tech Stack

- Language: Go
- TUI library: [gocui](https://github.com/jroimartin/gocui)
- AWS SDK: [aws-sdk-go-v2](https://github.com/aws/aws-sdk-go-v2)

## Installation

### Build from source

```sh
git clone https://github.com/myuron/lazysfn
cd lazysfn
go build -o ./dist/lazysfn ./cmd/lazysfn
```

### Nix

```sh
nix flake run .#build
```

## Keybindings

### Global

| Key | Action |
|---|---|
| `?` | Toggle keybinding help |
| `q` | Quit / Close popup |
| `R` | Refresh |

### Main View

| Key | Action |
|---|---|
| `j` / `k` | Cursor down / up |
| `h` / `l` | Focus left / right panel |
| `Tab` | Switch panel |
| `/` | Incremental search (left panel) |

### Search Mode

| Key | Action |
|---|---|
| Type | Filter state machines in real time |
| `Esc` | Cancel search (show all) |
| `Enter` | Confirm search (keep filter) |

### Profile Selection

| Key | Action |
|---|---|
| `j` / `k` | Cursor down / up |
| `Enter` | Select profile |
| `q` | Quit |

## Development

Requires [Nix](https://nixos.org/).

```sh
nix develop
```
