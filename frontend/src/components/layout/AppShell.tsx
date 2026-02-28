import React from 'react';
import { Activity, Radio, BarChart3, Settings, Zap, Wifi, WifiOff } from 'lucide-react';
import { useUIStore } from '../../store';

const navItems = [
    { id: 'dashboard', label: 'Dashboard', icon: Activity },
    { id: 'relay', label: 'Relay Monitor', icon: Radio },
    { id: 'analytics', label: 'Analytics', icon: BarChart3 },
    { id: 'settings', label: 'Settings', icon: Settings },
];

interface AppShellProps {
    children: React.ReactNode;
    chainCount: number;
    opportunityCount: number;
    wsConnected: boolean;
}

export default function AppShell({ children, chainCount, opportunityCount, wsConnected }: AppShellProps) {
    const { activeTab, setActiveTab } = useUIStore();

    return (
        <div style={{ display: 'flex', height: '100vh', width: '100vw', overflow: 'hidden' }}>
            {/* Sidebar */}
            <aside style={{
                width: 220,
                minWidth: 220,
                background: 'var(--bg-secondary)',
                borderRight: '1px solid var(--border-dim)',
                display: 'flex',
                flexDirection: 'column',
                padding: '16px 0',
            }}>
                {/* Logo */}
                <div style={{
                    padding: '0 20px 20px',
                    borderBottom: '1px solid var(--border-dim)',
                    marginBottom: 8,
                }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                        <Zap size={24} style={{ color: 'var(--accent-blue)' }} />
                        <div>
                            <div style={{ fontWeight: 700, fontSize: 15, letterSpacing: '-0.02em' }}>ArbEngine</div>
                            <div style={{ fontSize: 11, color: 'var(--text-dim)', fontWeight: 500 }}>COSMOS PRO</div>
                        </div>
                    </div>
                </div>

                {/* Nav Items */}
                <nav style={{ flex: 1, padding: '4px 8px' }}>
                    {navItems.map((item) => {
                        const Icon = item.icon;
                        const isActive = activeTab === item.id;
                        return (
                            <button
                                key={item.id}
                                onClick={() => setActiveTab(item.id)}
                                aria-label={item.label}
                                style={{
                                    display: 'flex',
                                    alignItems: 'center',
                                    gap: 10,
                                    width: '100%',
                                    padding: '10px 12px',
                                    border: 'none',
                                    borderRadius: 8,
                                    cursor: 'pointer',
                                    fontSize: 13,
                                    fontWeight: isActive ? 600 : 400,
                                    color: isActive ? 'var(--accent-blue)' : 'var(--text-secondary)',
                                    background: isActive ? 'var(--accent-blue-dim)' : 'transparent',
                                    marginBottom: 2,
                                    transition: 'all 0.15s ease',
                                }}
                            >
                                <Icon size={18} />
                                {item.label}
                            </button>
                        );
                    })}
                </nav>

                {/* Status Footer */}
                <div style={{
                    padding: '12px 16px',
                    borderTop: '1px solid var(--border-dim)',
                    fontSize: 11,
                    color: 'var(--text-dim)',
                }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 6, marginBottom: 6 }}>
                        {wsConnected
                            ? <Wifi size={12} style={{ color: 'var(--accent-green)' }} />
                            : <WifiOff size={12} style={{ color: 'var(--accent-red)' }} />
                        }
                        <span>{wsConnected ? 'Connected' : 'Disconnected'}</span>
                    </div>
                    <div>{chainCount} chains • {opportunityCount} ops</div>
                </div>
            </aside>

            {/* Main Content */}
            <main style={{
                flex: 1,
                overflow: 'auto',
                background: 'var(--bg-primary)',
            }}>
                {/* Top Bar */}
                <header style={{
                    height: 56,
                    background: 'var(--bg-secondary)',
                    borderBottom: '1px solid var(--border-dim)',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'space-between',
                    padding: '0 24px',
                    position: 'sticky',
                    top: 0,
                    zIndex: 50,
                }}>
                    <h1 style={{ fontSize: 16, fontWeight: 600, letterSpacing: '-0.01em' }}>
                        {navItems.find((n) => n.id === activeTab)?.label || 'Dashboard'}
                    </h1>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
                        <div style={{
                            display: 'flex',
                            alignItems: 'center',
                            gap: 6,
                            padding: '4px 12px',
                            borderRadius: 20,
                            fontSize: 12,
                            fontWeight: 600,
                            background: 'var(--accent-green-dim)',
                            color: 'var(--accent-green)',
                        }}>
                            <div style={{
                                width: 6, height: 6, borderRadius: '50%',
                                background: 'var(--accent-green)',
                                animation: 'pulse 2s ease-in-out infinite',
                            }} />
                            LIVE
                        </div>
                        <span style={{ fontSize: 12, color: 'var(--text-dim)' }}>
                            {new Date().toLocaleTimeString()}
                        </span>
                    </div>
                </header>

                <div style={{ padding: 24 }}>
                    {children}
                </div>
            </main>
        </div>
    );
}
