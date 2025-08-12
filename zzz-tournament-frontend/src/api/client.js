// src/api/client.js - ������ � ������� �� ����������� ������
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

// ������ �� ����������� ������
let isRefreshing = false
let refreshPromise = null
let refreshAttempts = 0
const MAX_REFRESH_ATTEMPTS = 3
const REFRESH_COOLDOWN = 5000 // 5 ������

// ������� instance axios
const apiClient = axios.create({
  baseURL: getApiUrl(),
  timeout: TIMEOUTS.DEFAULT,
  headers: DEFAULT_HEADERS
})

// Request interceptor
apiClient.interceptors.request.use(
  (config) => {
    // ��������� ����� �� localStorage ���� ����
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

    // ��������� request ID ��� ���������
    config.metadata = { 
      startTime: Date.now(),
      requestId: `req_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
    }

    logRequest(config)
    return config
  },
  (error) => {
    logError(error)
    return Promise.reject(error)
  }
)

// Response interceptor � ������� �� ������
apiClient.interceptors.response.use(
  (response) => {
    // ���������� ������� ��� �������� �������
    refreshAttempts = 0
    
    if (response.config.metadata) {
      const duration = Date.now() - response.config.metadata.startTime
      response.config.metadata.duration = duration
    }

    logResponse(response)
    return response
  },
  async (error) => {
    const originalRequest = error.config

    logError(error)

    // ��������� 401 ������ � ������� �� ������
    if (error.response?.status === HTTP_STATUS.UNAUTHORIZED && !originalRequest._retry) {
      originalRequest._retry = true

      // ��������� ����� ������� refresh
      if (refreshAttempts >= MAX_REFRESH_ATTEMPTS) {
        console.error(`? Max refresh attempts (${MAX_REFRESH_ATTEMPTS}) reached, logging out`)
        handleAuthFailure()
        return Promise.reject(error)
      }

      // ���� ��� ���� ������� refresh, ���� ��� ����������
      if (isRefreshing) {
        console.log('? Refresh already in progress, waiting...')
        try {
          const result = await refreshPromise
          if (result.success) {
            originalRequest.headers.Authorization = `Bearer ${result.token}`
            return apiClient(originalRequest)
          } else {
            return Promise.reject(error)
          }
        } catch (refreshError) {
          return Promise.reject(error)
        }
      }

      // �������� ������� refresh
      isRefreshing = true
      refreshAttempts++
      
      console.log(`?? Starting token refresh attempt ${refreshAttempts}/${MAX_REFRESH_ATTEMPTS}`)
      
      refreshPromise = refreshAuthToken()
      
      try {
        const refreshResult = await refreshPromise
        
        if (refreshResult.success) {
          console.log('? Token refresh successful')
          // ��������� ������������ ������ � ����� �������
          originalRequest.headers.Authorization = `Bearer ${refreshResult.token}`
          return apiClient(originalRequest)
        } else {
          console.error('? Token refresh failed')
          handleAuthFailure()
          return Promise.reject(error)
        }
      } catch (refreshError) {
        console.error('? Token refresh error:', refreshError)
        handleAuthFailure()
        return Promise.reject(error)
      } finally {
        // ���������� ���� � ������������� cooldown
        isRefreshing = false
        refreshPromise = null
        
        // ���� �������� ������ �������, ������������� cooldown
        if (refreshAttempts >= MAX_REFRESH_ATTEMPTS) {
          setTimeout(() => {
            refreshAttempts = 0
            console.log('?? Refresh attempts counter reset after cooldown')
          }, REFRESH_COOLDOWN)
        }
      }
    }

    // Retry ������ ��� ��������� ������
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

    const processedError = extractErrorFromResponse(error)
    return Promise.reject(processedError)
  }
)

// ���������� ������� ��� ���������� ������
const refreshAuthToken = async () => {
  try {
    const authData = localStorage.getItem('auth-storage')
    if (!authData) {
      console.warn('No auth data found for refresh')
      return { success: false, reason: 'no_auth_data' }
    }

    const parsedAuth = JSON.parse(authData)
    const refreshToken = parsedAuth.state?.refreshToken

    if (!refreshToken) {
      console.warn('No refresh token found')
      return { success: false, reason: 'no_refresh_token' }
    }

    console.log('?? Attempting to refresh token...')

    // ���������� ������ �� ���������� ������ � ���������
    const response = await axios.post(getApiUrl('/api/v1/auth/refresh'), {}, {
      headers: {
        Authorization: `Bearer ${refreshToken}`,
        'Content-Type': 'application/json'
      },
      timeout: TIMEOUTS.SHORT,
      // �����: �� ���������� apiClient ����� �������� ��������
      validateStatus: (status) => status < 500 // ��������� ��� ������� < 500
    })

    // ��������� ������ ������
    if (response.status === 401) {
      console.warn('Refresh token expired (401)')
      return { success: false, reason: 'refresh_token_expired' }
    }

    if (response.status >= 400) {
      console.error(`Refresh failed with status ${response.status}:`, response.data)
      return { success: false, reason: 'server_error', status: response.status }
    }

    // ������������ �������� �����
    if (response.data?.success && response.data?.data) {
      const { access_token, refresh_token, user } = response.data.data

      if (!access_token) {
        console.error('No access_token in refresh response')
        return { success: false, reason: 'no_access_token_in_response' }
      }

      // ��������� ������ � localStorage
      const newAuthData = {
        ...parsedAuth,
        state: {
          ...parsedAuth.state,
          token: access_token,
          refreshToken: refresh_token || refreshToken, // Fallback to old refresh token
          user: user || parsedAuth.state.user,
          lastRefresh: Date.now() // ��������� timestamp ���������� refresh
        }
      }
      
      localStorage.setItem('auth-storage', JSON.stringify(newAuthData))

      console.log('? Tokens refreshed and saved successfully')
      return { 
        success: true, 
        token: access_token,
        refreshToken: refresh_token || refreshToken
      }
    }

    console.error('Invalid refresh response structure:', response.data)
    return { success: false, reason: 'invalid_response_structure' }

  } catch (error) {
    console.error('Refresh token request failed:', error)
    
    // ����������� ��� ������
    if (error.code === 'ECONNABORTED') {
      return { success: false, reason: 'timeout' }
    }
    
    if (error.response?.status === 401) {
      return { success: false, reason: 'refresh_token_expired' }
    }
    
    if (error.response?.status >= 500) {
      return { success: false, reason: 'server_error' }
    }
    
    return { success: false, reason: 'network_error', error: error.message }
  }
}

// ������� ��� ��������� ������ �������������� � ������� �� �����
let authFailureHandled = false

const handleAuthFailure = () => {
  // ������������� ������������� ������
  if (authFailureHandled) {
    console.log('Auth failure already handled, skipping...')
    return
  }
  
  authFailureHandled = true
  
  console.log('?? Auth failure - clearing tokens and redirecting')
  
  // ������� ��� ���������
  localStorage.removeItem('auth-storage')
  isRefreshing = false
  refreshPromise = null
  refreshAttempts = 0
  
  // ���������� ������������ (���� ���� ���������� ������� �����������)
  if (window.showNotification) {
    window.showNotification('������ �������. ���������� ����� ������.', 'warning')
  }
  
  // �������������� �� �������� ����� � ��������� ��� �������������� �����
  setTimeout(() => {
    if (window.location.pathname !== '/login') {
      const currentPath = window.location.pathname
      window.location.href = `/login?from=${encodeURIComponent(currentPath)}`
    }
    
    // ���������� ���� ����� ��������� �����
    setTimeout(() => {
      authFailureHandled = false
    }, 2000)
  }, 100)
}

// ������� ��� ������� ������ ��������� refresh (��� �������)
export const resetRefreshState = () => {
  console.log('?? Manually resetting refresh state')
  isRefreshing = false
  refreshPromise = null
  refreshAttempts = 0
  authFailureHandled = false
}

// ������� ��� �������� ��������� refresh (��� �������)
export const getRefreshState = () => {
  return {
    isRefreshing,
    refreshAttempts,
    maxAttempts: MAX_REFRESH_ATTEMPTS,
    authFailureHandled,
    hasRefreshPromise: !!refreshPromise
  }
}

// ��������� ��� �������� ��� ���������...
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

export const setBaseURL = (baseURL) => {
  apiClient.defaults.baseURL = baseURL
}

export const setAuthToken = (token) => {
  if (token) {
    apiClient.defaults.headers.Authorization = `Bearer ${token}`
  } else {
    delete apiClient.defaults.headers.Authorization
  }
}

export const clearAuthToken = () => {
  delete apiClient.defaults.headers.Authorization
}

export default apiClient