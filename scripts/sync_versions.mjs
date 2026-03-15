import { readFile, writeFile } from "node:fs/promises";

const files = {
  npm: "npm/package.json",
  marketplace: ".claude-plugin/marketplace.json",
  plugin: ".claude-plugin/plugins/wework/.claude-plugin/plugin.json",
};

function assertVersion(version) {
  if (typeof version !== "string" || version.trim() === "") {
    throw new Error("npm/package.json is missing a version");
  }
}

async function readJSON(path) {
  return JSON.parse(await readFile(path, "utf8"));
}

async function writeJSON(path, value) {
  await writeFile(path, `${JSON.stringify(value, null, 2)}\n`);
}

const npmPackage = await readJSON(files.npm);
assertVersion(npmPackage.version);

const version = npmPackage.version;

const marketplace = await readJSON(files.marketplace);
marketplace.metadata.version = version;
if (Array.isArray(marketplace.plugins)) {
  for (const plugin of marketplace.plugins) {
    plugin.version = version;
  }
}

const plugin = await readJSON(files.plugin);
plugin.version = version;

await Promise.all([
  writeJSON(files.npm, npmPackage),
  writeJSON(files.marketplace, marketplace),
  writeJSON(files.plugin, plugin),
]);

console.log(`Synced version ${version}`);
