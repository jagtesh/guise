# Guise ![Release](https://github.com/jagtesh/guise/actions/workflows/release.yml/badge.svg)

**Guise** is a CLI identity manager for your developer tools. It allows you to seamlessly switch contexts (profiles) for tools like OpenAI Codex, Google Gemini, Anthropic Claude, and GitHub Copilot.

Think of it as a "vault" for your configuration files. Switch from "Personal" to "Work" credentials instantly with a beautiful TUI.

![Guise](https://via.placeholder.com/800x400?text=Guise+TUI+Screenshot)

## Features

- **Profile Switching**: Safely swap configuration directories (e.g., `~/.codex`, `~/.config/github-copilot`).
- **Context Awareness**: "Adopts" existing configurations so you don't lose your setup.
- **Cross-Platform**: Works on macOS, Linux, and Windows.
- **Beautiful TUI**: Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea).

## Installation

### Shell Script (macOS, Linux, WSL)
The easiest way to install Guise is via the install script:

```bash
curl -sL https://raw.githubusercontent.com/jagtesh/guise/main/install.sh | bash
```

### Windows (PowerShell)
```powershell
iwr -useb https://raw.githubusercontent.com/jagtesh/guise/main/install.ps1 | iex
```

### From Source (Go)
If you have Go installed:

```bash
go install github.com/jagtesh/guise@latest
```

## Usage

Run the tool:

```bash
guise
```

### Controls

- **Arrows (↑/↓)**: Navigate lists.
- **Arrows (←/→)**: Switch between Provider list and Profile list.
- **Enter**: Select Provider / Activate Profile.
- **n**: New Profile (Captures current config).
- **d**: Delete Profile.
- **q**: Quit.

## Configuration

Guise stores its data in `~/.guise`.
- `config.json`: Maps providers to their target directories.
- `store/`: Contains the backed-up configuration files for each profile.

### Supported Defaults
Guise comes pre-configured for:
- OpenAI Codex
- Google Gemini
- Anthropic Claude
- GitHub Copilot CLI

You can add custom providers by editing `~/.guise/config.json`.

## License
BSD 3-Clause
