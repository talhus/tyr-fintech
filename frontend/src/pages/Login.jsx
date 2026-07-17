import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Wallet, LogIn, Mail, Lock, User, UserPlus } from 'lucide-react';
import { useAuth } from '../context/AuthContext';

function Login() {
  const [isRegister, setIsRegister] = useState(false);
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const { login, register } = useAuth();
  const navigate = useNavigate();

  const handleSubmit = async (e) => {
    e.preventDefault();
    setIsLoading(true);
    setError('');
    setSuccess('');

    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailRegex.test(email)) {
      setError('Please enter a valid email address.');
      setIsLoading(false);
      return;
    }

    if (isRegister) {
      if (name.trim().length < 2) {
        setError('Full Name must be at least 2 characters long.');
        setIsLoading(false);
        return;
      }
      if (password.length < 6) {
        setError('Password must be at least 6 characters long.');
        setIsLoading(false);
        return;
      }

      const result = await register(name, email, password);
      if (result.success) {
        setSuccess('Registration successful! Please log in.');
        setIsRegister(false);
        setPassword('');
        setName('');
      } else {
        setError(result.error || 'Registration failed.');
      }
    } else {
      if (password.length === 0) {
        setError('Password is required.');
        setIsLoading(false);
        return;
      }

      const result = await login(email, password);
      if (result.success) {
        navigate('/dashboard');
      } else {
        setError(result.error || 'Login failed. Invalid email or password.');
      }
    }

    setIsLoading(false);
  };

  return (
    <div className="min-h-screen bg-background relative overflow-hidden font-sans flex items-center justify-center">
      <div className="absolute top-[-20%] left-[-10%] w-[50%] h-[50%] rounded-full bg-primary/20 blur-[120px] pointer-events-none" />
      <div className="absolute bottom-[-20%] right-[-10%] w-[50%] h-[50%] rounded-full bg-accent/20 blur-[120px] pointer-events-none" />

      <div className="w-full max-w-md px-4 relative z-10 animate-slide-up">
        <div className="flex flex-col items-center mb-8">
          <div className="w-16 h-16 rounded-2xl bg-gradient-to-tr from-primary to-accent flex items-center justify-center shadow-xl shadow-primary/20 mb-4">
            <Wallet className="w-8 h-8 text-white" />
          </div>
          <h1 className="text-3xl font-bold text-white mb-2">Welcome to TyrFintech</h1>
          <p className="text-white/60">Secure and fast money transfer</p>
        </div>

        <div className="glass-panel rounded-3xl p-8">
          <h2 className="text-xl font-semibold mb-6 text-white flex items-center gap-2">
            {isRegister ? (
              <>
                <UserPlus className="w-5 h-5 text-primary" /> Create Account
              </>
            ) : (
              <>
                <LogIn className="w-5 h-5 text-primary" /> Sign In
              </>
            )}
          </h2>

          {error && (
            <div className="bg-red-500/10 border border-red-500/50 text-red-200 px-4 py-3 rounded-xl mb-6 text-sm">
              {error}
            </div>
          )}

          {success && (
            <div className="bg-emerald-500/10 border border-emerald-500/50 text-emerald-200 px-4 py-3 rounded-xl mb-6 text-sm">
              {success}
            </div>
          )}

          <form onSubmit={handleSubmit} className="space-y-5">
            {isRegister && (
              <div>
                <label className="block text-sm text-white/60 mb-2">Full Name</label>
                <div className="relative">
                  <User className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-white/40" />
                  <input 
                    type="text" 
                    required
                    className="glass-input w-full !pl-11" 
                    placeholder="John Doe"
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                  />
                </div>
              </div>
            )}

            <div>
              <label className="block text-sm text-white/60 mb-2">Email Address</label>
              <div className="relative">
                <Mail className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-white/40" />
                <input 
                  type="email" 
                  required
                  className="glass-input w-full !pl-11" 
                  placeholder="example@mail.com"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                />
              </div>
            </div>

            <div>
              <label className="block text-sm text-white/60 mb-2">Password</label>
              <div className="relative">
                <Lock className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-white/40" />
                <input 
                  type="password" 
                  required
                  className="glass-input w-full !pl-11" 
                  placeholder="••••••••"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                />
              </div>
            </div>

            <button 
              type="submit" 
              disabled={isLoading}
              className="glass-button w-full flex items-center justify-center gap-2 mt-4"
            >
              {isLoading ? 'Processing...' : isRegister ? 'Register' : 'Log In'}
            </button>
          </form>

          <div className="mt-6 text-center">
            <p className="text-white/50 text-sm">
              {isRegister ? (
                <>
                  Already have an account?{' '}
                  <button 
                    onClick={() => { setIsRegister(false); setError(''); setSuccess(''); }} 
                    className="text-primary hover:text-primary/80 transition-colors font-medium bg-transparent border-0 cursor-pointer p-0"
                  >
                    Log In
                  </button>
                </>
              ) : (
                <>
                  Don't have an account?{' '}
                  <button 
                    onClick={() => { setIsRegister(true); setError(''); setSuccess(''); }} 
                    className="text-primary hover:text-primary/80 transition-colors font-medium bg-transparent border-0 cursor-pointer p-0"
                  >
                    Register
                  </button>
                </>
              )}
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}

export default Login;
