import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
  server: {
    host: "0.0.0.0",
    strictPort: true,
    port: 5173,
    allowedHosts: true,
    proxy: {
      "/code": {
        target: "http://localhost:8080",
        changeOrigin: true
      },
      "/api": {
        target: "http://localhost:8080",
        changeOrigin: true
      }
    }
  }
});
