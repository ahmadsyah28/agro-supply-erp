/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  // Proxy API requests to backend Go server
  async rewrites() {
    let backendUrl = process.env.BACKEND_URL;
    if (!backendUrl) {
      if (process.env.NODE_ENV === 'production') {
        return [];
      }
      backendUrl = 'http://127.0.0.1:8080';
    }

    // Cegah infinite redirect loop jika salah mengisi domain Vercel sendiri
    if (backendUrl.includes('vercel.app')) {
      return [];
    }

    // Render mengharuskan HTTPS, hindari redirect 301 dari http ke https
    if (backendUrl.startsWith('http://') && backendUrl.includes('onrender.com')) {
      backendUrl = backendUrl.replace('http://', 'https://');
    }

    const cleanUrl = backendUrl.replace(/\/$/, '');
    return [
      {
        source: '/api/:path*',
        destination: `${cleanUrl}/api/:path*`,
      },
      {
        source: '/health/:path*',
        destination: `${cleanUrl}/health/:path*`,
      },
    ];
  },
};

export default nextConfig;
