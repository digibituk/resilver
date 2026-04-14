# Resilver

Smart mirror dashboard. One binary, no dependencies, no headaches.

## Why Resilver?

Most smart mirror dashboards require Node.js, dozens of dependencies, and a lengthy setup process. Resilver takes a different approach. It's a compact single ~7 MB statically compiled binary that you download and run. No package manager, no dependencies, and sensible defaults out of the box. It's designed to be simple, reliable, and secure without the constant maintenance concerns.

### How it compares

|               | Resilver              | MagicMirror²              |
| ------------- | --------------------- | ------------------------- |
| Install       | Download one binary   | Node.js + npm install     |
| Dependencies  | 0                     | 400+                      |
| Binary size   | ~7 MB                 | ~200 MB installed         |
| Configuration | Single JSON file      | JS config + per-module JS |
| Updates       | Built-in self-update  | Manual git pull + npm     |
| Privacy       | Fully local, no cloud | Fully local, no cloud     |
| Language      | Go                    | Node.js                   |

MagicMirror² is a great project that pioneered the space. Resilver is for anyone who wants a similar result with less maintenance and fewer moving parts.

## What you get

- **Zero dependencies** — nothing to install, nothing to break on update
- **Privacy-first** — no accounts, no cloud, no telemetry. Everything runs locally
- **Self-updating** — automatic updates with checksum verification
- **Container-ready** — ships with a Containerfile and compose.yaml for Podman
- **Fully configurable** — one JSON file controls everything

## Current widgets

| Widget  | Description                                              |
| ------- | -------------------------------------------------------- |
| Clock   | 12h/24h format with optional seconds and date            |
| Weather | Current conditions via Open-Meteo. No API key needed     |
| News    | Multi-source RSS aggregator with images and auto-cycling |

More widgets are planned. Want to build your own? See [Contributing](#contributing).

## Requirements

- Any modern browser (Chromium recommended for kiosk mode)
- A linux-based OS (Raspberry Pi OS, Ubuntu, etc.)
- Go 1.26.2+ (only if building from source)

## Getting started

### 1. Download and run

```bash
curl -LO https://github.com/digibituk/resilver/releases/latest/download/resilver-linux-arm64
chmod +x resilver-linux-arm64
./resilver-linux-arm64
```

Open `http://localhost:8080` in a browser.

### 2. Set up kiosk mode

For a true smart mirror experience, launch Chromium in kiosk mode so the dashboard fills the entire screen with no browser UI:

```bash
chromium-browser --kiosk --noerrdialogs --disable-infobars http://localhost:8080
```

On a Raspberry Pi, add this to your autostart so it launches on boot. Pair it with a two-way mirror and a monitor, and you have a fully self-contained smart mirror.

### Alternative: Podman

```bash
podman compose up
```

### Alternative: Build from source

```bash
git clone https://github.com/digibituk/resilver.git
cd resilver
make run
```

## Configuration

Resilver works out of the box with sensible defaults. To customise, copy `config.json` from the source and pass it at startup:

```bash
./resilver --config /path/to/config.json
```

| Section   | What it controls                                                                                                                                                          |
| --------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `server`  | Port the dashboard is served on                                                                                                                                           |
| `layout`  | Configure grid layout order, direction (`row`/`column`), and max widget count. Widgets render in array order; last widget in an odd count auto-spans the remaining space. |
| `modules` | Per-widget settings. Each widget can be configured individually.                                                                                                          |
| `update`  | Self-update configuration. On by default.                                                                                                                                 |

## Contributing

Resilver is open source and contributions are welcome!

## License

[MIT](LICENSE)
