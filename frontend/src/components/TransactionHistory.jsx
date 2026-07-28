import React, { useState } from 'react';
import { ArrowUpRight, ArrowDownLeft, Calendar, Copy, Hash, FileDown, ShoppingBag } from 'lucide-react';
import api from '../lib/axios';
import { useTransactionHistory } from '../hooks/useQueries';

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

export default function TransactionHistory({ walletId, currency, addToast }) {
  const { data: transactions = [], isLoading } = useTransactionHistory(walletId);
  const [isExporting, setIsExporting] = useState(false);

  const handleCopyTxId = (id) => {
    navigator.clipboard.writeText(id);
    addToast('Transaction ID copied to clipboard', 'success');
  };

  const formatDate = (date) => {
    return new Intl.DateTimeFormat("en-US", {
      day: "2-digit",
      month: "2-digit",
      year: "numeric",
      hour: "2-digit",
      minute: "2-digit",
      hour12: false
    }).format(date);
  };

  const formatAmount = (tx) => {
    const isSent = tx.from_wallet_id?.toLowerCase() === walletId?.toLowerCase();
    const value = (isSent ? tx.amount : (tx.converted_amount || tx.amount)) / 100;
    const formattedValue = value.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 });
    const symbol = getCurrencySymbol(currency);
    return isSent ? `-${symbol}${formattedValue}` : `+${symbol}${formattedValue}`;
  };

  const handleExport = async (format) => {
    if (!walletId) return;
    setIsExporting(true);
    try {
      const response = await api.get(`/transactions/${walletId}/export`, {
        params: { format },
        responseType: 'blob',
      });
      const blob = new Blob([response.data], { 
        type: format === 'pdf' ? 'application/pdf' : 'text/csv' 
      });
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.setAttribute('download', `statement_${walletId.substring(0, 8)}.${format}`);
      document.body.appendChild(link);
      link.click();
      link.remove();
      window.URL.revokeObjectURL(url);
      addToast(`${format.toUpperCase()} statement downloaded successfully`, 'success');
    } catch {
      addToast('Failed to export statement', 'error');
    } finally {
      setIsExporting(false);
    }
  };

  return (
    <div className="glass-panel rounded-3xl p-6 shadow-xl animate-slide-up">
      <div className="flex justify-between items-center mb-6">
        <h3 className="text-lg font-semibold flex items-center gap-2">
          <Hash className="w-5 h-5 text-primary" /> Transaction History
        </h3>
        <div className="flex items-center gap-2">
          <button
            onClick={() => handleExport('csv')}
            disabled={isExporting || isLoading || transactions.length === 0}
            className="text-xs font-medium text-white/60 hover:text-white transition-colors bg-white/5 hover:bg-white/10 border border-white/10 px-3 py-1.5 rounded-lg cursor-pointer flex items-center gap-1 disabled:opacity-40 disabled:cursor-not-allowed"
          >
            <FileDown className="w-3.5 h-3.5" /> CSV
          </button>
          <button
            onClick={() => handleExport('pdf')}
            disabled={isExporting || isLoading || transactions.length === 0}
            className="text-xs font-medium text-white/60 hover:text-white transition-colors bg-white/5 hover:bg-white/10 border border-white/10 px-3 py-1.5 rounded-lg cursor-pointer flex items-center gap-1 disabled:opacity-40 disabled:cursor-not-allowed"
          >
            <FileDown className="w-3.5 h-3.5" /> PDF
          </button>
        </div>
      </div>

      {isLoading ? (
        <div className="text-white/60 text-sm">Loading transactions...</div>
      ) : transactions.length === 0 ? (
        <div className="text-white/40 text-sm py-4">No transactions found for this wallet.</div>
      ) : (
        <div className="space-y-3 max-h-[350px] overflow-y-auto pr-1 custom-scrollbar">
          {transactions.map((tx) => {
            const isSent = tx.from_wallet_id?.toLowerCase() === walletId?.toLowerCase();
            const isCardTx = !!(tx.card_id || tx.merchant_name);

            return (
              <div 
                key={tx.id} 
                className="flex items-center justify-between p-3.5 rounded-xl border border-white/5 bg-white/5 hover:bg-white/10 transition-colors"
              >
                <div className="flex items-center gap-3">
                  <div className={`w-9 h-9 rounded-lg flex items-center justify-center ${
                    isCardTx
                      ? 'bg-amber-500/10 text-amber-400'
                      : isSent
                      ? 'bg-red-500/10 text-red-400'
                      : 'bg-secondary/10 text-secondary'
                  }`}>
                    {isCardTx ? (
                      <ShoppingBag className="w-5 h-5" />
                    ) : isSent ? (
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
                          : isSent
                          ? 'Sent'
                          : 'Received'}
                      </span>
                      <span className="text-xs text-white/30 font-mono">({tx.id.substring(0, 8)})</span>
                      <Copy className="w-3 h-3 text-white/40" />
                    </button>
                    <div className="text-xs text-white/40 flex items-center gap-1 mt-0.5">
                      <Calendar className="w-3.5 h-3.5" />
                      <span>{formatDate(new Date(tx.created_at))}</span>
                    </div>
                  </div>
                </div>
                <div className="text-right">
                  <span className={`text-base font-bold ${
                    isSent ? 'text-red-400' : 'text-emerald-400'
                  }`}>
                    {formatAmount(tx)}
                  </span>
                  <div className="mt-0.5">
                    <span className={`text-[10px] font-semibold px-2 py-0.5 rounded-md ${
                      tx.status === 'COMPLETED' ? 'bg-secondary/20 text-secondary' :
                      tx.status === 'PENDING' ? 'bg-yellow-500/20 text-yellow-400' :
                      'bg-red-500/20 text-red-400'
                    }`}>
                      {tx.status}
                    </span>
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}
