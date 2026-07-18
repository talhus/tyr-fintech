import React, { useState, useEffect } from 'react';
import { X, ShoppingBag, Lock } from 'lucide-react';

export default function CardPaymentModal({ isOpen, onClose, card, onProcessPayment, isSubmitting }) {
  const [merchantName, setMerchantName] = useState('Amazon Store');
  const [amount, setAmount] = useState('25.00');
  const [cardNumber, setCardNumber] = useState('');
  const [cvv, setCvv] = useState('123');
  const [expiryMonth, setExpiryMonth] = useState('12');
  const [expiryYear, setExpiryYear] = useState('2029');

  const resetForm = () => {
    setCvv('');
    setAmount('25.00');
    setMerchantName('Amazon Store');
  };

  useEffect(() => {
    if (card) {
      setCardNumber(card.card_number || '');
      setExpiryMonth(card.expiry_month ? card.expiry_month.toString() : '12');
      setExpiryYear(card.expiry_year ? card.expiry_year.toString() : '2029');
      setCvv('');
    }
  }, [card]);

  if (!isOpen || !card) return null;

  const handleSubmit = (e) => {
    e.preventDefault();
    onProcessPayment({
      cardId: card.id,
      cardNumber: cardNumber,
      cvv: cvv,
      expiryMonth: expiryMonth,
      expiryYear: expiryYear,
      amount: amount,
      merchantName: merchantName,
    });
    resetForm();
  };

  return (
    <div 
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4 animate-fade-in" 
      onClick={(e) => e.target === e.currentTarget && onClose()}
    >
      <div className="glass-panel max-w-md w-full rounded-3xl p-6 shadow-2xl animate-slide-up">
        <div className="flex justify-between items-center mb-6">
          <h3 className="text-xl font-semibold text-white flex items-center gap-2">
            <ShoppingBag className="w-5 h-5 text-accent" /> Mock Online Purchase
          </h3>
          <button className="text-white/50 hover:text-white transition-colors cursor-pointer" onClick={onClose}>
            <X className="w-5 h-5" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label htmlFor="merchant-name" className="block text-sm text-white/60 mb-2">Merchant Name</label>
            <input
              type="text"
              id="merchant-name"
              required
              placeholder="e.g. Netflix, Amazon"
              value={merchantName}
              onChange={(e) => setMerchantName(e.target.value)}
              className="glass-input w-full"
            />
          </div>

          <div>
            <label htmlFor="payment-amount" className="block text-sm text-white/60 mb-2">Purchase Amount</label>
            <input
              type="number"
              id="payment-amount"
              required
              min="0.01"
              step="any"
              placeholder="25.00"
              value={amount}
              onChange={(e) => setAmount(e.target.value)}
              className="glass-input w-full font-medium"
            />
          </div>

          <div className="p-3 rounded-xl bg-white/5 border border-white/10 space-y-2 text-xs">
            <div className="flex items-center justify-between text-white/60">
              <span>Card Number:</span>
              <span className="font-mono text-white">{card.card_number}</span>
            </div>
            <div className="grid grid-cols-3 gap-2 pt-1 border-t border-white/5">
              <div>
                <label className="block text-white/40 text-[10px] mb-1">CVV</label>
                <input
                  type="text"
                  maxLength={3}
                  value={cvv}
                  onChange={(e) => setCvv(e.target.value)}
                  className="glass-input w-full text-center py-1 text-xs font-mono"
                />
              </div>
              <div>
                <label className="block text-white/40 text-[10px] mb-1">Expiry Month</label>
                <input
                  type="number"
                  min="1"
                  max="12"
                  value={expiryMonth}
                  onChange={(e) => setExpiryMonth(e.target.value)}
                  className="glass-input w-full text-center py-1 text-xs font-mono"
                />
              </div>
              <div>
                <label className="block text-white/40 text-[10px] mb-1">Expiry Year</label>
                <input
                  type="number"
                  value={expiryYear}
                  onChange={(e) => setExpiryYear(e.target.value)}
                  className="glass-input w-full text-center py-1 text-xs font-mono"
                />
              </div>
            </div>
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
              disabled={isSubmitting}
              className="glass-button text-sm px-5 py-2.5 cursor-pointer flex items-center gap-1.5 disabled:opacity-50"
            >
              <Lock className="w-3.5 h-3.5" />
              <span>{isSubmitting ? 'Processing...' : 'Pay Now'}</span>
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
