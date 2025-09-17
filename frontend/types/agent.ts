export interface AgentRow {
  chainId: string
  agentId: number
  registryAddr?: string
  domain: string
  addressCaip10: string
  card: Record<string, any>
  trustModels: string[]
  skills: Array<Record<string, any>>
  capabilities: Record<string, any>
  scoreAvg?: number
  validationsCnt: number
  feedbacksCnt: number
  lastSeenAt: string
}

export interface AgentsResponse {
  items: AgentRow[]
  nextCursor?: string
}

export interface SearchParams {
  q?: string
  network?: string
  capability?: string
  skill?: string
  tag?: string
  trustModel?: string
  cursor?: string
  limit?: number
}