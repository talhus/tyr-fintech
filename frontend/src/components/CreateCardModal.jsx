import React, { useState, useEffect, useRef } from 'react';
import { X, CreditCard, Shield, ChevronDown, Check } from 'lucide-react';

export default function CreateCardModal({ isOpen, onClose, wallets, existingCards = [], onCreateCard, isSubmitting, getCurrencySymbol }) {
  const [selectedWalletId, setSelectedWalletId] = useState('');
  const [limitAmount, setLimitAmount] = useState('500');
  const [isSelectOpen, setIsSelectOpen] = useState(false);
  const selectRef = useRef(null);

  // Filter out wallets that ALREADY have an active/frozen card
  const existingCardWalletIds = existingCards
    .filter((c) => c.status !== 'CLOSED')
    .map((c) => c.wallet_id);

  const existingCurrencies = existingCards
    .filter((c) => c.status !== 'CLOSED')
    .map((c) => c.currency?.toUpperCase());

  const availableWallets = wallets.filter(
    (w) => !existingCardWalletIds.includes(w.id) && !existingCurrencies.includes(w.currency?.toUpperCase())
  );

  useEffect(() => {
    // Default to first available wallet if available
    if (availableWallets.length > 0 && (!selectedWalletId || !availableWallets.some((w) => w.id === selectedWalletId))) {
      setSelectedWalletId(availableWallets[0].id);
    }
  }, [availableWallets, selectedWalletId]);

  useEffect(() => {
    const handleClickOutside = (event) => {
      if (selectRef.current && !selectRef.current.contains(event.target)) {
        setIsSelectOpen(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  if (!isOpen) return null;

  const selectedWallet = wallets.find((w) => w.id === selectedWalletId);

  const handleSubmit = (e) => {
    e.preventDefault();
    if (!selectedWalletId || !limitAmount) return;
    if (existingCardWalletIds.includes(selectedWalletId)) return;
    onCreateCard(selectedWalletId, limitAmount);
  };

  const symbol = (currency) => (getCurrencySymbol ? getCurrencySymbol(currency) : (currency || '$'));

  return (
    <div 
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4 animate-fade-in" 
      onClick={(e) => e.target === e.currentTarget && onClose()}
    >
      <div className="glass-panel max-w-md w-full rounded-3xl p-6 shadow-2xl animate-slide-up">
        <div className="flex justify-between items-center mb-6">
          <h3 className="text-xl font-semibold text-white flex items-center gap-2">
            <CreditCard className="w-5 h-5 text-primary" /> Issue Virtual Card
          </h3>
          <button className="text-white/50 hover:text-white transition-colors cursor-pointer" onClick={onClose}>
            <X className="w-5 h-5" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div ref={selectRef}>
            <label className="block text-sm text-white/60 mb-2">Select Linked Wallet</label>
            <div className="relative">
              <button
                type="button"
                onClick={() => setIsSelectOpen(!isSelectOpen)}
                disabled={availableWallets.length === 0}
                className="glass-input w-full flex items-center justify-between text-left cursor-pointer bg-[#1e293b]/50 border border-white/10 hover:border-primary/50 transition-all duration-300 rounded-xl px-4 py-3 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {selectedWallet ? (
                  <div className="flex items-center justify-between w-full pr-2">
                    <div className="flex items-center gap-2">
                      <span className="font-semibold text-white">{selectedWallet.currency} Wallet</span>
                      <span className="text-white/40 text-xs font-mono">No: {selectedWallet.wallet_number}</span>
                    </div>
                    <span className="font-bold text-secondary text-sm">
                      {symbol(selectedWallet.currency)}
                      {(selectedWallet.balance / 100).toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
                    </span>
                  </div>
                ) : (
                  <span className="text-white/40">
                    {availableWallets.length === 0 ? 'All wallets already have cards' : 'Select wallet to link...'}
                  </span>
                )}
                <ChevronDown className={`w-4 h-4 text-white/60 transition-transform duration-300 shrink-0 ${isSelectOpen ? 'rotate-180' : ''}`} />
              </button>

              {isSelectOpen && availableWallets.length > 0 && (
                <div className="absolute z-50 w-full mt-2 rounded-xl bg-[#1e293b] border border-white/15 shadow-2xl shadow-black/80 overflow-hidden max-h-60 overflow-y-auto animate-fade-in">
                  {availableWallets.map((wallet) => (
                    <button
                      key={wallet.id}
                      type="button"
                      onClick={() => {
                        setSelectedWalletId(wallet.id);
                        setIsSelectOpen(false);
                      }}
                      className={`w-full flex items-center justify-between text-left px-4 py-3 hover:bg-white/10 transition-colors cursor-pointer border-b border-white/5 last:border-b-0 ${
                        wallet.id === selectedWalletId ? 'bg-white/15 font-medium' : ''
                      }`}
                    >
                      <div className="flex flex-col">
                        <span className="font-semibold text-white">{wallet.currency} Wallet</span>
                        <span className="text-[10px] text-white/40 font-mono">No: {wallet.wallet_number}</span>
                      </div>
                      <span className="font-bold text-secondary text-sm flex items-center gap-2">
                        <span>
                          {symbol(wallet.currency)}
                          {(wallet.balance / 100).toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
                        </span>
                        {wallet.id === selectedWalletId && <Check className="w-4 h-4 text-primary" />}
                      </span>
                    </button>
                  ))}
                </div>
              )}
            </div>
          </div>

          <div>
            <label htmlFor="card-limit" className="block text-sm text-white/60 mb-2">Monthly Spending Limit</label>
            <div className="relative">
              <input
                type="number"
                id="card-limit"
                required
                min="1"
                step="any"
                placeholder="500.00"
                value={limitAmount}
                onChange={(e) => setLimitAmount(e.target.value)}
                className="glass-input w-full font-medium"
              />
            </div>
          </div>

          <div className="p-3 rounded-xl bg-primary/10 border border-primary/20 text-xs text-primary/90 flex items-start gap-2">
            <Shield className="w-4 h-4 shrink-0 mt-0.5" />
            <span>Virtual cards are limited to 1 card per wallet currency and linked directly to your balance.</span>
          </div>

          <div className="flex justify-end gap-3 pt-2">
            <button 
              type="button" 
              onClick={onClose} 
              className="glass-panel bg-white/5 hover:bg-white/10 text-white font-semibold rounded-xl px-5 py-2.5 transition-colors border-white/10 text-sm cursor-pointer"
            >
              Cancel
            </button>
            <button 
              type="submit" 
              disabled={isSubmitting || !selectedWalletId || availableWallets.length === 0}
              className="glass-button text-sm px-5 py-2.5 cursor-pointer disabled:opacity-50"
            >
              {isSubmitting ? 'Issuing...' : 'Issue Card'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
