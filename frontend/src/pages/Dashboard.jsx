import React, { useState, useCallback, useEffect } from 'react';
import { useAuth } from '../context/AuthContext';
import { TopBar, TransferForm, ToastContainer, WalletsSection, TransactionHistory } from '../components';
import { Navigate } from 'react-router-dom';
import { useWallets, useTransferMutation } from '../hooks/useQueries';

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

function Dashboard() {
  const { user, logout } = useAuth();
  const [selectedWalletId, setSelectedWalletId] = useState(null);
  const [toasts, setToasts] = useState([]);

  const { data: wallets = [] } = useWallets();
  const transferMutation = useTransferMutation((msg, type) => addToast(msg, type));

  const removeToast = useCallback((id) => {
    setToasts((prev) => prev.filter((t) => t.id !== id));
  }, []);

  const addToast = useCallback((message, type = 'info') => {
    const id = Date.now();
    setToasts((prev) => [...prev, { id, message, type }]);
    setTimeout(() => {
      removeToast(id);
    }, 4000);
  }, [removeToast]);

  // Synchronize selectedWalletId with loaded wallets
  useEffect(() => {
    if (wallets.length > 0) {
      if (!selectedWalletId || !wallets.some((w) => w.id === selectedWalletId)) {
        setSelectedWalletId(wallets[0].id);
      }
    } else {
      setSelectedWalletId(null);
    }
  }, [wallets, selectedWalletId]);

  const handleTransfer = useCallback(async (fromWalletNumber, toWalletNumber, amount, idempotencyKey) => {
    try {
      await transferMutation.mutateAsync({
        fromWalletNumber,
        toWalletNumber,
        amount,
        idempotencyKey,
      });
      return true;
    } catch {
      return false;
    }
  }, [transferMutation]);

  if (!user) {
    return <Navigate to="/login" replace />;
  }

  const selectedWalletCurrency = wallets.find((w) => w.id === selectedWalletId)?.currency;

  return (
    <div className="min-h-screen bg-background relative overflow-hidden font-sans">
      <div className="absolute top-[-10%] left-[-10%] w-[40%] h-[40%] rounded-full bg-primary/20 blur-[120px] pointer-events-none" />
      <div className="absolute bottom-[-10%] right-[-10%] w-[40%] h-[40%] rounded-full bg-accent/20 blur-[120px] pointer-events-none" />

      <TopBar user={user} onLogout={logout} />

      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 animate-fade-in relative z-10">
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          
          <div className="lg:col-span-2 space-y-8 animate-slide-up">
            <WalletsSection
              user={user}
              selectedWalletId={selectedWalletId}
              onSelectWallet={setSelectedWalletId}
              addToast={addToast}
            />

            {selectedWalletId && (
              <TransactionHistory
                walletId={selectedWalletId}
                currency={selectedWalletCurrency}
                addToast={addToast}
              />
            )}
          </div>

          <div className="animate-slide-up" style={{ animationDelay: '0.1s' }}>
            <TransferForm
              wallets={wallets}
              onTransfer={handleTransfer}
              getCurrencySymbol={getCurrencySymbol}
            />
          </div>
          
        </div>
      </main>

      <ToastContainer toasts={toasts} removeToast={removeToast} />
    </div>
  );
}

export default Dashboard;
