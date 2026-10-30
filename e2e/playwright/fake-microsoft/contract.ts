// Hand-written response contracts for the Microsoft Graph and Entra ID calls the
// plugin makes. Graph's published OpenAPI metadata marks nearly everything optional,
// so these list the fields the plugin actually reads. The fake checks each of its
// own responses, and the nightly job checks the live service against the same lists.

export type Contract = {name: string; required: string[]};

export const CONTRACTS: Record<string, Contract> = {
    token: {name: 'token response', required: ['token_type', 'access_token', 'expires_in', 'scope']},
    appToken: {name: 'app-only token response', required: ['token_type', 'access_token', 'expires_in']},
    user: {name: 'user', required: ['id', 'displayName', 'userPrincipalName', 'mail']},
    event: {
        name: 'event',
        required: [
            'id', 'iCalUId', 'subject', 'start.dateTime', 'start.timeZone', 'end.dateTime', 'end.timeZone',
            'isCancelled', 'isAllDay', 'showAs', 'organizer.emailAddress.address', 'responseStatus.response',
        ],
    },
    calendar: {name: 'calendar', required: ['id', 'name']},
    mailboxSettings: {name: 'mailbox settings', required: ['timeZone']},
    subscription: {
        name: 'subscription',
        required: ['id', 'resource', 'changeType', 'notificationUrl', 'expirationDateTime', 'clientState'],
    },
    notification: {
        name: 'change notification',
        required: ['subscriptionId', 'subscriptionExpirationDateTime', 'changeType', 'resource', 'resourceData.@odata.type'],
    },
    graphError: {name: 'Graph error', required: ['error.code', 'error.message']},
    createEventRequest: {name: 'create event request', required: ['subject', 'start.dateTime', 'start.timeZone', 'end.dateTime', 'end.timeZone']},
    createSubscriptionRequest: {name: 'create subscription request', required: ['resource', 'changeType', 'notificationUrl', 'expirationDateTime']},
};

function lookup(value: unknown, path: string): unknown {
    // A dot inside an "@odata.type" key is part of the key, not a path separator.
    const parts = path.split(/\.(?!type$)/);
    let current: any = value;
    for (const part of parts) {
        if (current === null || typeof current !== 'object' || !(part in current)) {
            return undefined;
        }
        current = current[part];
    }
    return current;
}

// Returns the required fields that are missing or null.
export function missingFields(contractKey: string, value: unknown): string[] {
    const contract = CONTRACTS[contractKey];
    return contract.required.filter((path) => {
        const found = lookup(value, path);
        return found === undefined || found === null;
    });
}
