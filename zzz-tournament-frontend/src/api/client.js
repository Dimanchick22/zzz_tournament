// API Client - базовый HTTP клиент на Axios
import axios from 'axios'
import { env, getApiUrl } from '@config/env'
import { 
  DEFAULT_HEADERS, 
  TIMEOUTS, 
  HTTP_STATUS,
  getErrorMessage,
  isRetryableError,
  RETRY_CONFIG,
  logRequest,
  logResponse,
  logError,
  extractErrorFromResponse
} from '@config/api'

// Создаем instance axios
const apiClient = axios.create({
  baseURL: getApiUrl(),
  timeout: TIMEOUTS.DEFAULT,
  headers: DEFAULT_HEADERS
})

// Request interceptor
apiClient.interceptors.request.use(
  (config) => {
    // Добавляем токен из localStorage если есть
    const authData = localStorage.getItem('auth-storage')
    if (authData) {
      try {
        const parsedAuth = JSON.parse(authData)
        if (parsedAuth.state?.token) {
          config.headers.Authorization = `Bearer ${parsedAuth.state.token}`
        }
      } catch (error) {
        console.warn('Failed to parse auth token:', error)
      }
    }

    // Добавляем request ID для трейсинга
    config.metadata = { 
      startTime: Date.now(),
      requestId: `req_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
    }

    // Логируем запрос в dev режиме
    logRequest(config)

    return config
  },
  (error) => {
    logError(error)
    return Promise.reject(error)
  }
)

// Response interceptor
apiClient.interceptors.response.use(
  (response) => {
    // Добавляем время выполнения запроса
    if (response.config.metadata) {
      const duration = Date.now() - response.config.metadata.startTime
      response.config.metadata.duration = duration
    }

    // Логируем ответ в dev режиме
    logResponse(response)

    return response
  },
  async (error) => {
    const originalRequest = error.config

    // Логируем ошибку
    logError(error)

    // Обработка 401 ошибки (токен истек)
    if (error.response?.status === HTTP_STATUS.UNAUTHORIZED && !originalRequest._retry) {
      originalRequest._retry = true

      try {
        // Пытаемся обновить токен
        const refreshResult = await refreshAuthToken()
        
        if (refreshResult.success) {
          // Повторяем оригинальный запрос с новым токеном
          originalRequest.headers.Authorization = `Bearer ${refreshResult.token}`
          return apiClient(originalRequest)
        } else {
          // Если не удалось обновить токен, выходим
          handleAuthFailure()
        }
      } catch (refreshError) {
        console.error('Token refresh failed:', refreshError)
        handleAuthFailure()
      }
    }

    // Retry логика для серверных ошибок
    if (
      isRetryableError(error.response?.status) &&
      originalRequest._retryCount < RETRY_CONFIG.MAX_RETRIES
    ) {
      originalRequest._retryCount = (originalRequest._retryCount || 0) + 1
      
      const delay = RETRY_CONFIG.RETRY_DELAY * Math.pow(RETRY_CONFIG.RETRY_MULTIPLIER, originalRequest._retryCount - 1)
      
      console.log(`Retrying request (${originalRequest._retryCount}/${RETRY_CONFIG.MAX_RETRIES}) after ${delay}ms`)
      
      await new Promise(resolve => setTimeout(resolve, delay))
      return apiClient(originalRequest)
    }

    // Возвращаем обработанную ошибку
    const processedError = extractErrorFromResponse(error)
    return Promise.reject(processedError)
  }
)

// Функция для обновления токена
const refreshAuthToken = async () => {
  try {
    const authData = localStorage.getItem('auth-storage')
    if (!authData) {
      return { success: false }
    }

    const parsedAuth = JSON.parse(authData)
    const refreshToken = parsedAuth.state?.refreshToken || parsedAuth.state?.token

    if (!refreshToken) {
      return { success: false }
    }

    // Отправляем запрос на обновление токена
    const response = await axios.post(getApiUrl('/api/v1/auth/refresh'), {}, {
      headers: {
        Authorization: `Bearer ${refreshToken}`,
        'Content-Type': 'application/json'
      },
      timeout: TIMEOUTS.SHORT
    })

    if (response.data.success && response.data.token) {
      // Обновляем токен в localStorage
      const newAuthData = {
        ...parsedAuth,
        state: {
          ...parsedAuth.state,
          token: response.data.token
        }
      }
      localStorage.setItem('auth-storage', JSON.stringify(newAuthData))

      return { 
        success: true, 
        token: response.data.token 
      }
    }

    return { success: false }
  } catch (error) {
    console.error('Refresh token request failed:', error)
    return { success: false }
  }
}

// Функция для обработки ошибок аутентификации
const handleAuthFailure = () => {
  // Очищаем localStorage
  localStorage.removeItem('auth-storage')
  
  // Перенаправляем на страницу входа
  if (window.location.pathname !== '/login') {
    window.location.href = '/login'
  }
}

// Вспомогательные функции для API запросов
export const apiRequest = {
  get: (url, config = {}) => {
    return apiClient.get(url, config)
  },

  post: (url, data = {}, config = {}) => {
    return apiClient.post(url, data, config)
  },

  put: (url, data = {}, config = {}) => {
    return apiClient.put(url, data, config)
  },

  patch: (url, data = {}, config = {}) => {
    return apiClient.patch(url, data, config)
  },

  delete: (url, config = {}) => {
    return apiClient.delete(url, config)
  }
}

// Функция для отправки файлов
export const uploadFile = async (url, file, onProgress) => {
  const formData = new FormData()
  formData.append('file', file)

  const config = {
    headers: {
      'Content-Type': 'multipart/form-data'
    },
    timeout: TIMEOUTS.UPLOAD,
    onUploadProgress: (progressEvent) => {
      if (onProgress) {
        const percentCompleted = Math.round(
          (progressEvent.loaded * 100) / progressEvent.total
        )
        onProgress(percentCompleted)
      }
    }
  }

  return apiClient.post(url, formData, config)
}

// Функция для проверки статуса API
export const checkApiHealth = async () => {
  try {
    const response = await apiClient.get('/health', { 
      timeout: TIMEOUTS.SHORT 
    })
    return {
      status: 'healthy',
      data: response.data
    }
  } catch (error) {
    return {
      status: 'unhealthy',
      error: error.message
    }
  }
}

// Функция для установки базового URL (для тестирования)
export const setBaseURL = (baseURL) => {
  apiClient.defaults.baseURL = baseURL
}

// Функция для установки токена авторизации
export const setAuthToken = (token) => {
  if (token) {
    apiClient.defaults.headers.Authorization = `Bearer ${token}`
  } else {
    delete apiClient.defaults.headers.Authorization
  }
}

// Функция для очистки токена
export const clearAuthToken = () => {
  delete apiClient.defaults.headers.Authorization
}

// Экспортируем основной клиент
export default apiClient