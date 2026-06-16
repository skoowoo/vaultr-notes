import esbuild from 'esbuild';
import { fileURLToPath } from 'url';
import { dirname, join } from 'path';

const __dirname = dirname(fileURLToPath(import.meta.url));
const outFile = join(__dirname, '../../internal/server/static/editor.js');

await esbuild.build({
  entryPoints: [join(__dirname, 'src/index.js')],
  bundle: true,
  format: 'esm',
  outfile: outFile,
  minify: true,
  target: ['chrome120'],
  treeShaking: true,
});

console.log('built →', outFile);
