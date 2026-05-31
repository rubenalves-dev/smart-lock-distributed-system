import { ref } from 'vue';
import { API_BASE_URL } from '@/config';

export interface MFARequest {
  id: number;
  rfid_uid: string;
  device_id: string;
  fails: number;
  distance_cm: number;
  light_level: number;
  classification: number;
  confidence: number;
  recommendation: string;
  status: string; // 'pending', 'approved', 'rejected'
  created_at: string;
  updated_at: string;
  user_name?: string;
}

// Shared state between components (Sidebar, Top Header notifications, page View)
const requests = ref<MFARequest[]>([]);
const loading = ref(false);
const wsConnected = ref(false);
const error = ref<string | null>(null);

let socket: WebSocket | null = null;

export function useMfa() {
  const fetchRequests = async () => {
    loading.value = true;
    error.value = null;
    try {
      const response = await fetch(`${API_BASE_URL}/mfa/requests`);
      if (!response.ok) {
        throw new Error('Failed to fetch MFA requests');
      }
      requests.value = await response.json();
    } catch (err: any) {
      error.value = err.message || 'Error fetching requests';
    } finally {
      loading.value = false;
    }
  };

  const approveRequest = async (id: number): Promise<boolean> => {
    try {
      const response = await fetch(`${API_BASE_URL}/mfa/requests/${id}/approve`, {
        method: 'POST',
      });
      if (response.ok) {
        const req = requests.value.find(r => r.id === id);
        if (req) {
          req.status = 'approved';
        }
        return true;
      }
      return false;
    } catch {
      return false;
    }
  };

  const rejectRequest = async (id: number): Promise<boolean> => {
    try {
      const response = await fetch(`${API_BASE_URL}/mfa/requests/${id}/reject`, {
        method: 'POST',
      });
      if (response.ok) {
        const req = requests.value.find(r => r.id === id);
        if (req) {
          req.status = 'rejected';
        }
        return true;
      }
      return false;
    } catch {
      return false;
    }
  };

  const connectWebSocket = () => {
    if (socket && (socket.readyState === WebSocket.OPEN || socket.readyState === WebSocket.CONNECTING)) {
      return;
    }

    const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    let wsHost = API_BASE_URL.replace(/^https?:\/\//, '');
    
    // Fallback if host is relative (e.g. /api in local dev proxy)
    if (wsHost.startsWith('/')) {
      wsHost = window.location.host + wsHost;
    }
    
    const wsUrl = `${wsProtocol}//${wsHost}/ws`;

    console.log('[MFA WS] Connecting to:', wsUrl);
    socket = new WebSocket(wsUrl);

    socket.onopen = () => {
      console.log('[MFA WS] Connected successfully');
      wsConnected.value = true;
    };

    socket.onmessage = (event) => {
      try {
        const payload = JSON.parse(event.data);
        console.log('[MFA WS] Event received:', payload);
        if (payload.type === 'mfa_request') {
          // Push new request to the top
          requests.value.unshift(payload.data);
        } else if (payload.type === 'mfa_approved') {
          const req = requests.value.find(r => r.id === payload.id);
          if (req) req.status = 'approved';
        } else if (payload.type === 'mfa_rejected') {
          const req = requests.value.find(r => r.id === payload.id);
          if (req) req.status = 'rejected';
        }
      } catch (err) {
        console.error('[MFA WS] Parse error:', err);
      }
    };

    socket.onclose = () => {
      console.log('[MFA WS] Connection closed. Reconnecting in 5 seconds...');
      wsConnected.value = false;
      setTimeout(connectWebSocket, 5000);
    };

    socket.onerror = (err) => {
      console.error('[MFA WS] WebSocket error:', err);
      socket?.close();
    };
  };

  return {
    requests,
    loading,
    error,
    wsConnected,
    fetchRequests,
    approveRequest,
    rejectRequest,
    connectWebSocket,
  };
}
