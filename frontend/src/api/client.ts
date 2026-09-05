const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080';

export class ApiError extends Error {
    status: number
    constructor(status: number, message: string) {
        super(message);
        this.status = status
    }
}

// credentials: 'include' が必須。これが無いとCookieベースセッションのCookieが送られず、
// どのAPIも401 Unauthorizedになる。
export async function apiFetch<T>(path: string, options: RequestInit = {}): Promise<T> {
    const res = await fetch(`${API_BASE_URL}${path}`, {
        ...options,
        credentials: 'include',
        headers: {
            'Content-Type': 'application/json',
            ...options.headers,
        },
    })

    if (!res.ok) {
        throw new ApiError(res.status, `Request failed: ${res.status}`)
    }

    if (res.status === 204) {
        return undefined as T
    }
    return res.json() as Promise<T>
}

export { API_BASE_URL }