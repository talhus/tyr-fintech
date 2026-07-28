import React from 'react';
import { X, Receipt, ShoppingBag, Calendar, CheckCircle2 } from 'lucide-react';
import { useCardTransactions } from '../hooks/useQueries';

export default function CardSpendingsModal({ isOpen, onClose, card, getCurrencySymbol }) {
  const { data: transactions = [], isLoading } = useCardTransactions(isOpen ? card?.id : null);

  if (!isOpen || !card) return null;

  const symbol = getCurrencySymbol ? getCurrencySymbol(card.currency) : (card.currency || '$');

  return (
    <div 
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4 animate-fade-in" 
      onClick={(e) => e.target === e.currentTarget && onClose()}
    >
      <div className="glass-panel max-w-lg w-full rounded-3xl p-6 shadow-2xl animate-slide-up flex flex-col max-h-[85vh]">
        
        {/* Header */}
        <div className="flex justify-between items-center pb-4 border-b border-white/10 shrink-0">
          <div>
            <h3 className="text-xl font-semibold text-white flex items-center gap-2">
              <Receipt className="w-5 h-5 text-accent" /> Card Spendings
            </h3>
            <p className="text-xs text-white/50 font-mono mt-0.5">
              {card.currency ? `${card.currency} Card` : 'Virtual Card'} • {card.card_number}
            </p>
          </div>
          <button 
            className="text-white/50 hover:text-white transition-colors cursor-pointer bg-white/5 hover:bg-white/10 p-2 rounded-xl border border-white/10"
            onClick={onClose}
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Transactions List */}
        <div className="overflow-y-auto py-4 space-y-3 flex-1 pr-1 custom-scrollbar">
          {isLoading ? (
            <div className="text-white/60 text-sm text-center py-8">Loading spendings...</div>
          ) : transactions.length === 0 ? (
            <div className="text-center py-12 space-y-2">
              <ShoppingBag className="w-10 h-10 text-white/20 mx-auto" />
              <p className="text-white/60 text-sm font-medium">No transactions on this card yet</p>
              <p className="text-white/40 text-xs">Test payments or merchant charges will appear here.</p>
            </div>
          ) : (
            transactions.map((tx) => {
              const formattedDate = new Date(tx.created_at).toLocaleString('en-US', {
                month: 'short',
                day: 'numeric',
                hour: '2-digit',
                minute: '2-digit',
              });
              const amount = (tx.amount / 100).toFixed(2);

              return (
                <div 
                  key={tx.id} 
                  className="glass-panel rounded-2xl p-4 flex items-center justify-between border border-white/10 hover:border-white/20 transition-all bg-white/5"
                >
                  <div className="flex items-center gap-3">
                    <div className="w-10 h-10 rounded-xl bg-accent/10 border border-accent/20 flex items-center justify-center text-accent">
                      <ShoppingBag className="w-5 h-5" />
                    </div>
                    <div>
                      <h4 className="text-sm font-semibold text-white">
                        {tx.merchant_name || 'Online Merchant'}
                      </h4>
                      <div className="flex items-center gap-2 text-xs text-white/40 mt-0.5">
                        <span className="flex items-center gap-1">
                          <Calendar className="w-3 h-3" /> {formattedDate}
                        </span>
                      </div>
                    </div>
                  </div>

                  <div className="text-right">
                    <div className="text-sm font-mono font-bold text-red-400">
                      -{symbol}{amount}
                    </div>
                    <span className="inline-flex items-center gap-1 text-[10px] font-semibold text-emerald-400">
                      <CheckCircle2 className="w-2.5 h-2.5" /> {tx.status}
                    </span>
                  </div>
                </div>
              );
            })
          )}
        </div>

        {/* Footer */}
        <div className="pt-4 border-t border-white/10 flex justify-end shrink-0">
          <button 
            onClick={onClose} 
            className="glass-panel bg-white/5 hover:bg-white/10 text-white font-semibold rounded-xl px-5 py-2.5 transition-colors border-white/10 text-sm cursor-pointer"
          >
            Close
          </button>
        </div>

      </div>
    </div>
  );
}
