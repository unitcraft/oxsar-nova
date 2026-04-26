/* eslint-env node */
module.exports = {
  root: true,
  env: { browser: true, es2022: true },
  parser: '@typescript-eslint/parser',
  parserOptions: { ecmaVersion: 'latest', sourceType: 'module', project: './tsconfig.json' },
  plugins: ['@typescript-eslint', 'react', 'react-hooks', 'import', 'jsx-a11y'],
  extends: [
    'eslint:recommended',
    'plugin:@typescript-eslint/recommended-type-checked',
    'plugin:react/recommended',
    'plugin:react-hooks/recommended',
    'plugin:jsx-a11y/recommended',
    'plugin:import/recommended',
    'plugin:import/typescript',
    'prettier',
  ],
  settings: { react: { version: 'detect' } },
  rules: {
    'react/react-in-jsx-scope': 'off',
    '@typescript-eslint/no-explicit-any': 'error',
    '@typescript-eslint/consistent-type-imports': 'error',
    'no-console': ['error', { allow: ['warn', 'error'] }],
    'import/no-default-export': 'error',
    'no-restricted-syntax': [
      'error',
      {
        // Кириллица в строковых литералах: 'Текст' или "Текст"
        selector: 'Literal[value=/[А-Яа-яЁё]/]',
        message: 'Hardcoded Cyrillic string — use t() from useTranslation instead.',
      },
      {
        // Кириллица в шаблонных литералах: `Текст ${x}`
        selector: 'TemplateLiteral > TemplateElement[value.raw=/[А-Яа-яЁё]/]',
        message: 'Hardcoded Cyrillic in template literal — use t() from useTranslation instead.',
      },
      {
        // Кириллица как JSX-текст: <div>Текст</div>
        selector: 'JSXText[value=/[А-Яа-яЁё]/]',
        message: 'Hardcoded Cyrillic JSX text — use {t()} from useTranslation instead.',
      },
    ],
  },
  overrides: [
    {
      files: ['**/routes/**/*.tsx', 'vite.config.ts'],
      rules: { 'import/no-default-export': 'off' },
    },
    {
      // i18n-модуль, тесты и spec-файлы освобождены от правила
      files: ['**/i18n/**', '**/*.test.ts', '**/*.test.tsx', '**/*.spec.ts', '**/*.spec.tsx'],
      rules: { 'no-restricted-syntax': 'off' },
    },
  ],
};
