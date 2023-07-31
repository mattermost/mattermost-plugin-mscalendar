export type CreateEventPayload = {
    all_day: boolean;
    attendees: string[]; // list of Mattermost UserIDs or email addresses
    start_time: string;
    end_time: string;
    reminder?: number;
    body?: string;
    subject: string;
    location?: EventLocation;
}

export type EventLocation = {
    display_name: string;
    street: string;
    city: string;
    state: string;
    postalcode: string;
    country: string;
}
