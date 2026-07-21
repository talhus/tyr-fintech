import React, { useState, useRef, useEffect } from 'react';
import { Wallet, LogOut, Bell, Trash2, X } from 'lucide-react';

export default function TopBar({ 
  user, 
  onLogout, 
  notifications = [], 
  unreadCount = 0, 
  onMarkAllAsRead, 
  onClearNotifications 
}) {
  const [isOpen, setIsOpen] = useState(false);
  const dropdownRef = useRef(null);

  useEffect(() => {
    function handleClickOutside(event) {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target)) {
        setIsOpen(false);
      }
    }

    if (isOpen) {
      document.addEventListener('mousedown', handleClickOutside);
    }

    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [isOpen]);

  return (
    <header className="glass-panel border-b-0 border-x-0 rounded-none px-8 py-4 flex justify-between items-center sticky top-0 z-50">
      <div className="flex items-center gap-3">
        <div className="w-10 h-10 rounded-xl bg-gradient-to-tr from-primary to-accent flex items-center justify-center shadow-lg shadow-primary/20">
          <Wallet className="w-6 h-6 text-white" />
        </div>
        <span className="text-xl font-bold bg-clip-text text-transparent bg-gradient-to-r from-white to-white/70 tracking-wider">
          TYR FINTECH
        </span>
      </div>
      <div className="flex items-center gap-4">
        {/* Notification Bell Dropdown */}
        <div className="relative" ref={dropdownRef}>
          <button
            onClick={() => {
              setIsOpen(!isOpen);
              if (!isOpen && unreadCount > 0 && onMarkAllAsRead) {
                onMarkAllAsRead();
              }
            }}
            className="relative p-2.5 rounded-xl glass-panel hover:bg-white/10 text-white/80 hover:text-white transition-all cursor-pointer flex items-center justify-center"
            title="Notifications"
          >
            <Bell className="w-5 h-5" />
            {unreadCount > 0 && (
              <span className="absolute -top-1 -right-1 w-5 h-5 rounded-full bg-primary text-white text-[10px] font-bold flex items-center justify-center animate-pulse border-2 border-slate-950">
                {unreadCount > 9 ? '9+' : unreadCount}
              </span>
            )}
          </button>

          {/* Notifications Dropdown Panel */}
          {isOpen && (
            <div className="absolute right-0 mt-3 w-80 sm:w-96 rounded-2xl border border-white/20 bg-[#1e293b] shadow-2xl shadow-black/80 z-50 p-4 animate-fade-in">
              <div className="flex items-center justify-between border-b border-white/10 pb-3 mb-3">
                <div className="flex items-center gap-2">
                  <Bell className="w-4 h-4 text-primary" />
                  <h4 className="text-sm font-semibold text-white">Live Notifications</h4>
                </div>
                <div className="flex items-center gap-2">
                  {notifications.length > 0 && (
                    <button
                      onClick={onClearNotifications}
                      className="text-xs text-white/40 hover:text-red-400 transition-colors flex items-center gap-1 cursor-pointer"
                      title="Clear All"
                    >
                      <Trash2 className="w-3.5 h-3.5" />
                      <span>Clear</span>
                    </button>
                  )}
                  <button
                    onClick={() => setIsOpen(false)}
                    className="text-white/40 hover:text-white transition-colors cursor-pointer"
                  >
                    <X className="w-4 h-4" />
                  </button>
                </div>
              </div>

              <div className="max-h-72 overflow-y-auto space-y-2.5 pr-1 custom-scrollbar">
                {notifications.length === 0 ? (
                  <div className="text-center py-6 text-xs text-white/40">
                    No notifications yet.
                  </div>
                ) : (
                  notifications.map((notif) => (
                    <div
                      key={notif.id}
                      className={`p-3.5 rounded-xl border text-xs transition-all bg-[#0f172a] shadow-md ${
                        notif.type === 'CARD_PAYMENT'
                          ? 'border-l-4 border-l-purple-500 border-white/10 text-slate-100'
                          : notif.type === 'TRANSACTION'
                          ? 'border-l-4 border-l-emerald-500 border-white/10 text-slate-100'
                          : 'border-l-4 border-l-blue-500 border-white/10 text-slate-100'
                      }`}
                    >
                      <div className="flex justify-between items-start mb-1.5 font-bold text-white">
                        <span>{notif.title}</span>
                        <span className="text-[10px] text-white/50 font-mono">{notif.timestamp}</span>
                      </div>
                      <p className="text-slate-300 leading-relaxed font-medium">{notif.message}</p>
                    </div>
                  ))
                )}
              </div>
            </div>
          )}
        </div>

        <div className="hidden md:flex flex-col text-right">
          <span className="text-sm font-semibold text-white">{user?.name || user?.username}</span>
          <span className="text-xs text-white/50">{user?.email}</span>
        </div>
        <button 
          onClick={onLogout} 
          className="glass-panel bg-red-500/10 hover:bg-red-500/20 text-red-200 border-red-500/30 font-semibold rounded-xl px-4 py-2 transition-colors flex items-center gap-2 text-sm cursor-pointer"
        >
          <LogOut className="w-4 h-4" />
          <span>Logout</span>
        </button>
      </div>
    </header>
  );
}

