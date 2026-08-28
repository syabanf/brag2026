import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  // Emits .next/standalone with a self-contained server and only the traced
  // dependencies, which keeps the runtime image small.
  output: "standalone",
  // PGlite ships WASM and a .tar.gz extension bundle that the bundler cannot
  // rewrite; keeping it external lets Node resolve those files from disk.
  serverExternalPackages: ["@electric-sql/pglite"],
  allowedDevOrigins: ["127.0.0.1"],
  experimental: {},
};

export default nextConfig;
