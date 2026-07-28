import React, { useState } from 'react';
import { CreditCard } from 'lucide-react';
import { useWallets, useCreateWalletMutation, useDeleteWalletMutation } from '../hooks/useQueries';
import WalletsGrid from './WalletsGrid';
import CreateWalletModal from './CreateWalletModal';

export default function WalletsSection({ user, selectedWalletId, onSelectWallet, addToast }) {
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [selectedCurrency, setSelectedCurrency] = useState('');

  const { data: wallets = [], isLoading } = useWallets();
  const createWalletMutation = useCreateWalletMutation(addToast);
  const deleteWalletMutation = useDeleteWalletMutation(addToast);

  const handleCreateWallet = async (currency) => {
    await createWalletMutation.mutateAsync({
      userId: user.id,
      currency,
    });
    setIsModalOpen(false);
    setSelectedCurrency('');
  };

  const handleCopy = (text) => {
    navigator.clipboard.writeText(text);
    addToast('Wallet number copied to clipboard', 'success');
  };

  const handleDeleteWallet = async (walletId) => {
    if (!window.confirm('Are you sure you want to delete this wallet?')) return;
    await deleteWalletMutation.mutateAsync(walletId);
  };

  const getCurrencySymbol = (currency) => {
    switch (currency?.toUpperCase()) {
      case 'TRY':
        return '₺';
      case 'USD':
        return '$';
      case 'EUR':
        return '€';
      default:
        return '¤';
    }
  };

  const getCurrencyIcon = (currency) => {
    switch (currency?.toUpperCase()) {
      case 'TRY':
        return 'fa-lira-sign';
      case 'USD':
        return 'fa-dollar-sign';
      case 'EUR':
        return 'fa-euro-sign';
      default:
        return 'fa-wallet';
    }
  };

  const activeCurrencies = wallets.map((w) => w?.currency?.toUpperCase() || '');
  const missingCurrencies = ['TRY', 'USD', 'EUR'].filter((c) => !activeCurrencies.includes(c));

  return (
    <div>
      <h2 className="text-xl font-semibold mb-4 flex items-center gap-2">
        <CreditCard className="w-5 h-5 text-primary" /> My Wallets
      </h2>
      {isLoading ? (
        <div className="text-white/60">Loading wallets...</div>
      ) : (
        <WalletsGrid
          wallets={wallets}
          onAddWalletClick={(currency) => {
            setSelectedCurrency(currency);
            setIsModalOpen(true);
          }}
          onCopy={handleCopy}
          onDelete={handleDeleteWallet}
          selectedWalletId={selectedWalletId}
          onSelectWallet={onSelectWallet}
          getCurrencySymbol={getCurrencySymbol}
          getCurrencyIcon={getCurrencyIcon}
        />
      )}

      <CreateWalletModal
        isOpen={isModalOpen}
        onClose={() => {
          setIsModalOpen(false);
          setSelectedCurrency('');
        }}
        selectedCurrency={selectedCurrency}
        setSelectedCurrency={setSelectedCurrency}
        onCreateWallet={handleCreateWallet}
        missingCurrencies={missingCurrencies}
      />
    </div>
  );
}
