import React, { useState } from 'react';
import { FileDown, Calendar, X } from 'lucide-react';
import api from '../lib/axios';

export default function ExportModal({ isOpen, onClose, walletId, walletName, addToast }) {
  const isAllAccounts = !walletId || walletId === 'all';
  const [format, setFormat] = useState(isAllAccounts ? 'csv' : 'pdf'); // 'pdf' or 'csv'
  const [startDate, setStartDate] = useState('');
  const [endDate, setEndDate] = useState('');
  const [isExporting, setIsExporting] = useState(false);

  if (!isOpen) return null;

  const handleSelectPDF = () => {
    if (isAllAccounts) {
      addToast('Official PDF bank statements can only be generated for a specific wallet account. Please select a wallet from the Filter bar.', 'warning');
      return;
    }
    setFormat('pdf');
  };

  const handleQuickRange = (rangeType) => {
    const today = new Date();
    const endStr = today.toISOString().split('T')[0];

    if (rangeType === 'all') {
      setStartDate('');
      setEndDate('');
      return;
    }

    let start = new Date();
    if (rangeType === '7days') {
      start.setDate(today.getDate() - 7);
    } else if (rangeType === '30days') {
      start.setDate(today.getDate() - 30);
    } else if (rangeType === 'thisMonth') {
      start = new Date(today.getFullYear(), today.getMonth(), 1);
    }
    setStartDate(start.toISOString().split('T')[0]);
    setEndDate(endStr);
  };

  const handleDownload = async (e) => {
    e.preventDefault();
    setIsExporting(true);
    const targetWallet = walletId || 'all';

    try {
      const response = await api.get(`/transactions/${targetWallet}/export`, {
        params: {
          format,
          start_date: startDate || undefined,
          end_date: endDate || undefined,
        },
        responseType: 'blob',
      });

      const blob = new Blob([response.data], {
        type: format === 'pdf' ? 'application/pdf' : 'text/csv',
      });
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.setAttribute('download', `statement_${targetWallet.substring(0, 8)}_${format}.${format}`);
      document.body.appendChild(link);
      link.click();
      link.remove();
      window.URL.revokeObjectURL(url);

      addToast(`${format.toUpperCase()} statement downloaded successfully`, 'success');
      onClose();
    } catch {
      addToast('Failed to export statement', 'error');
    } finally {
      setIsExporting(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm animate-fade-in">
      <div className="glass-panel w-full max-w-md rounded-3xl p-6 shadow-2xl relative">
        <button
          onClick={onClose}
          className="absolute top-4 right-4 text-white/40 hover:text-white transition-colors bg-transparent border-0 cursor-pointer"
        >
          <X className="w-5 h-5" />
        </button>

        <h3 className="text-xl font-semibold text-white mb-1 flex items-center gap-2">
          <FileDown className="w-5 h-5 text-primary" /> Export Statement
        </h3>
        <p className="text-xs text-white/60 mb-6">
          Account: <span className="text-primary font-semibold">{walletName || 'All Accounts'}</span>
        </p>

        <form onSubmit={handleDownload} className="space-y-5">
          {/* Format Picker */}
          <div>
            <label className="block text-xs font-semibold text-white/70 mb-2">Export Format</label>
            <div className="grid grid-cols-2 gap-3">
              <button
                type="button"
                onClick={handleSelectPDF}
                className={`py-2.5 px-4 rounded-xl text-xs font-semibold border transition-all cursor-pointer ${
                  format === 'pdf'
                    ? 'bg-primary/20 border-primary text-white shadow-lg shadow-primary/20'
                    : isAllAccounts
                    ? 'bg-white/5 border-white/5 text-white/30 hover:bg-white/5 cursor-not-allowed'
                    : 'bg-white/5 border-white/10 text-white/60 hover:text-white'
                }`}
              >
                PDF Document {isAllAccounts && '(Account Only)'}
              </button>
              <button
                type="button"
                onClick={() => setFormat('csv')}
                className={`py-2.5 px-4 rounded-xl text-xs font-semibold border transition-all cursor-pointer ${
                  format === 'csv'
                    ? 'bg-primary/20 border-primary text-white shadow-lg shadow-primary/20'
                    : 'bg-white/5 border-white/10 text-white/60 hover:text-white'
                }`}
              >
                CSV Spreadsheet
              </button>
            </div>
            {isAllAccounts && (
              <p className="text-[11px] text-amber-400/80 mt-2 bg-amber-500/10 border border-amber-500/20 p-2 rounded-xl">
                Official PDF Bank Statements can only be generated for a specific wallet account. Switch filter to a specific account to export PDF.
              </p>
            )}
          </div>

          {/* Quick Date Range Shortcuts */}
          <div>
            <label className="block text-xs font-semibold text-white/70 mb-2">Quick Date Shortcuts</label>
            <div className="flex flex-wrap gap-1.5">
              <button
                type="button"
                onClick={() => handleQuickRange('7days')}
                className="text-[11px] py-1 px-2.5 rounded-lg bg-white/5 hover:bg-white/10 border border-white/10 text-white/70 hover:text-white transition-colors cursor-pointer"
              >
                Last 7 Days
              </button>
              <button
                type="button"
                onClick={() => handleQuickRange('30days')}
                className="text-[11px] py-1 px-2.5 rounded-lg bg-white/5 hover:bg-white/10 border border-white/10 text-white/70 hover:text-white transition-colors cursor-pointer"
              >
                Last 30 Days
              </button>
              <button
                type="button"
                onClick={() => handleQuickRange('thisMonth')}
                className="text-[11px] py-1 px-2.5 rounded-lg bg-white/5 hover:bg-white/10 border border-white/10 text-white/70 hover:text-white transition-colors cursor-pointer"
              >
                This Month
              </button>
              <button
                type="button"
                onClick={() => handleQuickRange('all')}
                className="text-[11px] py-1 px-2.5 rounded-lg bg-white/5 hover:bg-white/10 border border-white/10 text-white/70 hover:text-white transition-colors cursor-pointer"
              >
                All Time
              </button>
            </div>
          </div>

          {/* Date Range Inputs */}
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-xs text-white/60 mb-1 flex items-center gap-1">
                <Calendar className="w-3.5 h-3.5 text-primary" /> Start Date
              </label>
              <input
                type="date"
                value={startDate}
                onChange={(e) => setStartDate(e.target.value)}
                className="glass-input w-full text-xs"
              />
            </div>
            <div>
              <label className="block text-xs text-white/60 mb-1 flex items-center gap-1">
                <Calendar className="w-3.5 h-3.5 text-primary" /> End Date
              </label>
              <input
                type="date"
                value={endDate}
                onChange={(e) => setEndDate(e.target.value)}
                className="glass-input w-full text-xs"
              />
            </div>
          </div>

          <div className="pt-2">
            <button
              type="submit"
              disabled={isExporting}
              className="glass-button w-full py-3 text-sm font-semibold flex items-center justify-center gap-2"
            >
              {isExporting ? 'Generating Statement...' : `Download ${format.toUpperCase()}`}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
