'use client'

import { useState, useEffect } from 'react'
import { useParams } from 'next/navigation'
import Link from 'next/link'
import Header from '@/components/Header'
import { getAgent } from '@/lib/api'
import { AgentRow } from '@/types/agent'
import { formatDate, isOnline, getChainName, truncateAddress, extractSkillName, extractSkillTags } from '@/lib/utils'

export default function AgentDetailPage() {
  const rawParams = useParams() as any
  const params = (rawParams || {}) as { chainId?: string; agentId?: string }
  const [agent, setAgent] = useState<AgentRow | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [activeTab, setActiveTab] = useState<'overview' | 'skills' | 'capabilities' | 'raw'>('overview')

  useEffect(() => {
    const fetchAgent = async () => {
      try {
        setLoading(true)
        setError(null)
        const data = await getAgent(
          String(params.chainId || ''),
          String(params.agentId || '')
        )
        setAgent(data)
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to fetch agent')
      } finally {
        setLoading(false)
      }
    }

    if (params?.chainId && params?.agentId) {
      fetchAgent()
    }
  }, [params?.chainId, params?.agentId])

  if (loading) {
    return (
      <div className="min-h-screen">
        <Header />
        <div className="flex items-center justify-center min-h-[60vh]">
          <div className="relative">
            <div className="w-16 h-16 border-4 border-prxs-charcoal border-t-prxs-orange rounded-full animate-spin" />
            <div className="absolute inset-0 w-16 h-16 border-4 border-transparent border-b-prxs-cyan rounded-full animate-spin animate-reverse" />
          </div>
        </div>
      </div>
    )
  }

  if (error || !agent) {
    return (
      <div className="min-h-screen">
        <Header />
        <div className="flex items-center justify-center min-h-[60vh]">
          <div className="text-center">
            <div className="inline-flex items-center gap-3 px-6 py-3 bg-red-500/10 border border-red-500/20 rounded-lg mb-4">
              <svg className="w-5 h-5 text-red-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <span className="text-red-500">{error || 'Agent not found'}</span>
            </div>
            <div>
              <Link href="/" className="text-prxs-orange hover:text-prxs-orange/80 transition-colors">
                ← Back to Explorer
              </Link>
            </div>
          </div>
        </div>
      </div>
    )
  }

  const online = isOnline(agent.lastSeenAt)
  const agentName = agent.card?.name || agent.domain.split('.')[0]
  const description = agent.card?.description || 'No description available'
  const verified = agent.agentId > 0

  return (
    <div className="min-h-screen">
      <Header />
      
      <main className="pt-32 pb-20">
        <div className="section-container">
          <Link
            href="/"
            className="inline-flex items-center gap-2 text-prxs-gray-light hover:text-white transition-colors mb-8"
          >
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
            </svg>
            Back to Explorer
          </Link>

          <div className="bg-gradient-to-br from-prxs-black-secondary to-prxs-black border border-prxs-charcoal rounded-3xl overflow-hidden">
            <div className="relative p-8 pb-0">
              <div className="absolute inset-0 bg-gradient-to-br from-prxs-orange/10 via-transparent to-prxs-cyan/10 opacity-50" />
              
              <div className="relative z-10">
                <div className="flex items-start justify-between mb-6">
                  <div>
                    <div className="flex items-center gap-3 mb-2">
                      <h1 className="text-4xl font-bold text-white">
                        {agentName}
                      </h1>
                      {verified && (
                        <div className="flex items-center gap-1 px-3 py-1 bg-prxs-cyan/10 border border-prxs-cyan/20 rounded-full">
                          <svg className="w-4 h-4 text-prxs-cyan" fill="currentColor" viewBox="0 0 20 20">
                            <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.857-9.809a.75.75 0 00-1.214-.882l-3.483 4.79-1.88-1.88a.75.75 0 10-1.06 1.061l2.5 2.5a.75.75 0 001.137-.089l4-5.5z" clipRule="evenodd" />
                          </svg>
                          <span className="text-xs text-prxs-cyan">Verified</span>
                        </div>
                      )}
                    </div>
                    <p className="text-xl text-prxs-gray-light mb-4">{agent.domain}</p>
                    <p className="text-prxs-gray-light max-w-3xl">{description}</p>
                  </div>
                  
                  <div className="flex items-center gap-3">
                    <div className={`flex items-center gap-2 px-4 py-2 rounded-full ${
                      online ? 'bg-green-500/10 border border-green-500/20' : 'bg-prxs-charcoal/50 border border-prxs-charcoal'
                    }`}>
                      <div className={`w-2 h-2 rounded-full ${online ? 'bg-green-500 animate-pulse' : 'bg-prxs-gray'}`} />
                      <span className={`text-sm ${online ? 'text-green-500' : 'text-prxs-gray'}`}>
                        {online ? 'Online' : `Last seen ${formatDate(agent.lastSeenAt)}`}
                      </span>
                    </div>
                  </div>
                </div>

                <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-8">
                  <div className="bg-prxs-black/50 border border-prxs-charcoal rounded-xl p-4">
                    <p className="text-sm text-prxs-gray mb-1">Network</p>
                    <p className="text-lg font-semibold text-white">{getChainName(agent.chainId)}</p>
                  </div>
                  
                  <div className="bg-prxs-black/50 border border-prxs-charcoal rounded-xl p-4">
                    <p className="text-sm text-prxs-gray mb-1">Agent ID</p>
                    <p className="text-lg font-semibold text-white">#{agent.agentId}</p>
                  </div>
                  
                  <div className="bg-prxs-black/50 border border-prxs-charcoal rounded-xl p-4">
                    <p className="text-sm text-prxs-gray mb-1">Address</p>
                    <p className="text-lg font-semibold text-white font-mono">
                      {truncateAddress(agent.addressCaip10, 6)}
                    </p>
                  </div>
                  
                  {agent.scoreAvg && (
                    <div className="bg-prxs-black/50 border border-prxs-charcoal rounded-xl p-4">
                      <p className="text-sm text-prxs-gray mb-1">Score</p>
                      <div className="flex items-center gap-2">
                        <svg className="w-5 h-5 text-yellow-500" fill="currentColor" viewBox="0 0 20 20">
                          <path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z" />
                        </svg>
                        <span className="text-lg font-semibold text-white">{agent.scoreAvg.toFixed(1)}</span>
                      </div>
                    </div>
                  )}
                </div>

                <div className="flex flex-wrap gap-2 mb-8">
                  {agent.trustModels.map((model) => (
                    <span
                      key={model}
                      className="px-4 py-2 bg-gradient-to-r from-prxs-orange/10 to-prxs-cyan/10 border border-prxs-orange/20 rounded-full text-sm font-medium text-white"
                    >
                      {model}
                    </span>
                  ))}
                </div>

                <div className="flex gap-2 border-b border-prxs-charcoal">
                  {(['overview', 'skills', 'capabilities', 'raw'] as const).map((tab) => (
                    <button
                      key={tab}
                      onClick={() => setActiveTab(tab)}
                      className={`px-6 py-3 font-medium capitalize transition-all duration-200 ${
                        activeTab === tab
                          ? 'text-prxs-orange border-b-2 border-prxs-orange'
                          : 'text-prxs-gray-light hover:text-white'
                      }`}
                    >
                      {tab}
                    </button>
                  ))}
                </div>
              </div>
            </div>

            <div className="p-8">
              {activeTab === 'overview' && (
                <div className="space-y-6 animate-fade-in">
                  <div>
                    <h3 className="text-lg font-semibold text-white mb-3">Statistics</h3>
                    <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
                      <div className="bg-prxs-black/30 border border-prxs-charcoal/50 rounded-lg p-4">
                        <p className="text-sm text-prxs-gray mb-1">Validations</p>
                        <p className="text-2xl font-bold text-white">{agent.validationsCnt}</p>
                      </div>
                      <div className="bg-prxs-black/30 border border-prxs-charcoal/50 rounded-lg p-4">
                        <p className="text-sm text-prxs-gray mb-1">Feedbacks</p>
                        <p className="text-2xl font-bold text-white">{agent.feedbacksCnt}</p>
                      </div>
                      <div className="bg-prxs-black/30 border border-prxs-charcoal/50 rounded-lg p-4">
                        <p className="text-sm text-prxs-gray mb-1">Skills</p>
                        <p className="text-2xl font-bold text-white">{agent.skills.length}</p>
                      </div>
                    </div>
                  </div>

                  {agent.card?.homepage && (
                    <div>
                      <h3 className="text-lg font-semibold text-white mb-3">Links</h3>
                      <a
                        href={agent.card.homepage}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="inline-flex items-center gap-2 text-prxs-orange hover:text-prxs-orange/80 transition-colors"
                      >
                        Homepage
                        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
                        </svg>
                      </a>
                    </div>
                  )}
                </div>
              )}

              {activeTab === 'skills' && (
                <div className="space-y-4 animate-fade-in">
                  {agent.skills.length > 0 ? (
                    agent.skills.map((skill, idx) => (
                      <div
                        key={idx}
                        className="bg-prxs-black/30 border border-prxs-charcoal/50 rounded-xl p-6"
                      >
                        <h4 className="text-lg font-semibold text-white mb-2">
                          {extractSkillName(skill)}
                        </h4>
                        {skill.description && (
                          <p className="text-prxs-gray-light mb-4">{skill.description}</p>
                        )}
                        {extractSkillTags(skill).length > 0 && (
                          <div className="flex flex-wrap gap-2">
                            {extractSkillTags(skill).map((tag, tagIdx) => (
                              <span
                                key={tagIdx}
                                className="px-3 py-1 bg-prxs-charcoal/50 text-prxs-gray-light rounded-md text-sm"
                              >
                                {tag}
                              </span>
                            ))}
                          </div>
                        )}
                      </div>
                    ))
                  ) : (
                    <p className="text-prxs-gray-light">No skills defined</p>
                  )}
                </div>
              )}

              {activeTab === 'capabilities' && (
                <div className="animate-fade-in">
                  {Object.keys(agent.capabilities).length > 0 ? (
                    <div className="bg-prxs-black/30 border border-prxs-charcoal/50 rounded-xl p-6">
                      <pre className="text-sm text-prxs-gray-light overflow-x-auto">
                        {JSON.stringify(agent.capabilities, null, 2)}
                      </pre>
                    </div>
                  ) : (
                    <p className="text-prxs-gray-light">No capabilities defined</p>
                  )}
                </div>
              )}

              {activeTab === 'raw' && (
                <div className="animate-fade-in">
                  <div className="bg-prxs-black/30 border border-prxs-charcoal/50 rounded-xl p-6">
                    <pre className="text-sm text-prxs-gray-light overflow-x-auto">
                      {JSON.stringify(agent.card, null, 2)}
                    </pre>
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>
      </main>
    </div>
  )
}
