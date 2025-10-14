'use client';

import { http, createStorage, cookieStorage } from 'wagmi'
import { sepolia } from 'wagmi/chains'
import { Chain, getDefaultConfig } from '@rainbow-me/rainbowkit'

const projectId = "60956e570d8c93562759ed10c881a156";

const supportedChains: Chain[] = [sepolia];

export const config = getDefaultConfig({
   appName: 'Praxis Explorer',
   projectId,
   chains: supportedChains as any,
   ssr: true,
   storage: createStorage({
    storage: cookieStorage,
   }),
  transports: supportedChains.reduce((obj, chain) => ({ ...obj, [chain.id]: http() }), {})
 });
