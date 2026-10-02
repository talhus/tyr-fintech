import React, { useState } from 'react';
import { CreditCard, TrendingUp } from 'lucide-react';
import { useWallets, useCreateWalletMutation, useDeleteWalletMutation } from '../hooks/useQueries';
import WalletsGrid from './WalletsGrid';
import CreateWalletModal from './CreateWalletModal';

export default function WalletsSection({ user, selectedWalletId, onSelectWallet, addToast }) {
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [selectedCurrency, setSelectedCurrency] = useState('');
  const [displayCurrency, setDisplayCurrency] = useState('EUR');

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

  const getExchangeRate = (from, to) => {
    if (from === to) return 1.0;
    const rates = {
      'TRY_USD': 0.027,
      'TRY_EUR': 0.025,
      'USD_TRY': 36.5,
      'USD_EUR': 0.92,
      'EUR_TRY': 40.0,
      'EUR_USD': 1.09,
    };
    return rates[`${from}_${to}`] || 1.0;
  };

  const calculateTotalSavings = () => {
    return wallets.reduce((total, wallet) => {
      const balanceInMajor = wallet.balance / 100;
      const rate = getExchangeRate(wallet.currency.toUpperCase(), displayCurrency);
      return total + (balanceInMajor * rate);
    }, 0);
  };

  const activeCurrencies = wallets.map((w) => w?.currency?.toUpperCase() || '');
  const missingCurrencies = ['TRY', 'USD', 'EUR'].filter((c) => !activeCurrencies.includes(c));

  return (
    <div>
      {/* Total Savings / Net Worth Card */}
      {!isLoading && wallets.length > 0 && (
        <div className="glass-panel rounded-3xl p-6 mb-6 shadow-xl relative overflow-hidden bg-gradient-to-r from-primary/15 via-accent/10 to-transparent border border-white/10">
          <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
            <div>
              <div className="flex items-center gap-2 text-white/60 text-xs font-semibold uppercase tracking-wider mb-1">
                <TrendingUp className="w-4 h-4 text-emerald-400" />
                <span>Total Net Savings (All Balances)</span>
              </div>
              <div className="text-3xl font-extrabold text-white tracking-tight">
                {getCurrencySymbol(displayCurrency)}
                {calculateTotalSavings().toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
                <span className="text-sm font-medium text-white/50 ml-2">{displayCurrency}</span>
              </div>
              <p className="text-xs text-white/40 mt-1">
                Combined value of all your active wallets converted into {displayCurrency}
              </p>
            </div>

            {/* Currency Switcher Pill Group */}
            <div className="flex items-center gap-1 bg-black/20 p-1 rounded-2xl border border-white/10 backdrop-blur-md">
              {['EUR', 'USD', 'TRY'].map((curr) => (
                <button
                  key={curr}
                  onClick={() => setDisplayCurrency(curr)}
                  className={`py-1.5 px-3 rounded-xl text-xs font-bold transition-all cursor-pointer ${
                    displayCurrency === curr
                      ? 'bg-gradient-to-r from-primary to-accent text-white shadow-md shadow-primary/20'
                      : 'text-white/50 hover:text-white hover:bg-white/5'
                  }`}
                >
                  {curr}
                </button>
              ))}
            </div>
          </div>
        </div>
      )}

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
