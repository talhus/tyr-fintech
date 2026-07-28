import React from 'react';

export default function CreateWalletModal({
  isOpen,
  onClose,
  selectedCurrency,
  setSelectedCurrency,
  onCreateWallet,
  missingCurrencies,
}) {
  if (!isOpen) return null;

  const handleSubmit = (e) => {
    e.preventDefault();
    onCreateWallet(selectedCurrency);
  };

  return (
    <div 
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4" 
      onClick={(e) => e.target === e.currentTarget && onClose()}
    >
      <div className="glass-panel max-w-md w-full rounded-3xl p-6 shadow-2xl animate-slide-up">
        <div className="flex justify-between items-center mb-6">
          <h3 className="text-xl font-semibold text-white">Create New Wallet</h3>
          <button className="text-white/50 hover:text-white transition-colors cursor-pointer" onClick={onClose}>
            <i className="fa-solid fa-xmark text-lg"></i>
          </button>
        </div>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label htmlFor="wallet-currency" className="block text-sm text-white/60 mb-2">Wallet Currency</label>
            <select
              id="wallet-currency"
              required
              value={selectedCurrency}
              onChange={(e) => setSelectedCurrency(e.target.value)}
              className="glass-input w-full bg-[#1e293b] text-white cursor-pointer"
            >
              <option value="" disabled>Select Currency...</option>
              {missingCurrencies.includes('TRY') && <option value="TRY">TRY - Turkish Lira</option>}
              {missingCurrencies.includes('USD') && <option value="USD">USD - US Dollar</option>}
              {missingCurrencies.includes('EUR') && <option value="EUR">EUR - Euro</option>}
            </select>
          </div>
          <p className="text-xs text-white/40 leading-relaxed">
            You can create one wallet per currency. New wallets are initialized with 0 balance.
          </p>
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
              className="glass-button text-sm px-5 py-2.5 cursor-pointer"
            >
              Create Wallet
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
