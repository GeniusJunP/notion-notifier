import js from "@eslint/js";
import svelte from "eslint-plugin-svelte";
import svelteParser from "svelte-eslint-parser";
import tsParser from "@typescript-eslint/parser";
import tailwind from "eslint-plugin-tailwindcss";
import globals from "globals";

export default [
  js.configs.recommended,
  ...svelte.configs["flat/recommended"],
  ...tailwind.configs["flat/recommended"],
  {
    languageOptions: {
      globals: {
        ...globals.browser,
        ...globals.node,
      },
    },
  },
  {
    files: ["**/*.svelte", "**/*.html", "**/*.ts", "**/*.tsx", "**/*.js", "**/*.jsx"],
    rules: {
      "tailwindcss/no-custom-classname": "error",
    }
  },
  {
    files: ["**/*.svelte"],
    languageOptions: {
      parser: svelteParser,
      parserOptions: {
        parser: tsParser,
      },
    },
  },
  {
    files: ["**/*.ts"],
    languageOptions: {
      parser: tsParser,
    },
  },
  {
    ignores: ["dist/**", "node_modules/**", "build/**"],
  }
];
