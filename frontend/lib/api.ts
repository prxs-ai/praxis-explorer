import { AgentRow, AgentsResponse, SearchParams } from '@/types/agent'

const API_BASE_URL = (process.env.NEXT_PUBLIC_API_URL || process.env.NEXT_PUBLIC_EXPLORER_URL || 'http://localhost:8080').replace(/\/$/, '')

export async function searchAgents(params: SearchParams): Promise<AgentsResponse> {
  const searchParams = new URLSearchParams()

  if (params.q) searchParams.set('q', params.q)
  if (params.network) searchParams.set('network', params.network)
  if (params.capability) searchParams.set('capability', params.capability)
  if (params.skill) searchParams.set('skill', params.skill)
  if (params.tag) searchParams.set('tag', params.tag)
  if (params.trustModel) searchParams.set('trustModel', params.trustModel)
  if (params.cursor) searchParams.set('cursor', params.cursor)
  if (params.limit) searchParams.set('limit', params.limit.toString())

  const response = await fetch(`${API_BASE_URL}/agents?${searchParams}`, {
    headers: {
      'Content-Type': 'application/json',
    },
  })

  if (!response.ok) {
    const txt = await response.text().catch(() => `${response.status}`)
    throw new Error(`API error: ${response.status} ${txt}`)
  }

  const data = await response.json().catch(() => ({} as any))
  const items = Array.isArray(data.items) ? data.items : []
  const nextCursor = typeof data.nextCursor === 'string' ? data.nextCursor : undefined
  return { items, nextCursor }
}

export async function getAgent(chainId: string, agentId: string): Promise<AgentRow> {
  const response = await fetch(`${API_BASE_URL}/agents/${chainId}/${agentId}`, {
    headers: {
      'Content-Type': 'application/json',
    },
  })

  if (!response.ok) {
    const txt = await response.text().catch(() => `${response.status}`)
    throw new Error(`API error: ${response.status} ${txt}`)
  }

  const data = await response.json().catch(() => (null as any))
  if (!data) throw new Error('Empty response')
  return data
}

export async function refreshAgent(chainId: string, domain: string, agentId: number, registryAddr?: string) {
  const response = await fetch(`${API_BASE_URL}/admin/refresh`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      chainId,
      domain,
      agentId,
      registryAddr,
    }),
  })

  if (!response.ok) {
    const txt = await response.text().catch(() => `${response.status}`)
    throw new Error(`API error: ${response.status} ${txt}`)
  }

  const data = await response.json().catch(() => (null as any))
  if (!data) throw new Error('Empty response')
  return data
}
