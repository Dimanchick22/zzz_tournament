// Auth Store - управление состоянием аутентификации
import { create } from 'zustand'
import { persist } from 'zustand/middleware'

export const useAuthStore = create(
  persist(
    (set, get) => ({
      // State
      isAuthenticated: false,
      isLoading: false,
      user: null,
      token: null,
      error: null,

      // Actions
      setLoading: (loading) => set({ isLoading: loading }),
      
      setError: (error) => set({ error }),
      
      clearError: () => set({ error: null }),

      // Login action
      login: (userData, token) => set({
        isAuthenticated: true,
        user: userData,
        token: token,
        error: null,
        isLoading: false
      }),

      // Logout action
      logout: () => set({
        isAuthenticated: false,
        user: null,
        token: null,
        error: null,
        isLoading: false
      }),

      // Update user profile
      updateUser: (userData) => set(state => ({
        user: { ...state.user, ...userData }
      })),

      // Initialize auth (будет вызываться при запуске приложения)
      initAuth: () => {
        const state = get()
        if (state.token && state.user) {
          set({ isAuthenticated: true, isLoading: false })
        } else {
          set({ isAuthenticated: false, isLoading: false })
        }
      }
    }),
    {
      name: 'auth-storage', // localStorage key
      partialize: (state) => ({
        user: state.user,
        token: state.token,
        isAuthenticated: state.isAuthenticated
      })
    }
  )
)