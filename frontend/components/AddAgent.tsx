'use client';

import { useState, useEffect } from 'react';
import { useAccount, useSimulateContract, useWaitForTransactionReceipt, useWriteContract } from 'wagmi';
import { type Address } from 'viem';
import { identityRegistryAbi } from '@/lib/abis';
import { isIPFSUrl } from '@/lib/utils';

export default function AddAgent({
  registry,
  onSuccess,
}: {
  registry: Address;          // IdentityRegistry contract address
  onSuccess?: () => void;
}) {
  const { address: userAddress, isConnected } = useAccount();

  const [tokenURI, setTokenURI] = useState('');

  const tokenURIOk = tokenURI.trim().length > 0;

  // Check if it's an IPFS URL
  const isIPFS = isIPFSUrl(tokenURI);

  // Simulate contract call (only when we have inputs + connected wallet)
  const { data: sim, isLoading: simLoading, error: simError } = useSimulateContract({
    address: registry,
    abi: identityRegistryAbi,
    functionName: 'register',
    args: tokenURIOk ? [tokenURI.trim()] : undefined,
    query: {
      enabled: Boolean(tokenURIOk && isConnected),
    },
  });

  // Write contract + wait for transaction
  const { writeContract, data: hash, isPending: sending, error: writeError } = useWriteContract();
  const { isLoading: confirming, isSuccess, data: receipt } = useWaitForTransactionReceipt({ 
    hash
  });

  // Handle success
  useEffect(() => {
    if (isSuccess && hash) {
      setTokenURI(''); // Reset form
      onSuccess?.();
    }
  }, [isSuccess, hash, onSuccess]);

  const canSubmit = isConnected && tokenURIOk && !!sim && !sending;

  if (!isConnected) {
    return (
      <div className="text-center py-6">
        <p className="text-prxs-gray mb-4">Connect your wallet to register a new agent</p>
        <div className="w-full bg-prxs-charcoal/30 border border-prxs-charcoal rounded-lg p-4">
          <div className="flex items-center justify-center gap-2 text-prxs-gray">
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
            </svg>
            Wallet not connected
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <div className="text-center mb-4">
        <h3 className="text-lg font-semibold text-white mb-2">Register New Agent</h3>
        <p className="text-sm text-prxs-gray">Agent NFT will be minted to your connected wallet</p>
      </div>

      <div className="space-y-1">
        <label className="text-sm text-prxs-gray">Token URI</label>
        <input
          className="w-full bg-prxs-black border border-prxs-charcoal rounded-lg px-4 py-3 text-white placeholder-prxs-gray outline-none focus:border-prxs-orange transition-colors"
          placeholder="e.g. ipfs://QmHash..."
          value={tokenURI}
          onChange={(e) => setTokenURI(e.target.value)}
        />
        <div className="flex items-start gap-2">
          <p className="text-xs text-prxs-gray flex-1">
            Enter full URL, or IPFS hash where your agent card is hosted
          </p>
          {isIPFS && (
            <div className="flex items-center gap-1 px-2 py-1 bg-purple-500/10 border border-purple-500/20 rounded-full">
              <svg className="w-3 h-3 text-purple-400" fill="currentColor" viewBox="0 0 20 20">
                <path d="M10 2L3 7l7 5 7-5-7-5zM3 13l7 5 7-5M3 10l7 5 7-5" />
              </svg>
              <span className="text-xs text-purple-400">IPFS</span>
            </div>
          )}
        </div>
        {!tokenURIOk && tokenURI.length > 0 && <p className="text-xs text-red-400">Token URI is required.</p>}
      </div>


      {isIPFS && (
        <div className="bg-purple-500/5 border border-purple-500/20 rounded-lg p-4">
          <div className="flex items-start gap-3">
            <svg className="w-5 h-5 text-purple-400 mt-0.5" fill="currentColor" viewBox="0 0 20 20">
              <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clipRule="evenodd" />
            </svg>
            <div>
              <h4 className="text-sm font-medium text-purple-400 mb-1">IPFS Detected</h4>
              <p className="text-xs text-prxs-gray">
                Your agent metadata will be fetched from IPFS using multiple gateways for reliability.
                Make sure your IPFS content includes a valid agent card JSON.
              </p>
            </div>
          </div>
        </div>
      )}

      <button
        onClick={() => sim && writeContract(sim.request)}
        disabled={!canSubmit}
        className={`w-full py-3 px-4 rounded-lg font-medium transition-all ${
          canSubmit
            ? 'bg-prxs-orange hover:bg-prxs-orange/90 text-black'
            : 'bg-prxs-charcoal text-prxs-gray cursor-not-allowed'
        }`}
      >
        {sending ? (
          <div className="flex items-center justify-center gap-2">
            <div className="w-4 h-4 border-2 border-current border-t-transparent rounded-full animate-spin" />
            Sending...
          </div>
        ) : confirming ? (
          <div className="flex items-center justify-center gap-2">
            <div className="w-4 h-4 border-2 border-current border-t-transparent rounded-full animate-spin" />
            Confirming...
          </div>
        ) : (
          'Register Agent'
        )}
      </button>

      {(simError || writeError) && (
        <div className="bg-red-500/10 border border-red-500/20 rounded-lg p-3">
          <p className="text-sm text-red-400">
            {simError?.message || writeError?.message || 'Transaction failed'}
          </p>
        </div>
      )}

      {hash && (
        <div className="bg-green-500/10 border border-green-500/20 rounded-lg p-3">
          <p className="text-sm text-green-400">
            Transaction submitted! Hash: {hash.slice(0, 10)}...
          </p>
        </div>
      )}
    </div>
  );
}