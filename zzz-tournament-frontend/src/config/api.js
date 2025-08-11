import { env } from './env.js'

// API Endpoints
export const API_ENDPOINTS = {
  // Auth endpoints
  AUTH: {
    REGISTER: '/api/v1/auth/register',
    LOGIN: '/api/v1/auth/login',
    REFRESH: '/api/v1/auth/refresh',
    LOGOUT: '/api/v1/auth/logout'
  },
  
  // User endpoints
  USERS: {
    PROFILE: '/api/v1/users/profile',
    UPDATE_PROFILE: '/api/v1/users/profile',
    LEADERBOARD: '/api/v1/users/leaderboard'
  },
  
  // Hero endpoints
  HEROES: {
    LIST: '/api/v1/heroes',
    CREATE: '/api/v1/heroes',
    UPDATE: (id) => `/api/v1/heroes/${id}`,
    DELETE: (id) => `/api/v1/heroes/${id}`,
    GET: (id) => `/api/v1/heroes/${id}`
  },
  
  // Room endpoints
  ROOMS: {
    LIST: '/api/v1/rooms',
    CREATE: '/api/v1/rooms',
    GET: (id) => `/api/v1/rooms/${id}`,
    UPDATE: (id) => `/api/v1/rooms/${id}`,
    DELETE: (id) => `/api/v1/rooms/${id}`,
    JOIN: (id) => `/api/v1/rooms/${id}/join`,
    LEAVE: (id) => `/api/v1/rooms/${id}/leave`
  },
  
  // Tournament endpoints
  TOURNAMENTS: {
    START: (roomId) => `/api/v1/rooms/${roomId}/tournament/start`,
    GET: (id) => `/api/v1/tournaments/${id}`,
    SUBMIT_RESULT: (tournamentId, matchId) => `/api/v1/tournaments/${tournamentId}/matches/${matchId}/result`
  },
  
  // Chat endpoints
  CHAT: {
    MESSAGES: (roomId) => `/api/v1/rooms/${roomId}/messages`,
    // Отправка сообщений через WebSocket
  },
  
  // WebSocket endpoint
  WEBSOCKET: '/ws'
}

// HTTP Methods
export const HTTP_METHODS = {
  GET: 'GET',
  POST: 'POST',
  PUT: 'PUT',
  PATCH: 'PATCH',
  DELETE: 'DELETE'
}

// HTTP Status Codes
export const HTTP_STATUS = {
  // Success
  OK: 200,
  CREATED: 201,
  NO_CONTENT: 204,
  
  // Client Errors
  BAD_REQUEST: 400,
  UNAUTHORIZED: 401,
  FORBIDDEN: 403,
  NOT_FOUND: 404,
  CONFLICT: 409,
  UNPROCESSABLE_ENTITY: 422,
  TOO_MANY_REQUESTS: 429,
  
  // Server Errors
  INTERNAL_SERVER_ERROR: 500,
  BAD_GATEWAY: 502,
  SERVICE_UNAVAILABLE: 503,
  GATEWAY_TIMEOUT: 504
}

// Request timeouts (в миллисекундах)
export const TIMEOUTS = {
  DEFAULT: 10000,    // 10 seconds
  LONG: 30000,       // 30 seconds
  SHORT: 5000,       // 5 seconds
  UPLOAD: 60000      // 60 seconds для загрузки файлов
}

// Headers
export const HEADERS = {
  CONTENT_TYPE: 'Content-Type',
  AUTHORIZATION: 'Authorization',
  ACCEPT: 'Accept',
  USER_AGENT: 'User-Agent',
  X_REQUESTED_WITH: 'X-Requested-With',
  X_CSRF_TOKEN: 'X-CSRF-Token'
}

// Content Types
export const CONTENT_TYPES = {
  JSON: 'application/json',
  FORM_DATA: 'multipart/form-data',
  URL_ENCODED: 'application/x-www-form-urlencoded',
  TEXT: 'text/plain'
}

// Default headers для всех запросов
export const DEFAULT_HEADERS = {
  [HEADERS.CONTENT_TYPE]: CONTENT_TYPES.JSON,
  [HEADERS.ACCEPT]: CONTENT_TYPES.JSON,
  [HEADERS.X_REQUESTED_WITH]: 'XMLHttpRequest'
}

// Retry configuration
export const RETRY_CONFIG = {
  MAX_RETRIES: 3,
  RETRY_DELAY: 1000,     // Начальная задержка
  RETRY_MULTIPLIER: 2,   // Множитель для экспоненциального backoff
  RETRYABLE_STATUSES: [
    HTTP_STATUS.INTERNAL_SERVER_ERROR,
    HTTP_STATUS.BAD_GATEWAY,
    HTTP_STATUS.SERVICE_UNAVAILABLE,
    HTTP_STATUS.GATEWAY_TIMEOUT
  ]
}

