import { useState, useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import { fetchOpportunityHistory, exportOpportunitiesCSV } from '../services/api';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, BarChart, Bar, AreaChart, Area } from 'recharts';
import { Download, Calendar, TrendingUp } from 'lucide-react';
import type { Opportunity } from '../types';

export default function Analytics() {
    const [days, setDays] = useState(7);
    const [selectedPair, setSelectedPair] = useState('');

    const from = new Date(Date.now() - days * 24 * 60 * 60 * 1000).toISOString();
    const to = new Date().toISOString();

    const { data: history = [] } = useQuery<Opportunity[]>({
        queryKey: ['opportunity-history', from, to, selectedPair],
        queryFn: () => fetchOpportunityHistory(from, to, selectedPair, 500),
        refetchInterval: 30000,
    });

    // Group by hour for charts
    const hourlyData = useMemo(() => {
        const buckets = new Map<string, { count: number; totalSpread: number; totalProfit: number }>();
        history.forEach((opp: Opportunity) => {
            const hour = new Date(opp.discoveredAt).toISOString().slice(0, 13) + ':00';
            const existing = buckets.get(hour) || { count: 0, totalSpread: 0, totalProfit: 0 };
            existing.count++;
            existing.totalSpread += opp.spreadPct;
            existing.totalProfit += opp.netProfitUsd;
            buckets.set(hour, existing);
        });
        return Array.from(buckets.entries())
            .map(([hour, data]) => ({
                hour: new Date(hour).toLocaleString(undefined, { month: 'short', day: 'numeric', hour: '2-digit' }),
                count: data.count,
                avgSpread: data.totalSpread / data.count,
                totalProfit: data.totalProfit,
            }))
            .sort((a, b) => a.hour.localeCompare(b.hour));
    }, [history]);

    const assetPairs = useMemo(() => {
        const pairs = new Set(history.map((o: Opportunity) => o.assetPair));
        return Array.from(pairs).sort();
    }, [history]);

    return (
        <div>
            {/* Controls */}
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 20 }}>
                <div style={{ display: 'flex', gap: 8 }}>
                    {[1, 7, 14, 30].map((d) => (
                        <button
                            key={d}
                            onClick={() => setDays(d)}
                            style={{
                                padding: '6px 14px',
                                borderRadius: 6,
                                border: `1px solid ${days === d ? 'var(--accent-blue)' : 'var(--border-dim)'}`,
                                background: days === d ? 'var(--accent-blue-dim)' : 'var(--bg-card)',
                                color: days === d ? 'var(--accent-blue)' : 'var(--text-secondary)',
                                cursor: 'pointer',
                                fontSize: 12,
                                fontWeight: 600,
                            }}
                        >
                            {d}D
                        </button>
                    ))}
                    <select
                        value={selectedPair}
                        onChange={(e) => setSelectedPair(e.target.value)}
                        style={{
                            padding: '6px 12px',
                            borderRadius: 6,
                            border: '1px solid var(--border-dim)',
                            background: 'var(--bg-card)',
                            color: 'var(--text-primary)',
                            fontSize: 12,
                            cursor: 'pointer',
                        }}
                    >
                        <option value="">All Pairs</option>
                        {assetPairs.map((p: string) => <option key={p} value={p}>{p}</option>)}
                    </select>
                </div>
                <button
                    onClick={() => exportOpportunitiesCSV(from, to, selectedPair)}
                    style={{
                        display: 'flex', alignItems: 'center', gap: 6,
                        padding: '8px 16px',
                        borderRadius: 8,
                        border: '1px solid var(--accent-blue)',
                        background: 'var(--accent-blue-dim)',
                        color: 'var(--accent-blue)',
                        cursor: 'pointer',
                        fontSize: 12,
                        fontWeight: 600,
                    }}
                >
                    <Download size={14} />
                    Export CSV
                </button>
            </div>

            {/* Summary Cards */}
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 12, marginBottom: 20 }}>
                <div className="glass-card" style={{ padding: 16 }}>
                    <div style={{ fontSize: 11, color: 'var(--text-dim)', marginBottom: 6, textTransform: 'uppercase' }}>
                        <Calendar size={12} style={{ marginRight: 4, verticalAlign: 'middle' }} />
                        Total Opportunities
                    </div>
                    <div style={{ fontSize: 28, fontWeight: 700, color: 'var(--accent-blue)' }}>{history.length}</div>
                </div>
                <div className="glass-card" style={{ padding: 16 }}>
                    <div style={{ fontSize: 11, color: 'var(--text-dim)', marginBottom: 6, textTransform: 'uppercase' }}>
                        <TrendingUp size={12} style={{ marginRight: 4, verticalAlign: 'middle' }} />
                        Avg Spread
                    </div>
                    <div style={{ fontSize: 28, fontWeight: 700, color: 'var(--accent-green)' }}>
                        {history.length > 0 ? (history.reduce((s: number, o: Opportunity) => s + o.spreadPct, 0) / history.length).toFixed(2) : '0'}%
                    </div>
                </div>
                <div className="glass-card" style={{ padding: 16 }}>
                    <div style={{ fontSize: 11, color: 'var(--text-dim)', marginBottom: 6, textTransform: 'uppercase' }}>
                        Total Net Profit
                    </div>
                    <div style={{ fontSize: 28, fontWeight: 700, color: 'var(--accent-green)' }}>
                        ${history.reduce((s: number, o: Opportunity) => s + o.netProfitUsd, 0).toFixed(2)}
                    </div>
                </div>
            </div>

            {/* Charts */}
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16 }}>
                {/* Opportunity Volume */}
                <div className="glass-card" style={{ padding: 20 }}>
                    <h3 style={{ fontSize: 13, fontWeight: 600, marginBottom: 16, color: 'var(--text-secondary)' }}>
                        Opportunity Volume
                    </h3>
                    <ResponsiveContainer width="100%" height={250}>
                        <BarChart data={hourlyData}>
                            <CartesianGrid strokeDasharray="3 3" stroke="var(--border-dim)" />
                            <XAxis dataKey="hour" tick={{ fontSize: 10, fill: 'var(--text-dim)' }} />
                            <YAxis tick={{ fontSize: 10, fill: 'var(--text-dim)' }} />
                            <Tooltip
                                contentStyle={{ background: 'var(--bg-secondary)', border: '1px solid var(--border-bright)', borderRadius: 8, fontSize: 12 }}
                            />
                            <Bar dataKey="count" fill="var(--accent-blue)" radius={[4, 4, 0, 0]} />
                        </BarChart>
                    </ResponsiveContainer>
                </div>

                {/* Spread Over Time */}
                <div className="glass-card" style={{ padding: 20 }}>
                    <h3 style={{ fontSize: 13, fontWeight: 600, marginBottom: 16, color: 'var(--text-secondary)' }}>
                        Average Spread (%)
                    </h3>
                    <ResponsiveContainer width="100%" height={250}>
                        <AreaChart data={hourlyData}>
                            <defs>
                                <linearGradient id="spreadGrad" x1="0" y1="0" x2="0" y2="1">
                                    <stop offset="0%" stopColor="var(--accent-green)" stopOpacity={0.3} />
                                    <stop offset="100%" stopColor="var(--accent-green)" stopOpacity={0} />
                                </linearGradient>
                            </defs>
                            <CartesianGrid strokeDasharray="3 3" stroke="var(--border-dim)" />
                            <XAxis dataKey="hour" tick={{ fontSize: 10, fill: 'var(--text-dim)' }} />
                            <YAxis tick={{ fontSize: 10, fill: 'var(--text-dim)' }} />
                            <Tooltip
                                contentStyle={{ background: 'var(--bg-secondary)', border: '1px solid var(--border-bright)', borderRadius: 8, fontSize: 12 }}
                            />
                            <Area type="monotone" dataKey="avgSpread" stroke="var(--accent-green)" fill="url(#spreadGrad)" strokeWidth={2} />
                        </AreaChart>
                    </ResponsiveContainer>
                </div>

                {/* Profit Over Time */}
                <div className="glass-card" style={{ padding: 20, gridColumn: 'span 2' }}>
                    <h3 style={{ fontSize: 13, fontWeight: 600, marginBottom: 16, color: 'var(--text-secondary)' }}>
                        Cumulative Profit (USD)
                    </h3>
                    <ResponsiveContainer width="100%" height={250}>
                        <LineChart data={hourlyData}>
                            <CartesianGrid strokeDasharray="3 3" stroke="var(--border-dim)" />
                            <XAxis dataKey="hour" tick={{ fontSize: 10, fill: 'var(--text-dim)' }} />
                            <YAxis tick={{ fontSize: 10, fill: 'var(--text-dim)' }} />
                            <Tooltip
                                contentStyle={{ background: 'var(--bg-secondary)', border: '1px solid var(--border-bright)', borderRadius: 8, fontSize: 12 }}
                            />
                            <Line type="monotone" dataKey="totalProfit" stroke="var(--accent-purple)" strokeWidth={2} dot={false} />
                        </LineChart>
                    </ResponsiveContainer>
                </div>
            </div>
        </div>
    );
}
