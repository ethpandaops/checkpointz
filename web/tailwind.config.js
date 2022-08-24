/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./src/**/*.{js,jsx,ts,tsx}",
  ],
  theme: {
    extend: {
      keyframes: {
        'fade': {
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
        'fade': 'fade 0.5s ease-in',
        'spin-slow': 'spin 4s linear infinite',
        'spin-slower': 'spin 6s linear infinite',
      },
      colors: {
        'scooter': {
            '50': '#f2fdff', 
            '100': '#e6fbfe', 
            '200': '#c0f6fe', 
            '300': '#9bf0fd', 
            '400': '#4fe5fb', 
            '500': '#04daf9', 
            '600': '#04c4e0', 
            '700': '#03a4bb', 
            '800': '#028395', 
            '900': '#026b7a'
        }
      }
    },
  },
  plugins: [
    require('@tailwindcss/typography'),
  ],
}
