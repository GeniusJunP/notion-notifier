import { defineConfig } from "@pandacss/dev";

export default defineConfig({
  preflight: false,
  strictTokens: true,
  jsxFramework: "none",
  include: ["./src/**/*.{ts,svelte}"],
  outdir: "src/styled-system",
  theme: {
    extend: {
      tokens: {
        colors: {
          brand: {
            50: { value: "#f4f6f8" },
            100: { value: "#e8edf2" },
            200: { value: "#d4dce4" },
            300: { value: "#b1bcc8" },
            400: { value: "#8795a6" },
            500: { value: "#677689" },
            600: { value: "#526071" },
            700: { value: "#444f5d" },
            800: { value: "#38414c" },
            900: { value: "#272d35" },
            950: { value: "#171b20" },
          },
        },
      },
    },
  },
});
