const colors = require('tailwindcss/colors')
const defaultTheme = require('tailwindcss/defaultTheme')

/** @type {import('tailwindcss').Config} */
module.exports = {
    content: [
        "./content/articles/*.md",
        "./content/drafts/*.md",
        "./content/markdown/**/*.md",
        "./layouts/**/*.{html,js}",
        "./pages/**/*.{html,js}",
        "./views/_*.ace",
        "./views/**/*.{html,js}"
    ],
    darkMode: 'selector',
    theme: {
        extend: {
            colors: {
                proseBody: '#374151',       // --tw-prose-body
                proseLinks: '#111827',      // --tw-prose-links
                proseInvertBody: '#d1d5db', // --tw-prose-invert-body
                proseInvertLinks: '#fff',   // --tw-prose-invert-links
            },
            fontFamily: {
            },
            typography: {
                DEFAULT: {
                    css: {
                        blockquote: {
                            // Disables the quotes around blockquotes that
                            // Tailwind includes by default. They look decent,
                            // but turn into a real mess if you do things like
                            // cite a source (tick appears after the source's
                            // name) or include a list.
                            quotes: "none",
                        },
                    },
                },
            },
        }
    },
    plugins: [
        require( '@tailwindcss/typography' )
    ],
}
