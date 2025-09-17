/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  output: 'standalone',
  async rewrites() {
    return [
      {
        source: '/api/explorer/:path*',
        destination: `${process.env.API_URL || 'http://localhost:8080'}/:path*`,
      },
    ]
  },
}

module.exports = nextConfig