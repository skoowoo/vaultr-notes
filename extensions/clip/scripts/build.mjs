import * as esbuild from "esbuild";
import { fileURLToPath } from "node:url";
import { dirname, join } from "node:path";
import { copyFileSync, mkdirSync, existsSync, readdirSync, statSync } from "node:fs";

const __dirname = dirname(fileURLToPath(import.meta.url));
const root = join(__dirname, "..");
const watch = process.argv.includes("--watch");
const dist = join(root, "dist");

const common = {
  bundle: true,
  sourcemap: watch ? "inline" : false,
  minify: !watch,
  target: "chrome120",
  logLevel: "info",
  format: "iife",
  platform: "browser",
};

function copyIconsDir() {
  const srcDir = join(root, "icons");
  if (!existsSync(srcDir)) return;
  const destDir = join(dist, "icons");
  mkdirSync(destDir, { recursive: true });
  for (const name of readdirSync(srcDir)) {
    const from = join(srcDir, name);
    if (statSync(from).isFile()) {
      copyFileSync(from, join(destDir, name));
    }
  }
}

function copyAssets() {
  mkdirSync(dist, { recursive: true });
  copyIconsDir();
  copyFileSync(join(root, "manifest.json"), join(dist, "manifest.json"));
  copyFileSync(join(root, "static/popup.html"), join(dist, "popup.html"));
  copyFileSync(join(root, "static/options.html"), join(dist, "options.html"));
  const css = join(root, "static/popup.css");
  if (existsSync(css)) {
    copyFileSync(css, join(dist, "popup.css"));
  }
}

const ctx = await esbuild.context({
  ...common,
  entryPoints: {
    background: join(root, "src/background.ts"),
    content: join(root, "src/content.ts"),
    popup: join(root, "src/popup.ts"),
    options: join(root, "src/options.ts"),
  },
  outdir: dist,
});

if (watch) {
  copyAssets();
  await ctx.watch();
  console.log("watching…");
} else {
  await ctx.rebuild();
  copyAssets();
  await ctx.dispose();
}
