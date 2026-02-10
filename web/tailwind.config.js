/** @type {import('tailwindcss').Config} */
export default {
    content: [
        "./index.html",
        "./src/**/*.{svelte,js,ts,jsx,tsx}",
    ],
    theme: {
        extend: {
            colors: {
                brand: {
                    50: '#f5f7ff',
                    100: '#ebf0ff',
                    200: '#d6e0ff',
                    300: '#b3c7ff',
                    400: '#85a2ff',
                    500: '#5c7aff',
                    600: '#3d51ff',
                    700: '#2e39eb',
                    800: '#272ec2',
                    900: '#262d98',
                    950: '#171a58',
                },
            }
        },
    },
    plugins: [],
}
