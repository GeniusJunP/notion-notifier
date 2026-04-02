/** @type {import('tailwindcss').Config} */
export default {
    content: [
        "./index.html",
        "./src/**/*.{svelte,js,ts,jsx,tsx}",
    ],
    darkMode: 'class',
    theme: {
        extend: {
            colors: {
                brand: {
                    50: '#f4f6f8',
                    100: '#e8edf2',
                    200: '#d4dce4',
                    300: '#b1bcc8',
                    400: '#8795a6',
                    500: '#677689',
                    600: '#526071',
                    700: '#444f5d',
                    800: '#38414c',
                    900: '#272d35',
                    950: '#171b20',
                },
            }
        },
    },
    plugins: [],
}
