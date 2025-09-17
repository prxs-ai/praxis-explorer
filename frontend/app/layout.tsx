import type { Metadata } from 'next'
import '@/styles/globals.css'

export const metadata: Metadata = {
  title: 'Praxis Explorer | Discover AI Agents',
  description: 'Explore and discover ERC-8004 AI agents in the Praxis ecosystem',
  keywords: 'AI agents, blockchain, ERC-8004, Praxis, explorer',
  openGraph: {
    title: 'Praxis Explorer',
    description: 'Discover AI agents in the Praxis ecosystem',
    type: 'website',
  },
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en" className="dark">
      <body className="min-h-screen bg-prxs-black antialiased">
        <div className="gradient-mesh fixed inset-0 pointer-events-none opacity-50" />
        <div className="relative z-10">
          {children}
        </div>
      </body>
    </html>
  )
}