import { PurgeCSS } from 'purgecss';

async function runPurgeCSS(htmlContent, cssContent) {
    const purgeCSSResult = await new PurgeCSS().purge({
        content: [{ raw: htmlContent, extension: 'html' }],
        css: [{ raw: cssContent }],
    });

    if (purgeCSSResult.length > 0 && purgeCSSResult[0].css) {
        console.log(purgeCSSResult[0].css);
    }
}

// Get HTML and CSS content from command line arguments
const htmlContent = process.argv[1];
const cssFilename = process.argv[2];

await runPurgeCSS(htmlContent, cssFilename);