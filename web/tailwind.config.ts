import type { Config } from 'tailwindcss'

export default {
  content: ['./index.html', './src/**/*.{vue,ts}'],
  theme: {
    extend: {
      colors: {
        paper: '#f6efe2',
        ink: '#231813',
        brass: '#8a5a1f',
        ember: '#bc6c25',
        moss: '#606c38',
      },
      boxShadow: {
        panel: '0 24px 60px -36px rgba(74, 53, 34, 0.45)',
      },
    },
  },
  plugins: [],
} satisfies Config

