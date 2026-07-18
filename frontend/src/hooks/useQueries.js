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

// Fetch user's virtual cards
export function useCards() {
  return useQuery({
    queryKey: ['cards'],
    queryFn: async () => {
      const response = await api.get('/cards');
      return response.data.data || response.data || [];
    },
  });
}

// Create virtual card mutation
export function useCreateCardMutation(addToast) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({ walletId, limitAmount }) => {
      const limitInCents = Math.round(parseFloat(limitAmount) * 100);
      const response = await api.post('/cards', {
        wallet_id: walletId,
        limit_amount: limitInCents,
      });
      return response.data;
    },
    onSuccess: () => {
      addToast('Virtual card created successfully', 'success');
      queryClient.invalidateQueries({ queryKey: ['cards'] });
    },
    onError: (error) => {
      addToast(error.response?.data?.error || 'Failed to create card', 'error');
    },
  });
}

// Freeze card mutation
export function useFreezeCardMutation(addToast) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (cardId) => {
      const response = await api.post(`/cards/${cardId}/freeze`);
      return response.data;
    },
    onSuccess: () => {
      addToast('Card frozen successfully', 'info');
      queryClient.invalidateQueries({ queryKey: ['cards'] });
    },
    onError: (error) => {
      addToast(error.response?.data?.error || 'Failed to freeze card', 'error');
    },
  });
}

// Unfreeze card mutation
export function useUnfreezeCardMutation(addToast) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (cardId) => {
      const response = await api.post(`/cards/${cardId}/unfreeze`);
      return response.data;
    },
    onSuccess: () => {
      addToast('Card unfrozen successfully', 'success');
      queryClient.invalidateQueries({ queryKey: ['cards'] });
    },
    onError: (error) => {
      addToast(error.response?.data?.error || 'Failed to unfreeze card', 'error');
    },
  });
}

// Close/Delete card mutation
export function useCloseCardMutation(addToast) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (cardId) => {
      const response = await api.delete(`/cards/${cardId}`);
      return response.data;
    },
    onSuccess: () => {
      addToast('Card closed successfully', 'success');
      queryClient.invalidateQueries({ queryKey: ['cards'] });
    },
    onError: (error) => {
      addToast(error.response?.data?.error || 'Failed to close card', 'error');
    },
  });
}

// Process mock card payment mutation
export function useProcessCardPaymentMutation(addToast) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({ cardId, cardNumber, cvv, expiryMonth, expiryYear, amount, merchantName }) => {
      const amountInCents = Math.round(parseFloat(amount) * 100);
      const response = await api.post(`/cards/${cardId}/process-payment`, {
        card_number: cardNumber,
        cvv: cvv,
        expiry_month: parseInt(expiryMonth, 10),
        expiry_year: parseInt(expiryYear, 10),
        amount: amountInCents,
        merchant_name: merchantName || 'Online Merchant',
      });
      return response.data;
    },
    onSuccess: () => {
      addToast('Card payment processed successfully', 'success');
      queryClient.invalidateQueries({ queryKey: ['cards'] });
      queryClient.invalidateQueries({ queryKey: ['wallets'] });
      queryClient.invalidateQueries({ queryKey: ['transactions'] });
      queryClient.invalidateQueries({ queryKey: ['cardTransactions'] });
    },
    onError: (error) => {
      addToast(error.response?.data?.error || 'Payment failed', 'error');
    },
  });
}

// Fetch transactions for a specific virtual card
export function useCardTransactions(cardId) {
  return useQuery({
    queryKey: ['cardTransactions', cardId],
    queryFn: async () => {
      const response = await api.get(`/cards/${cardId}/transactions`);
      return response.data.data || response.data || [];
    },
    enabled: !!cardId,
  });
}

// Fetch unmasked card details (full card number & CVV)
export function useCardDetails(cardId, enabled = false) {
  return useQuery({
    queryKey: ['cardDetails', cardId],
    queryFn: async () => {
      const response = await api.get(`/cards/${cardId}/details`);
      return response.data.data || response.data;
    },
    enabled: enabled && !!cardId,
  });
}
