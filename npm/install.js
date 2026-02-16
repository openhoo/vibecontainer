#!/usr/bin/env node
"use strict";

const https = require("https");
const fs = require("fs");
const path = require("path");
const crypto = require("crypto");
const { version } = require("./package.json");

const PLATFORM_MAP = {
  darwin: "darwin",
  linux: "linux",
  win32: "windows",
};

const ARCH_MAP = {
  x64: "amd64",
  arm64: "arm64",
};

function getBinaryName() {
  const platform = PLATFORM_MAP[process.platform];
  const arch = ARCH_MAP[process.arch];

  if (!platform || !arch) {
    throw new Error(
      `Unsupported platform: ${process.platform} ${process.arch}`
    );
  }

  const ext = process.platform === "win32" ? ".exe" : "";
  return `vibecontainer-${platform}-${arch}${ext}`;
}

function getBinaryPath() {
  const ext = process.platform === "win32" ? ".exe" : "";
  return path.join(__dirname, `vibecontainer${ext}`);
}

function fetch(url) {
  return new Promise((resolve, reject) => {
    https
      .get(url, (res) => {
        if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
          return fetch(res.headers.location).then(resolve, reject);
        }
        if (res.statusCode !== 200) {
          return reject(new Error(`HTTP ${res.statusCode} for ${url}`));
        }
        const chunks = [];
        res.on("data", (chunk) => chunks.push(chunk));
        res.on("end", () => resolve(Buffer.concat(chunks)));
        res.on("error", reject);
      })
      .on("error", reject);
  });
}

async function verifyChecksum(buffer, binaryName) {
  const checksumsUrl = `https://github.com/openhoo/vibecontainer/releases/download/v${version}/checksums.txt`;
  let checksumsData;
  try {
    checksumsData = await fetch(checksumsUrl);
  } catch {
    console.warn("Warning: could not download checksums.txt, skipping verification");
    return;
  }

  const lines = checksumsData.toString("utf8").trim().split("\n");
  let expectedHash = null;
  for (const line of lines) {
    const [hash, name] = line.trim().split(/\s+/);
    if (name === binaryName) {
      expectedHash = hash;
      break;
    }
  }

  if (!expectedHash) {
    throw new Error(`Checksum for ${binaryName} not found in checksums.txt`);
  }

  const actualHash = crypto.createHash("sha256").update(buffer).digest("hex");
  if (actualHash !== expectedHash) {
    throw new Error(
      `Checksum mismatch for ${binaryName}:\n  expected: ${expectedHash}\n  actual:   ${actualHash}`
    );
  }
}

async function main() {
  const binaryName = getBinaryName();
  const binaryPath = getBinaryPath();
  const url = `https://github.com/openhoo/vibecontainer/releases/download/v${version}/${binaryName}`;

  console.log(`Downloading ${binaryName}...`);
  const buffer = await fetch(url);

  await verifyChecksum(buffer, binaryName);

  fs.writeFileSync(binaryPath, buffer);
  fs.chmodSync(binaryPath, 0o755);

  console.log(`Installed vibecontainer to ${binaryPath}`);
}

main().catch((err) => {
  console.error(`Failed to install vibecontainer: ${err.message}`);
  process.exit(1);
});
