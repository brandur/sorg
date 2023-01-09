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
                'inter': ['Inter var', 'Inter', ...defaultTheme.fontFamily.sans],
                // 'sans': ['Inter var', 'Inter', ...defaultTheme.fontFamily.sans],
                // 'serif': ['Cardo', ...defaultTheme.fontFamily.serif],
            },
        }
    },
    plugins: [
        require( '@tailwindcss/typography' )
    ],
}
