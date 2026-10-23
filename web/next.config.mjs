/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  // Proxy API requests to backend Go server running on :8080 during development
  async rewrites() {
    return [
      {
        source: '/api/:path*',
        destination: 'http://127.0.0.1:8080/api/:path*',
      },
      {
        source: '/health/:path*',
        destination: 'http://127.0.0.1:8080/health/:path*',
      },
    ];
  },
};

export default nextConfig;
