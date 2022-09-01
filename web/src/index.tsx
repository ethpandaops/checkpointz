import React from 'react';

import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import TimeAgo from 'javascript-time-ago';
import en from 'javascript-time-ago/locale/en.json';
import ReactDOM from 'react-dom/client';

import './index.css';
import App from './App';
import ErrorBoundary from './ErrorBoundary';

const queryClient = new QueryClient();
TimeAgo.addDefaultLocale(en);

const root = ReactDOM.createRoot(document.getElementById('root') as HTMLElement);
root.render(
  <React.StrictMode>
    <ErrorBoundary>
      <QueryClientProvider client={queryClient}>
        <App />
      </QueryClientProvider>
    </ErrorBoundary>
  </React.StrictMode>,
);