// Cache configuration
export const CACHE_CONFIG = {
  DEFAULT_TTL: env.CACHE_DURATION,
  MAX_SIZE: 100,         // Максимальное количество кэшированных запросов
  CACHEABLE_METHODS: [HTTP_METHODS.GET],
  CACHE_HEADERS: {
    'Cache-Control': `max-age=${env.CACHE_DURATION / 1000}`
  }
}

// WebSocket configuration
export const WS_CONFIG = {
  RECONNECT_INTERVAL: 5000,    // 5 seconds
  MAX_RECONNECT_ATTEMPTS: 10,
  HEARTBEAT_INTERVAL: 30000,   // 30 seconds
  CONNECTION_TIMEOUT: 10000,   // 10 seconds
  
  // Message types
  MESSAGE_TYPES: {
    // Outgoing (client -> server)
    JOIN_ROOM: 'join_room',
    LEAVE_ROOM: 'leave_room',
    CHAT_MESSAGE: 'chat_message',
    HEARTBEAT: 'heartbeat',
    
    // Incoming (server -> client)
    ROOM_UPDATED: 'room_updated',
    TOURNAMENT_STARTED: 'tournament_started',
    TOURNAMENT_UPDATED: 'tournament_updated',
    CHAT_MESSAGE_RECEIVED: 'chat_message',
    MATCH_ASSIGNED: 'match_assigned',
    USER_JOINED: 'user_joined',
    USER_LEFT: 'user_left',
    ERROR: 'error'
  }
}

// Error messages
export const ERROR_MESSAGES = {
  NETWORK_ERROR: 'Ошибка сети. Проверьте подключение к интернету.',
  TIMEOUT_ERROR: 'Превышено время ожидания ответа сервера.',
  UNAUTHORIZED: 'Необходима авторизация.',
  FORBIDDEN: 'Недостаточно прав доступа.',
  NOT_FOUND: 'Ресурс не найден.',
  VALIDATION_ERROR: 'Ошибка валидации данных.',
  SERVER_ERROR: 'Внутренняя ошибка сервера.',
  TOO_MANY_REQUESTS: 'Слишком много запросов. Попробуйте позже.',
  UNKNOWN_ERROR: 'Произошла неизвестная ошибка.'
}

// Rate limiting
export const RATE_LIMIT = {
  MAX_REQUESTS: env.API_RATE_LIMIT,
  WINDOW_SIZE: 60000,    // 1 minute
  WARNING_THRESHOLD: 0.8 // Предупреждение при 80% лимита
}

// File upload configuration
export const UPLOAD_CONFIG = {
  MAX_FILE_SIZE: env.MAX_FILE_SIZE,
  ALLOWED_TYPES: env.ALLOWED_FILE_TYPES,
  CHUNK_SIZE: 1024 * 1024, // 1MB chunks для больших файлов
  TIMEOUT: TIMEOUTS.UPLOAD
}

// Development configuration
export const DEV_CONFIG = {
  MOCK_DELAY: 500,       // Задержка для mock API
  LOG_REQUESTS: env.DEBUG_MODE,
  LOG_RESPONSES: env.DEBUG_MODE,
  LOG_ERRORS: true
}

// Utility functions
export const isRetryableError = (status) => {
  return RETRY_CONFIG.RETRYABLE_STATUSES.includes(status)
}

export const isClientError = (status) => {
  return status >= 400 && status < 500
}

export const isServerError = (status) => {
  return status >= 500
}

export const getErrorMessage = (status) => {
  switch (status) {
    case HTTP_STATUS.UNAUTHORIZED:
      return ERROR_MESSAGES.UNAUTHORIZED
    case HTTP_STATUS.FORBIDDEN:
      return ERROR_MESSAGES.FORBIDDEN
    case HTTP_STATUS.NOT_FOUND:
      return ERROR_MESSAGES.NOT_FOUND
    case HTTP_STATUS.UNPROCESSABLE_ENTITY:
      return ERROR_MESSAGES.VALIDATION_ERROR
    case HTTP_STATUS.TOO_MANY_REQUESTS:
      return ERROR_MESSAGES.TOO_MANY_REQUESTS
    case HTTP_STATUS.INTERNAL_SERVER_ERROR:
      return ERROR_MESSAGES.SERVER_ERROR
    default:
      if (isServerError(status)) {
        return ERROR_MESSAGES.SERVER_ERROR
      }
      return ERROR_MESSAGES.UNKNOWN_ERROR
  }
}

// Helper для создания URL с параметрами
export const buildUrlWithParams = (baseUrl, params = {}) => {
  const url = new URL(baseUrl, window.location.origin)
  Object.entries(params).forEach(([key, value]) => {
    if (value !== null && value !== undefined && value !== '') {
      url.searchParams.append(key, String(value))
    }
  })
  return url.toString()
}

