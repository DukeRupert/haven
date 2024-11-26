/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./view/**/*.{templ,tpl,html}"],
  theme: {
    extend: {
      colors: {
        'picton-blue': {
          '50': '#f0faff',
          '100': '#e0f5fe',
          '200': '#bae8fd',
          '300': '#7dd5fc',
          '400': '#38bcf8',
          '500': '#0ea5e9',
          '600': '#028ac7',
          '700': '#0370a1',
          '800': '#075e85',
          '900': '#0c506e',
          '950': '#083549',
        },
      },
      keyframes: {
        'fade-in': {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' },
        },
        'fade-out': {
          '0%': { opacity: '1' },
          '100%': { opacity: '0' },
        }
      },
      animation: {
        'fade-in': 'fade-in 0.3s ease-in',
        'fade-out': 'fade-out 0.3s ease-out'
      }
    },
  },
  plugins: []
}

