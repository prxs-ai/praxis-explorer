import type { Config } from 'tailwindcss'

const config: Config = {
  content: [
    './pages/**/*.{js,ts,jsx,tsx,mdx}',
    './components/**/*.{js,ts,jsx,tsx,mdx}',
    './app/**/*.{js,ts,jsx,tsx,mdx}',
  ],
  theme: {
    extend: {
      colors: {
        'prxs-black': '#000000',
        'prxs-black-secondary': '#010101',
        'prxs-orange': '#FF8562',
        'prxs-cyan': '#96EEEA',
        'prxs-blue': '#9393FF',
        'prxs-gray-dark': '#555555',
        'prxs-gray': '#838383',
        'prxs-gray-light': '#959595',
        'prxs-charcoal': '#3C3C3C',
      },
      fontFamily: {
        'sans': ['Lato', 'Arial', 'sans-serif'],
      },
      borderRadius: {
        'pill': '30px',
      },
      animation: {
        'fade-in': 'fadeIn 0.5s ease-in-out',
        'slide-up': 'slideUp 0.3s ease-out',
        'pulse-slow': 'pulse 3s infinite',
      },
      keyframes: {
        fadeIn: {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' },
        },
        slideUp: {
          '0%': { transform: 'translateY(10px)', opacity: '0' },
          '100%': { transform: 'translateY(0)', opacity: '1' },
        }
      }
    },
  },
  plugins: [],
}
export default config