// Helper для создания FormData
export const createFormData = (data) => {
  const formData = new FormData()
  Object.entries(data).forEach(([key, value]) => {
    if (value instanceof File || value instanceof Blob) {
      formData.append(key, value)
    } else if (Array.isArray(value)) {
      value.forEach((item, index) => {
        formData.append(`${key}[${index}]`, item)
      })
    } else if (typeof value === 'object' && value !== null) {
      formData.append(key, JSON.stringify(value))
    } else {
      formData.append(key, String(value))
    }
  })
  return formData
}

// Validation helpers
export const validateFile = (file) => {
  const errors = []
  
  // Проверяем размер
  if (file.size > UPLOAD_CONFIG.MAX_FILE_SIZE) {
    errors.push(`Размер файла превышает ${UPLOAD_CONFIG.MAX_FILE_SIZE / 1024 / 1024}MB`)
  }
  
  // Проверяем тип
  if (!UPLOAD_CONFIG.ALLOWED_TYPES.includes(file.type)) {
    errors.push(`Неподдерживаемый тип файла. Разрешены: ${UPLOAD_CONFIG.ALLOWED_TYPES.join(', ')}`)
  }
  
  return {
    isValid: errors.length === 0,
    errors
  }
}

// Response helpers
export const isSuccessResponse = (status) => {
  return status >= 200 && status < 300
}

export const extractErrorFromResponse = (error) => {
  // Если есть response с данными об ошибке
  if (error.response?.data) {
    const { data } = error.response
    
    // Если есть детали ошибок валидации
    if (data.details && Array.isArray(data.details)) {
      return {
        message: data.error || ERROR_MESSAGES.VALIDATION_ERROR,
        details: data.details,
        status: error.response.status
      }
    }
    
    // Обычная ошибка с сообщением
    if (data.error) {
      return {
        message: data.error,
        status: error.response.status
      }
    }
  }
  
  // Ошибка сети или таймаут
  if (error.code === 'NETWORK_ERROR' || error.code === 'ERR_NETWORK') {
    return {
      message: ERROR_MESSAGES.NETWORK_ERROR,
      status: 0
    }
  }
  
  if (error.code === 'ECONNABORTED') {
    return {
      message: ERROR_MESSAGES.TIMEOUT_ERROR,
      status: 0
    }
  }
  
  // Fallback
  return {
    message: error.message || ERROR_MESSAGES.UNKNOWN_ERROR,
    status: error.response?.status || 0
  }
}

// Auth token helpers
export const getAuthHeader = (token) => {
  return token ? `Bearer ${token}` : null
}

export const createAuthHeaders = (token) => {
  const headers = { ...DEFAULT_HEADERS }
  if (token) {
    headers[HEADERS.AUTHORIZATION] = getAuthHeader(token)
  }
  return headers
}

// Pagination helpers
export const createPaginationParams = (page = 1, limit = env.DEFAULT_PAGE_SIZE) => {
  return {
    page: Math.max(1, page),
    limit: Math.min(limit, env.MAX_PAGE_SIZE)
  }
}

// Cache key helpers
export const createCacheKey = (method, url, params = {}) => {
  const key = `${method}:${url}`
  const paramString = Object.keys(params).length > 0 
    ? `:${JSON.stringify(params)}`
    : ''
  return `${key}${paramString}`
}

// Request ID для трейсинга
export const generateRequestId = () => {
  return `req_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
}

// Debug helpers
export const logRequest = (config) => {
  if (DEV_CONFIG.LOG_REQUESTS) {
    console.group(`🚀 API Request: ${config.method?.toUpperCase()} ${config.url}`)
    console.log('Config:', config)
    console.log('Headers:', config.headers)
    if (config.data) {
      console.log('Data:', config.data)
    }
    console.groupEnd()
  }
}

export const logResponse = (response) => {
  if (DEV_CONFIG.LOG_RESPONSES) {
    console.group(`✅ API Response: ${response.status} ${response.config.url}`)
    console.log('Status:', response.status)
    console.log('Headers:', response.headers)
    console.log('Data:', response.data)
    console.groupEnd()
  }
}

export const logError = (error) => {
  if (DEV_CONFIG.LOG_ERRORS) {
    console.group(`❌ API Error: ${error.config?.url || 'Unknown'}`)
    console.error('Error:', error)
    if (error.response) {
      console.log('Status:', error.response.status)
      console.log('Data:', error.response.data)
    }
    console.groupEnd()
  }
}

export default {
  API_ENDPOINTS,
  HTTP_METHODS,
  HTTP_STATUS,
  TIMEOUTS,
  HEADERS,
  CONTENT_TYPES,
  DEFAULT_HEADERS,
  RETRY_CONFIG,
  CACHE_CONFIG,
  WS_CONFIG,
  ERROR_MESSAGES,
  RATE_LIMIT,
  UPLOAD_CONFIG,
  DEV_CONFIG
}