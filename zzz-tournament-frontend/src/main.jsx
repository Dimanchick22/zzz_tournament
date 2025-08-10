import React from 'react'
import ReactDOM from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import { Toaster } from 'react-hot-toast'

import App from './App.jsx'
import './index.css'

// Настройка для React 18 StrictMode
const root = ReactDOM.createRoot(document.getElementById('root'))

// Глобальная обработка ошибок
window.addEventListener('error', (event) => {
  console.error('Global error:', event.error)
})

window.addEventListener('unhandledrejection', (event) => {
  console.error('Unhandled promise rejection:', event.reason)
})

root.render(
  <React.StrictMode>
    <BrowserRouter>
      <App />
      
      {/* Глобальные уведомления */}
      <Toaster
        position="top-right"
        reverseOrder={false}
        gutter={8}
        containerClassName=""
        containerStyle={{}}
        toastOptions={{
          // Глобальные настройки для всех тостов
          duration: 4000,
          style: {
            background: 'var(--color-surface)',
            color: 'var(--color-text-primary)',
            border: '1px solid var(--color-border)',
            borderRadius: '8px',
            fontSize: '14px',
            maxWidth: '420px',
          },
          
          // Настройки для разных типов
          success: {
            iconTheme: {
              primary: 'var(--color-success)',
              secondary: 'var(--color-surface)',
            },
          },
          
          error: {
            iconTheme: {
              primary: 'var(--color-error)',
              secondary: 'var(--color-surface)',
            },
            duration: 6000,
          },
          
          loading: {
            iconTheme: {
              primary: 'var(--color-primary)',
              secondary: 'var(--color-surface)',
            },
          },
        }}
      />
    </BrowserRouter>
  </React.StrictMode>
)