import { useQuery } from '@tanstack/react-query';
import { fetchRelayChannels } from '../services/api';
import { CHAIN_NAMES } from '../types';
import type { RelayChannel } from '../types';
import { Radio, AlertTriangle, CheckCircle2, XCircle, Clock } from 'lucide-react';

const statusConfig: Record<string, { color: string; bg: string; icon: typeof CheckCircle2 }> = {
    healthy: { color: 'var(--accent-green)', bg: 'var(--accent-green-dim)', icon: CheckCircle2 },
    backlogged: { color: 'var(--accent-yellow)', bg: 'var(--accent-yellow-dim)', icon: Clock },
    stuck: { color: 'var(--accent-red)', bg: 'var(--accent-red-dim)', icon: AlertTriangle },
    closed: { color: 'var(--text-dim)', bg: 'var(--bg-card)', icon: XCircle },
};

function formatAge(seconds?: number): string {
    if (seconds == null) return '—';
    if (seconds < 60) return `${seconds}s`;
    if (seconds < 3600) return `${Math.floor(seconds / 60)}m`;
    return `${Math.floor(seconds / 3600)}h ${Math.floor((seconds % 3600) / 60)}m`;
}

export default function RelayMonitor() {
    const { data: channels = [], isLoading } = useQuery<RelayChannel[]>({
        queryKey: ['relay-channels'],
        queryFn: fetchRelayChannels,
        refetchInterval: 15000,
    });

    const stats = {
        total: channels.length,
        healthy: channels.filter((c) => c.status === 'healthy').length,
        backlogged: channels.filter((c) => c.status === 'backlogged').length,
        stuck: channels.filter((c) => c.status === 'stuck').length,
    };

    return (
        <div>
            {/* Stats */}
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: 12, marginBottom: 20 }}>
                {[
                    { label: 'Total Channels', value: stats.total, color: 'var(--accent-blue)', icon: Radio },
                    { label: 'Healthy', value: stats.healthy, color: 'var(--accent-green)', icon: CheckCircle2 },
                    { label: 'Backlogged', value: stats.backlogged, color: 'var(--accent-yellow)', icon: Clock },
                    { label: 'Stuck', value: stats.stuck, color: 'var(--accent-red)', icon: AlertTriangle },
                ].map((stat, i) => {
                    const Icon = stat.icon;
                    return (
                        <div key={i} className="glass-card" style={{ padding: 16 }}>
                            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: 8 }}>
                                <span style={{ fontSize: 11, color: 'var(--text-dim)', fontWeight: 500, textTransform: 'uppercase', letterSpacing: '0.04em' }}>{stat.label}</span>
                                <Icon size={14} style={{ color: stat.color }} />
                            </div>
                            <div style={{ fontSize: 28, fontWeight: 700, color: stat.color }}>{stat.value}</div>
                        </div>
                    );
                })}
            </div>

            {/* Channel Matrix Grid */}
            <div className="glass-card" style={{ padding: 20 }}>
                <h3 style={{ fontSize: 14, fontWeight: 600, marginBottom: 16 }}>IBC Channel Health Matrix</h3>

                {isLoading ? (
                    <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))', gap: 12 }}>
                        {Array.from({ length: 6 }).map((_, i) => (
                            <div key={i} className="skeleton" style={{ height: 80, borderRadius: 8 }} />
                        ))}
                    </div>
                ) : (
                    <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))', gap: 12 }}>
                        {channels.map((ch) => (
                            <ChannelCard key={`${ch.sourceChain}-${ch.channelId}`} channel={ch} />
                        ))}
                    </div>
                )}
            </div>
        </div>
    );
}

function ChannelCard({ channel: ch }: { channel: RelayChannel }) {
    const cfg = statusConfig[ch.status] || statusConfig.healthy;
    const Icon = cfg.icon;

    return (
        <div
            className="animate-fade-in"
            style={{
                background: 'var(--bg-card)',
                border: `1px solid ${ch.status === 'stuck' ? 'var(--accent-red)' : 'var(--border-dim)'}`,
                borderRadius: 10,
                padding: 14,
                transition: 'all 0.2s ease',
                cursor: 'pointer',
            }}
            onMouseEnter={(e) => { e.currentTarget.style.borderColor = cfg.color; e.currentTarget.style.background = 'var(--bg-hover)'; }}
            onMouseLeave={(e) => { e.currentTarget.style.borderColor = ch.status === 'stuck' ? 'var(--accent-red)' : 'var(--border-dim)'; e.currentTarget.style.background = 'var(--bg-card)'; }}
        >
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: 10 }}>
                <div>
                    <div style={{ fontSize: 13, fontWeight: 600 }}>
                        {CHAIN_NAMES[ch.sourceChain as keyof typeof CHAIN_NAMES] || ch.sourceChain}
                        {' → '}
                        {CHAIN_NAMES[ch.destChain as keyof typeof CHAIN_NAMES] || ch.destChain}
                    </div>
                    <div style={{ fontSize: 11, color: 'var(--text-dim)', marginTop: 2 }}>{ch.channelId} / {ch.portId}</div>
                </div>
                <span style={{
                    display: 'flex', alignItems: 'center', gap: 4,
                    padding: '3px 8px', borderRadius: 6,
                    fontSize: 10, fontWeight: 700,
                    background: cfg.bg, color: cfg.color,
                    textTransform: 'uppercase',
                }}>
                    <Icon size={10} />
                    {ch.status}
                </span>
            </div>
            <div style={{ display: 'flex', gap: 16, fontSize: 11 }}>
                <div>
                    <span style={{ color: 'var(--text-dim)' }}>Pending: </span>
                    <span style={{ fontWeight: 600, color: ch.pendingPackets > 10 ? 'var(--accent-red)' : 'var(--text-primary)' }}>
                        {ch.pendingPackets}
                    </span>
                </div>
                <div>
                    <span style={{ color: 'var(--text-dim)' }}>Oldest: </span>
                    <span style={{ fontWeight: 600 }}>{formatAge(ch.oldestPacketAgeS)}</span>
                </div>
            </div>
        </div>
    );
}
