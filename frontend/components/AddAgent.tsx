'use client';

import { useState, useEffect } from 'react';
import { useAccount, useSimulateContract, useWaitForTransactionReceipt, useWriteContract } from 'wagmi';
import { type Address, isAddress } from 'viem';
import { identityRegistryAbi } from '@/lib/abis';

export default function AddAgent({
  registry,
  onSuccess,
}: {
  registry: Address;          // IdentityRegistry contract address
  onSuccess?: () => void;
}) {
  const { address: userAddress, isConnected } = useAccount();

  const [domain, setDomain] = useState('');
  const [useCustomAddress, setUseCustomAddress] = useState(false);
  const [customAddress, setCustomAddress] = useState('');

  const domainOk = domain.trim().length > 0;
  const agentAddress = useCustomAddress ? customAddress : userAddress;
  const addressOk = agentAddress && isAddress(agentAddress);

  // Simulate contract call (only when we have inputs + connected wallet)
  const { data: sim, isLoading: simLoading, error: simError } = useSimulateContract({
    address: registry,
    abi: identityRegistryAbi,
    functionName: 'newAgent',
    args: domainOk && addressOk ? [domain.trim(), agentAddress as Address] : undefined,
    query: {
      enabled: Boolean(domainOk && addressOk && isConnected),
    },
  });

  // 3) write + wait
  const { writeContract, data: hash, isPending: sending, error: writeError } = useWriteContract();
  const { isLoading: confirming, isSuccess, data: receipt } = useWaitForTransactionReceipt({
    hash
  });

  // Handle success
  useEffect(() => {
    if (isSuccess && hash) {
      setDomain(''); // Reset form
      setCustomAddress('');
      setUseCustomAddress(false);
      onSuccess?.();
    }
  }, [isSuccess, hash, onSuccess]);

  const canSubmit = isConnected && domainOk && addressOk && !!sim && !sending;

  if (!isConnected) {
    return (
      <div className="text-center py-6">
        <p className="text-prxs-gray mb-4">Connect your wallet to register a new agent</p>
        <div className="w-full bg-prxs-charcoal/30 border border-prxs-charcoal rounded-lg p-4">
          <div className="flex items-center justify-center gap-2 text-prxs-gray">
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
            </svg>
            <span>Wallet connection required</span>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-4 w-full max-w-md mx-auto">
      <div className="text-center mb-4">
        <h3 className="text-lg font-semibold text-white mb-1">Register New Agent</h3>
        <p className="text-sm text-prxs-gray">Agent will be registered to your connected wallet</p>
      </div>

      <div className="space-y-1">
        <label className="text-sm text-prxs-gray">Agent domain</label>
        <input
          className="w-full bg-prxs-black border border-prxs-charcoal rounded-lg px-4 py-3 text-white placeholder-prxs-gray outline-none focus:border-prxs-orange transition-colors"
          placeholder="e.g. alice.agent"
          value={domain}
          onChange={(e) => setDomain(e.target.value)}
        />
        {!domainOk && domain.length > 0 && <p className="text-xs text-red-400">Domain is required.</p>}
      </div>

      <div className="space-y-1">
        <label className="text-sm text-prxs-gray">Agent address</label>
        <div className="w-full bg-prxs-charcoal/30 border border-prxs-charcoal rounded-lg px-4 py-3 text-prxs-gray">
          {userAddress || 'Connect wallet to see address'}
        </div>
        <p className="text-xs text-prxs-gray">Using your connected wallet address</p>
      </div>

      <button
        onClick={() => sim && writeContract(sim.request)}
        disabled={!canSubmit}
        className="w-full px-4 py-3 rounded-lg bg-prxs-orange/90 hover:bg-prxs-orange text-black font-medium disabled:opacity-50 disabled:cursor-not-allowed transition-all"
      >
        {sending ? 'Confirm in wallet…' : confirming ? 'Waiting for confirmations…' : 'Register Agent'}
      </button>

      {/* Status / errors */}
      {simError && <p className="text-xs text-red-400 bg-red-500/10 border border-red-500/20 rounded p-2">Simulation failed: {simError.message}</p>}
      {writeError && <p className="text-xs text-red-400 bg-red-500/10 border border-red-500/20 rounded p-2">Transaction failed: {writeError.message}</p>}
      {hash && <p className="text-xs text-prxs-gray break-all bg-prxs-charcoal/30 rounded p-2">Transaction: {hash}</p>}
      {isSuccess && (
        <p className="text-green-400 text-sm bg-green-500/10 border border-green-500/20 rounded p-2 text-center">
          ✅ Agent registered successfully in block #{receipt?.blockNumber?.toString()}
        </p>
      )}
    </div>
  );
}
