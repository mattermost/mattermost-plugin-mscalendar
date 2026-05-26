export type CreateEventPayload = {
    all_day: boolean;
    attendees: string[];
    date: string;
    start_time: string;
    end_time: string;
    reminder?: number;
    description?: string;
    subject: string;
    location?: string;
    channel_id?: string;
}
