# Contributing to Resilver

Thanks for considering contributing to Resilver! All contributions, improvements, and suggestions are welcome!

## Project structure

```
resilver/
├── cmd/resilver/         # Entry point
├── internal/             # Core Go packages
│   ├── config/           # JSON config loader
│   ├── server/           # HTTP server and API routes
│   ├── news/             # News RSS fetcher
│   ├── weather/          # Weather client
│   └── update/           # Self-update mechanism
├── web/                  # Frontend assets (embedded at build time)
│   ├── css/              # Stylesheets
│   ├── font/             # Typefaces
│   └── js/
│       ├── app.js        # Grid layout and widget loader
│       └── widgets/      # One directory per widget
│           ├── clock/
│           ├── weather/
│           └── news/
├── e2e/                  # Playwright end-to-end tests
├── config.json           # Resilver configuration
├── Makefile              # Build, test, lint targets, etc
├── Containerfile         # Podman container definition
└── compose.yaml          # Podman Compose for local dev
```

## Prerequisites

- Go 1.26.2+
- Node.js 24+ (for e2e tests only)
- Podman (optional, for container builds)

## Getting started

```bash
git clone https://github.com/digibituk/resilver.git
cd resilver
make run
```

This builds the binary and starts the server on `http://localhost:8080`.

## Makefile targets

| Target           | What it does                         |
| ---------------- | ------------------------------------ |
| `make build`     | Compile the binary to `bin/resilver` |
| `make run`       | Build and run                        |
| `make test`      | Run all Go unit tests                |
| `make lint`      | Run `go vet`                         |
| `make test-e2e`  | Run Playwright end-to-end tests      |
| `make test-all`  | Unit tests + e2e tests               |
| `make container` | Build a Podman container image       |

## How it all fits together

Resilver is a single Go binary that serves a static frontend. The frontend fetches its configuration from `/api/config` and dynamically loads widgets based on what's defined in the layout.

Each widget is a [Web Component](https://developer.mozilla.org/en-US/docs/Web/API/Web_components) (a custom HTML element). The grid layout in `app.js` loads each widget's script on demand and renders `<resilver-{name}>` elements into the page. Widget config is passed via a `data-config` attribute.

On the backend, widgets that need server-side data, get their own API route and cached client in `internal/`. Purely client-side widgets (like the clock) don't need any backend code at all.

## Adding a new widget

### 1. Create the frontend widget

Add a new directory under `web/js/widgets/` with a single JS file:

```
web/js/widgets/yourwidget/yourwidget.js
```

Your widget must be a custom element that follows this pattern:

```javascript
class ResilverYourwidget extends HTMLElement {
  connectedCallback() {
    const cfg = JSON.parse(this.dataset.config || "{}");
    // Build your UI here
  }

  disconnectedCallback() {
    // Clean up intervals, listeners, etc.
  }
}

customElements.define("resilver-yourwidget", ResilverYourwidget);
```

Use Tailwind CSS classes for styling (Tailwind is available globally). Use container query units (`cqmin`) for responsive sizing so your widget scales with the grid cell.

### 2. Add a backend API (if needed)

If your widget needs server-side data fetching:

1. Create a new package under `internal/yourwidget/` with a cached client
2. Register the API route in `internal/server/server.go`
3. Include unit tests alongside the implementation (`yourwidget_test.go`)

If your widget is purely client-side, skip this step entirely. Ensure you keep your widget naming conventions consistent between the frontend and backend.

### 3. Add default configuration

Add your widget to `config.json`:

- Add an entry to `layout.widgets` to include it in the grid
- Add a `modules.yourwidget` block with sensible defaults

### 4. Add an e2e test

Create `e2e/yourwidget.spec.js` to verify your widget renders correctly. Look at the existing specs for examples.

### 5. Update the README

Add your widget to the "Current widgets" table in `README.md`.

## Code style

- **Go** — standard library only, no external dependencies. Run `make lint` before committing.
- **JavaScript** — vanilla JS, no frameworks, no build step. Widgets are Web Components.
- **CSS** — Tailwind utility classes. Custom styles go in `web/css/main.css`.
- **Tests** — every Go package has a `_test.go` file. Every widget needs an e2e spec.

## Running tests

Before opening a pull request, make sure everything passes:

```bash
make test-all      # Runs unit tests and e2e tests (e2e tests require npm ci && npx playwright install)
```

CI runs both of these automatically on every pull request.

## Submitting a pull request

1. Fork the repository and create a feature branch
2. Make your changes, keeping commits focused and descriptive
3. Make sure `make test` and `make test-e2e` pass
4. Open a pull request against `main`

Keep it simple. Resilver's whole point is being minimal, compact and self-sufficient. If a feature needs a bloated external dependency, it probably doesn't belong /.

## License

By contributing, you agree that your contributions will be licensed under the [MIT License](LICENSE).
