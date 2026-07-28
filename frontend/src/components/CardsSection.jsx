import React, { useState } from 'react';
import { CreditCard, Plus, Eye, EyeOff, ChevronLeft, ChevronRight } from 'lucide-react';
import { useCards, useCreateCardMutation, useFreezeCardMutation, useUnfreezeCardMutation, useCloseCardMutation, useProcessCardPaymentMutation, useCardDetails } from '../hooks/useQueries';
import CreateCardModal from './CreateCardModal';
import CardPaymentModal from './CardPaymentModal';
import CardSpendingsModal from './CardSpendingsModal';

export default function CardsSection({ user, wallets, addToast, getCurrencySymbol }) {
  const { data: cards = [], isLoading } = useCards();

  const [currentIndex, setCurrentIndex] = useState(0);
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [selectedCardForPayment, setSelectedCardForPayment] = useState(null);
  const [selectedCardForHistory, setSelectedCardForHistory] = useState(null);
  const [revealedCardIds, setRevealedCardIds] = useState([]);

  const createCardMutation = useCreateCardMutation(addToast);
  const freezeCardMutation = useFreezeCardMutation(addToast);
  const unfreezeCardMutation = useUnfreezeCardMutation(addToast);
  const closeCardMutation = useCloseCardMutation(addToast);
  const processPaymentMutation = useProcessCardPaymentMutation(addToast);

  // Active card in carousel
  const safeIndex = Math.min(currentIndex, Math.max(0, cards.length - 1));
  const activeCard = cards[safeIndex];
  const isCurrentCardRevealed = activeCard ? revealedCardIds.includes(activeCard.id) : false;

  // Query unmasked details when revealed
  const { data: unmaskedDetails } = useCardDetails(activeCard?.id, isCurrentCardRevealed);

  const toggleRevealCard = (cardId) => {
    setRevealedCardIds((prev) =>
      prev.includes(cardId) ? prev.filter((id) => id !== cardId) : [...prev, cardId]
    );
  };

  const handleNext = () => {
    if (cards.length > 0) {
      setCurrentIndex((prev) => (prev + 1) % cards.length);
    }
  };

  const handlePrev = () => {
    if (cards.length > 0) {
      setCurrentIndex((prev) => (prev - 1 + cards.length) % cards.length);
    }
  };

  const handleCreateCard = async (walletId, limitAmount) => {
    await createCardMutation.mutateAsync({ walletId, limitAmount });
    setIsCreateModalOpen(false);
    setCurrentIndex(cards.length); // Slide to newly created card
  };

  const handleFreezeToggle = async (card) => {
    if (card.status === 'ACTIVE') {
      await freezeCardMutation.mutateAsync(card.id);
    } else if (card.status === 'FROZEN') {
      await unfreezeCardMutation.mutateAsync(card.id);
    }
  };

  const handleCloseCard = async (cardId) => {
    if (!window.confirm('Are you sure you want to terminate this virtual card?')) return;
    await closeCardMutation.mutateAsync(cardId);
    setCurrentIndex(0);
  };

  const handleProcessPayment = async (paymentData) => {
    await processPaymentMutation.mutateAsync(paymentData);
    setSelectedCardForPayment(null);
  };

  return (
    <div className="space-y-4">
      <div className="flex justify-between items-center">
        <h2 className="text-xl font-semibold flex items-center gap-2 text-white">
          <CreditCard className="w-5 h-5 text-accent" /> Virtual Cards
        </h2>
        <button
          onClick={() => setIsCreateModalOpen(true)}
          disabled={wallets.length === 0}
          className="glass-button text-xs py-2 px-4 flex items-center gap-1.5 cursor-pointer disabled:opacity-50"
        >
          <Plus className="w-4 h-4" /> Issue Virtual Card
        </button>
      </div>

      {isLoading ? (
        <div className="text-white/60 text-sm">Loading virtual cards...</div>
      ) : cards.length === 0 ? (
        <div className="glass-panel rounded-3xl p-6 text-center text-white/50 text-sm">
          No virtual cards issued yet. Click "Issue Virtual Card" to create your first card!
        </div>
      ) : (
        <div className="space-y-4">
          {/* Carousel View (1 Card at a Time) */}
          <div className="relative max-w-lg mx-auto">
            {cards.length > 1 && (
              <button
                onClick={handlePrev}
                className="absolute left-[-20px] sm:left-[-24px] top-1/2 -translate-y-1/2 z-20 w-10 h-10 rounded-full bg-slate-900/80 border border-white/20 text-white flex items-center justify-center shadow-xl hover:bg-slate-800 transition-colors cursor-pointer"
                title="Previous card"
              >
                <ChevronLeft className="w-5 h-5" />
              </button>
            )}

            {cards.length > 1 && (
              <button
                onClick={handleNext}
                className="absolute right-[-20px] sm:right-[-24px] top-1/2 -translate-y-1/2 z-20 w-10 h-10 rounded-full bg-slate-900/80 border border-white/20 text-white flex items-center justify-center shadow-xl hover:bg-slate-800 transition-colors cursor-pointer"
                title="Next card"
              >
                <ChevronRight className="w-5 h-5" />
              </button>
            )}

            {(() => {
              const card = activeCard;
              if (!card) return null;

              const isFrozen = card.status === 'FROZEN';
              const isClosed = card.status === 'CLOSED';
              const isRevealed = revealedCardIds.includes(card.id);
              const limit = (card.limit_amount / 100) || 0;
              const spent = (card.spent_amount / 100) || 0;
              const percentageSpent = limit > 0 ? Math.min(100, Math.round((spent / limit) * 100)) : 0;
              const symbol = getCurrencySymbol ? getCurrencySymbol(card.currency) : (card.currency || '$');

              const cardNumberDisplay = isRevealed && unmaskedDetails?.card_number
                ? unmaskedDetails.card_number
                : card.card_number;

              const cvvDisplay = isRevealed && unmaskedDetails?.cvv
                ? unmaskedDetails.cvv
                : '•••';

              return (
                <div 
                  key={card.id}
                  className={`relative overflow-hidden rounded-3xl p-6 sm:p-8 shadow-2xl transition-all duration-500 border animate-fade-in ${
                    isFrozen 
                      ? 'bg-gradient-to-br from-slate-900/90 via-slate-800/80 to-amber-950/40 border-amber-500/30' 
                      : isClosed 
                      ? 'bg-gradient-to-br from-slate-950 via-slate-900 to-red-950/30 border-red-500/20 opacity-60' 
                      : 'bg-gradient-to-br from-[#1e1e38] via-[#161d33] to-[#0f172a] border-white/15 hover:border-primary/40'
                  }`}
                >
                  {/* Background glow & Chip */}
                  <div className="absolute top-0 right-0 w-32 h-32 bg-accent/10 blur-3xl pointer-events-none" />

                  <div className="flex justify-between items-start mb-6 relative z-10">
                    <div className="flex items-center gap-2">
                      <span className="font-bold text-white tracking-wider text-sm">
                        {card.currency ? `${card.currency} VIRTUAL CARD` : 'TYR VIRTUAL'}
                      </span>
                      <span className={`text-[10px] font-bold px-2 py-0.5 rounded-full ${
                        isFrozen 
                          ? 'bg-amber-500/20 text-amber-300 border border-amber-500/30' 
                          : isClosed 
                          ? 'bg-red-500/20 text-red-300 border border-red-500/30' 
                          : 'bg-emerald-500/20 text-emerald-300 border border-emerald-500/30'
                      }`}>
                        {card.status}
                      </span>
                    </div>

                    <button
                      onClick={() => toggleRevealCard(card.id)}
                      className="text-white/70 hover:text-white transition-colors cursor-pointer bg-white/10 hover:bg-white/20 p-2 rounded-xl border border-white/10 flex items-center gap-1.5 text-xs"
                      title={isRevealed ? "Hide details" : "Show full unmasked details"}
                    >
                      {isRevealed ? <EyeOff className="w-4 h-4 text-accent" /> : <Eye className="w-4 h-4 text-accent" />}
                      <span>{isRevealed ? 'Hide' : 'Reveal'}</span>
                    </button>
                  </div>

                  {/* Card Number */}
                  <div className="mb-6 relative z-10">
                    <div className="text-white/40 text-[10px] uppercase tracking-widest mb-1 font-mono">Card Number</div>
                    <div className="font-mono text-xl sm:text-2xl tracking-widest font-bold text-white flex items-center justify-between">
                      <span>{cardNumberDisplay}</span>
                    </div>
                  </div>

                  {/* Card Expiry & CVV & Cardholder */}
                  <div className="grid grid-cols-3 gap-4 mb-6 relative z-10 text-xs">
                    <div>
                      <span className="text-white/40 block text-[10px] uppercase tracking-wider mb-0.5">Expires</span>
                      <span className="font-mono font-medium text-white">
                        {String(card.expiry_month).padStart(2, '0')} / {String(card.expiry_year).slice(-2)}
                      </span>
                    </div>
                    <div>
                      <span className="text-white/40 block text-[10px] uppercase tracking-wider mb-0.5">CVV</span>
                      <span className="font-mono font-medium text-amber-300">{cvvDisplay}</span>
                    </div>
                    <div>
                      <span className="text-white/40 block text-[10px] uppercase tracking-wider mb-0.5">Cardholder</span>
                      <span className="font-medium text-white truncate block">{user?.name || 'Valued Customer'}</span>
                    </div>
                  </div>

                  {/* Spending Limit Progress Bar */}
                  <div className="space-y-1.5 mb-6 relative z-10">
                    <div className="flex justify-between text-xs">
                      <span className="text-white/50 text-[11px]">Spent Limit ({card.currency || 'USD'})</span>
                      <span className="font-mono text-white/80 font-medium">
                        {symbol}{spent.toFixed(2)} / {symbol}{limit.toFixed(2)}
                      </span>
                    </div>
                    <div className="w-full bg-white/10 h-1.5 rounded-full overflow-hidden">
                      <div 
                        className={`h-full transition-all duration-500 rounded-full ${
                          percentageSpent > 85 ? 'bg-red-500' : 'bg-gradient-to-r from-primary to-accent'
                        }`} 
                        style={{ width: `${percentageSpent}%` }}
                      />
                    </div>
                  </div>

                  {/* Card Action Buttons */}
                  {!isClosed && (
                    <div className="flex items-center gap-2 pt-2 border-t border-white/10 relative z-10">
                      <button
                        onClick={() => handleFreezeToggle(card)}
                        className={`flex-1 text-xs py-2 px-3 rounded-xl border font-semibold flex items-center justify-center cursor-pointer transition-colors ${
                          isFrozen
                            ? 'bg-emerald-500/20 border-emerald-500/30 text-emerald-300 hover:bg-emerald-500/30'
                            : 'bg-amber-500/10 border-amber-500/20 text-amber-300 hover:bg-amber-500/20'
                        }`}
                      >
                        {isFrozen ? 'Unfreeze' : 'Freeze'}
                      </button>

                      <button
                        onClick={() => setSelectedCardForHistory(card)}
                        className="text-xs py-2 px-3 rounded-xl bg-white/5 hover:bg-white/10 border border-white/10 text-white font-semibold flex items-center justify-center cursor-pointer transition-colors"
                      >
                        Spendings
                      </button>

                      <button
                        onClick={() => setSelectedCardForPayment(card)}
                        disabled={isFrozen}
                        className="flex-1 text-xs py-2 px-3 rounded-xl bg-white/5 hover:bg-white/10 border border-white/10 text-white font-semibold flex items-center justify-center cursor-pointer transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
                      >
                        Test Pay
                      </button>

                      <button
                        onClick={() => handleCloseCard(card.id)}
                        className="text-xs py-2 px-3 rounded-xl bg-red-500/10 hover:bg-red-500/20 border border-red-500/20 text-red-400 font-semibold cursor-pointer transition-colors"
                      >
                        Terminate
                      </button>
                    </div>
                  )}
                </div>
              );
            })()}
          </div>

          {/* Carousel Dots & Controls */}
          {cards.length > 1 && (
            <div className="flex items-center justify-center gap-2 pt-1">
              {cards.map((card, idx) => (
                <button
                  key={card.id}
                  onClick={() => setCurrentIndex(idx)}
                  className={`h-2 rounded-full transition-all duration-300 cursor-pointer ${
                    idx === safeIndex ? 'w-6 bg-primary' : 'w-2 bg-white/20 hover:bg-white/40'
                  }`}
                  title={`Card ${idx + 1}`}
                />
              ))}
            </div>
          )}
        </div>
      )}

      {/* Modals */}
      <CreateCardModal
        isOpen={isCreateModalOpen}
        onClose={() => setIsCreateModalOpen(false)}
        wallets={wallets}
        existingCards={cards}
        onCreateCard={handleCreateCard}
        isSubmitting={createCardMutation.isPending}
        getCurrencySymbol={getCurrencySymbol}
      />

      <CardPaymentModal
        isOpen={!!selectedCardForPayment}
        onClose={() => setSelectedCardForPayment(null)}
        card={selectedCardForPayment}
        onProcessPayment={handleProcessPayment}
        isSubmitting={processPaymentMutation.isPending}
      />

      <CardSpendingsModal
        isOpen={!!selectedCardForHistory}
        onClose={() => setSelectedCardForHistory(null)}
        card={selectedCardForHistory}
        getCurrencySymbol={getCurrencySymbol}
      />
    </div>
  );
}
