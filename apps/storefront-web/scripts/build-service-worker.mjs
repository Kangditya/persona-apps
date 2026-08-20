import { glob, mkdir, readFile, writeFile } from "node:fs/promises";
import path from "node:path";

const appRoot = path.resolve(import.meta.dirname, "..");
const nextRoot = path.join(appRoot, ".next");
const staticRoot = path.join(nextRoot, "static");

const files = [];
for await (const file of glob("static/**/*.{css,js,woff,woff2}", {
  cwd: nextRoot,
})) {
  files.push(`/_next/${file.replaceAll(path.sep, "/")}`);
}

const buildId = (
  await readFile(path.join(nextRoot, "BUILD_ID"), "utf8")
).trim();
const precacheUrls = [
  ...new Set([
    ...files.sort(),
    "/icon.svg",
    "/icon-192.svg",
    "/icon-512.svg",
    "/offline.html",
  ]),
];
const sourcePath = path.join(appRoot, "src/pwa/service-worker.js");
const source = await readFile(sourcePath, "utf8");
if (
  !source.includes("__CACHE_NAME__") ||
  !source.includes("__PRECACHE_URLS__")
) {
  throw new Error(
    `${sourcePath} is missing a service-worker build placeholder`,
  );
}

const worker = source
  .replace(
    "__CACHE_NAME__",
    JSON.stringify(`qurban-storefront-shell-${buildId}`),
  )
  .replace("__PRECACHE_URLS__", JSON.stringify(precacheUrls));
await mkdir(staticRoot, { recursive: true });
await writeFile(path.join(staticRoot, "sw.js"), worker);
