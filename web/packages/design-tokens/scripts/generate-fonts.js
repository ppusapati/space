import fs from 'fs/promises';
import path from 'path';
import { fileURLToPath } from 'url';
import { createRequire } from 'module';
const require = createRequire(import.meta.url);
const { fontless } = require('fontless');

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const DIST_DIR = path.resolve(__dirname, '../dist/fonts');
const FONTS_DIR = path.resolve(__dirname, '../public/fonts');

// Google Fonts to download and self-host
const FONTS = [
  {
    name: 'Inter',
    weights: [400, 500, 600, 700, 800],
    subsets: ['latin'],
    display: 'swap',
  },
  {
    name: 'Inter Display',
    weights: [400, 500, 600, 700, 800],
    subsets: ['latin'],
    display: 'swap',
  },
  {
    name: 'Roboto Mono',
    weights: [400, 500, 700],
    subsets: ['latin'],
    display: 'swap',
  },
];

async function generateFonts() {
  try {
    // Create directories if they don't exist
    await fs.mkdir(DIST_DIR, { recursive: true });
    await fs.mkdir(FONTS_DIR, { recursive: true });

    console.log('🚀 Generating font files...');
    
    // Generate and save each font
    for (const font of FONTS) {
      console.log(`\n📝 Processing ${font.name}...`);
      
      // fontless doesn't download fonts, it just generates CSS
      // For now, let's create a simple CSS file with Google Fonts imports
      const fontFamily = font.name.toLowerCase().replace(/\s+/g, '-');
      const cssContent = `@import url('https://fonts.googleapis.com/css2?family=${font.name.replace(/\s+/g, '+')}:wght@${font.weights.join(';')}&display=${font.display}&subset=${font.subsets.join(',')}');`;

      await fs.writeFile(
        path.join(DIST_DIR, `${fontFamily}.css`),
        cssContent
      );

      console.log(`✅ Generated ${font.name} (${font.weights.length} weights)`);
    }

    console.log('\n✨ All fonts generated successfully!');
  } catch (error) {
    console.error('❌ Error generating fonts:', error);
    process.exit(1);
  }
}

generateFonts();
