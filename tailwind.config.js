const defaultTheme = require( 'tailwindcss/defaultTheme' )

/** @type {import('tailwindcss').Config} */
module.exports = {
    content: [
        "./content/markdown/**/*.md",
        "./layouts/**/*.{html,js}",
        "./pages/**/*.{html,js}",
        "./views/_*.ace",
        "./views/**/*.{html,js}"
    ],
    theme: {
        extend: {
            fontFamily: {
            },
        }
    },
    plugins: [
        require( '@tailwindcss/typography' )
    ],
}
