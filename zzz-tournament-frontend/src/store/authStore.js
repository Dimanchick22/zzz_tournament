// Auth Store - управление состоянием аутентификации (исправлено)
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
      login: (userData, token) => {
        console.log('🔑 Logging in user:', userData?.username, 'with token:', !!token)
        
        set({
          isAuthenticated: true,
          user: userData,
          token: token,
          error: null,
          isLoading: false
        })
      },

      // Logout action
      logout: () => {
        console.log('🚪 Logging out user')
        
        set({
          isAuthenticated: false,
          user: null,
          token: null,
          error: null,
          isLoading: false
        })
        
        // Также очищаем localStorage явно
        localStorage.removeItem('auth-storage')
      },

      // Update user profile
      updateUser: (userData) => {
        const currentUser = get().user
        if (!currentUser) {
          console.warn('Trying to update user but no user is logged in')
          return
        }
        
        const updatedUser = { ...currentUser, ...userData }
        console.log('👤 Updating user data:', updatedUser.username)
        
        set(state => ({
          user: updatedUser
        }))
      },

      // Initialize auth (будет вызываться при запуске приложения)
      initAuth: () => {
        const state = get()
        console.log('🚀 Initializing auth. Token exists:', !!state.token, 'User exists:', !!state.user)
        
        if (state.token && state.user) {
          set({ 
            isAuthenticated: true, 
            isLoading: false 
          })
        } else {
          set({ 
            isAuthenticated: false, 
            isLoading: false,
            user: null,
            token: null 
          })
        }
      },

      // Reset store (для отладки)
      reset: () => {
        set({
          isAuthenticated: false,
          isLoading: false,
          user: null,
          token: null,
          error: null
        })
        localStorage.removeItem('auth-storage')
      }
    }),
    {
      name: 'auth-storage', // localStorage key
      partialize: (state) => {
        // Сохраняем только нужные поля в localStorage
        return {
          user: state.user,
          token: state.token,
          isAuthenticated: state.isAuthenticated
        }
      },
      onRehydrateStorage: () => (state, error) => {
        if (error) {
          console.error('Failed to rehydrate auth store:', error)
          return
        }
        
        if (state) {
          console.log('💾 Auth store rehydrated:', {
            hasUser: !!state.user,
            hasToken: !!state.token,
            username: state.user?.username
          })
          
          // Убеждаемся что состояние корректное после восстановления
          if (state.user && state.token) {
            state.isAuthenticated = true
          } else {
            state.isAuthenticated = false
            state.user = null
            state.token = null
          }
        }
      }
    }
  )
)