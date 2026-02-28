import { X, Copy, ArrowRight, Check } from 'lucide-react';
import type { Opportunity } from '../../types';
import { CHAIN_NAMES, CHAIN_COLORS } from '../../types';
import { useState } from 'react';

interface Props {
    opportunity: Opportunity;
    onClose: () => void;
}

export default function OpportunityDrawer({ opportunity: opp, onClose }: Props) {
    const [copied, setCopied] = useState(false);

    const handleCopy = () => {
        navigator.clipboard.writeText(typeof opp.calldataJson === 'string' ? opp.calldataJson : JSON.stringify(opp.calldataJson, null, 2));
        setCopied(true);
        setTimeout(() => setCopied(false), 2000);
    };

    return (
        <div
            className="animate-fade-in"
            style={{
                width: 400,
                minWidth: 400,
                background: 'var(--bg-secondary)',
                borderLeft: '1px solid var(--border-dim)',
                height: '100%',
                overflow: 'auto',
                padding: 20,
            }}
        >
            {/* Header */}
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 20 }}>
                <h2 style={{ fontSize: 16, fontWeight: 700 }}>{opp.assetPair}</h2>
                <button
                    onClick={onClose}
                    aria-label="Close drawer"
                    style={{
                        background: 'var(--bg-card)',
                        border: '1px solid var(--border-dim)',
                        borderRadius: 6,
                        padding: 6,
                        cursor: 'pointer',
                        color: 'var(--text-secondary)',
                    }}
                >
                    <X size={16} />
                </button>
            </div>

            {/* Spread and Profit */}
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12, marginBottom: 20 }}>
                <div className="glass-card" style={{ padding: 14, textAlign: 'center' }}>
                    <div style={{ fontSize: 11, color: 'var(--text-dim)', marginBottom: 4, textTransform: 'uppercase' }}>Spread</div>
                    <div style={{
                        fontSize: 24, fontWeight: 800,
                        color: opp.spreadPct >= 2 ? 'var(--accent-green)' : opp.spreadPct >= 0.5 ? 'var(--accent-yellow)' : 'var(--text-dim)',
                    }}>
                        {opp.spreadPct.toFixed(3)}%
                    </div>
                </div>
                <div className="glass-card" style={{ padding: 14, textAlign: 'center' }}>
                    <div style={{ fontSize: 11, color: 'var(--text-dim)', marginBottom: 4, textTransform: 'uppercase' }}>Net Profit</div>
                    <div style={{ fontSize: 24, fontWeight: 800, color: 'var(--accent-green)' }}>
                        ${opp.netProfitUsd.toFixed(2)}
                    </div>
                </div>
            </div>

            {/* Path Diagram */}
            <div className="glass-card" style={{ padding: 16, marginBottom: 16 }}>
                <div style={{ fontSize: 12, fontWeight: 600, color: 'var(--text-dim)', marginBottom: 12, textTransform: 'uppercase' }}>
                    Path ({opp.pathHops} hop{opp.pathHops > 1 ? 's' : ''})
                </div>
                <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 8, flexWrap: 'wrap' }}>
                    {opp.path.map((node, i) => (
                        <div key={i} style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                            <div style={{
                                padding: '8px 14px',
                                borderRadius: 8,
                                border: `2px solid ${CHAIN_COLORS[node.chain] || '#444'}`,
                                background: 'var(--bg-card)',
                                textAlign: 'center',
                            }}>
                                <div style={{ fontSize: 11, fontWeight: 700, color: CHAIN_COLORS[node.chain] || '#eee' }}>
                                    {CHAIN_NAMES[node.chain]}
                                </div>
                                <div style={{ fontSize: 10, color: 'var(--text-dim)', marginTop: 2 }}>{node.dex}</div>
                                <div style={{ fontSize: 12, fontWeight: 600, color: 'var(--text-primary)', marginTop: 4 }}>
                                    ${node.price.toFixed(4)}
                                </div>
                            </div>
                            {i < opp.path.length - 1 && (
                                <ArrowRight size={16} style={{ color: 'var(--accent-blue)' }} />
                            )}
                        </div>
                    ))}
                </div>
            </div>

            {/* Fee Breakdown */}
            <div className="glass-card" style={{ padding: 16, marginBottom: 16 }}>
                <div style={{ fontSize: 12, fontWeight: 600, color: 'var(--text-dim)', marginBottom: 12, textTransform: 'uppercase' }}>
                    Fee Breakdown
                </div>
                <table style={{ width: '100%', fontSize: 12 }}>
                    <tbody>
                        {opp.feeBreakdown.map((fee, i) => (
                            <tr key={i} style={{ borderBottom: '1px solid var(--border-dim)' }}>
                                <td style={{ padding: '6px 0', color: 'var(--text-secondary)' }}>
                                    {CHAIN_NAMES[fee.chain]} — {fee.feeType.replace('_', ' ')}
                                </td>
                                <td style={{ padding: '6px 0', textAlign: 'right', fontWeight: 600 }}>
                                    ${fee.amountUsd.toFixed(2)}
                                </td>
                            </tr>
                        ))}
                        <tr>
                            <td style={{ padding: '8px 0', fontWeight: 700, fontSize: 13 }}>Slippage Est.</td>
                            <td style={{ padding: '8px 0', textAlign: 'right', fontWeight: 700, fontSize: 13 }}>
                                {opp.slippageEstPct.toFixed(2)}%
                            </td>
                        </tr>
                        <tr style={{ borderTop: '2px solid var(--border-bright)' }}>
                            <td style={{ padding: '8px 0', fontWeight: 700, fontSize: 13, color: 'var(--accent-green)' }}>Net Profit</td>
                            <td style={{ padding: '8px 0', textAlign: 'right', fontWeight: 800, fontSize: 14, color: 'var(--accent-green)' }}>
                                ${opp.netProfitUsd.toFixed(2)}
                            </td>
                        </tr>
                    </tbody>
                </table>
            </div>

            {/* Details */}
            <div className="glass-card" style={{ padding: 16, marginBottom: 16 }}>
                <div style={{ fontSize: 12, fontWeight: 600, color: 'var(--text-dim)', marginBottom: 12, textTransform: 'uppercase' }}>
                    Details
                </div>
                <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 8, fontSize: 12 }}>
                    <div style={{ color: 'var(--text-dim)' }}>Input Amount</div>
                    <div style={{ textAlign: 'right', fontWeight: 600 }}>${opp.inputAmountUsd.toFixed(0)}</div>
                    <div style={{ color: 'var(--text-dim)' }}>Gross Profit</div>
                    <div style={{ textAlign: 'right', fontWeight: 600 }}>${opp.grossProfitUsd.toFixed(2)}</div>
                    <div style={{ color: 'var(--text-dim)' }}>Status</div>
                    <div style={{ textAlign: 'right' }}>
                        <span style={{
                            padding: '2px 8px',
                            borderRadius: 4,
                            fontSize: 11,
                            fontWeight: 600,
                            background: opp.status === 'live' ? 'var(--accent-green-dim)' : 'var(--accent-red-dim)',
                            color: opp.status === 'live' ? 'var(--accent-green)' : 'var(--accent-red)',
                        }}>
                            {opp.status.toUpperCase()}
                        </span>
                    </div>
                    <div style={{ color: 'var(--text-dim)' }}>Discovered</div>
                    <div style={{ textAlign: 'right', fontWeight: 500 }}>{new Date(opp.discoveredAt).toLocaleTimeString()}</div>
                </div>
            </div>

            {/* Calldata */}
            <div className="glass-card" style={{ padding: 16 }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 8 }}>
                    <span style={{ fontSize: 12, fontWeight: 600, color: 'var(--text-dim)', textTransform: 'uppercase' }}>
                        Execution Calldata (IBC MsgTransfer)
                    </span>
                    <button
                        onClick={handleCopy}
                        aria-label="Copy calldata"
                        style={{
                            display: 'flex', alignItems: 'center', gap: 4,
                            background: copied ? 'var(--accent-green-dim)' : 'var(--bg-card)',
                            border: `1px solid ${copied ? 'var(--accent-green)' : 'var(--border-dim)'}`,
                            borderRadius: 6,
                            padding: '4px 10px',
                            cursor: 'pointer',
                            color: copied ? 'var(--accent-green)' : 'var(--text-secondary)',
                            fontSize: 11,
                            fontWeight: 500,
                            transition: 'all 0.2s',
                        }}
                    >
                        {copied ? <Check size={12} /> : <Copy size={12} />}
                        {copied ? 'Copied' : 'Copy'}
                    </button>
                </div>
                <pre style={{
                    background: 'var(--bg-primary)',
                    borderRadius: 6,
                    padding: 12,
                    fontSize: 10,
                    lineHeight: 1.5,
                    overflow: 'auto',
                    maxHeight: 200,
                    color: 'var(--text-secondary)',
                    border: '1px solid var(--border-dim)',
                }}>
                    {typeof opp.calldataJson === 'string' ? opp.calldataJson : JSON.stringify(opp.calldataJson, null, 2)}
                </pre>
            </div>
        </div>
    );
}
