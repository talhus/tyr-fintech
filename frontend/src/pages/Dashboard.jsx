import React, { useState, useCallback, useEffect } from 'react';
import { useAuth } from '../context/AuthContext';
import { TopBar, TransferForm, ToastContainer, WalletsSection, TransactionHistory, CardsSection } from '../components';
import { Navigate } from 'react-router-dom';
import { useWallets, useTransferMutation, useCards } from '../hooks/useQueries';
import { useNotificationStream } from '../hooks/useNotificationStream';
import { Send, CreditCard } from 'lucide-react';

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
  const [activeTab, setActiveTab] = useState('transfer'); // 'transfer' or 'cards'
  const [toasts, setToasts] = useState([]);

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

  const { notifications, unreadCount, markAllAsRead, clearNotifications } = useNotificationStream(addToast);

  const { data: wallets = [] } = useWallets();
  const { data: cards = [] } = useCards();
  const transferMutation = useTransferMutation((msg, type) => addToast(msg, type));

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

      <TopBar 
        user={user} 
        onLogout={logout} 
        notifications={notifications}
        unreadCount={unreadCount}
        onMarkAllAsRead={markAllAsRead}
        onClearNotifications={clearNotifications}
      />

      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 animate-fade-in relative z-10">
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          
          {/* Main Left Section: Wallets & History */}
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

          {/* Right Column: Tabbed Action Switcher (Transfer Funds | Virtual Cards) */}
          <div className="space-y-4 animate-slide-up" style={{ animationDelay: '0.1s' }}>
            {/* Tab Navigation Pill Bar */}
            <div className="p-1.5 rounded-2xl bg-white/5 border border-white/10 flex items-center gap-1 backdrop-blur-md">
              <button
                onClick={() => setActiveTab('transfer')}
                className={`flex-1 py-2.5 px-4 rounded-xl text-xs font-semibold flex items-center justify-center gap-2 transition-all duration-300 cursor-pointer ${
                  activeTab === 'transfer'
                    ? 'bg-gradient-to-r from-primary to-accent text-white shadow-lg shadow-primary/25 font-bold'
                    : 'text-white/60 hover:text-white hover:bg-white/5'
                }`}
              >
                <Send className="w-3.5 h-3.5" />
                <span>Transfer Funds</span>
              </button>

              <button
                onClick={() => setActiveTab('cards')}
                className={`flex-1 py-2.5 px-4 rounded-xl text-xs font-semibold flex items-center justify-center gap-2 transition-all duration-300 cursor-pointer relative ${
                  activeTab === 'cards'
                    ? 'bg-gradient-to-r from-primary to-accent text-white shadow-lg shadow-primary/25 font-bold'
                    : 'text-white/60 hover:text-white hover:bg-white/5'
                }`}
              >
                <CreditCard className="w-3.5 h-3.5" />
                <span>Virtual Cards</span>
                {cards.length > 0 && (
                  <span className={`text-[10px] px-1.5 py-0.2 rounded-full font-mono font-bold ${
                    activeTab === 'cards' ? 'bg-white/20 text-white' : 'bg-primary/30 text-primary'
                  }`}>
                    {cards.length}
                  </span>
                )}
              </button>
            </div>

            {/* Tab Panel Contents */}
            {activeTab === 'transfer' ? (
              <div className="animate-fade-in">
                <TransferForm
                  wallets={wallets}
                  onTransfer={handleTransfer}
                  getCurrencySymbol={getCurrencySymbol}
                />
              </div>
            ) : (
              <div className="animate-fade-in">
                <CardsSection
                  user={user}
                  wallets={wallets}
                  addToast={addToast}
                  getCurrencySymbol={getCurrencySymbol}
                />
              </div>
            )}
          </div>
          
        </div>
      </main>

      <ToastContainer toasts={toasts} removeToast={removeToast} />
    </div>
  );
}

export default Dashboard;
