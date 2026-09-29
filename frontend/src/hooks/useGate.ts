import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  getGateTokenStatus,
  getNotificationSettings,
  issueGateToken,
  NotificationMode,
  updateNotificationMode,
} from '../api/gate';

const SETTINGS_KEY = ['notification-settings'] as const;
const GATE_TOKEN_KEY = ['gate-token'] as const;

export function useNotificationSettings(enabled: boolean) {
  return useQuery({
    queryKey: SETTINGS_KEY,
    queryFn: getNotificationSettings,
    enabled,
  });
}

export function useUpdateNotificationMode() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (mode: NotificationMode) => updateNotificationMode(mode),
    onSuccess: (data) => {
      queryClient.setQueryData(SETTINGS_KEY, data);
    },
  });
}

export function useGateTokenStatus(enabled: boolean) {
  return useQuery({
    queryKey: GATE_TOKEN_KEY,
    queryFn: getGateTokenStatus,
    enabled,
  });
}

export function useIssueGateToken() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: issueGateToken,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: GATE_TOKEN_KEY });
    },
  });
}
