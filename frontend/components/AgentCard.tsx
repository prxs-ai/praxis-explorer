'use client'

import { AgentRow } from '@/types/agent'
import { formatDate, isOnline, getChainName, extractSkillName, getTrustModelColor } from '@/lib/utils'
import Link from 'next/link'

interface AgentCardProps {
  agent: AgentRow
}

export default function AgentCard({ agent }: AgentCardProps) {
  const online = isOnline(agent.lastSeenAt)
  const agentName = agent.card?.name || agent.domain.split('.')[0]
  const description = agent.card?.description || 'No description available'
  const verified = agent.agentId > 0

  return (
    <Link href={`/agent/${agent.chainId}/${agent.agentId}`}>
      <div className="group relative bg-gradient-to-br from-prxs-black-secondary to-prxs-black border border-prxs-charcoal rounded-2xl p-6 card-hover cursor-pointer overflow-hidden">
        <div className="absolute inset-0 bg-gradient-to-br from-prxs-orange/5 via-transparent to-prxs-cyan/5 opacity-0 group-hover:opacity-100 transition-opacity duration-500" />
        
        <div className="relative z-10">
          <div className="flex items-start justify-between mb-4">
            <div className="flex-1">
              <div className="flex items-center gap-2 mb-1">
                <h3 className="text-xl font-bold text-white group-hover:text-prxs-orange transition-colors">
                  {agentName}
                </h3>
                {verified && (
                  <svg className="w-5 h-5 text-prxs-cyan" fill="currentColor" viewBox="0 0 20 20">
                    <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.857-9.809a.75.75 0 00-1.214-.882l-3.483 4.79-1.88-1.88a.75.75 0 10-1.06 1.061l2.5 2.5a.75.75 0 001.137-.089l4-5.5z" clipRule="evenodd" />
                  </svg>
                )}
              </div>
              <p className="text-prxs-gray text-sm">{agent.domain}</p>
            </div>
            
            <div className="flex items-center gap-2">
              <div className={`w-2 h-2 rounded-full ${online ? 'bg-green-500 animate-pulse' : 'bg-prxs-gray'}`} />
              <span className="text-xs text-prxs-gray-light">
                {online ? 'Online' : formatDate(agent.lastSeenAt)}
              </span>
            </div>
          </div>

          <p className="text-prxs-gray-light text-sm mb-4 line-clamp-2">
            {description}
          </p>

          <div className="flex flex-wrap gap-2 mb-4">
            {agent.trustModels.map((model) => (
              <span
                key={model}
                className={`badge badge-${getTrustModelColor(model)}`}
              >
                {model}
              </span>
            ))}
          </div>

          {agent.skills.length > 0 && (
            <div className="space-y-1">
              <p className="text-xs text-prxs-gray uppercase tracking-wider mb-2">Skills</p>
              <div className="flex flex-wrap gap-1">
                {agent.skills.slice(0, 3).map((skill, idx) => (
                  <span
                    key={idx}
                    className="text-xs px-2 py-1 bg-prxs-charcoal/50 text-prxs-gray-light rounded-md"
                  >
                    {extractSkillName(skill)}
                  </span>
                ))}
                {agent.skills.length > 3 && (
                  <span className="text-xs px-2 py-1 text-prxs-gray">
                    +{agent.skills.length - 3} more
                  </span>
                )}
              </div>
            </div>
          )}

          <div className="flex items-center justify-between mt-4 pt-4 border-t border-prxs-charcoal/50">
            <div className="flex items-center gap-4">
              {agent.scoreAvg && (
                <div className="flex items-center gap-1">
                  <svg className="w-4 h-4 text-yellow-500" fill="currentColor" viewBox="0 0 20 20">
                    <path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z" />
                  </svg>
                  <span className="text-sm text-white">{agent.scoreAvg.toFixed(1)}</span>
                </div>
              )}
              <div className="flex items-center gap-1 text-xs text-prxs-gray">
                <span>{agent.validationsCnt} validations</span>
                <span>•</span>
                <span>{agent.feedbacksCnt} feedbacks</span>
              </div>
            </div>
            
            <div className="text-xs text-prxs-gray">
              {getChainName(agent.chainId)} #{agent.agentId}
            </div>
          </div>
        </div>

        <div className="absolute bottom-0 left-0 right-0 h-1 bg-gradient-to-r from-prxs-orange via-prxs-cyan to-prxs-blue transform scale-x-0 group-hover:scale-x-100 transition-transform duration-500" />
      </div>
    </Link>
  )
}
