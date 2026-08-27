import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  // PGlite ships WASM and a .tar.gz extension bundle that the bundler cannot
  // rewrite; keeping it external lets Node resolve those files from disk.
  serverExternalPackages: ["@electric-sql/pglite"],
  allowedDevOrigins: ["127.0.0.1"],
  experimental: {},
};

export default nextConfig;
