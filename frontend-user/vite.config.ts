import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import { fileURLToPath } from "node:url";

const root = fileURLToPath(new URL(".", import.meta.url));

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      "@": root + "src",
      "@creator": root + "../frontend-creator/src",
      "@admin": root + "../frontend-admin/src",
      react: root + "node_modules/react",
      "react-dom": root + "node_modules/react-dom",
      "react-router-dom": root + "node_modules/react-router-dom",
    },
    dedupe: ["react", "react-dom", "react-router-dom"],
  },
  server: { port: 5173, proxy: { "/api": "http://localhost:29472" } },
});
