import React, { useState, useEffect } from 'react';
import { ArrowUpRight, ArrowDownLeft, Calendar, Copy, Hash, FileDown, ShoppingBag, Filter, ChevronLeft, ChevronRight } from 'lucide-react';
import { useTransactionHistory } from '../hooks/useQueries';
import ExportModal from './ExportModal';

const ITEMS_PER_PAGE = 20;

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

export default function TransactionHistory({ walletId, wallets = [], onSelectWallet, addToast }) {
  const activeWalletId = walletId || 'all';
  const { data: rawTxns = [], isLoading } = useTransactionHistory(activeWalletId);
  const transactions = Array.isArray(rawTxns) ? rawTxns : [];
  const safeWallets = Array.isArray(wallets) ? wallets : [];

  const [currentPage, setCurrentPage] = useState(1);
  const [isExportModalOpen, setIsExportModalOpen] = useState(false);

  // Reset pagination on account filter change
  useEffect(() => {
    setCurrentPage(1);
  }, [activeWalletId]);

  const handleCopyTxId = (id) => {
    navigator.clipboard.writeText(id);
    addToast('Transaction ID copied to clipboard', 'success');
  };

  const formatDate = (dateStr) => {
    try {
      const date = new Date(dateStr);
      return new Intl.DateTimeFormat("en-US", {
        day: "2-digit",
        month: "2-digit",
        year: "numeric",
        hour: "2-digit",
        minute: "2-digit",
        hour12: false
      }).format(date);
    } catch {
      return dateStr;
    }
  };

  const formatAmount = (tx) => {
    const isOutflow = activeWalletId !== 'all'
      ? tx.from_wallet_id?.toLowerCase() === activeWalletId?.toLowerCase()
      : (tx.card_id || tx.from_wallet_id);

    const value = (isOutflow ? tx.amount : (tx.converted_amount || tx.amount)) / 100;
    const formattedValue = value.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 });
    
    // Get symbol
    const activeWallet = safeWallets.find((w) => w.id === activeWalletId);
    const symbol = getCurrencySymbol(activeWallet?.currency || 'TRY');
    return isOutflow ? `-${symbol}${formattedValue}` : `+${symbol}${formattedValue}`;
  };

  // Pagination calculations
  const totalPages = Math.ceil(transactions.length / ITEMS_PER_PAGE) || 1;
  const startIndex = (currentPage - 1) * ITEMS_PER_PAGE;
  const currentTransactions = transactions.slice(startIndex, startIndex + ITEMS_PER_PAGE);

  const activeWalletObj = safeWallets.find((w) => w.id === activeWalletId);
  const activeWalletName = activeWalletObj ? `${activeWalletObj.currency} Wallet (#${activeWalletObj.wallet_number})` : 'All Accounts';

  return (
    <div className="glass-panel rounded-3xl p-6 shadow-xl animate-slide-up">
      {/* Top Bar: Header, Account Filters, Export Button */}
      <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 mb-6">
        <div>
          <h3 className="text-lg font-semibold flex items-center gap-2">
            <Hash className="w-5 h-5 text-primary" /> Transaction History
          </h3>
          <p className="text-xs text-white/50 mt-0.5">
            Showing transactions for <span className="text-primary font-medium">{activeWalletName}</span>
          </p>
        </div>

        <button
          onClick={() => setIsExportModalOpen(true)}
          disabled={isLoading || transactions.length === 0}
          className="text-xs font-semibold text-white/80 hover:text-white transition-all bg-white/5 hover:bg-white/10 border border-white/15 px-3.5 py-2 rounded-xl cursor-pointer flex items-center gap-1.5 shadow-md disabled:opacity-40 disabled:cursor-not-allowed"
        >
          <FileDown className="w-4 h-4 text-primary" /> Export Statement
        </button>
      </div>

      {/* Account Switcher Filter Tabs */}
      <div className="flex items-center gap-1.5 overflow-x-auto pb-3 mb-4 border-b border-white/10 custom-scrollbar">
        <span className="text-xs font-medium text-white/40 flex items-center gap-1 pr-1">
          <Filter className="w-3.5 h-3.5" /> Filter:
        </span>
        <button
          onClick={() => onSelectWallet('all')}
          className={`px-3 py-1.5 rounded-xl text-xs font-semibold border transition-all cursor-pointer whitespace-nowrap ${
            activeWalletId === 'all'
              ? 'bg-primary/25 border-primary text-white shadow-md shadow-primary/20'
              : 'bg-white/5 border-white/10 text-white/60 hover:text-white hover:bg-white/10'
          }`}
        >
          All Accounts
        </button>
        {safeWallets.map((w) => (
          <button
            key={w.id}
            onClick={() => onSelectWallet(w.id)}
            className={`px-3 py-1.5 rounded-xl text-xs font-semibold border transition-all cursor-pointer whitespace-nowrap ${
              activeWalletId === w.id
                ? 'bg-primary/25 border-primary text-white shadow-md shadow-primary/20'
                : 'bg-white/5 border-white/10 text-white/60 hover:text-white hover:bg-white/10'
            }`}
          >
            {w.currency} Wallet (#{w.wallet_number || w.id.substring(0, 6)})
          </button>
        ))}
      </div>

      {/* Transactions List */}
      {isLoading ? (
        <div className="text-white/60 text-sm py-8 text-center">Loading transaction records...</div>
      ) : transactions.length === 0 ? (
        <div className="text-white/40 text-sm py-8 text-center">No transactions recorded for this account.</div>
      ) : (
        <div className="space-y-2.5">
          {currentTransactions.map((tx) => {
            const isOutflow = activeWalletId !== 'all'
              ? tx.from_wallet_id?.toLowerCase() === activeWalletId?.toLowerCase()
              : (tx.card_id || tx.from_wallet_id);

            const isCardTx = !!(tx.card_id || tx.merchant_name);

            return (
              <div
                key={tx.id}
                className="flex items-center justify-between p-3.5 rounded-xl border border-white/5 bg-white/5 hover:bg-white/10 transition-colors"
              >
                <div className="flex items-center gap-3">
                  <div
                    className={`w-9 h-9 rounded-lg flex items-center justify-center ${
                      isCardTx
                        ? 'bg-amber-500/10 text-amber-400'
                        : isOutflow
                        ? 'bg-red-500/10 text-red-400'
                        : 'bg-emerald-500/10 text-emerald-400'
                    }`}
                  >
                    {isCardTx ? (
                      <ShoppingBag className="w-5 h-5" />
                    ) : isOutflow ? (
                      <ArrowUpRight className="w-5 h-5" />
                    ) : (
                      <ArrowDownLeft className="w-5 h-5" />
                    )}
                  </div>
                  <div>
                    <button
                      onClick={() => handleCopyTxId(tx.id)}
                      className="text-sm font-semibold text-white flex items-center gap-1.5 hover:text-primary transition-colors bg-transparent border-0 cursor-pointer"
                    >
                      <span>
                        {isCardTx
                          ? `${tx.merchant_name || 'Card Merchant'} (Virtual Card)`
                          : isOutflow
                          ? 'Sent Transfer'
                          : 'Received Deposit'}
                      </span>
                      <span className="text-xs text-white/30 font-mono">({tx.id.substring(0, 8)})</span>
                      <Copy className="w-3 h-3 text-white/40" />
                    </button>
                    <div className="text-xs text-white/40 flex items-center gap-1.5 mt-0.5">
                      <Calendar className="w-3.5 h-3.5" />
                      <span>{formatDate(tx.created_at)}</span>
                    </div>
                  </div>
                </div>

                <div className="text-right">
                  <span className={`text-base font-bold ${isOutflow ? 'text-red-400' : 'text-emerald-400'}`}>
                    {formatAmount(tx)}
                  </span>
                  <div className="mt-0.5">
                    <span
                      className={`text-[10px] font-semibold px-2 py-0.5 rounded-md ${
                        tx.status === 'COMPLETED'
                          ? 'bg-emerald-500/20 text-emerald-300'
                          : tx.status === 'PENDING'
                          ? 'bg-amber-500/20 text-amber-300'
                          : 'bg-red-500/20 text-red-400'
                      }`}
                    >
                      {tx.status}
                    </span>
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      )}

      {/* Pagination Bar (Max 20 per page) */}
      {!isLoading && transactions.length > 0 && (
        <div className="flex flex-col sm:flex-row items-center justify-between gap-3 pt-5 mt-4 border-t border-white/10 text-xs">
          <div className="text-white/40">
            Showing <span className="text-white font-medium">{startIndex + 1}</span> -{' '}
            <span className="text-white font-medium">{Math.min(startIndex + ITEMS_PER_PAGE, transactions.length)}</span> of{' '}
            <span className="text-white font-medium">{transactions.length}</span> transactions
          </div>

          {totalPages > 1 && (
            <div className="flex items-center gap-2">
              <button
                disabled={currentPage === 1}
                onClick={() => setCurrentPage((p) => Math.max(p - 1, 1))}
                className="px-3 py-1.5 rounded-xl bg-white/5 hover:bg-white/10 border border-white/10 text-white/70 hover:text-white disabled:opacity-30 disabled:cursor-not-allowed transition-all cursor-pointer flex items-center gap-1"
              >
                <ChevronLeft className="w-3.5 h-3.5" /> Previous
              </button>
              <span className="text-white/60 font-medium px-2">
                Page {currentPage} of {totalPages}
              </span>
              <button
                disabled={currentPage >= totalPages}
                onClick={() => setCurrentPage((p) => Math.min(p + 1, totalPages))}
                className="px-3 py-1.5 rounded-xl bg-white/5 hover:bg-white/10 border border-white/10 text-white/70 hover:text-white disabled:opacity-30 disabled:cursor-not-allowed transition-all cursor-pointer flex items-center gap-1"
              >
                Next <ChevronRight className="w-3.5 h-3.5" />
              </button>
            </div>
          )}
        </div>
      )}

      {/* Date Range Statement Export Modal */}
      <ExportModal
        isOpen={isExportModalOpen}
        onClose={() => setIsExportModalOpen(false)}
        walletId={activeWalletId}
        walletName={activeWalletName}
        addToast={addToast}
      />
    </div>
  );
}
