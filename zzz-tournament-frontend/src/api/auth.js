// Auth API methods
import { apiRequest } from './client'
import { API_ENDPOINTS } from '@config/api'

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
      const { token, user } = response.data.data

      if (!token) {
        return {
          success: false,
          error: 'Токен не получен от сервера'
        }
      }

      return {
        success: true,
        data: response.data,
        token: token,
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
      error: error.response?.data?.message || error.message || 'Ошибка регистрации',
      details: error.response?.data?.details || []
    }
  }
}

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

    // Обрабатываем стандартную структуру ответа бэкенда
    if (response.data?.success && response.data?.data) {
      const { token, user } = response.data.data

      if (!token) {
        return {
          success: false,
          error: 'Токен не получен от сервера'
        }
      }

      return {
        success: true,
        data: response.data,
        token: token,
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
 * Обновление токена
 * @returns {Promise<Object>} результат обновления токена
 */
export const refreshToken = async () => {
  try {
    const response = await apiRequest.post(API_ENDPOINTS.AUTH.REFRESH)

    return {
      success: true,
      token: response.data.token,
      user: response.data.user
    }
  } catch (error) {
    return {
      success: false,
      error: error.message || 'Ошибка обновления токена'
    }
  }
}

/**
 * Выход из системы
 * @returns {Promise<Object>} результат выхода
 */
export const logoutUser = async () => {
  try {
    // Отправляем запрос на сервер для инвалидации токена
    await apiRequest.post(API_ENDPOINTS.AUTH.LOGOUT)

    return {
      success: true
    }
  } catch (error) {
    // Даже если запрос не прошел, считаем выход успешным
    // так как мы все равно удалим токен локально
    return {
      success: true,
      warning: 'Не удалось уведомить сервер о выходе'
    }
  }
}

/**
 * Проверка валидности токена
 * @returns {Promise<Object>} результат проверки
 */
export const validateToken = async () => {
  try {
    // Проверяем токен через запрос профиля
    const response = await apiRequest.get(API_ENDPOINTS.USERS.PROFILE)

    return {
      success: true,
      valid: true,
      user: response.data.user || response.data
    }
  } catch (error) {
    return {
      success: false,
      valid: false,
      error: error.message
    }
  }
}

/**
 * Сброс пароля (отправка email)
 * @param {string} email - email для сброса пароля
 * @returns {Promise<Object>} результат отправки
 */
export const requestPasswordReset = async (email) => {
  try {
    const response = await apiRequest.post('/api/v1/auth/forgot-password', {
      email
    })

    return {
      success: true,
      message: response.data.message || 'Инструкции отправлены на email'
    }
  } catch (error) {
    return {
      success: false,
      error: error.message || 'Ошибка отправки инструкций'
    }
  }
}

/**
 * Подтверждение сброса пароля
 * @param {Object} resetData - данные для сброса
 * @param {string} resetData.token - токен сброса
 * @param {string} resetData.password - новый пароль
 * @returns {Promise<Object>} результат сброса
 */
export const resetPassword = async (resetData) => {
  try {
    const response = await apiRequest.post('/api/v1/auth/reset-password', {
      token: resetData.token,
      password: resetData.password
    })

    return {
      success: true,
      message: response.data.message || 'Пароль успешно изменен'
    }
  } catch (error) {
    return {
      success: false,
      error: error.message || 'Ошибка смены пароля'
    }
  }
}

/**
 * Изменение пароля (для авторизованного пользователя)
 * @param {Object} passwordData - данные для смены пароля
 * @param {string} passwordData.currentPassword - текущий пароль
 * @param {string} passwordData.newPassword - новый пароль
 * @returns {Promise<Object>} результат изменения
 */
export const changePassword = async (passwordData) => {
  try {
    const response = await apiRequest.post('/api/v1/auth/change-password', {
      current_password: passwordData.currentPassword,
      new_password: passwordData.newPassword
    })

    return {
      success: true,
      message: response.data.message || 'Пароль успешно изменен'
    }
  } catch (error) {
    return {
      success: false,
      error: error.message || 'Ошибка изменения пароля'
    }
  }
}

// Экспорт всех методов
export const authAPI = {
  register: registerUser,
  login: loginUser,
  refresh: refreshToken,
  logout: logoutUser,
  validate: validateToken,
  requestPasswordReset,
  resetPassword,
  changePassword
}

export default authAPI