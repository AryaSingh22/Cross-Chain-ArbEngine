export type ChainId = 'osmosis' | 'injective' | 'neutron' | 'stride' | 'juno' | 'cosmoshub' | 'akash';

export interface PathNode {
    chain: ChainId;
    dex: string;
    assetIn: string;
    assetOut: string;
    price: number;
    poolId?: string;
}

export interface FeeEntry {
    chain: ChainId;
    feeType: string;
    amountUsd: number;
    asset: string;
}

export interface Opportunity {
    id: string;
    assetPair: string;
    sourceChain: ChainId;
    destChain: ChainId;
    spreadPct: number;
    grossProfitUsd: number;
    netProfitUsd: number;
    pathHops: number;
    discoveredAt: string;
    path: PathNode[];
    feeBreakdown: FeeEntry[];
    calldataJson: string;
    inputAmountUsd: number;
    slippageEstPct: number;
    status: 'live' | 'expired' | 'executed';
    expiresAt?: string;
}

export interface ChainStatus {
    id: ChainId;
    name: string;
    connected: boolean;
    feedCount: number;
}

export interface RelayChannel {
    id: number;
    sourceChain: string;
    destChain: string;
    channelId: string;
    portId: string;
    status: 'healthy' | 'backlogged' | 'stuck' | 'closed';
    pendingPackets: number;
    oldestPacketAgeS?: number;
    lastCheckedAt: string;
    lastRelayAt?: string;
}

export interface RelayEvent {
    id: string;
    eventAt: string;
    channelId: string;
    sourceChain: string;
    destChain: string;
    eventType: string;
    packetSequence?: number;
    relayLatencyMs?: number;
}

export interface PriceData {
    chain: ChainId;
    assetPair: string;
    priceUsd: number;
    sourceDex: string;
    poolId?: string;
    timestamp: string;
}

export const CHAIN_NAMES: Record<ChainId, string> = {
    osmosis: 'Osmosis',
    injective: 'Injective',
    neutron: 'Neutron',
    stride: 'Stride',
    juno: 'Juno',
    cosmoshub: 'Cosmos Hub',
    akash: 'Akash',
};

export const CHAIN_COLORS: Record<ChainId, string> = {
    osmosis: '#6F4CFF',
    injective: '#00F2FE',
    neutron: '#1752F0',
    stride: '#E50571',
    juno: '#F0827D',
    cosmoshub: '#2E3148',
    akash: '#FF414C',
};
