import { create } from 'zustand';
import type { Opportunity } from '../types';

interface OpportunityStore {
    opportunities: Opportunity[];
    setOpportunities: (opps: Opportunity[]) => void;
    upsert: (opp: Opportunity) => void;
    removeExpired: () => void;
}

export const useOpportunityStore = create<OpportunityStore>((set) => ({
    opportunities: [],
    setOpportunities: (opps) => set({ opportunities: opps }),
    upsert: (opp) =>
        set((state) => {
            const idx = state.opportunities.findIndex((o) => o.id === opp.id);
            if (idx >= 0) {
                const updated = [...state.opportunities];
                updated[idx] = opp;
                return { opportunities: updated };
            }
            return { opportunities: [opp, ...state.opportunities].slice(0, 200) };
        }),
    removeExpired: () =>
        set((state) => ({
            opportunities: state.opportunities.filter((o) => o.status === 'live'),
        })),
}));

interface UIStore {
    selectedOpportunityId: string | null;
    drawerOpen: boolean;
    activeTab: string;
    filterChain: string;
    filterAssetPair: string;
    setSelectedOpportunity: (id: string | null) => void;
    setDrawerOpen: (open: boolean) => void;
    setActiveTab: (tab: string) => void;
    setFilterChain: (chain: string) => void;
    setFilterAssetPair: (pair: string) => void;
}

export const useUIStore = create<UIStore>((set) => ({
    selectedOpportunityId: null,
    drawerOpen: false,
    activeTab: 'dashboard',
    filterChain: '',
    filterAssetPair: '',
    setSelectedOpportunity: (id) => set({ selectedOpportunityId: id, drawerOpen: !!id }),
    setDrawerOpen: (open) => set({ drawerOpen: open }),
    setActiveTab: (tab) => set({ activeTab: tab }),
    setFilterChain: (chain) => set({ filterChain: chain }),
    setFilterAssetPair: (pair) => set({ filterAssetPair: pair }),
}));
