'use client'

import { useState } from 'react'
import { SearchParams } from '@/types/agent'

interface SearchBarProps {
  onSearch: (params: SearchParams) => void
  initialParams?: SearchParams
}

export default function SearchBar({ onSearch, initialParams = {} }: SearchBarProps) {
  const [searchQuery, setSearchQuery] = useState(initialParams.q || '')
  const [showFilters, setShowFilters] = useState(false)
  const [filters, setFilters] = useState<SearchParams>({
    network: initialParams.network || '',
    capability: initialParams.capability || '',
    skill: initialParams.skill || '',
    tag: initialParams.tag || '',
    trustModel: initialParams.trustModel || '',
  })

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault()
    onSearch({ q: searchQuery, ...filters })
  }

  const handleFilterChange = (key: keyof SearchParams, value: string) => {
    const newFilters = { ...filters, [key]: value }
    setFilters(newFilters)
    onSearch({ q: searchQuery, ...newFilters })
  }

  const clearFilters = () => {
    const clearedFilters = {
      network: '',
      capability: '',
      skill: '',
      tag: '',
      trustModel: '',
    }
    setFilters(clearedFilters)
    onSearch({ q: searchQuery })
  }

  const activeFiltersCount = Object.values(filters).filter(v => v).length

  return (
    <div className="w-full space-y-4">
      <form onSubmit={handleSearch} className="relative">
        <div className="relative group">
          <input
            type="text"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            placeholder="Search agents by name, domain, or skill..."
            className="w-full pl-12 pr-32 py-4 bg-prxs-black-secondary border border-prxs-charcoal rounded-2xl text-white placeholder-prxs-gray-light focus:border-prxs-orange focus:outline-none focus:ring-2 focus:ring-prxs-orange/20 transition-all duration-200 text-lg"
          />
          <svg
            className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-prxs-gray group-focus-within:text-prxs-orange transition-colors"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
          
          <div className="absolute right-2 top-1/2 -translate-y-1/2 flex items-center gap-2">
            <button
              type="button"
              onClick={() => setShowFilters(!showFilters)}
              className={`relative px-4 py-2 rounded-lg border ${
                showFilters || activeFiltersCount > 0
                  ? 'bg-prxs-orange/10 border-prxs-orange text-prxs-orange'
                  : 'bg-prxs-charcoal/50 border-prxs-charcoal text-prxs-gray-light hover:text-white'
              } transition-all duration-200`}
            >
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 6V4m0 2a2 2 0 100 4m0-4a2 2 0 110 4m-6 8a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4m6 6v10m6-2a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4" />
              </svg>
              {activeFiltersCount > 0 && (
                <span className="absolute -top-1 -right-1 w-5 h-5 bg-prxs-orange text-white text-xs rounded-full flex items-center justify-center">
                  {activeFiltersCount}
                </span>
              )}
            </button>
            
            <button
              type="submit"
              className="px-6 py-2 bg-gradient-to-r from-prxs-orange to-[#FF6B47] text-white font-semibold rounded-lg hover:shadow-lg hover:shadow-prxs-orange/30 transition-all duration-300 transform hover:scale-105"
            >
              Search
            </button>
          </div>
        </div>
      </form>

      {showFilters && (
        <div className="animate-slide-up bg-prxs-black-secondary border border-prxs-charcoal rounded-2xl p-6 space-y-4">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-lg font-semibold text-white">Advanced Filters</h3>
            {activeFiltersCount > 0 && (
              <button
                onClick={clearFilters}
                className="text-sm text-prxs-orange hover:text-prxs-orange/80 transition-colors"
              >
                Clear all filters
              </button>
            )}
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            <div>
              <label className="block text-sm font-medium text-prxs-gray-light mb-2">
                Network
              </label>
              <select
                value={filters.network || ''}
                onChange={(e) => handleFilterChange('network', e.target.value)}
                className="w-full px-4 py-2 bg-prxs-black border border-prxs-charcoal rounded-lg text-white focus:border-prxs-orange focus:outline-none focus:ring-2 focus:ring-prxs-orange/20"
              >
                <option value="">All Networks</option>
                <option value="ethereum">Ethereum</option>
                <option value="sepolia">Sepolia</option>
                <option value="base">Base</option>
                <option value="base-sepolia">Base Sepolia</option>
                <option value="arbitrum">Arbitrum</option>
                <option value="arbitrum-sepolia">Arbitrum Sepolia</option>
              </select>
            </div>

            <div>
              <label className="block text-sm font-medium text-prxs-gray-light mb-2">
                Trust Model
              </label>
              <select
                value={filters.trustModel || ''}
                onChange={(e) => handleFilterChange('trustModel', e.target.value)}
                className="w-full px-4 py-2 bg-prxs-black border border-prxs-charcoal rounded-lg text-white focus:border-prxs-orange focus:outline-none focus:ring-2 focus:ring-prxs-orange/20"
              >
                <option value="">All Models</option>
                <option value="feedback">Feedback</option>
                <option value="reputation">Reputation</option>
                <option value="economic">Economic</option>
                <option value="hybrid">Hybrid</option>
              </select>
            </div>

            <div>
              <label className="block text-sm font-medium text-prxs-gray-light mb-2">
                Capability
              </label>
              <input
                type="text"
                value={filters.capability || ''}
                onChange={(e) => handleFilterChange('capability', e.target.value)}
                placeholder="e.g., natural-language"
                className="w-full px-4 py-2 bg-prxs-black border border-prxs-charcoal rounded-lg text-white placeholder-prxs-gray focus:border-prxs-orange focus:outline-none focus:ring-2 focus:ring-prxs-orange/20"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-prxs-gray-light mb-2">
                Skill
              </label>
              <input
                type="text"
                value={filters.skill || ''}
                onChange={(e) => handleFilterChange('skill', e.target.value)}
                placeholder="e.g., code-generation"
                className="w-full px-4 py-2 bg-prxs-black border border-prxs-charcoal rounded-lg text-white placeholder-prxs-gray focus:border-prxs-orange focus:outline-none focus:ring-2 focus:ring-prxs-orange/20"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-prxs-gray-light mb-2">
                Tag
              </label>
              <input
                type="text"
                value={filters.tag || ''}
                onChange={(e) => handleFilterChange('tag', e.target.value)}
                placeholder="e.g., ai, blockchain"
                className="w-full px-4 py-2 bg-prxs-black border border-prxs-charcoal rounded-lg text-white placeholder-prxs-gray focus:border-prxs-orange focus:outline-none focus:ring-2 focus:ring-prxs-orange/20"
              />
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
