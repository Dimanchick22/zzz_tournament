// UI Store - управление состоянием интерфейса
import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import { env } from '@config/env'

export const useUIStore = create(
  persist(
    (set, get) => ({
      // State
      theme: env.DEFAULT_THEME,
      sidebarCollapsed: false,
      isMobile: false,
      notifications: [],
      modals: {
        createRoom: false,
        userProfile: false,
        settings: false
      },

      // Actions
      setTheme: (theme) => set({ theme }),
      
      toggleTheme: () => set(state => ({
        theme: state.theme === 'dark' ? 'light' : 'dark'
      })),

      setSidebarCollapsed: (collapsed) => set({ sidebarCollapsed: collapsed }),
      
      toggleSidebar: () => set(state => ({
        sidebarCollapsed: !state.sidebarCollapsed
      })),

      setIsMobile: (isMobile) => set({ isMobile }),

      // Notifications
      addNotification: (notification) => set(state => ({
        notifications: [
          ...state.notifications,
          {
            id: Date.now(),
            timestamp: new Date(),
            ...notification
          }
        ]
      })),

      removeNotification: (id) => set(state => ({
        notifications: state.notifications.filter(n => n.id !== id)
      })),

      clearNotifications: () => set({ notifications: [] }),

      // Modals
      openModal: (modalName) => set(state => ({
        modals: { ...state.modals, [modalName]: true }
      })),

      closeModal: (modalName) => set(state => ({
        modals: { ...state.modals, [modalName]: false }
      })),

      closeAllModals: () => set(state => ({
        modals: Object.keys(state.modals).reduce((acc, key) => {
          acc[key] = false
          return acc
        }, {})
      })),

      // Initialize theme
      initializeTheme: () => {
        const state = get()
        // Проверяем системную тему если не установлена
        if (!state.theme) {
          const systemTheme = window.matchMedia('(prefers-color-scheme: dark)').matches 
            ? 'dark' 
            : 'light'
          set({ theme: systemTheme })
        }
      }
    }),
    {
      name: 'ui-storage',
      partialize: (state) => ({
        theme: state.theme,
        sidebarCollapsed: state.sidebarCollapsed
      })
    }
  )
)