import { PUBLIC_BASE_URL } from '$env/static/public';

export const apiEndpoint = (param: string) => {
    const cleanParam = param.replace(/^[/]*(api\/?)+/, '').replace(/^\/+/, '');
    return `${PUBLIC_BASE_URL}/api/${cleanParam}`;
}

export const eventEndpoint = (param: string) => {
    const cleanParam = param.replace(/^[/]*(event\/?)+/, '').replace(/^\/+/, '');
    return `${PUBLIC_BASE_URL}/events/${cleanParam}`;
}