/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./src/**/*.{html,js}"
  ],
  theme: {
    extend: {
        colors: {
            'gray-900': "#1A1A1A",
            'blue-vallem': "#183CDD"
        },
        boxShadow: {
            'box': '0 0 10px 0 rgba(0, 0, 0, 0.1)',
        }
    },
  },
  plugins: [],
}

