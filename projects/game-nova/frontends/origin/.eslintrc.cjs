// ESLint-конфиг origin-фронта. По спеке плана 72 — те же правила,
// что и у nova-фронта (TypeScript strict, без default exports кроме
// route-модулей, без any/console.log/@ts-ignore).
module.exports = {
  root: true,
  env: { browser: true, es2022: true },
  extends: [
    'eslint:recommended',
    'plugin:@typescript-eslint/recommended',
    'plugin:react/recommended',
    'plugin:react-hooks/recommended',
    'plugin:jsx-a11y/recommended',
    'plugin:import/recommended',
    'plugin:import/typescript',
  ],
  parser: '@typescript-eslint/parser',
  parserOptions: { ecmaVersion: 'latest', sourceType: 'module' },
  plugins: ['@typescript-eslint', 'react', 'react-hooks', 'jsx-a11y', 'import'],
  settings: {
    react: { version: 'detect' },
    'import/resolver': {
      typescript: { project: './tsconfig.json' },
    },
  },
  rules: {
    '@typescript-eslint/no-explicit-any': 'error',
    '@typescript-eslint/ban-ts-comment': [
      'error',
      { 'ts-expect-error': 'allow-with-description', 'ts-ignore': true },
    ],
    'no-console': ['error', { allow: ['warn', 'error'] }],
    'import/no-default-export': 'error',
    'react/react-in-jsx-scope': 'off',
    'react/prop-types': 'off',
  },
  overrides: [
    {
      files: ['vite.config.ts', 'src/main.tsx'],
      rules: { 'import/no-default-export': 'off' },
    },
  ],
  ignorePatterns: ['dist', 'node_modules', 'src/api/schema.d.ts'],
};
