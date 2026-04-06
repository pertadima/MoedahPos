import { defineConfig, globalIgnores } from 'eslint/config';
import nextVitals from 'eslint-config-next/core-web-vitals';
import nextTs from 'eslint-config-next/typescript';
import tseslint from 'typescript-eslint';

const eslintConfig = defineConfig([
  // ── Base configs ───────────────────────────────────────────────────────────
  ...nextVitals,
  ...nextTs,

  // ── TypeScript strict rules ────────────────────────────────────────────────
  ...tseslint.configs.recommended,

  // ── Project-specific overrides ─────────────────────────────────────────────
  {
    rules: {
      // ── TypeScript ─────────────────────────────────────────────────────────
      '@typescript-eslint/no-explicit-any': 'warn', // Flag `any` types
      '@typescript-eslint/no-unused-vars': [
        'error',
        {
          // No unused vars
          argsIgnorePattern: '^_',
          varsIgnorePattern: '^_',
        },
      ],
      '@typescript-eslint/consistent-type-imports': 'warn', // Use `import type`
      '@typescript-eslint/no-non-null-assertion': 'warn', // Flag ! operator
      '@typescript-eslint/no-empty-object-type': 'warn',

      // ── React ──────────────────────────────────────────────────────────────
      'react/no-unused-state': 'warn',
      'react/self-closing-comp': 'warn', // <Foo /> not <Foo></Foo>
      'react/jsx-no-useless-fragment': 'warn', // No empty <>...</>

      // ── General JS quality ─────────────────────────────────────────────────
      'no-console': ['warn', { allow: ['warn', 'error'] }], // No console.log in prod
      'no-debugger': 'error',
      'prefer-const': 'error',
      'no-var': 'error',
      eqeqeq: ['error', 'always', { null: 'ignore' }], // Always ===
      'no-duplicate-imports': 'error',
      'no-unused-expressions': 'error',
    },
  },

  // ── Ignores ────────────────────────────────────────────────────────────────
  globalIgnores([
    '.next/**',
    'out/**',
    'build/**',
    'next-env.d.ts',
    'node_modules/**',
    '*.config.*', // Skip config files
  ]),
]);

export default eslintConfig;
