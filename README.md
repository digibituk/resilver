# Resilver

Smart mirror dashboard. One binary, no dependencies, no headaches.

## Why do we need another smart mirror dashboard?

The most popular open source smart mirror project currently is [MagicMirror²](https://github.com/MagicMirrorOrg/MagicMirror). It's a fantastic project with a large community and a rich ecosystem of modules. However, it has some drawbacks. The main one being its dependency on Node.js, which can be a pain to maintain. As long-term support for older versions ends and new versions require dependencies that only exist on updated Raspberry Pi OS distros. It also depends on a large number of npm packages, which need to be updated regularly to avoid security vulnerabilities. This, in my experience, is a lot of maintenance overhead for dashboard that rarely needs changing once configured.

**Resilver** tries to take a different approach. It's a compact single ~7 MB statically compiled binary that you download and run. No package manager, no dependencies, and sensible defaults out of the box. It's designed to be minimalistic, reliable, and self sufficient without the constant need of maintenance.

### How it compares

|               | Resilver             | MagicMirror²              |
| ------------- | -------------------- | ------------------------- |
| Install       | Download one binary  | Node.js + npm install     |
| Dependencies  | 0                    | 400+                      |
| Binary size   | ~7 MB                | ~200 MB installed         |
| Configuration | Single JSON file     | JS config + per-module JS |
| Updates       | Built-in self-update | Manual git pull + npm     |
| Language      | Go                   | Node.js                   |

MagicMirror² is a awesome solution for those who want lots of module options and a great community, as long as they don't mind the maintenance overhead or potential breaking changes. Resilver is for those who crave a minimal setup with little to no maintenance, and are happy with a smaller selection of built-in widgets. If would like to contribute to expanding our widget selection, see our [contributing](#contributing) guidelines.

## What you get

- **Zero dependencies** — nothing to install, just download and run
- **Self-updating** — automatic updates with checksum verification
- **Modular design** — add new custom widgets to fit your needs
- **Dynamic layout** — responsive grid with auto-spanning
- **Fully configurable** — single source of truth for all settings, layout and widgets

## Current widgets

| Widget  | Description                                              |
| ------- | -------------------------------------------------------- |
| Clock   | 12h/24h format with optional seconds and date            |
| Weather | Current conditions via Open-Meteo. No API key needed     |
| News    | Multi-source RSS aggregator with images and auto-cycling |

More widgets comming soon but if you can't wait, and want to build your own? See our [contributing](#contributing) guidelines.

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

For a true smart mirror experience, download and launch Chromium in kiosk mode.

```bash
chromium-browser --kiosk --noerrdialogs --disable-infobars http://localhost:8080
```

On a Raspberry Pi, you can add this to your autostart so it launches on boot up.

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
| `theme`   | Theme settings, currently only supports accent color.                                                                                                                     |
| `layout`  | Configure grid layout order, direction (`row`/`column`), and max widget count. Widgets render in array order; last widget in an odd count auto-spans the remaining space. |
| `modules` | Per-widget settings. Each widget can be configured individually.                                                                                                          |
| `update`  | Self-update configuration. On by default.                                                                                                                                 |

## Contributing

Resilver is open source and any contributions are welcome!

## License

[MIT](LICENSE)
