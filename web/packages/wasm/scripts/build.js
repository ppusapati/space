#!/usr/bin/env node
/**
 * Samavaya WASM Build Script (Node.js wrapper)
 * Cross-platform build script for WASM modules
 */

import { spawn } from 'child_process';
import { platform } from 'os';
import { dirname, join } from 'path';
import { fileURLToPath } from 'url';

const __dirname = dirname(fileURLToPath(import.meta.url));
const wasmDir = join(__dirname, '..');

const args = process.argv.slice(2);
const isDev = args.includes('--dev');
const crate = args.find((arg, i) => args[i - 1] === '--crate');

const isWindows = platform() === 'win32';

let command;
let commandArgs;

if (isWindows) {
  command = 'powershell.exe';
  commandArgs = [
    '-ExecutionPolicy', 'Bypass',
    '-File', join(wasmDir, 'build.ps1'),
  ];
  if (isDev) commandArgs.push('-Dev');
  if (crate) commandArgs.push('-Crate', crate);
} else {
  command = 'bash';
  commandArgs = [join(wasmDir, 'build.sh')];
  if (isDev) commandArgs.push('--dev');
  if (crate) commandArgs.push('--crate', crate);
}

console.log(`Running: ${command} ${commandArgs.join(' ')}`);

const child = spawn(command, commandArgs, {
  cwd: wasmDir,
  stdio: 'inherit',
  shell: false,
});

child.on('error', (err) => {
  console.error('Failed to start build process:', err);
  process.exit(1);
});

child.on('close', (code) => {
  process.exit(code ?? 0);
});
