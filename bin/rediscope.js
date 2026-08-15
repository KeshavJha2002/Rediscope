#!/usr/bin/env node
const { spawnSync } = require("node:child_process");
const fs = require("node:fs");
const path = require("node:path");

const root = path.resolve(__dirname, "..");
const exe = process.platform === "win32" ? "rediscope.exe" : "rediscope";
const platformKey = `${process.platform}-${process.arch}`;
const candidates = [
  path.join(root, "prebuilds", platformKey, exe),
  path.join(root, "dist", exe),
];

const binary = candidates.find((item) => fs.existsSync(item));
if (!binary) {
  console.error("rediscope binary was not found.");
  console.error("Run `npm run build` for local development, or install a package with prebuilt binaries.");
  process.exit(1);
}

const result = spawnSync(binary, process.argv.slice(2), { stdio: "inherit" });
if (result.error) {
  console.error(result.error.message);
  process.exit(1);
}
process.exit(result.status ?? 0);
