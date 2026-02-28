import { useEffect, useRef, useCallback } from 'react';
import { useOpportunityStore } from '../store';
import type { Opportunity } from '../types';
import { parseWSOpportunity } from '../services/api';

interface UseWebSocketOptions {
    url: string;
    onOpportunity?: (opp: Opportunity) => void;
}

export function useWebSocket({ url, onOpportunity }: UseWebSocketOptions) {
    const wsRef = useRef<WebSocket | null>(null);
    const reconnectTimerRef = useRef<number | undefined>(undefined);
    const reconnectCountRef = useRef(0);
    const maxReconnects = 10;
    const upsert = useOpportunityStore((s) => s.upsert);

    const connect = useCallback(() => {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${protocol}//${window.location.host}${url}`;

        try {
            const ws = new WebSocket(wsUrl);
            wsRef.current = ws;

            ws.onopen = () => {
                reconnectCountRef.current = 0;
            };

            ws.onmessage = (event) => {
                try {
                    const data = JSON.parse(event.data);
                    if (data.type === 'opportunity' && data.payload) {
                        const opp = parseWSOpportunity(data.payload);
                        upsert(opp);
                        onOpportunity?.(opp);
                    }
                } catch {
                    // ignore parse errors
                }
            };

            ws.onclose = () => {
                if (reconnectCountRef.current < maxReconnects) {
                    const delay = Math.min(2000 * Math.pow(2, reconnectCountRef.current), 30000);
                    reconnectTimerRef.current = window.setTimeout(() => {
                        reconnectCountRef.current++;
                        connect();
                    }, delay);
                }
            };

            ws.onerror = () => {
                ws.close();
            };
        } catch {
            // WebSocket construction failed, retry
            if (reconnectCountRef.current < maxReconnects) {
                reconnectTimerRef.current = window.setTimeout(() => {
                    reconnectCountRef.current++;
                    connect();
                }, 3000);
            }
        }
    }, [url, upsert, onOpportunity]);

    useEffect(() => {
        connect();
        return () => {
            if (reconnectTimerRef.current) clearTimeout(reconnectTimerRef.current);
            wsRef.current?.close();
        };
    }, [connect]);

    return {
        isConnected: wsRef.current?.readyState === WebSocket.OPEN,
    };
}
