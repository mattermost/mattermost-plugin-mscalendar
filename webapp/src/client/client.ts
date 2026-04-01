import {RemoteEvent} from '@/types/calendar';

import manifest from '../manifest';

class Client {
    private baseUrl: string;
    private pluginUrl: string;
    private pluginApiUrl: string;

    constructor() {
        this.baseUrl = window.location.origin;
        this.pluginUrl = this.baseUrl + '/plugins/' + manifest.id;
        this.pluginApiUrl = this.pluginUrl + '/api/v1';
    }

    getProviderConfiguration = async (): Promise<any> => {
        return this.doGet(`${this.pluginApiUrl}/provider`);
    };

    getCalendarEvents = async (from: string, to: string): Promise<RemoteEvent[]> => {
        const params = new URLSearchParams({from, to});
        return this.doGet<RemoteEvent[]>(`${this.pluginApiUrl}/events/view?${params.toString()}`);
    };

    private doGet = async <T>(url: string): Promise<T> => {
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

        return response.json();
    };
}

const client = new Client();
export default client;
