import useSWR from 'swr'
import { searchAgents } from '@/lib/api'
import { SearchParams } from '@/types/agent'

export function useAgents(params: SearchParams) {
  const queryKey = `/agents?${new URLSearchParams(
    Object.entries(params)
      .filter(([_, v]) => v !== undefined && v !== '')
      .map(([k, v]) => [k, String(v)])
  ).toString()}`

  const { data, error, isLoading, mutate } = useSWR(
    queryKey,
    () => searchAgents(params),
    {
      revalidateOnFocus: false,
      revalidateOnReconnect: false,
    }
  )

  return {
    agents: data?.items || [],
    nextCursor: data?.nextCursor,
    isLoading,
    error,
    mutate,
  }
}
