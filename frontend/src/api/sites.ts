import { apiFetch } from './client';

export type Site = {
  id: string;
  name: string;
  url: string;
  intervalHours: number;
  hoursSinceLastCheck: number;
  isOverdue: boolean;
  siteStreak: number;
};

export type SitesResponse = {
  sites: Site[];
  userStreak: number;
};

export type RegisterSiteInput = {
  name: string;
  url: string;
  intervalHours: number;
};

export function fetchSites(): Promise<SitesResponse> {
  return apiFetch<SitesResponse>('/api/sites');
}

export function registerSite(
  input: RegisterSiteInput,
): Promise<{ id: string }> {
  return apiFetch<{ id: string }>('/api/sites', {
    method: 'POST',
    body: JSON.stringify(input),
  });
}

export function deleteSite(id: string): Promise<void> {
  return apiFetch<void>(`/api/sites/${id}`, { method: 'DELETE' });
}
