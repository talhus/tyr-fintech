import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import api from '../lib/axios';

// Fetch all wallets belonging to the user
export function useWallets() {
  return useQuery({
    queryKey: ['wallets'],
    queryFn: async () => {
      const response = await api.get('/wallets');
      return response.data.wallets || [];
    },
  });
}

// Fetch transaction history for a specific wallet
export function useTransactionHistory(walletId) {
  return useQuery({
    queryKey: ['transactions', walletId],
    queryFn: async () => {
      const response = await api.get(`/transactions/${walletId}`);
      return response.data.data || [];
    },
    enabled: !!walletId,
  });
}

// Create a new wallet mutation
export function useCreateWalletMutation(addToast) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({ userId, currency }) => {
      const response = await api.post('/wallets', { user_id: userId, currency });
      return response.data;
    },
    onSuccess: (data, variables) => {
      addToast(`${variables.currency} wallet activated successfully`, 'success');
      queryClient.invalidateQueries({ queryKey: ['wallets'] });
    },
    onError: (error) => {
      addToast(error.response?.data?.error || 'Failed to activate wallet', 'error');
    },
  });
}

// Delete a wallet mutation
export function useDeleteWalletMutation(addToast) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (walletId) => {
      const response = await api.delete(`/wallets/${walletId}`);
      return response.data;
    },
    onSuccess: () => {
      addToast('Wallet deleted successfully', 'success');
      queryClient.invalidateQueries({ queryKey: ['wallets'] });
    },
    onError: (error) => {
      addToast(error.response?.data?.error || 'Failed to delete wallet', 'error');
    },
  });
}

// Transfer funds mutation
export function useTransferMutation(addToast) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({ fromWalletNumber, toWalletNumber, amount, idempotencyKey }) => {
      const amountInCents = Math.round(parseFloat(amount) * 100);
      const response = await api.post(
        '/transfer',
        {
          from_wallet_number: fromWalletNumber,
          to_wallet_number: toWalletNumber,
          amount: amountInCents,
        },
        {
          headers: {
            'X-Idempotency-Key': idempotencyKey,
          },
        }
      );
      return response.data;
    },
    onSuccess: () => {
      addToast('Transfer completed successfully', 'success');
      // Refetch all wallets and transaction histories
      queryClient.invalidateQueries({ queryKey: ['wallets'] });
      queryClient.invalidateQueries({ queryKey: ['transactions'] });
    },
    onError: (error) => {
      addToast(error.response?.data?.error || 'Transfer failed', 'error');
    },
  });
}
