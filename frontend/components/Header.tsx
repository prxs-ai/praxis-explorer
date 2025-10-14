'use client'

import Link from 'next/link'
import { useState, useEffect } from 'react'
import { ConnectButton } from '@rainbow-me/rainbowkit';

export default function Header() {
  const [scrolled, setScrolled] = useState(false)

  useEffect(() => {
    const handleScroll = () => {
      setScrolled(window.scrollY > 20)
    }
    window.addEventListener('scroll', handleScroll)
    return () => window.removeEventListener('scroll', handleScroll)
  }, [])

  return (
    <header className={`fixed top-0 left-0 right-0 z-50 transition-all duration-300 ${
      scrolled ? 'bg-prxs-black/80 backdrop-blur-xl border-b border-prxs-charcoal' : 'bg-transparent'
    }`}>
      <nav className="section-container">
        <div className="flex items-center justify-between h-20">
          <Link href="/" className="flex items-center gap-3 group">
            <div className="relative">
              <div className="absolute inset-0 bg-gradient-to-r from-prxs-orange to-prxs-cyan rounded-lg blur-lg opacity-50 group-hover:opacity-100 transition-opacity" />
              <div className="relative bg-prxs-black border border-prxs-orange/20 rounded-lg p-2">
                <svg className="w-6 h-6 text-prxs-orange" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
                </svg>
              </div>
            </div>
            <div>
              <h1 className="text-2xl font-bold text-white group-hover:text-prxs-orange transition-colors">
                Praxis Explorer
              </h1>
              <p className="text-xs text-prxs-gray">Discover AI Agents</p>
            </div>
          </Link>

          <div className="flex items-center gap-6">
            <Link
              href="/"
              className="text-white hover:text-prxs-orange transition-colors font-medium"
            >
              Explore
            </Link>
            <Link
              href="/docs"
              className="text-white hover:text-prxs-cyan transition-colors font-medium"
            >
              Docs
            </Link>
            <Link
              href="https://prxs.ai"
              target="_blank"
              rel="noopener noreferrer"
              className="text-white hover:text-prxs-blue transition-colors font-medium flex items-center gap-1"
            >
              prxs.ai
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
              </svg>
            </Link>
            <ConnectButton accountStatus="address" />
          </div>
        </div>
      </nav>
    </header>
  )
}
