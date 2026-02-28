import { useEffect } from 'react';
import { QueryClient, QueryClientProvider, useQuery } from '@tanstack/react-query';
import { Toaster, toast } from 'sonner';
import AppShell from './components/layout/AppShell';
import OpportunityTable from './components/opportunity/OpportunityTable';
import RelayMonitor from './pages/RelayMonitor';
import Analytics from './pages/Analytics';
import Settings from './pages/Settings';
import { useWebSocket } from './hooks/useWebSocket';
import { useOpportunityStore, useUIStore } from './store';
import { fetchOpportunities, fetchChains } from './services/api';
import type { Opportunity } from './types';
import './index.css';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false,
      retry: 2,
      staleTime: 10000,
    },
  },
});

function Dashboard() {
  const setOpportunities = useOpportunityStore((s) => s.setOpportunities);
  const opportunities = useOpportunityStore((s) => s.opportunities);
  const activeTab = useUIStore((s) => s.activeTab);

  // Fetch initial opportunities
  const { data: initialOpps } = useQuery({
    queryKey: ['opportunities'],
    queryFn: () => fetchOpportunities('live', 50),
    refetchInterval: 15000,
  });

  // Fetch chain status
  const { data: chains = [] } = useQuery({
    queryKey: ['chains'],
    queryFn: fetchChains,
    refetchInterval: 30000,
  });

  useEffect(() => {
    if (initialOpps && initialOpps.length > 0) {
      setOpportunities(initialOpps);
    }
  }, [initialOpps, setOpportunities]);

  // WebSocket for live updates
  const { isConnected } = useWebSocket({
    url: '/ws/opportunities',
    onOpportunity: (opp: Opportunity) => {
      if (opp.netProfitUsd > 50) {
        toast.success(`High-value opportunity: ${opp.assetPair}`, {
          description: `$${opp.netProfitUsd.toFixed(2)} net profit (${opp.spreadPct.toFixed(2)}% spread)`,
          duration: 5000,
        });
      }
    },
  });

  const renderPage = () => {
    switch (activeTab) {
      case 'relay': return <RelayMonitor />;
      case 'analytics': return <Analytics />;
      case 'settings': return <Settings />;
      default: return <OpportunityTable />;
    }
  };

  return (
    <AppShell
      chainCount={chains.length || 7}
      opportunityCount={opportunities.length}
      wsConnected={isConnected}
    >
      {renderPage()}
    </AppShell>
  );
}

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <Dashboard />
      <Toaster
        theme="dark"
        position="top-right"
        toastOptions={{
          style: {
            background: 'var(--bg-secondary)',
            border: '1px solid var(--border-bright)',
            color: 'var(--text-primary)',
          },
        }}
      />
    </QueryClientProvider>
  );
}
