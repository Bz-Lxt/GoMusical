/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{ts,tsx}",
    "../frontend-creator/src/**/*.{ts,tsx}",
    "../frontend-admin/src/**/*.{ts,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        ink: "#140f0c",
        paper: "#f3e6d2",
        amber: "#c9a36a",
        rust: "#b85c38",
        moss: "#6b8f71",
      },
      fontFamily: {
        display: ["Instrument Serif", "serif"],
        sans: ["Manrope", "ui-sans-serif", "system-ui"],
      },
    },
  },
  plugins: [],
};
