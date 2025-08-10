// useAuth Hook - логика аутентификации с реальным API
import { useCallback } from 'react'
import { useAuthStore } from '@store/authStore'
import { useUIStore } from '@store/uiStore'
import { authAPI } from '@api/auth'
import { usersAPI } from '@api/users'
import { setAuthToken, clearAuthToken } from '@api/client'

export const useAuth = () => {
  const {
    isAuthenticated,
    isLoading,
    user,
    token,
    error,
    setLoading,
    setError,
    clearError,
    login: loginAction,
    logout: logoutAction,
    updateUser,
    initAuth
  } = useAuthStore()

  const { addNotification } = useUIStore()

  // Проверка аутентификации при запуске приложения
  const checkAuth = useCallback(async () => {
    setLoading(true)
    
    try {
      // Проверяем есть ли токен в localStorage
      const authData = localStorage.getItem('auth-storage')
      if (!authData) {
        setLoading(false)
        return
      }

      const parsedAuth = JSON.parse(authData)
      const storedToken = parsedAuth.state?.token

      if (!storedToken) {
        setLoading(false)
        return
      }

      // Устанавливаем токен в API клиент
      setAuthToken(storedToken)

      // Проверяем валидность токена через запрос профиля
      const result = await authAPI.validate()

      if (result.success && result.valid) {
        // Токен валидный, инициализируем auth state
        loginAction(result.user, storedToken)
        
        addNotification({
          type: 'success',
          title: 'С возвращением!',
          message: `Привет, ${result.user.username}!`
        })
      } else {
        // Токен невалидный, очищаем все
        clearAuthToken()
        logoutAction()
      }
    } catch (error) {
      console.error('Auth check failed:', error)
      clearAuthToken()
      logoutAction()
      
      addNotification({
        type: 'warning',
        title: 'Сессия истекла',
        message: 'Пожалуйста, войдите снова'
      })
    } finally {
      setLoading(false)
    }
  }, [setLoading, loginAction, logoutAction, addNotification])

  // Вход в систему
  const login = useCallback(async (credentials) => {
    setLoading(true)
    clearError()
    
    try {
      const result = await authAPI.login(credentials)
      
      if (result.success) {
        // Устанавливаем токен в API клиент
        setAuthToken(result.token)
        
        // Сохраняем данные в store
        loginAction(result.user, result.token)
        
        addNotification({
          type: 'success',
          title: 'Добро пожаловать!',
          message: `Привет, ${result.user.username}!`
        })
        
        return { success: true }
      } else {
        setError(result.error)
        
        addNotification({
          type: 'error',
          title: 'Ошибка входа',
          message: result.error
        })
        
        return { success: false, error: result.error, details: result.details }
      }
    } catch (error) {
      const errorMessage = error.message || 'Неожиданная ошибка при входе'
      setError(errorMessage)
      
      addNotification({
        type: 'error',
        title: 'Ошибка входа',
        message: errorMessage
      })
      
      return { success: false, error: errorMessage }
    } finally {
      setLoading(false)
    }
  }, [setLoading, clearError, loginAction, setError, addNotification])

  // Регистрация
  const register = useCallback(async (userData) => {
    setLoading(true)
    clearError()
    
    try {
      const result = await authAPI.register(userData)
      
      if (result.success) {
        // Устанавливаем токен в API клиент
        setAuthToken(result.token)
        
        // Сохраняем данные в store
        loginAction(result.user, result.token)
        
        addNotification({
          type: 'success',
          title: 'Регистрация успешна!',
          message: 'Добро пожаловать в ZZZ Tournament!'
        })
        
        return { success: true }
      } else {
        setError(result.error)
        
        addNotification({
          type: 'error',
          title: 'Ошибка регистрации',
          message: result.error
        })
        
        return { success: false, error: result.error, details: result.details }
      }
    } catch (error) {
      const errorMessage = error.message || 'Неожиданная ошибка при регистрации'
      setError(errorMessage)
      
      addNotification({
        type: 'error',
        title: 'Ошибка регистрации',
        message: errorMessage
      })
      
      return { success: false, error: errorMessage }
    } finally {
      setLoading(false)
    }
  }, [setLoading, clearError, loginAction, setError, addNotification])

  // Выход из системы
  const logout = useCallback(async () => {
    setLoading(true)
    
    try {
      // Уведомляем сервер о выходе
      await authAPI.logout()
    } catch (error) {
      console.warn('Logout API call failed:', error)
    }
    
    // В любом случае очищаем локальные данные
    clearAuthToken()
    logoutAction()
    
    addNotification({
      type: 'info',
      title: 'Выход выполнен',
      message: 'До свидания!'
    })
    
    setLoading(false)
  }, [logoutAction, addNotification, setLoading])

  // Обновление профиля
  const updateProfile = useCallback(async (profileData) => {
    setLoading(true)
    
    try {
      const result = await usersAPI.updateProfile(profileData)
      
      if (result.success) {
        updateUser(result.user)
        
        addNotification({
          type: 'success',
          title: 'Профиль обновлен',
          message: result.message || 'Изменения сохранены'
        })
        
        return { success: true }
      } else {
        addNotification({
          type: 'error',
          title: 'Ошибка обновления',
          message: result.error
        })
        
        return { success: false, error: result.error, details: result.details }
      }
    } catch (error) {
      const errorMessage = error.message || 'Ошибка обновления профиля'
      
      addNotification({
        type: 'error',
        title: 'Ошибка',
        message: errorMessage
      })
      
      return { success: false, error: errorMessage }
    } finally {
      setLoading(false)
    }
  }, [setLoading, updateUser, addNotification])

  // Смена пароля
  const changePassword = useCallback(async (passwordData) => {
    setLoading(true)
    
    try {
      const result = await authAPI.changePassword(passwordData)
      
      if (result.success) {
        addNotification({
          type: 'success',
          title: 'Пароль изменен',
          message: result.message
        })
        
        return { success: true }
      } else {
        addNotification({
          type: 'error',
          title: 'Ошибка смены пароля',
          message: result.error
        })
        
        return { success: false, error: result.error }
      }
    } catch (error) {
      const errorMessage = error.message || 'Ошибка смены пароля'
      
      addNotification({
        type: 'error',
        title: 'Ошибка',
        message: errorMessage
      })
      
      return { success: false, error: errorMessage }
    } finally {
      setLoading(false)
    }
  }, [setLoading, addNotification])

  // Загрузка аватара
  const uploadAvatar = useCallback(async (file, onProgress) => {
    setLoading(true)
    
    try {
      const result = await usersAPI.uploadAvatar(file, onProgress)
      
      if (result.success) {
        // Обновляем аватар в профиле пользователя
        updateUser({ avatar: result.avatarUrl })
        
        addNotification({
          type: 'success',
          title: 'Аватар загружен',
          message: result.message
        })
        
        return { success: true, avatarUrl: result.avatarUrl }
      } else {
        addNotification({
          type: 'error',
          title: 'Ошибка загрузки',
          message: result.error
        })
        
        return { success: false, error: result.error }
      }
    } catch (error) {
      const errorMessage = error.message || 'Ошибка загрузки аватара'
      
      addNotification({
        type: 'error',
        title: 'Ошибка',
        message: errorMessage
      })
      
      return { success: false, error: errorMessage }
    } finally {
      setLoading(false)
    }
  }, [setLoading, updateUser, addNotification])

  return {
    // State
    isAuthenticated,
    isLoading,
    user,
    token,
    error,
    
    // Actions
    login,
    register,
    logout,
    checkAuth,
    updateProfile,
    changePassword,
    uploadAvatar,
    clearError
  }
}