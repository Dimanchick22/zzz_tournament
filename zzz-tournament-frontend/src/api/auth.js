// src/api/auth.js - исправленная версия

import { apiRequest } from './client'
import { API_ENDPOINTS } from '@config/api'

/**
 * Вход в систему
 * @param {Object} credentials - данные для входа
 * @param {string} credentials.username - имя пользователя
 * @param {string} credentials.password - пароль
 * @returns {Promise<Object>} результат входа
 */
export const loginUser = async (credentials) => {
  try {
    const response = await apiRequest.post(API_ENDPOINTS.AUTH.LOGIN, {
      username: credentials.username,
      password: credentials.password
    })

    // Обрабатываем реальную структуру ответа с бэкенда
    if (response.data?.success && response.data?.data) {
      const { access_token, refresh_token, user } = response.data.data

      if (!access_token) {
        return {
          success: false,
          error: 'Токен не получен от сервера'
        }
      }

      return {
        success: true,
        data: response.data,
        token: access_token,        // ✅ Используем access_token
        refreshToken: refresh_token, // ✅ Сохраняем refresh_token
        user: user
      }
    }

    // Если структура ответа неожиданная
    return {
      success: false,
      error: response.data?.message || 'Неожиданная структура ответа сервера'
    }

  } catch (error) {
    return {
      success: false,
      error: error.response?.data?.message || error.message || 'Ошибка входа в систему',
      details: error.response?.data?.details || []
    }
  }
}

/**
 * Регистрация нового пользователя
 * @param {Object} userData - данные пользователя
 * @param {string} userData.username - имя пользователя
 * @param {string} userData.email - email
 * @param {string} userData.password - пароль
 * @returns {Promise<Object>} результат регистрации
 */
export const registerUser = async (userData) => {
  try {
    const response = await apiRequest.post(API_ENDPOINTS.AUTH.REGISTER, {
      username: userData.username,
      email: userData.email,
      password: userData.password
    })

    // Обрабатываем стандартную структуру ответа бэкенда
    if (response.data?.success && response.data?.data) {
      const { access_token, refresh_token, user } = response.data.data

      if (!access_token) {
        return {
          success: false,
          error: 'Токен не получен от сервера'
        }
      }

      return {
        success: true,
        data: response.data,
        token: access_token,
        refreshToken: refresh_token,
        user: user
      }
    }

    return {
      success: false,
      error: response.data?.message || 'Неожиданная структура ответа сервера'
    }

  } catch (error) {
    return {
      success: false,
      error: error.response?.data?.message || error.message || 'Ошибка регистрации',
      details: error.response?.data?.details || []
    }
  }
}

/**
 * Обновление токена
 * @returns {Promise<Object>} результат обновления токена
 */
export const refreshToken = async () => {
  try {
    const response = await apiRequest.post(API_ENDPOINTS.AUTH.REFRESH)

    if (response.data?.success && response.data?.data) {
      const { access_token, refresh_token, user } = response.data.data

      return {
        success: true,
        token: access_token,
        refreshToken: refresh_token,
        user: user
      }
    }

    return {
      success: false,
      error: 'Неожиданная структура ответа сервера'
    }
  } catch (error) {
    return {
      success: false,
      error: error.message || 'Ошибка обновления токена'
    }
  }
}

// Остальные методы остаются без изменений...
export const logoutUser = async () => {
  try {
    await apiRequest.post(API_ENDPOINTS.AUTH.LOGOUT)
    return { success: true }
  } catch (error) {
    return {
      success: true,
      warning: 'Не удалось уведомить сервер о выходе'
    }
  }
}

export const validateToken = async () => {
  try {
    const response = await apiRequest.get(API_ENDPOINTS.USERS.PROFILE)

    let user = null
    if (response.data?.data) {
      user = response.data.data
    } else if (response.data?.user) {
      user = response.data.user
    } else if (response.data?.username) {
      user = response.data
    }

    return {
      success: true,
      valid: true,
      user: user
    }
  } catch (error) {
    console.error('Token validation failed:', error)
    return {
      success: false,
      valid: false,
      error: error.message
    }
  }
}

// Экспорт всех методов
export const authAPI = {
  register: registerUser,
  login: loginUser,
  refresh: refreshToken,
  logout: logoutUser,
  validate: validateToken
}

export default authAPI