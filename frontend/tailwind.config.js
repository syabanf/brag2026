/** @type {import('tailwindcss').Config} */
export default {
  content: ["./index.html", "./src/**/*.{js,ts,jsx,tsx}"],
  theme: {
    extend: {
      colors: {
        brand: {
          50: "#fff1f2",
          100: "#ffe4e6",
          // The step between the tint and the full red. Three places already
          // asked for it, and without it here those classes did nothing.
          200: "#fecdd3",
          500: "#e11d2e",
          600: "#c8102e",
          700: "#a60b24",
          900: "#520713",
        },
        ember: "#ff5a1f",
        ink: "#171923",
        muted: "#687082",
      },
      boxShadow: {
        glass: "0 22px 60px rgba(80, 0, 18, 0.10)",
        lift: "0 14px 36px rgba(23, 25, 35, 0.10)",
      },
      fontFamily: {
        sans: ['"DM Sans"', "-apple-system", "BlinkMacSystemFont", "Segoe UI", "sans-serif"],
      },
    },
  },
  plugins: [],
};
