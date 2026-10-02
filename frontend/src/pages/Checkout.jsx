import React, { useState, useEffect } from 'react';
import { useSearchParams, useNavigate } from 'react-router-dom';
import { 
  ShieldCheck, 
  Wallet, 
  CheckCircle2, 
  AlertCircle, 
  ArrowRight, 
  Lock, 
  Store, 
  Receipt, 
  Loader2,
  Clock,
  Sparkles
} from 'lucide-react';
import api from '../lib/axios';
import { useAuth } from '../context/AuthContext';

export default function Checkout() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const sessionId = searchParams.get('session_id');

  const { user, login } = useAuth();

  // Session state
  const [session, setSession] = useState(null);
  const [sessionLoading, setSessionLoading] = useState(true);
  const [sessionError, setSessionError] = useState('');

  // Wallets state
  const [wallets, setWallets] = useState([]);
  const [selectedWalletId, setSelectedWalletId] = useState('');
  const [walletsLoading, setWalletsLoading] = useState(false);

  // Login form state (if not authenticated)
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [loginLoading, setLoginLoading] = useState(false);
  const [loginError, setLoginError] = useState('');

  // Payment execution state
  const [payLoading, setPayLoading] = useState(false);
  const [payError, setPayError] = useState('');
  const [paySuccess, setPaySuccess] = useState(null);

  // 1. Fetch checkout session details
  useEffect(() => {
    if (!sessionId) {
      setSessionError('No session_id provided in URL.');
      setSessionLoading(false);
      return;
    }

    const fetchSession = async () => {
      try {
        setSessionLoading(true);
        const res = await api.get(`/checkout/sessions/${sessionId}`);
        setSession(res.data);
      } catch (err) {
        setSessionError(err.response?.data?.error || 'Failed to load checkout session');
      } finally {
        setSessionLoading(false);
      }
    };

    fetchSession();
  }, [sessionId]);

  // 2. Fetch user's wallets when user is authenticated
  useEffect(() => {
    if (!user) return;

    const fetchWallets = async () => {
      try {
        setWalletsLoading(true);
        const res = await api.get('/wallets');
        const userWallets = res.data.data?.wallets || res.data?.wallets || [];
        setWallets(userWallets);

        // Pre-select matching currency wallet
        if (session && userWallets.length > 0) {
          const matching = userWallets.find((w) => w.currency === session.currency);
          if (matching) {
            setSelectedWalletId(matching.id);
          } else {
            setSelectedWalletId(userWallets[0].id);
          }
        }
      } catch (err) {
        console.error('Failed to fetch user wallets:', err);
      } finally {
        setWalletsLoading(false);
      }
    };

    fetchWallets();
  }, [user, session]);

  // Handle Quick Demo Login
  const handleQuickDemoLogin = async () => {
    setLoginLoading(true);
    setLoginError('');
    const res = await login('demo@tyr.com', 'demo123456');
    if (!res.success) {
      setLoginError(res.error || 'Demo login failed');
    }
    setLoginLoading(false);
  };

  // Handle Normal Login
  const handleLoginSubmit = async (e) => {
    e.preventDefault();
    setLoginLoading(true);
    setLoginError('');
    const res = await login(email, password);
    if (!res.success) {
      setLoginError(res.error || 'Invalid credentials');
    }
    setLoginLoading(false);
  };

  // Handle Payment Confirmation
  const handlePayConfirm = async () => {
    if (!selectedWalletId) {
      setPayError('Please select a wallet to complete payment.');
      return;
    }

    setPayLoading(true);
    setPayError('');

    try {
      const res = await api.post(`/checkout/sessions/${sessionId}/pay`, {
        wallet_id: selectedWalletId,
      });

      setPaySuccess(res.data);

      // Redirect user back to Foodeli callback URL after brief confirmation
      setTimeout(() => {
        if (res.data?.redirect_url) {
          window.location.href = res.data.redirect_url;
        }
      }, 1500);
    } catch (err) {
      setPayError(err.response?.data?.error || 'Payment failed. Please try again.');
      setPayLoading(false);
    }
  };

  // Format minor units to currency display
  const formatAmount = (minorUnits, curr) => {
    if (minorUnits === undefined || minorUnits === null) return '0.00';
    return (minorUnits / 100).toLocaleString('tr-TR', {
      minimumFractionDigits: 2,
      maximumFractionDigits: 2,
    }) + ' ' + (curr || 'TRY');
  };

  // Selected wallet object
  const selectedWallet = wallets.find((w) => w.id === selectedWalletId);
  const isSufficientBalance = selectedWallet && session && selectedWallet.balance >= session.amount;

  // Render: Loading session
  if (sessionLoading) {
    return (
      <div className="min-h-screen bg-slate-950 flex flex-col items-center justify-center text-white">
        <Loader2 className="w-10 h-10 animate-spin text-cyan-400 mb-4" />
        <p className="text-slate-400 font-medium">Securing checkout session...</p>
      </div>
    );
  }

  // Render: Error loading session
  if (sessionError || !session) {
    return (
      <div className="min-h-screen bg-slate-950 flex items-center justify-center p-4">
        <div className="max-w-md w-full bg-slate-900 border border-red-500/30 rounded-2xl p-6 text-center shadow-2xl">
          <AlertCircle className="w-12 h-12 text-red-400 mx-auto mb-3" />
          <h2 className="text-xl font-bold text-white mb-2">Invalid Checkout Session</h2>
          <p className="text-slate-400 text-sm mb-6">{sessionError || 'This session is not available.'}</p>
          <button
            onClick={() => navigate('/dashboard')}
            className="w-full py-2.5 px-4 bg-slate-800 hover:bg-slate-700 text-white rounded-xl font-medium transition"
          >
            Go to TyrFintech
          </button>
        </div>
      </div>
    );
  }

  // Render: Payment already completed
  if (session.status === 'PAID' || paySuccess) {
    return (
      <div className="min-h-screen bg-slate-950 flex items-center justify-center p-4">
        <div className="max-w-md w-full bg-slate-900 border border-emerald-500/40 rounded-3xl p-8 text-center shadow-2xl animate-in fade-in zoom-in duration-300">
          <div className="w-16 h-16 bg-emerald-500/20 text-emerald-400 rounded-full flex items-center justify-center mx-auto mb-4">
            <CheckCircle2 className="w-10 h-10" />
          </div>
          <h2 className="text-2xl font-bold text-white mb-1">Payment Successful!</h2>
          <p className="text-slate-400 text-sm mb-6">
            Your payment of <span className="text-white font-semibold">{formatAmount(session.amount, session.currency)}</span> to <span className="text-white font-semibold">{session.merchant_name}</span> has been confirmed.
          </p>
          <div className="bg-slate-950/60 rounded-xl p-3 text-xs text-slate-400 flex items-center justify-center gap-2 mb-6">
            <Loader2 className="w-4 h-4 animate-spin text-cyan-400" />
            <span>Redirecting you back to {session.merchant_name}...</span>
          </div>
          {paySuccess?.redirect_url && (
            <a
              href={paySuccess.redirect_url}
              className="inline-flex items-center justify-center gap-2 text-sm text-cyan-400 hover:text-cyan-300 font-medium"
            >
              Click here if you are not redirected automatically
              <ArrowRight className="w-4 h-4" />
            </a>
          )}
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 flex flex-col justify-between p-4 sm:p-6 lg:p-8">
      {/* Top Header */}
      <div className="max-w-4xl mx-auto w-full flex items-center justify-between py-4 border-b border-slate-800/80">
        <div className="flex items-center gap-2.5">
          <div className="w-8 h-8 rounded-lg bg-gradient-to-tr from-cyan-500 to-blue-600 flex items-center justify-center shadow-lg shadow-cyan-500/20">
            <Wallet className="w-4 h-4 text-white" />
          </div>
          <span className="font-bold tracking-tight text-lg text-white">TyrFintech</span>
          <span className="text-xs px-2 py-0.5 rounded-full bg-cyan-500/10 text-cyan-400 border border-cyan-500/20 font-medium">Checkout</span>
        </div>
        <div className="flex items-center gap-1.5 text-xs text-slate-400">
          <ShieldCheck className="w-4 h-4 text-emerald-400" />
          <span>256-Bit Encrypted Escrow</span>
        </div>
      </div>

      {/* Main Checkout Container */}
      <div className="max-w-4xl mx-auto w-full my-auto py-8 grid grid-cols-1 md:grid-cols-12 gap-8 items-start">
        {/* Left Column: Order Summary */}
        <div className="md:col-span-5 bg-slate-900/60 border border-slate-800 rounded-3xl p-6 backdrop-blur-sm">
          <div className="flex items-center gap-3 pb-5 border-b border-slate-800">
            <div className="w-12 h-12 rounded-2xl bg-cyan-500/10 text-cyan-400 border border-cyan-500/20 flex items-center justify-center">
              <Store className="w-6 h-6" />
            </div>
            <div>
              <p className="text-xs uppercase tracking-wider text-slate-400 font-semibold">Paying Merchant</p>
              <h3 className="text-lg font-bold text-white">{session.merchant_name}</h3>
            </div>
          </div>

          <div className="py-5 space-y-3.5 text-sm">
            <div className="flex items-center justify-between text-slate-400">
              <span className="flex items-center gap-1.5">
                <Receipt className="w-4 h-4 text-slate-500" /> Order Reference
              </span>
              <span className="font-mono text-slate-200">{session.order_id}</span>
            </div>
            <div className="flex items-center justify-between text-slate-400">
              <span className="flex items-center gap-1.5">
                <Clock className="w-4 h-4 text-slate-500" /> Session Status
              </span>
              <span className="px-2 py-0.5 rounded-md bg-amber-500/10 text-amber-400 text-xs font-medium border border-amber-500/20">
                {session.status}
              </span>
            </div>
          </div>

          <div className="pt-5 border-t border-slate-800 flex items-baseline justify-between">
            <span className="text-slate-400 font-medium">Total Amount</span>
            <span className="text-2xl font-extrabold text-white tracking-tight">
              {formatAmount(session.amount, session.currency)}
            </span>
          </div>
        </div>

        {/* Right Column: Authentication & Payment Form */}
        <div className="md:col-span-7 bg-slate-900 border border-slate-800 rounded-3xl p-6 sm:p-8 shadow-xl">
          {!user ? (
            /* Case A: User Not Logged In */
            <div>
              <div className="flex items-center gap-2 mb-2">
                <Lock className="w-5 h-5 text-cyan-400" />
                <h2 className="text-xl font-bold text-white">Sign in to Pay</h2>
              </div>
              <p className="text-sm text-slate-400 mb-6">
                Log in to your TyrFintech account to authorize this payment directly from your wallet balance.
              </p>

              {/* Quick Demo Login Button */}
              <button
                type="button"
                onClick={handleQuickDemoLogin}
                disabled={loginLoading}
                className="w-full mb-5 py-3 px-4 rounded-xl bg-gradient-to-r from-cyan-600 to-blue-600 hover:from-cyan-500 hover:to-blue-500 text-white font-semibold text-sm flex items-center justify-center gap-2 shadow-lg shadow-cyan-500/20 transition disabled:opacity-50"
              >
                <Sparkles className="w-4 h-4" />
                <span>One-Click Demo Login (demo@tyr.com)</span>
              </button>

              <div className="relative flex py-2 items-center mb-5">
                <div className="flex-grow border-t border-slate-800"></div>
                <span className="flex-shrink mx-4 text-xs uppercase tracking-wider text-slate-500">or sign in manually</span>
                <div className="flex-grow border-t border-slate-800"></div>
              </div>

              {loginError && (
                <div className="mb-4 p-3 rounded-xl bg-red-500/10 border border-red-500/30 text-red-400 text-xs flex items-center gap-2">
                  <AlertCircle className="w-4 h-4 flex-shrink-0" />
                  <span>{loginError}</span>
                </div>
              )}

              <form onSubmit={handleLoginSubmit} className="space-y-4">
                <div>
                  <label className="block text-xs font-semibold text-slate-300 mb-1.5">Email Address</label>
                  <input
                    type="email"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    required
                    placeholder="you@domain.com"
                    className="w-full bg-slate-950 border border-slate-800 focus:border-cyan-500 rounded-xl px-3.5 py-2.5 text-sm text-white focus:outline-none transition"
                  />
                </div>
                <div>
                  <label className="block text-xs font-semibold text-slate-300 mb-1.5">Password</label>
                  <input
                    type="password"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    required
                    placeholder="••••••••"
                    className="w-full bg-slate-950 border border-slate-800 focus:border-cyan-500 rounded-xl px-3.5 py-2.5 text-sm text-white focus:outline-none transition"
                  />
                </div>
                <button
                  type="submit"
                  disabled={loginLoading}
                  className="w-full py-2.5 px-4 bg-slate-800 hover:bg-slate-700 text-white rounded-xl text-sm font-medium transition disabled:opacity-50 flex items-center justify-center gap-2"
                >
                  {loginLoading && <Loader2 className="w-4 h-4 animate-spin" />}
                  <span>Sign In</span>
                </button>
              </form>
            </div>
          ) : (
            /* Case B: User Logged In -> Wallet Selection & Payment Confirmation */
            <div>
              <div className="flex items-center justify-between mb-4">
                <div>
                  <h2 className="text-xl font-bold text-white">Select Wallet</h2>
                  <p className="text-xs text-slate-400">Signed in as <span className="text-slate-200">{user.email || user.name}</span></p>
                </div>
                <div className="w-8 h-8 rounded-full bg-slate-800 flex items-center justify-center text-xs font-bold text-cyan-400">
                  {user.name ? user.name[0].toUpperCase() : 'U'}
                </div>
              </div>

              {payError && (
                <div className="mb-4 p-3 rounded-xl bg-red-500/10 border border-red-500/30 text-red-400 text-xs flex items-center gap-2">
                  <AlertCircle className="w-4 h-4 flex-shrink-0" />
                  <span>{payError}</span>
                </div>
              )}

              {walletsLoading ? (
                <div className="py-8 flex justify-center text-cyan-400">
                  <Loader2 className="w-6 h-6 animate-spin" />
                </div>
              ) : wallets.length === 0 ? (
                <div className="p-4 rounded-xl bg-slate-950 border border-slate-800 text-center text-sm text-slate-400">
                  No wallets found. Please create a wallet in your TyrFintech dashboard.
                </div>
              ) : (
                <div className="space-y-3 mb-6">
                  {wallets.map((w) => {
                    const isSufficient = w.balance >= session.amount && w.currency === session.currency;
                    const isSelected = selectedWalletId === w.id;

                    return (
                      <div
                        key={w.id}
                        onClick={() => setSelectedWalletId(w.id)}
                        className={`p-4 rounded-2xl border cursor-pointer transition flex items-center justify-between ${
                          isSelected
                            ? 'bg-cyan-500/10 border-cyan-500 shadow-md shadow-cyan-500/10'
                            : 'bg-slate-950/60 border-slate-800 hover:border-slate-700'
                        }`}
                      >
                        <div className="flex items-center gap-3">
                          <div className={`w-10 h-10 rounded-xl flex items-center justify-center font-bold text-xs ${
                            isSelected ? 'bg-cyan-500 text-slate-950' : 'bg-slate-800 text-slate-300'
                          }`}>
                            {w.currency}
                          </div>
                          <div>
                            <div className="flex items-center gap-2">
                              <span className="font-semibold text-sm text-white">{w.currency} Wallet</span>
                              {w.wallet_number && (
                                <span className="text-xs text-slate-500 font-mono">#{w.wallet_number}</span>
                              )}
                            </div>
                            <span className="text-xs text-slate-400">
                              Balance: {formatAmount(w.balance, w.currency)}
                            </span>
                          </div>
                        </div>

                        <div>
                          {!isSufficient ? (
                            <span className="text-xs text-amber-400 font-medium">Insufficient</span>
                          ) : (
                            <div className={`w-5 h-5 rounded-full border flex items-center justify-center ${
                              isSelected ? 'border-cyan-500 bg-cyan-500' : 'border-slate-600'
                            }`}>
                              {isSelected && <div className="w-2 h-2 rounded-full bg-slate-950" />}
                            </div>
                          )}
                        </div>
                      </div>
                    );
                  })}
                </div>
              )}

              {/* Confirm & Pay Button */}
              <button
                type="button"
                onClick={handlePayConfirm}
                disabled={payLoading || !isSufficientBalance}
                className="w-full py-3.5 px-4 rounded-2xl bg-gradient-to-r from-emerald-500 to-teal-600 hover:from-emerald-400 hover:to-teal-500 text-white font-bold text-sm flex items-center justify-center gap-2 shadow-lg shadow-emerald-500/20 transition disabled:opacity-40 disabled:cursor-not-allowed"
              >
                {payLoading ? (
                  <>
                    <Loader2 className="w-4 h-4 animate-spin" />
                    <span>Processing Payment...</span>
                  </>
                ) : !isSufficientBalance ? (
                  <span>Insufficient Balance</span>
                ) : (
                  <>
                    <ShieldCheck className="w-4 h-4" />
                    <span>Confirm & Pay {formatAmount(session.amount, session.currency)}</span>
                  </>
                )}
              </button>
            </div>
          )}
        </div>
      </div>

      {/* Footer */}
      <div className="max-w-4xl mx-auto w-full py-4 text-center text-xs text-slate-500 border-t border-slate-800/80">
        Secured by TyrFintech Autonomous Payment Engine &bull; Zero Float Financial Settlement
      </div>
    </div>
  );
}
