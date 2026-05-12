import js from "@eslint/js";
import tsPlugin from "@typescript-eslint/eslint-plugin";
import tsParser from "@typescript-eslint/parser";
import simpleImportSort from "eslint-plugin-simple-import-sort";
import svelte from "eslint-plugin-svelte";
import tailwind from "eslint-plugin-tailwindcss";
import globals from "globals";
import svelteParser from "svelte-eslint-parser";

export default [
  js.configs.recommended,
  ...svelte.configs["flat/recommended"],
  ...tailwind.configs["flat/recommended"],
  ...tsPlugin.configs["flat/strict"],
  {
    plugins: {
      "simple-import-sort": simpleImportSort,
    },
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
      "tailwindcss/enforces-shorthand": "warn",
      "tailwindcss/no-contradicting-classname": "error",
      "simple-import-sort/imports": "error",
      "simple-import-sort/exports": "error",
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
      "svelte/valid-compile": "error",
      "@typescript-eslint/no-invalid-void-type": "off",
      "no-restricted-syntax": [
        "warn",
        {
          selector: "SvelteAttribute[key.name='class'] Literal[value=/(^|\\s)(text-(xs|sm|base|lg|xl|2xl|3xl|4xl)|font-(bold|semibold|medium))(\\s|$)/]",
          message: "フォントサイズやウェイトの直書きは避け、<Typography> コンポーネントの variant を使用してください。"
        }
      ]
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
      }],
      "@typescript-eslint/no-invalid-void-type": "off",
      "@typescript-eslint/no-non-null-assertion": "off"
    },
  },
  {
    ignores: ["dist/**", "node_modules/**", "build/**", "**/*.html"],
  }
];
