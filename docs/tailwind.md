# Tailwind

Install static binary:

    curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-macos-arm64
    chmod +x tailwindcss-macos-arm64
    mv tailwindcss-macos-arm64 /usr/local/bin/tailwindcss

Start watch process:

    tailwindcss -i ./content/stylesheets-modular/tailwind_base.css -o ./content/stylesheets-modular/tailwind.css --watch

Minify for production:

    tailwindcss -i ./content/stylesheets-modular/tailwind_base.css -o ./content/stylesheets-modular/tailwind.min.css --minify
