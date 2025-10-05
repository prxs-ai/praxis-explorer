'use client'

import { useState, useEffect, useCallback } from 'react'
import Link from 'next/link'
import Header from '@/components/Header'
import SearchBar from '@/components/SearchBar'
import AgentCard from '@/components/AgentCard'
import AddAgent from '@/components/AddAgent'
import { searchAgents } from '@/lib/api'
import { AgentRow, SearchParams } from '@/types/agent'

export default function HomePage() {
  const [agents, setAgents] = useState<AgentRow[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [searchParams, setSearchParams] = useState<SearchParams>({})
  const [nextCursor, setNextCursor] = useState<string | undefined>()
  const [loadingMore, setLoadingMore] = useState(false)
  const [showAddForm, setShowAddForm] = useState(false)

  const fetchAgents = useCallback(async (params: SearchParams, append = false) => {
    try {
      if (!append) {
        setLoading(true)
        setError(null)
      } else {
        setLoadingMore(true)
      }

      const response = await searchAgents({ ...params, limit: 12 })
      
      if (append) {
        setAgents(prev => [...prev, ...response.items])
      } else {
        setAgents(response.items)
      }
      
      setNextCursor(response.nextCursor)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch agents')
    } finally {
      setLoading(false)
      setLoadingMore(false)
    }
  }, [])

  useEffect(() => {
    fetchAgents({})
  }, [fetchAgents])

  const handleSearch = (params: SearchParams) => {
    setSearchParams(params)
    fetchAgents(params)
  }

  const loadMore = () => {
    if (nextCursor && !loadingMore) {
      fetchAgents({ ...searchParams, cursor: nextCursor }, true)
    }
  }

  return (
    <div className="min-h-screen">
      <Header />
      
      <main className="pt-32 pb-20">
        <div className="section-container">
          <div className="text-center mb-12 animate-fade-in">
            <h1 className="text-5xl md:text-6xl font-bold text-white mb-4">
              Discover <span className="text-gradient">AI Agents</span>
            </h1>
            <p className="text-xl text-prxs-gray-light max-w-2xl mx-auto">
              Explore the Praxis ecosystem of ERC-8004 agents with advanced search and filtering
            </p>
          </div>

          <div className="max-w-4xl mx-auto mb-12">
            <SearchBar onSearch={handleSearch} initialParams={searchParams} />
          </div>

          {loading && !agents.length ? (
            <div className="flex items-center justify-center py-20">
              <div className="relative">
                <div className="w-16 h-16 border-4 border-prxs-charcoal border-t-prxs-orange rounded-full animate-spin" />
                <div className="absolute inset-0 w-16 h-16 border-4 border-transparent border-b-prxs-cyan rounded-full animate-spin animate-reverse" />
              </div>
            </div>
          ) : error ? (
            <div className="text-center py-20">
              <div className="inline-flex items-center gap-3 px-6 py-3 bg-red-500/10 border border-red-500/20 rounded-lg">
                <svg className="w-5 h-5 text-red-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                <span className="text-red-500">{error}</span>
              </div>
            </div>
          ) : agents.length === 0 ? (
            <div className="text-center py-20">
              <div className="inline-flex flex-col items-center gap-4">
                <div className="w-24 h-24 bg-prxs-charcoal/30 rounded-full flex items-center justify-center">
                  <svg className="w-12 h-12 text-prxs-gray" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4" />
                  </svg>
                </div>
                <div>
                  <h3 className="text-xl font-semibold text-white mb-2">No agents found</h3>
                  <p className="text-prxs-gray-light mb-4">Try adjusting your search filters or register a new agent</p>
                  
                  <button
                    onClick={() => setShowAddForm(!showAddForm)}
                    className="inline-flex items-center gap-2 px-4 py-2 bg-prxs-orange/90 hover:bg-prxs-orange text-black font-medium rounded-lg transition-all"
                  >
                    <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
                    </svg>
                    {showAddForm ? 'Cancel' : 'Register Agent'}
                  </button>
                </div>
              </div>
              
              {showAddForm && (
                <div className="mt-8 animate-fade-in">
                  <AddAgent 
                    registry='0xeFbcfaB3547EF997A747FeA1fCfBBb2fd3912445' 
                    onSuccess={() => {
                      setShowAddForm(false);
                      fetchAgents({});
                    }}
                  />
                </div>
              )}
            </div>
          ) : (
            <>
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                {agents.map((agent, index) => (
                  <div
                    key={`${agent.chainId}-${agent.agentId}-${index}`}
                    className="animate-fade-in"
                    style={{ animationDelay: `${index * 50}ms` }}
                  >
                    <AgentCard agent={agent} />
                  </div>
                ))}
              </div>

              {nextCursor && (
                <div className="text-center mt-12">
                  <button
                    onClick={loadMore}
                    disabled={loadingMore}
                    className="btn-secondary inline-flex items-center gap-2"
                  >
                    {loadingMore ? (
                      <>
                        <div className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                        Loading...
                      </>
                    ) : (
                      <>
                        Load More
                        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
                        </svg>
                      </>
                    )}
                  </button>
                </div>
              )}

              {/* Add Agent Section */}
              <div className="text-center mt-12">
                <button
                  onClick={() => setShowAddForm(!showAddForm)}
                  className="inline-flex items-center gap-2 px-6 py-3 bg-prxs-orange/10 hover:bg-prxs-orange/20 border border-prxs-orange/30 hover:border-prxs-orange/50 text-prxs-orange font-medium rounded-lg transition-all"
                >
                  <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
                  </svg>
                  {showAddForm ? 'Cancel Registration' : 'Register New Agent'}
                </button>
                
                {showAddForm && (
                  <div className="mt-8 animate-fade-in">
                    <div className="max-w-md mx-auto bg-prxs-black-secondary border border-prxs-charcoal rounded-2xl p-6">
                      <AddAgent 
                        registry='0xeFbcfaB3547EF997A747FeA1fCfBBb2fd3912445' 
                        onSuccess={() => {
                          setShowAddForm(false);
                          fetchAgents({});
                        }}
                      />
                    </div>
                  </div>
                )}
              </div>
            </>
          )}

          <div className="mt-20 text-center">
            <div className="inline-flex items-center gap-4 text-sm text-prxs-gray">
              <span>Found {agents.length} agents</span>
              {nextCursor && <span>•</span>}
              {nextCursor && <span>More available</span>}
            </div>
          </div>
        </div>
      </main>

      <footer className="border-t border-prxs-charcoal py-8">
        <div className="section-container">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <span className="text-prxs-gray-light">Powered by</span>
              <Link
                href="https://prxs.ai"
                target="_blank"
                rel="noopener noreferrer"
                className="font-bold text-prxs-orange hover:text-prxs-orange/80 transition-colors"
              >
                PRXS.AI
              </Link>
            </div>
            <div className="flex items-center gap-6">
              <Link href="/docs" className="text-prxs-gray-light hover:text-white transition-colors">
                API Docs
              </Link>
              <Link href="https://github.com/prxs" target="_blank" className="text-prxs-gray-light hover:text-white transition-colors">
                GitHub
              </Link>
            </div>
          </div>
        </div>
      </footer>
    </div>
  )
}