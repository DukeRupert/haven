/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./view/**/*.{templ,tpl,html}"],
  theme: {
    extend: {
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

