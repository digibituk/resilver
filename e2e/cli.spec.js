const { test, expect } = require("@playwright/test");
const { execFileSync, execSync } = require("child_process");
const fs = require("fs");
const path = require("path");
const os = require("os");

const binary = path.resolve(__dirname, "../bin/resilver");

test.describe("CLI", () => {
  test("--version prints version and exits", () => {
    const output = execFileSync(binary, ["--version"], {
      encoding: "utf-8",
    }).trim();
    expect(output).not.toBe("");
    expect(output).not.toContain("listening");
  });

  test("--version output does not contain log prefix", () => {
    const output = execFileSync(binary, ["--version"], {
      encoding: "utf-8",
    }).trim();
    expect(output).not.toMatch(/^\d{4}\/\d{2}\/\d{2}/);
  });
});

test.describe("update validation", () => {
  let devBinary;
  let tmpDir;

  test.beforeAll(() => {
    tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), "resilver-test-"));
    devBinary = path.join(tmpDir, "resilver-dev");
    execSync(
      `go build -o ${devBinary} ./cmd/resilver`,
      { cwd: path.resolve(__dirname, "..") }
    );
  });

  test.afterAll(() => {
    fs.rmSync(tmpDir, { recursive: true });
  });

  test("logs warning when update enabled but version is not semver", () => {
    const configPath = path.join(tmpDir, "config-update.json");
    fs.writeFileSync(configPath, JSON.stringify({
      server: { port: 0 },
      layout: { maxWidgets: 8, direction: "row", widgets: [] },
      modules: {},
      update: { enabled: true, intervalHours: 24 },
    }));

    // Binary starts server which blocks, so we kill it after capturing stderr
    try {
      execFileSync(devBinary, ["--config", configPath, "--port", "0"], {
        encoding: "utf-8",
        timeout: 2000,
      });
    } catch (e) {
      // timeout expected — server blocks
      const output = (e.stderr || "") + (e.stdout || "");
      expect(output).toContain("auto-update disabled");
      expect(output).toContain("not a valid semver");
    }
  });

  test("logs warning when intervalHours is less than 1", () => {
    const configPath = path.join(tmpDir, "config-interval.json");
    fs.writeFileSync(configPath, JSON.stringify({
      server: { port: 0 },
      layout: { maxWidgets: 8, direction: "row", widgets: [] },
      modules: {},
      update: { enabled: true, intervalHours: 0 },
    }));

    // Build a binary with a valid semver for this test
    const semverBinary = path.join(tmpDir, "resilver-semver");
    execSync(
      `go build -ldflags "-X main.version=v1.0.0" -o ${semverBinary} ./cmd/resilver`,
      { cwd: path.resolve(__dirname, "..") }
    );

    try {
      execFileSync(semverBinary, ["--config", configPath, "--port", "0"], {
        encoding: "utf-8",
        timeout: 2000,
      });
    } catch (e) {
      const output = (e.stderr || "") + (e.stdout || "");
      expect(output).toContain("auto-update disabled");
      expect(output).toContain("intervalHours must be at least 1");
    }
  });
});
