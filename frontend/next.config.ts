import type { NextConfig } from 'next'

const isDev = process.env.NODE_ENV === 'development'

const config: NextConfig = {
  reactStrictMode: false,
  // dev 模式不设置 output: export，否则 rewrites 会被禁用
  ...(isDev ? {} : { output: 'export' }),
  distDir: 'dist',
  basePath: '/profile',
  trailingSlash: true,
  skipTrailingSlashRedirect: true,
  images: { unoptimized: true },
  // 仅开发时生效，将 API 请求代理到 Go 后端
  async rewrites() {
    if (!isDev) return []
    const prefixes = ['authserver', 'users', 'sessionserver', 'textures', 'api', 'minecraftservices', 'yggdrasil']
    return prefixes.map(prefix => ({
      source: `/${prefix}/:path*`,
      basePath: false,
      destination: `http://localhost:8080/${prefix}/:path*`,
    }))
  },
}
export default config
