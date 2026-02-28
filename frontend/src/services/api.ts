import axios from 'axios';
import type { Opportunity, ChainStatus, RelayChannel, RelayEvent, PriceData } from '../types';

const api = axios.create({
    baseURL: '/api/v1',
    timeout: 10000,
    headers: { 'Content-Type': 'application/json' },
});

// Parse decimal strings from shopspring/decimal to JS numbers
function parseOpportunity(raw: Record<string, unknown>): Opportunity {
    return {
        ...raw,
        spreadPct: Number(raw.spreadPct) || 0,
        grossProfitUsd: Number(raw.grossProfitUsd) || 0,
        netProfitUsd: Number(raw.netProfitUsd) || 0,
        inputAmountUsd: Number(raw.inputAmountUsd) || 0,
        slippageEstPct: Number(raw.slippageEstPct) || 0,
        pathHops: Number(raw.pathHops) || 0,
        path: Array.isArray(raw.path)
            ? raw.path.map((node: Record<string, unknown>) => ({
                ...node,
                price: Number(node.price) || 0,
            }))
            : [],
        feeBreakdown: Array.isArray(raw.feeBreakdown)
            ? raw.feeBreakdown.map((fee: Record<string, unknown>) => ({
                ...fee,
                amountUsd: Number(fee.amountUsd) || 0,
            }))
            : [],
        calldataJson:
            typeof raw.calldataJson === 'string'
                ? raw.calldataJson
                : JSON.stringify(raw.calldataJson, null, 2),
    } as Opportunity;
}

export const fetchOpportunities = async (status = 'live', limit = 50): Promise<Opportunity[]> => {
    const { data } = await api.get('/opportunities', { params: { status, limit } });
    return (data.data || []).map(parseOpportunity);
};

export const fetchOpportunityHistory = async (
    from: string, to: string, assetPair?: string, limit = 100, offset = 0
): Promise<Opportunity[]> => {
    const { data } = await api.get('/opportunities/history', {
        params: { from, to, assetPair, limit, offset },
    });
    return (data.data || []).map(parseOpportunity);
};

export const fetchChains = async (): Promise<ChainStatus[]> => {
    const { data } = await api.get('/chains');
    return data.data || [];
};

export const fetchChainPrices = async (): Promise<PriceData[]> => {
    const { data } = await api.get('/chains/prices');
    return data.data || [];
};

export const fetchRelayChannels = async (): Promise<RelayChannel[]> => {
    const { data } = await api.get('/relay/channels');
    return data.data || [];
};

export const fetchRelayEvents = async (channelId: string, limit = 50): Promise<RelayEvent[]> => {
    const { data } = await api.get(`/relay/channels/${channelId}/events`, { params: { limit } });
    return data.data || [];
};

export const exportOpportunitiesCSV = (from: string, to: string, assetPair?: string) => {
    const params = new URLSearchParams({ from, to });
    if (assetPair) params.set('assetPair', assetPair);
    window.open(`/api/v1/opportunities/export?${params.toString()}`);
};

// Parse incoming WebSocket opportunity
export function parseWSOpportunity(raw: Record<string, unknown>): Opportunity {
    return parseOpportunity(raw);
}

export default api;
