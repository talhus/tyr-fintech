import React, { useState, useEffect, useRef } from 'react';
import { Send, Wallet, CreditCard, ChevronDown } from 'lucide-react';
import api from '../lib/axios';

export default function TransferForm({ wallets, onTransfer, getCurrencySymbol }) {
  const [fromWalletNumber, setFromWalletNumber] = useState('');
  const [toWalletNumber, setToWalletNumber] = useState('');
  const [transferAmount, setTransferAmount] = useState('');
  const [idempotencyKey, setIdempotencyKey] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  // States for target wallet verification
  const [recipientInfo, setRecipientInfo] = useState(null);
  const [isVerifying, setIsVerifying] = useState(false);
  const [verifyError, setVerifyError] = useState(null);

  // States for exchange rate (cross-currency transfers)
  const [exchangeRate, setExchangeRate] = useState(null);
  const [isFetchingRate, setIsFetchingRate] = useState(false);

  // States & Ref for custom source wallet combobox
  const [isSelectOpen, setIsSelectOpen] = useState(false);
  const selectRef = useRef(null);

  useEffect(() => {
    function handleClickOutside(event) {
      if (selectRef.current && !selectRef.current.contains(event.target)) {
        setIsSelectOpen(false);
      }
    }
    document.addEventListener('mousedown', handleClickOutside);
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, []);

  const generateIdempotencyKey = () => {
    const key = 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (c) => {
      const r = (Math.random() * 16) | 0;
      const v = c === 'x' ? r : (r & 0x3) | 0x8;
      return v.toString(16);
    });
    setIdempotencyKey(key);
  };

  useEffect(() => {
    generateIdempotencyKey();
  }, []);

  useEffect(() => {
    if (wallets.length > 0 && !fromWalletNumber) {
      setFromWalletNumber(wallets[0].wallet_number?.toString() || '');
    }
  }, [wallets, fromWalletNumber]);

  useEffect(() => {
    // Check if the wallet number digit is 10; if not, do not request to backend
    if (toWalletNumber.trim().length === 10) {
      setIsVerifying(true);
      setVerifyError(null);

      const delayDebounceFn = setTimeout(async () => {
        try {
          const res = await api.get(`/wallets/verify/${toWalletNumber}`);
          if (res.data && res.data.success) {
            setRecipientInfo(res.data.data);
          } else {
            setVerifyError('Wallet verification failed');
            setRecipientInfo(null);
          }
        } catch (err) {
          setVerifyError(err.response?.data?.error || 'Wallet not found');
          setRecipientInfo(null);
        } finally {
          setIsVerifying(false);
        }
      }, 500); // 500ms debounce

      return () => clearTimeout(delayDebounceFn);
    } else {
      setRecipientInfo(null);
      setIsVerifying(false);
      setVerifyError(null);
    }
  }, [toWalletNumber]);

  const selectedWallet = wallets.find((w) => w.wallet_number?.toString() === fromWalletNumber);

  useEffect(() => {
    const sourceCurrency = selectedWallet?.currency;
    const targetCurrency = recipientInfo?.Currency || recipientInfo?.currency;

    if (sourceCurrency && targetCurrency) {
      if (sourceCurrency === targetCurrency) {
        setExchangeRate(1);
        return;
      }

      setIsFetchingRate(true);
      api.get(`/exchange-rate?from=${sourceCurrency}&to=${targetCurrency}`)
        .then((res) => {
          if (res.data && res.data.success) {
            setExchangeRate(res.data.data.rate);
          } else {
            setExchangeRate(null);
          }
        })
        .catch(() => {
          setExchangeRate(null);
        })
        .finally(() => {
          setIsFetchingRate(false);
        });
    } else {
      setExchangeRate(null);
    }
  }, [selectedWallet?.currency, recipientInfo]);

  const walletBalance = selectedWallet ? selectedWallet.balance / 100 : 0;
  const enteredAmount = parseFloat(transferAmount) || 0;
  const isInsufficient = !!(selectedWallet && enteredAmount > walletBalance);

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (isSubmitting || isInsufficient) return;
    setIsSubmitting(true);

    const success = await onTransfer(
      parseInt(fromWalletNumber, 10),
      parseInt(toWalletNumber, 10),
      transferAmount,
      idempotencyKey
    );
    if (success) {
      setFromWalletNumber(wallets[0]?.wallet_number?.toString() || '');
      setToWalletNumber('');
      setTransferAmount('');
    }
    generateIdempotencyKey();
    setIsSubmitting(false);
  };

  return (
    <div className="glass-panel rounded-3xl p-6 sticky top-28">
      <h3 className="text-lg font-semibold mb-2 flex items-center gap-2">
        <Send className="w-5 h-5 text-accent" /> Transfer Funds
      </h3>
      <p className="text-xs text-white/50 mb-6">Send multi-currency balances instantly.</p>

      <form onSubmit={handleSubmit} className="space-y-4">
        <div ref={selectRef}>
          <label className="block text-sm text-white/60 mb-2">Source Wallet</label>
          <div className="relative">
            <button
              type="button"
              onClick={() => setIsSelectOpen(!isSelectOpen)}
              className="glass-input w-full flex items-center justify-between text-left cursor-pointer bg-[#1e293b]/50 border border-white/10 hover:border-primary/50 transition-all duration-300 rounded-xl px-4 py-3"
            >
              {selectedWallet ? (
                <div className="flex items-center justify-between w-full pr-2">
                  <div className="flex items-center gap-2">
                    <span className="font-semibold text-white">{selectedWallet.currency} Wallet</span>
                    <span className="text-white/40 text-xs font-mono">No: {selectedWallet.wallet_number}</span>
                  </div>
                  <span className="font-bold text-secondary text-sm">
                    {getCurrencySymbol(selectedWallet.currency)}
                    {(selectedWallet.balance / 100).toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
                  </span>
                </div>
              ) : (
                <span className="text-white/40">Select sending account...</span>
              )}
              <ChevronDown className={`w-4 h-4 text-white/60 transition-transform duration-300 shrink-0 ${isSelectOpen ? 'rotate-180' : ''}`} />
            </button>

            {isSelectOpen && (
              <div className="absolute z-50 w-full mt-2 rounded-xl bg-[#1e293b] border border-white/15 shadow-2xl shadow-black/80 overflow-hidden max-h-60 overflow-y-auto animate-fade-in">
                {wallets.map((wallet) => (
                  <button
                    key={wallet.id}
                    type="button"
                    onClick={() => {
                      setFromWalletNumber(wallet.wallet_number?.toString() || '');
                      setIsSelectOpen(false);
                    }}
                    className={`w-full flex items-center justify-between text-left px-4 py-3 hover:bg-white/10 transition-colors cursor-pointer border-b border-white/5 last:border-b-0 ${
                      wallet.wallet_number?.toString() === fromWalletNumber ? 'bg-white/15 font-medium' : ''
                    }`}
                  >
                    <div className="flex flex-col">
                      <span className="font-semibold text-white">{wallet.currency} Wallet</span>
                      <span className="text-[10px] text-white/40 font-mono">No: {wallet.wallet_number}</span>
                    </div>
                    <span className="font-bold text-secondary text-sm">
                      {getCurrencySymbol(wallet.currency)}
                      {(wallet.balance / 100).toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
                    </span>
                  </button>
                ))}
              </div>
            )}
          </div>
        </div>

        <div>
          <label htmlFor="transfer-to" className="block text-sm text-white/60 mb-2">Destination Wallet Number</label>
          <div className="relative">
            <Wallet className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-white/40" />
            <input
              type="number"
              id="transfer-to"
              required
              placeholder="e.g. 1000000000"
              value={toWalletNumber}
              onChange={(e) => setToWalletNumber(e.target.value)}
              className="glass-input w-full !pl-11"
            />
          </div>
          {isVerifying && (
            <p className="text-xs text-white/60 mt-2 flex items-center gap-2 animate-pulse">
              <span className="w-1.5 h-1.5 rounded-full bg-primary animate-ping" />
              Verifying wallet...
            </p>
          )}
          {verifyError && (
            <p className="text-xs text-red-400 mt-2 font-medium">
              {verifyError}
            </p>
          )}
          {recipientInfo && (
            <div className="mt-2.5 p-3 rounded-xl bg-white/5 border border-white/10 space-y-1 text-xs animate-fade-in">
              <div className="flex justify-between items-center">
                <span className="text-white/50">Owner:</span>
                <span className="font-semibold text-white">
                  {recipientInfo.OwnerName || recipientInfo.owner_name}
                </span>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-white/50">Currency:</span>
                <span className="font-semibold text-secondary flex items-center gap-1">
                  {recipientInfo.Currency || recipientInfo.currency} ({getCurrencySymbol(recipientInfo.Currency || recipientInfo.currency)})
                </span>
              </div>
            </div>
          )}
        </div>

        <div>
          <label htmlFor="transfer-amount" className="block text-sm text-white/60 mb-2">Amount</label>
          <div className="relative">
            <CreditCard className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-white/40" />
            <input
              type="number"
              id="transfer-amount"
              required
              min="0.01"
              step="any"
              placeholder="0.00"
              value={transferAmount}
              onChange={(e) => setTransferAmount(e.target.value)}
              className="glass-input w-full !pl-11 text-lg font-medium"
            />
          </div>
          {isInsufficient && (
            <p className="text-xs text-red-400 mt-1.5 font-medium">
              Insufficient funds in the selected wallet.
            </p>
          )}
          {isFetchingRate && (
            <p className="text-xs text-white/60 mt-1.5 animate-pulse">
              Fetching exchange rate...
            </p>
          )}
          {recipientInfo && exchangeRate && selectedWallet?.currency !== (recipientInfo?.Currency || recipientInfo?.currency) && !isFetchingRate && (
            <div className="mt-2.5 p-3 rounded-xl bg-secondary/10 border border-secondary/20 space-y-1 text-xs animate-fade-in text-secondary">
              <div className="flex justify-between items-center">
                <span>Exchange Rate:</span>
                <span className="font-semibold">
                  1 {selectedWallet.currency} = {exchangeRate} {recipientInfo?.Currency || recipientInfo?.currency}
                </span>
              </div>
              {enteredAmount > 0 && (
                <div className="flex justify-between items-center pt-1 border-t border-secondary/10">
                  <span className="text-white/50">Recipient Receives:</span>
                  <span className="font-bold text-white">
                    {getCurrencySymbol(recipientInfo?.Currency || recipientInfo?.currency)}
                    {(enteredAmount * exchangeRate).toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
                  </span>
                </div>
              )}
            </div>
          )}
        </div>

        <button 
          type="submit" 
          disabled={isSubmitting || isInsufficient || isVerifying || isFetchingRate || !recipientInfo}
          className="glass-button w-full flex items-center justify-center gap-2 mt-6 cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
        >
          <span>{isSubmitting ? 'Sending...' : 'Send Balance'}</span>
          <Send className="w-4 h-4" />
        </button>
      </form>
    </div>
  );
}
