export type EventLocation = {
    display_name: string;
    street: string;
    city: string;
    state: string;
    postalcode: string;
    country: string;
}

export type CreateEventPayload = {
    all_day: boolean;
    attendees: string[]; // list of Mattermost UserIDs or email addresses
    date: string;
    start_time: string;
    end_time: string;
    reminder?: number;
    description?: string;
    subject: string;
    location?: EventLocation;
}
