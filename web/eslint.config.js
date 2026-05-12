import js from "@eslint/js";
import svelte from "eslint-plugin-svelte";
import svelteParser from "svelte-eslint-parser";
import tsParser from "@typescript-eslint/parser";
import tsPlugin from "@typescript-eslint/eslint-plugin";
import tailwind from "eslint-plugin-tailwindcss";
import globals from "globals";

export default [
  js.configs.recommended,
  ...svelte.configs["flat/recommended"],
  ...tailwind.configs["flat/recommended"],
  ...tsPlugin.configs["flat/recommended"],
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
      "tailwindcss/classnames-order": "warn",
    }
  },
  {
    files: ["**/*.svelte"],
    languageOptions: {
      parser: svelteParser,
      parserOptions: {
        parser: tsParser,
        extraFileExtensions: [".svelte"],
      },
    },
    rules: {
      "no-unused-vars": "off",
      "@typescript-eslint/no-unused-vars": ["error", {
        "varsIgnorePattern": "^_|^\\$\\$Props",
        "argsIgnorePattern": "^_"
      }],
      "svelte/valid-compile": "error"
    }
  },
  {
    files: ["**/*.ts"],
    languageOptions: {
      parser: tsParser,
    },
    rules: {
      "no-unused-vars": "off",
      "@typescript-eslint/no-unused-vars": ["error", {
        "varsIgnorePattern": "^_",
        "argsIgnorePattern": "^_",
        "caughtErrorsIgnorePattern": "^_"
      }]
    },
  },
  {
    ignores: ["dist/**", "node_modules/**", "build/**", "**/*.html"],
  }
];
