/**
 * Real-world test cases for web page clipping
 * Add problematic URLs here to ensure they are handled correctly
 */

export interface WebsiteTestCase {
    url: string;
    description: string;
    expectations: {
        titleContains?: string;
        mustContain?: string[];
        mustNotContain?: string[];
        imageCount?: number;
        tableCount?: number;
        codeBlockCount?: number;
    };
}

export const WEBSITE_TEST_CASES: WebsiteTestCase[] = [
    {
        url: 'https://www.mattkeeter.com/blog/2026-04-05-tailcall/',
        description: 'Matt Keeter blog post with SVG object embed and performance tables',
        expectations: {
            titleContains: 'tail-call interpreter',
            mustContain: [
                // SVG image
                '![',
                'uxn.svg',
                'https://www.mattkeeter.com/blog/2026-04-05-tailcall/uxn.svg',
                // Tables
                '| Fibonacci | Mandelbrot |',
                '| VM |',
                '| Assembly |',
                '| Tailcall |',
                // Code blocks
                '```',
                'fn run',
            ],
            mustNotContain: [
                '<object',
                '<img',
                '<table',
            ],
        },
    },
    {
        url: 'https://www.viktorcessan.com/the-economics-of-software-teams/',
        description: 'Viktor Cessan article with interactive calculators and canvas chart',
        expectations: {
            titleContains: 'economics',
            mustContain: [
                // Core article prose that must survive extraction
                'software team',
                // Input values must be preserved (inlineInputValues)
                // The default team size is hardcoded in the HTML
                '130',   // cost-per-engineer default value appears somewhere
                // No raw HTML artefacts
            ],
            mustNotContain: [
                '<table',
                '<input',
                '<canvas',
                '<div',
            ],
        },
    },
    {
        url: 'https://www.tobyord.com/writing/hourly-costs-for-ai-agents',
        description: 'Toby Ord Squarespace article with intro paragraphs in low-scoring block and images in button wrappers',
        expectations: {
            titleContains: 'Costs of AI Agents',
            mustContain: [
                // Intro paragraphs (in a separate sqs-row that Readability would drop without the fix)
                'extremely important question',
                'METR',
                // Images (inside <button> lightbox wrappers that Readability strips without the fix)
                '![',
                'squarespace-cdn.com',
            ],
            mustNotContain: [
                '<button',
                '<img',
            ],
        },
    },
    {
        url: 'https://unsung.aresluna.org/plain-text-has-been-around-for-decades-and-its-here-to-stay/',
        description: 'Ares Luna article with inline <video> inside <figure> using relative src and poster',
        expectations: {
            titleContains: 'plain text',
            mustContain: [
                // Video must be preserved as a thumbnail link with resolved absolute URLs
                '[![Video](https://unsung.aresluna.org/_media/plain-text-has-been-around-for-decades-and-its-here-to-stay/2-thumbnail.avif)](https://unsung.aresluna.org/_media/plain-text-has-been-around-for-decades-and-its-here-to-stay/2.1088w.mp4)',
            ],
            mustNotContain: [
                '<video',
                '<figure',
            ],
        },
    },
    // Add more test cases here as you discover problematic websites
    // Example:
    // {
    //   url: 'https://example.com/article',
    //   description: 'Article with lazy-loaded images',
    //   expectations: {
    //     titleContains: 'Example Article',
    //     mustContain: ['![', 'image.jpg'],
    //   },
    // },
];
