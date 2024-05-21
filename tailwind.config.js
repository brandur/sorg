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
                proseBody: '#374151', // --tw-prose-body
                proseLinks: '#111827', // --tw-prose-links
                proseInvertBody: '#d1d5db', // --tw-prose-invert-body
                proseInvertLinks: '#fff', // --tw-prose-invert-links
                // darkBorder: colors["border-slate-800"],
            },
            fontFamily: {
            },
        }
    },
    plugins: [
        require( '@tailwindcss/typography' )
    ],
}
