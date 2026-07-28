import React from 'react';

export default function ToastContainer({ toasts, removeToast }) {
  return (
    <div className="fixed top-6 right-6 z-[100] flex flex-col gap-3 max-w-sm w-full pointer-events-none">
      {toasts.map((toast) => (
        <div 
          key={toast.id} 
          className={`pointer-events-auto flex items-center gap-3 px-4 py-3 rounded-xl border glass-panel animate-slide-up ${
            toast.type === 'success' ? 'border-secondary/30 bg-secondary/10 text-emerald-200' :
            toast.type === 'error' ? 'border-red-500/30 bg-red-500/10 text-red-200' : 
            'border-primary/30 bg-primary/10 text-blue-200'
          }`}
        >
          <i className={`fa-solid ${
            toast.type === 'success' ? 'fa-circle-check text-secondary' :
            toast.type === 'error' ? 'fa-circle-exclamation text-red-400' : 
            'fa-info-circle text-primary'
          }`}></i>
          <div className="text-sm font-medium">{toast.message}</div>
          <button 
            className="text-white/40 hover:text-white transition-colors cursor-pointer ml-auto" 
            onClick={() => removeToast(toast.id)}
          >
            <i className="fa-solid fa-xmark text-xs"></i>
          </button>
        </div>
      ))}
    </div>
  );
}
