import React from 'react';
import { Copy, Trash2 } from 'lucide-react';

export default function WalletCard({ wallet, onCopy, onDelete, selected, onSelect, getCurrencySymbol, getCurrencyIcon }) {
  const currency = wallet?.currency?.toUpperCase() || '';
  const symbol = getCurrencySymbol(currency);

  const borderClass = 
    selected 
      ? 'border-primary ring-2 ring-primary/40 bg-gradient-to-br from-primary/15'
      : currency === 'TRY' ? 'border-primary/30 from-primary/10 hover:border-primary/50' :
        currency === 'USD' ? 'border-secondary/30 from-secondary/10 hover:border-secondary/50' :
        'border-accent/30 from-accent/10 hover:border-accent/50';
  
  const badgeBg =
    currency === 'TRY' ? 'bg-primary/20 text-primary' :
    currency === 'USD' ? 'bg-secondary/20 text-secondary' :
    'bg-accent/20 text-accent';

  return (
    <div 
      className={`glass-panel rounded-2xl p-6 bg-gradient-to-br to-transparent ${borderClass} relative overflow-hidden transition-all duration-300 hover:scale-[1.01] cursor-pointer`}
      onClick={() => onSelect && onSelect(wallet.id)}
    >
      <div className="flex justify-between items-center mb-6">
        <div className={`flex items-center gap-1.5 text-xs font-semibold px-2.5 py-1 rounded-md ${badgeBg}`}>
          <i className={`fa-solid ${getCurrencyIcon(currency)}`}></i>
          <span>{currency} Wallet</span>
        </div>
        <div className="flex items-center gap-3">
          <button 
            onClick={(e) => { e.stopPropagation(); onCopy(wallet.wallet_number); }} 
            className="flex items-center gap-1.5 text-xs text-white/40 hover:text-white transition-colors bg-transparent border-0 cursor-pointer"
          >
            <span>No: {wallet.wallet_number}</span>
            <Copy className="w-3.5 h-3.5" />
          </button>
          <button 
            onClick={(e) => { e.stopPropagation(); onDelete(wallet.id); }} 
            className="text-white/40 hover:text-red-400 transition-colors bg-transparent border-0 cursor-pointer"
          >
            <Trash2 className="w-3.5 h-3.5" />
          </button>
        </div>
      </div>
      <div>
        <div className="text-xs text-white/50 mb-1 font-medium">Available Balance</div>
        <div className="text-2xl font-bold tracking-tight text-white">
          <span className="text-white/70 mr-1">{symbol}</span>
          {(wallet.balance / 100).toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
        </div>
      </div>
    </div>
  );
}
