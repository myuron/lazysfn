# lazysfn

A TUI tool for browsing AWS Step Functions from your terminal. Inspired by [lazygit](https://github.com/jesseduffield/lazygit).

[https://github.com/user-attachments/assets/8f76fb05-b507-4118-8d61-6bbaa8145417](https://github.com/user-attachments/assets/39484e9b-fa0a-4abf-abaa-779c115b5f41)

[日本語版 README](docs/README.ja.md)

## Features

- AWS profile selection (loaded from `~/.aws/config`)
- State machine listing (Standard type only, sorted by name)
- Latest execution status shown with colored indicator `●`
- Execution history (execution ID, status, failed state, start/stop time, duration, input params)
- Pagination for execution history (auto-loads next page on scroll)
- Input parameter detail modal (Enter on execution → pretty-printed JSON, j/k scroll)
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

### go install

```sh
go install github.com/myuron/lazysfn@latest
```

### GitHub Releases

Download a prebuilt binary from the [Releases](https://github.com/myuron/lazysfn/releases/latest) page.

```sh
# Example: macOS (Apple Silicon)
curl -Lo lazysfn https://github.com/myuron/lazysfn/releases/latest/download/lazysfn-darwin-arm64
chmod +x lazysfn
sudo mv lazysfn /usr/local/bin/
```

Available binaries: `lazysfn-darwin-amd64`, `lazysfn-darwin-arm64`, `lazysfn-linux-amd64`, `lazysfn-windows-amd64.exe`

### Build from source

```sh
git clone https://github.com/myuron/lazysfn
cd lazysfn
go build -o ./dist/lazysfn .
```

### Nix

```sh
nix run github:myuron/lazysfn
```

To add as a flake input:

```nix
{
  inputs = {
    lazysfn.url = "github:myuron/lazysfn";
  };

  outputs = { self, nixpkgs, lazysfn, ... }: {
    # Use the overlay to add pkgs.lazysfn
    nixosConfigurations.example = nixpkgs.lib.nixosSystem {
      modules = [{
        nixpkgs.overlays = [ lazysfn.overlays.default ];
        environment.systemPackages = [ pkgs.lazysfn ];
      }];
    };

    # Or use the package directly
    # lazysfn.packages.${system}.default
  };
}
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
| `Enter` | Open input parameter detail (right panel) |
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
