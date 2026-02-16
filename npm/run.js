#!/usr/bin/env node
"use strict";

const { execFileSync } = require("child_process");
const path = require("path");
const fs = require("fs");

const ext = process.platform === "win32" ? ".exe" : "";
const binaryPath = path.join(__dirname, `vibecontainer${ext}`);

if (!fs.existsSync(binaryPath)) {
  console.error(
    "vibecontainer binary not found. Try running: npm rebuild @openhoo/vibecontainer"
  );
  process.exit(1);
}

try {
  execFileSync(binaryPath, process.argv.slice(2), { stdio: "inherit" });
} catch (err) {
  process.exit(err.status ?? 1);
}
