import { useState } from 'react';
import { Bell, Webhook, MessageCircle, Monitor, Save, Key, Plus, Shield } from 'lucide-react';

export default function Settings() {
    const [notifType, setNotifType] = useState('inapp');
    const [minProfit, setMinProfit] = useState('10');
    const [minSpread, setMinSpread] = useState('0.5');
    const [webhookUrl, setWebhookUrl] = useState('');
    const [telegramId, setTelegramId] = useState('');
    const [saved, setSaved] = useState(false);

    const handleSave = () => {
        setSaved(true);
        setTimeout(() => setSaved(false), 2000);
    };

    return (
        <div style={{ maxWidth: 700 }}>
            {/* Alert Configuration */}
            <div className="glass-card" style={{ padding: 24, marginBottom: 20 }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 20 }}>
                    <Bell size={18} style={{ color: 'var(--accent-blue)' }} />
                    <h2 style={{ fontSize: 16, fontWeight: 700 }}>Alert Configuration</h2>
                </div>

                <div style={{ display: 'grid', gap: 16 }}>
                    {/* Min Profit */}
                    <div>
                        <label style={{ display: 'block', fontSize: 12, fontWeight: 600, color: 'var(--text-dim)', marginBottom: 6, textTransform: 'uppercase' }}>
                            Minimum Net Profit (USD)
                        </label>
                        <input
                            type="number"
                            value={minProfit}
                            onChange={(e) => setMinProfit(e.target.value)}
                            style={{
                                width: '100%', padding: '10px 14px',
                                background: 'var(--bg-card)', border: '1px solid var(--border-dim)',
                                borderRadius: 8, color: 'var(--text-primary)',
                                fontSize: 14, outline: 'none',
                            }}
                        />
                    </div>

                    {/* Min Spread */}
                    <div>
                        <label style={{ display: 'block', fontSize: 12, fontWeight: 600, color: 'var(--text-dim)', marginBottom: 6, textTransform: 'uppercase' }}>
                            Minimum Spread (%)
                        </label>
                        <input
                            type="number"
                            step="0.1"
                            value={minSpread}
                            onChange={(e) => setMinSpread(e.target.value)}
                            style={{
                                width: '100%', padding: '10px 14px',
                                background: 'var(--bg-card)', border: '1px solid var(--border-dim)',
                                borderRadius: 8, color: 'var(--text-primary)',
                                fontSize: 14, outline: 'none',
                            }}
                        />
                    </div>

                    {/* Notification Type */}
                    <div>
                        <label style={{ display: 'block', fontSize: 12, fontWeight: 600, color: 'var(--text-dim)', marginBottom: 8, textTransform: 'uppercase' }}>
                            Notification Method
                        </label>
                        <div style={{ display: 'flex', gap: 8 }}>
                            {[
                                { id: 'inapp', label: 'In-App', icon: Monitor },
                                { id: 'webhook', label: 'Webhook', icon: Webhook },
                                { id: 'telegram', label: 'Telegram', icon: MessageCircle },
                            ].map(({ id, label, icon: Icon }) => (
                                <button
                                    key={id}
                                    onClick={() => setNotifType(id)}
                                    style={{
                                        flex: 1,
                                        display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 6,
                                        padding: '10px 16px',
                                        borderRadius: 8,
                                        border: `1px solid ${notifType === id ? 'var(--accent-blue)' : 'var(--border-dim)'}`,
                                        background: notifType === id ? 'var(--accent-blue-dim)' : 'var(--bg-card)',
                                        color: notifType === id ? 'var(--accent-blue)' : 'var(--text-secondary)',
                                        cursor: 'pointer',
                                        fontSize: 13,
                                        fontWeight: 600,
                                    }}
                                >
                                    <Icon size={16} />
                                    {label}
                                </button>
                            ))}
                        </div>
                    </div>

                    {notifType === 'webhook' && (
                        <div>
                            <label style={{ display: 'block', fontSize: 12, fontWeight: 600, color: 'var(--text-dim)', marginBottom: 6, textTransform: 'uppercase' }}>
                                Webhook URL
                            </label>
                            <input
                                type="url"
                                value={webhookUrl}
                                onChange={(e) => setWebhookUrl(e.target.value)}
                                placeholder="https://your-webhook.example.com/hook"
                                style={{
                                    width: '100%', padding: '10px 14px',
                                    background: 'var(--bg-card)', border: '1px solid var(--border-dim)',
                                    borderRadius: 8, color: 'var(--text-primary)',
                                    fontSize: 14, outline: 'none',
                                }}
                            />
                        </div>
                    )}

                    {notifType === 'telegram' && (
                        <div>
                            <label style={{ display: 'block', fontSize: 12, fontWeight: 600, color: 'var(--text-dim)', marginBottom: 6, textTransform: 'uppercase' }}>
                                Telegram Chat ID
                            </label>
                            <input
                                type="text"
                                value={telegramId}
                                onChange={(e) => setTelegramId(e.target.value)}
                                placeholder="Enter your Telegram chat ID"
                                style={{
                                    width: '100%', padding: '10px 14px',
                                    background: 'var(--bg-card)', border: '1px solid var(--border-dim)',
                                    borderRadius: 8, color: 'var(--text-primary)',
                                    fontSize: 14, outline: 'none',
                                }}
                            />
                        </div>
                    )}

                    <button
                        onClick={handleSave}
                        style={{
                            display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 8,
                            padding: '12px',
                            borderRadius: 8,
                            border: 'none',
                            background: saved ? 'var(--accent-green)' : 'var(--accent-blue)',
                            color: '#fff',
                            cursor: 'pointer',
                            fontSize: 14,
                            fontWeight: 700,
                            transition: 'background 0.2s',
                        }}
                    >
                        <Save size={16} />
                        {saved ? 'Saved!' : 'Save Alert Config'}
                    </button>
                </div>
            </div>

            {/* API Key Management */}
            <div className="glass-card" style={{ padding: 24 }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 20 }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                        <Shield size={18} style={{ color: 'var(--accent-purple)' }} />
                        <h2 style={{ fontSize: 16, fontWeight: 700 }}>API Keys</h2>
                    </div>
                    <button
                        style={{
                            display: 'flex', alignItems: 'center', gap: 4,
                            padding: '6px 12px',
                            borderRadius: 6,
                            border: '1px solid var(--accent-blue)',
                            background: 'var(--accent-blue-dim)',
                            color: 'var(--accent-blue)',
                            cursor: 'pointer',
                            fontSize: 12,
                            fontWeight: 600,
                        }}
                    >
                        <Plus size={14} />
                        Generate Key
                    </button>
                </div>

                <div style={{
                    padding: 32,
                    textAlign: 'center',
                    color: 'var(--text-dim)',
                    fontSize: 13,
                    border: '1px dashed var(--border-dim)',
                    borderRadius: 8,
                }}>
                    <Key size={24} style={{ margin: '0 auto 8px', opacity: 0.5 }} />
                    <div>No API keys configured</div>
                    <div style={{ fontSize: 11, marginTop: 4 }}>Generate a key to access the REST API and gRPC endpoints</div>
                </div>
            </div>
        </div>
    );
}
