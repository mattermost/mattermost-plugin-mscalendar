import {RemoteEvent} from '@/types/calendar';
import type {ProviderConfig} from '@/reducers';
import manifest from '../manifest';

class Client {
    private baseUrl: string;
    private pluginUrl: string;
    private pluginApiUrl: string;

    constructor() {
        this.baseUrl = window.location.origin;
        const basePath = (window as Window & {basename?: string}).basename ?? '';
        this.pluginUrl = this.baseUrl + basePath + '/plugins/' + manifest.id;
        this.pluginApiUrl = this.pluginUrl + '/api/v1';
    }

    getProviderConfiguration = async (): Promise<ProviderConfig | null> => {
        const config = await this.doGet<ProviderConfig>(`${this.pluginApiUrl}/provider`);
        return config || null;
    };

    getCalendarEvents = async (from: string, to: string): Promise<RemoteEvent[] | null> => {
        const params = new URLSearchParams({from, to});
        const events = await this.doGet<RemoteEvent[]>(`${this.pluginApiUrl}/events/view?${params.toString()}`);
        return events || [];
    };

    private doGet = async <T>(url: string): Promise<T | null> => {
        const response = await fetch(url, {
            method: 'GET',
            credentials: 'same-origin',
            headers: {
                'Content-Type': 'application/json',
                'X-Requested-With': 'XMLHttpRequest',
            },
        });

        if (!response.ok) {
            const errorData = await response.json().catch(() => ({}));
            throw new Error(errorData.error || `Request failed with status ${response.status}`);
        }

        const text = await response.text();
        if (!text.trim()) {
            return null;
        }
        try {
            return JSON.parse(text) as T;
        } catch {
            throw new Error(`Failed to parse response JSON: ${text}`);
        }
    };
}

const client = new Client();
export default client;
