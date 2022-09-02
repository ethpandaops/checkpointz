/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ['./src/**/*.{js,jsx,ts,tsx}'],
  theme: {
    extend: {
      screens: {
        '2xl': '1680px',
      },
      keyframes: {
        fade: {
          '0%': {
            opacity: '1',
          },
          '50%': {
            opacity: '0',
          },
          '100%': {
            opacity: '1',
          },
        },
      },
      animation: {
        fade: 'fade 0.5s ease-in',
        'spin-slow': 'spin 4s linear infinite',
        'spin-slower': 'spin 6s linear infinite',
      },
    },
  },
  plugins: [require('@tailwindcss/typography')],
};
