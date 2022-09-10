const defaultTheme = require( 'tailwindcss/defaultTheme' )

/** @type {import('tailwindcss').Config} */
module.exports = {
    content: [
        "./content/markdown/**/*.md",
        "./layouts/**/*.{html,js}",
        "./pages/**/*.{html,js}",
        "./views/_*.ace"
    ],
    theme: {
        extend: {
            fontFamily: {
                // 'sans': ['Inter var', 'Inter', ...defaultTheme.fontFamily.sans],
            },
        }
    },
    plugins: [
        require( '@tailwindcss/typography' )
    ],
}