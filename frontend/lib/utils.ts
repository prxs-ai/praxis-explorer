export function formatDate(dateString: string): string {
  const date = new Date(dateString)
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffSecs = Math.floor(diffMs / 1000)
  const diffMins = Math.floor(diffSecs / 60)
  const diffHours = Math.floor(diffMins / 60)
  const diffDays = Math.floor(diffHours / 24)

  if (diffSecs < 60) return 'just now'
  if (diffMins < 60) return `${diffMins}m ago`
  if (diffHours < 24) return `${diffHours}h ago`
  if (diffDays < 7) return `${diffDays}d ago`

  return date.toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    year: date.getFullYear() !== now.getFullYear() ? 'numeric' : undefined
  })
}

export function isOnline(lastSeenAt: string): boolean {
  const date = new Date(lastSeenAt)
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffMins = Math.floor(diffMs / 60000)
  return diffMins < 5
}

export function getChainName(chainId: string): string {
  const chains: Record<string, string> = {
    '1': 'Ethereum',
    '11155111': 'Sepolia',
    '8453': 'Base',
    '84532': 'Base Sepolia',
    '42161': 'Arbitrum',
    '421614': 'Arbitrum Sepolia',
    '10': 'Optimism',
    '11155420': 'Optimism Sepolia',
  }
  return chains[chainId] || `Chain ${chainId}`
}

export function truncateAddress(address: string, chars = 4): string {
  if (!address) return ''
  return `${address.slice(0, chars + 2)}...${address.slice(-chars)}`
}

export function getTrustModelColor(model: string): string {
  const colors: Record<string, string> = {
    'feedback': 'orange',
    'reputation': 'cyan',
    'economic': 'blue',
    'hybrid': 'purple',
  }
  return colors[model.toLowerCase()] || 'gray'
}

export function extractSkillName(skill: any): string {
  return skill?.name || skill?.id || 'Unknown Skill'
}

export function extractSkillTags(skill: any): string[] {
  if (!skill?.tags) return []
  if (Array.isArray(skill.tags)) return skill.tags
  if (typeof skill.tags === 'string') return [skill.tags]
  return []
}
