import { useMemo, useState } from 'react';
import { ArrowUpRight, Copy, ExternalLink, Clock, TrendingUp } from 'lucide-react';
import { useOpportunityStore, useUIStore } from '../../store';
import type { Opportunity } from '../../types';
import { CHAIN_NAMES, CHAIN_COLORS } from '../../types';
import OpportunityDrawer from './OpportunityDrawer';

function getSpreadClass(spread: number): string {
    if (spread >= 2) return 'spread-high';
    if (spread >= 0.5) return 'spread-medium';
    return 'spread-low';
}

function getSpreadBg(spread: number): string {
    if (spread >= 2) return 'var(--accent-green-dim)';
    if (spread >= 0.5) return 'var(--accent-yellow-dim)';
    return 'transparent';
}

function formatUSD(val: number): string {
    return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD', minimumFractionDigits: 2 }).format(val);
}

function timeAgo(dateStr: string): string {
    const seconds = Math.floor((Date.now() - new Date(dateStr).getTime()) / 1000);
    if (seconds < 60) return `${seconds}s ago`;
    if (seconds < 3600) return `${Math.floor(seconds / 60)}m ago`;
    return `${Math.floor(seconds / 3600)}h ago`;
}

export default function OpportunityTable() {
    const opportunities = useOpportunityStore((s) => s.opportunities);
    const { selectedOpportunityId, setSelectedOpportunity, filterChain, filterAssetPair } = useUIStore();
    const [sortField, setSortField] = useState<keyof Opportunity>('netProfitUsd');
    const [sortDir, setSortDir] = useState<'asc' | 'desc'>('desc');

    const filtered = useMemo(() => {
        let list = opportunities;
        if (filterChain) {
            list = list.filter((o) => o.sourceChain === filterChain || o.destChain === filterChain);
        }
        if (filterAssetPair) {
            list = list.filter((o) => o.assetPair === filterAssetPair);
        }
        return [...list].sort((a, b) => {
            const aVal = a[sortField] as number;
            const bVal = b[sortField] as number;
            return sortDir === 'desc' ? bVal - aVal : aVal - bVal;
        });
    }, [opportunities, filterChain, filterAssetPair, sortField, sortDir]);

    const selectedOpp = selectedOpportunityId
        ? opportunities.find((o) => o.id === selectedOpportunityId)
        : null;

    const handleSort = (field: keyof Opportunity) => {
        if (sortField === field) {
            setSortDir(sortDir === 'desc' ? 'asc' : 'desc');
        } else {
            setSortField(field);
            setSortDir('desc');
        }
    };

    const SortHeader = ({ field, children }: { field: keyof Opportunity; children: React.ReactNode }) => (
        <th
            onClick={() => handleSort(field)}
            style={{
                padding: '10px 12px',
                textAlign: 'left',
                cursor: 'pointer',
                userSelect: 'none',
                fontSize: 11,
                fontWeight: 600,
                color: 'var(--text-dim)',
                textTransform: 'uppercase',
                letterSpacing: '0.05em',
                borderBottom: '1px solid var(--border-dim)',
                whiteSpace: 'nowrap',
            }}
        >
            {children} {sortField === field && (sortDir === 'desc' ? '↓' : '↑')}
        </th>
    );

    return (
        <div style={{ display: 'flex', gap: 0, height: 'calc(100vh - 140px)' }}>
            <div style={{ flex: 1, overflow: 'auto' }}>
                {/* Stats Row */}
                <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: 12, marginBottom: 16 }}>
                    {[
                        { label: 'Live Opportunities', value: opportunities.filter((o) => o.status === 'live').length, icon: TrendingUp, color: 'var(--accent-green)' },
                        { label: 'Avg Spread', value: opportunities.length > 0 ? `${(opportunities.reduce((s, o) => s + o.spreadPct, 0) / opportunities.length).toFixed(2)}%` : '0%', icon: ArrowUpRight, color: 'var(--accent-blue)' },
                        { label: 'Top Profit', value: opportunities.length > 0 ? formatUSD(Math.max(...opportunities.map((o) => o.netProfitUsd))) : '$0', icon: ArrowUpRight, color: 'var(--accent-green)' },
                        { label: 'Paths Monitored', value: '17', icon: ExternalLink, color: 'var(--accent-purple)' },
                    ].map((stat, i) => {
                        const Icon = stat.icon;
                        return (
                            <div key={i} className="glass-card" style={{ padding: 16 }}>
                                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: 8 }}>
                                    <span style={{ fontSize: 11, color: 'var(--text-dim)', fontWeight: 500, textTransform: 'uppercase', letterSpacing: '0.04em' }}>{stat.label}</span>
                                    <Icon size={14} style={{ color: stat.color }} />
                                </div>
                                <div style={{ fontSize: 22, fontWeight: 700, color: stat.color }}>{stat.value}</div>
                            </div>
                        );
                    })}
                </div>

                {/* Table */}
                <div className="glass-card" style={{ overflow: 'hidden' }}>
                    <table style={{ width: '100%', borderCollapse: 'collapse' }}>
                        <thead>
                            <tr>
                                <SortHeader field="assetPair">Pair</SortHeader>
                                <th style={{ padding: '10px 12px', textAlign: 'left', fontSize: 11, fontWeight: 600, color: 'var(--text-dim)', textTransform: 'uppercase', letterSpacing: '0.05em', borderBottom: '1px solid var(--border-dim)' }}>Route</th>
                                <SortHeader field="spreadPct">Spread</SortHeader>
                                <SortHeader field="grossProfitUsd">Gross</SortHeader>
                                <SortHeader field="netProfitUsd">Net Profit</SortHeader>
                                <SortHeader field="pathHops">Hops</SortHeader>
                                <SortHeader field="discoveredAt">Time</SortHeader>
                                <th style={{ padding: '10px 12px', borderBottom: '1px solid var(--border-dim)' }}></th>
                            </tr>
                        </thead>
                        <tbody>
                            {filtered.length === 0 ? (
                                <tr>
                                    <td colSpan={8} style={{ textAlign: 'center', padding: 48, color: 'var(--text-dim)' }}>
                                        <div style={{ fontSize: 14, marginBottom: 8 }}>Scanning for opportunities...</div>
                                        <div className="animate-pulse-slow" style={{ fontSize: 12 }}>
                                            Monitoring 7 chains • 17 paths
                                        </div>
                                    </td>
                                </tr>
                            ) : (
                                filtered.map((opp, idx) => (
                                    <tr
                                        key={opp.id}
                                        onClick={() => setSelectedOpportunity(opp.id)}
                                        className="animate-slide-down"
                                        style={{
                                            cursor: 'pointer',
                                            background: selectedOpportunityId === opp.id ? 'var(--bg-hover)' : 'transparent',
                                            borderBottom: '1px solid var(--border-dim)',
                                            transition: 'background 0.15s ease',
                                            animationDelay: `${idx * 30}ms`,
                                        }}
                                        onMouseEnter={(e) => { if (selectedOpportunityId !== opp.id) e.currentTarget.style.background = 'var(--bg-hover)'; }}
                                        onMouseLeave={(e) => { if (selectedOpportunityId !== opp.id) e.currentTarget.style.background = 'transparent'; }}
                                    >
                                        <td style={{ padding: '10px 12px', fontWeight: 600, fontSize: 13 }}>
                                            {opp.assetPair}
                                        </td>
                                        <td style={{ padding: '10px 12px', fontSize: 12 }}>
                                            <div style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
                                                <span style={{
                                                    display: 'inline-block', width: 8, height: 8, borderRadius: '50%',
                                                    background: CHAIN_COLORS[opp.sourceChain] || '#666',
                                                }} />
                                                <span style={{ color: 'var(--text-secondary)' }}>{CHAIN_NAMES[opp.sourceChain]}</span>
                                                <ArrowUpRight size={12} style={{ color: 'var(--text-dim)' }} />
                                                <span style={{
                                                    display: 'inline-block', width: 8, height: 8, borderRadius: '50%',
                                                    background: CHAIN_COLORS[opp.destChain] || '#666',
                                                }} />
                                                <span style={{ color: 'var(--text-secondary)' }}>{CHAIN_NAMES[opp.destChain]}</span>
                                            </div>
                                        </td>
                                        <td style={{ padding: '10px 12px' }}>
                                            <span
                                                className={getSpreadClass(opp.spreadPct)}
                                                style={{
                                                    fontWeight: 700,
                                                    fontSize: 13,
                                                    padding: '2px 8px',
                                                    borderRadius: 4,
                                                    background: getSpreadBg(opp.spreadPct),
                                                }}
                                            >
                                                {opp.spreadPct.toFixed(2)}%
                                            </span>
                                        </td>
                                        <td style={{ padding: '10px 12px', fontSize: 13, color: 'var(--text-secondary)' }}>
                                            {formatUSD(opp.grossProfitUsd)}
                                        </td>
                                        <td style={{ padding: '10px 12px' }}>
                                            <span style={{
                                                fontSize: 13,
                                                fontWeight: 700,
                                                color: opp.netProfitUsd > 50 ? 'var(--accent-green)' : opp.netProfitUsd > 10 ? 'var(--accent-yellow)' : 'var(--text-secondary)',
                                            }}>
                                                {formatUSD(opp.netProfitUsd)}
                                            </span>
                                        </td>
                                        <td style={{ padding: '10px 12px', fontSize: 12, color: 'var(--text-dim)', textAlign: 'center' }}>
                                            {opp.pathHops}
                                        </td>
                                        <td style={{ padding: '10px 12px', fontSize: 11, color: 'var(--text-dim)' }}>
                                            <div style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
                                                <Clock size={10} />
                                                {timeAgo(opp.discoveredAt)}
                                            </div>
                                        </td>
                                        <td style={{ padding: '10px 12px' }}>
                                            <button
                                                onClick={(e) => { e.stopPropagation(); navigator.clipboard.writeText(opp.calldataJson); }}
                                                aria-label="Copy calldata"
                                                style={{
                                                    background: 'var(--bg-card)',
                                                    border: '1px solid var(--border-dim)',
                                                    borderRadius: 4,
                                                    padding: '4px 6px',
                                                    cursor: 'pointer',
                                                    color: 'var(--text-dim)',
                                                }}
                                            >
                                                <Copy size={12} />
                                            </button>
                                        </td>
                                    </tr>
                                ))
                            )}
                        </tbody>
                    </table>
                </div>
            </div>

            {/* Drawer */}
            {selectedOpp && (
                <OpportunityDrawer
                    opportunity={selectedOpp}
                    onClose={() => setSelectedOpportunity(null)}
                />
            )}
        </div>
    );
}
