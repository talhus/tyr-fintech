import React from 'react';
import WalletCard from './WalletCard';
import { PlusCircle } from 'lucide-react';

export default function WalletsGrid({
  wallets,
  onAddWalletClick,
  onCopy,
  onDelete,
  selectedWalletId,
  onSelectWallet,
  getCurrencySymbol,
  getCurrencyIcon,
}) {
  const activeCurrencies = wallets.map((w) => w.currency.toUpperCase());
  const missingCurrencies = ['TRY', 'USD', 'EUR'].filter((c) => !activeCurrencies.includes(c));

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
      {wallets.map((wallet) => (
        <WalletCard
          key={wallet.id}
          wallet={wallet}
          onCopy={onCopy}
          onDelete={onDelete}
          selected={wallet.id === selectedWalletId}
          onSelect={onSelectWallet}
          getCurrencySymbol={getCurrencySymbol}
          getCurrencyIcon={getCurrencyIcon}
        />
      ))}

      {missingCurrencies.map((currency) => (
        <button
          key={currency}
          className="glass-panel rounded-2xl p-6 border-dashed border-white/20 hover:border-primary/50 hover:bg-white/5 transition-all duration-300 flex flex-col items-center justify-center gap-3 min-h-[140px] cursor-pointer group text-left w-full"
          onClick={() => onAddWalletClick(currency)}
        >
          <PlusCircle className="w-8 h-8 text-white/40 group-hover:text-primary transition-colors duration-300" />
          <span className="text-sm font-semibold text-white/60 group-hover:text-white transition-colors duration-300">
            Activate {currency} Wallet
          </span>
        </button>
      ))}
    </div>
  );
}
