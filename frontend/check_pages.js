const fs = require('fs');
const glob = require('glob'); // use standard fs for simplicity
const path = require('path');

const dir = './src/app/(protected)';

function walk(directory) {
  let results = [];
  const list = fs.readdirSync(directory);
  list.forEach(file => {
    const filePath = path.join(directory, file);
    const stat = fs.statSync(filePath);
    if (stat && stat.isDirectory()) {
      results = results.concat(walk(filePath));
    } else if (file === 'page.tsx') {
      results.push(filePath);
    }
  });
  return results;
}

const pages = walk(dir);

pages.forEach(p => {
  const content = fs.readFileSync(p, 'utf8');
  const paddingMatch = content.match(/padding:\s*['"]?([^'"}]+)['"]?/);
  const classMatch = content.match(/className="([^"]*p-[^"]*)"/);
  const titleMatch = content.match(/<h1[^>]*>(.*?)<\/h1>/);
  
  const h1Content = titleMatch ? titleMatch[1].trim() : 'NO H1';
  let padding = 'NONE';
  if (classMatch) {
     padding = classMatch[1];
  } else if (paddingMatch) {
     padding = paddingMatch[1];
  }
  
  console.log(`\nFile: ${p}`);
  console.log(`Padding/Class: ${padding}`);
  console.log(`Title: ${h1Content.substring(0, 80)}`);
});
