// lib/abis.ts
import { parseAbi } from 'viem';

export const identityRegistryAbi = parseAbi([
  "function register(string tokenURI_) returns (uint256 agentId)",
  "function totalAgents() view returns (uint256 count)",
  "function ownerOf(uint256 tokenId) view returns (address owner)",
  "function tokenURI(uint256 tokenId) view returns (string)",
  "function getMetadata(uint256 agentId, string key) view returns (bytes value)",
  "function agentExists(uint256 agentId) view returns (bool exists)",
  "event Registered(uint256 indexed agentId, string tokenURI, address indexed owner)",
]);
