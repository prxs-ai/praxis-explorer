// lib/abis.ts
import { parseAbi } from 'viem';

export const identityRegistryAbi = parseAbi([
  "function REGISTRATION_FEE() view returns (uint256)",
  "function newAgent(string agentDomain, address agentAddress) payable returns (uint256 agentId)",
]);
