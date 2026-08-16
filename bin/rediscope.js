#!/usr/bin/env node
"use strict";

const { spawnSync } = require("node:child_process");
const fs = require("node:fs");
const path = require("node:path");

function main() {
  try {
    const root = path.resolve(__dirname, "..");
    const exe = process.platform === "win32" ? "rediscope.exe" : "rediscope";
    const platformKey = `${process.platform}-${process.arch}`;
    const candidates = [
      path.join(root, "prebuilds", platformKey, exe),
      path.join(root, "dist", exe),
    ];

    let binary = null;
    for (const candidate of candidates) {
      try {
        if (fs.existsSync(candidate)) {
          binary = candidate;
          break;
        }
      } catch {
        // Ignore filesystem check errors and continue checking next candidate
      }
    }

    if (!binary) {
      process.stderr.write("rediscope binary was not found.\n");
      process.stderr.write("Run `npm run build` for local development, or install a package with prebuilt binaries.\n");
      process.exit(1);
    }

    const result = spawnSync(binary, process.argv.slice(2), { stdio: "inherit" });

    if (result.error) {
      process.stderr.write(`rediscope process error: ${result.error.message}\n`);
      process.exit(1);
    }

    if (result.signal) {
      process.exit(128 + (result.signal === "SIGINT" ? 2 : 15));
    }

    process.exit(typeof result.status === "number" ? result.status : 0);
  } catch (err) {
    process.stderr.write(`rediscope launcher fatal error: ${err && err.message ? err.message : String(err)}\n`);
    process.exit(1);
  }
}

main();
