import type { Config } from "tailwindcss";

export default {
  content: ["./index.html", "./src/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        ink: "#0f172a",
        muted: "#64748b",
        line: "#e2e8f0",
        brand: {
          blue: "#2563eb",
          purple: "#7c3aed",
          violet: "#8b5cf6",
        },
      },
      boxShadow: {
        panel: "0 20px 50px rgba(15, 23, 42, 0.08)",
        soft: "0 10px 25px rgba(15, 23, 42, 0.06)",
        glow: "0 24px 54px rgba(37, 99, 235, 0.22)",
      },
      borderRadius: {
        panel: "26px",
        hero: "34px",
      },
      fontFamily: {
        sans: [
          "Inter",
          "ui-sans-serif",
          "system-ui",
          "-apple-system",
          "BlinkMacSystemFont",
          "Segoe UI",
          "sans-serif",
        ],
      },
    },
  },
  plugins: [],
} satisfies Config;
