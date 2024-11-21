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
        'slide-in-right': {
          '0%': { transform: 'translateX(100%)', opacity: '0' },
          '100%': { transform: 'translateX(0)', opacity: '1' },
        },
        'slide-out-right': {
          '0%': { transform: 'translateX(0)', opacity: '1' },
          '100%': { transform: 'translateX(100%)', opacity: '0' },
        }
      },
      animation: {
        'slide-in': 'slide-in-right 0.3s ease-out forwards',
        'slide-out': 'slide-out-right 0.3s ease-in forwards'
      }
    },
  },
  plugins: []
}

