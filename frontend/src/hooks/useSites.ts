import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  deleteSite,
  fetchSites,
  registerSite,
  RegisterSiteInput,
} from '../api/sites';

const SITES_KEY = ['sites'] as const;

export function useSites() {
  return useQuery({
    queryKey: SITES_KEY,
    queryFn: fetchSites,
  });
}

export function useRegisterSite() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: RegisterSiteInput) => registerSite(input),
    onSuccess: () => {
      // 登録成功後にサイト一覧を再取得
      queryClient.invalidateQueries({ queryKey: SITES_KEY });
    },
  });
}

export function useDeleteSite() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => deleteSite(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: SITES_KEY });
    },
  });
}
