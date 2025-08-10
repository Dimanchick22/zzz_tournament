// Environment configuration

// Получаем переменные окружения с дефолтными значениями
const getEnvVar = (key, defaultValue = '') => {
    const value = import.meta.env[key]
    return value !== undefined ? value : defaultValue
  }
  
  // Проверяем обязательные переменные
  const requiredEnvVars = [
    'VITE_API_BASE_URL',
    'VITE_WS_BASE_URL'
  ]
  
  // Валидация обязательных переменных
  const validateEnv = () => {
    const missing = requiredEnvVars.filter(key => !import.meta.env[key])
    
    if (missing.length > 0) {
      console.error('Missing required environment variables:', missing)
      throw new Error(`Missing required environment variables: ${missing.join(', ')}`)
    }
  }
  
  // Валидируем при импорте (только в dev режиме)
  if (import.meta.env.DEV) {
    validateEnv()
  }
  
  // Экспортируем конфигурацию
  export const env = {
    // Основные настройки
    NODE_ENV: import.meta.env.MODE || 'development',
    DEV: import.meta.env.DEV,
    PROD: import.meta.env.PROD,
    
    // API Configuration
    API_BASE_URL: getEnvVar('VITE_API_BASE_URL', 'http://localhost:8080'),
    WS_BASE_URL: getEnvVar('VITE_WS_BASE_URL', 'ws://localhost:8080'),
    
    // App Configuration
    APP_NAME: getEnvVar('VITE_APP_NAME', 'ZZZ Tournament'),
    APP_VERSION: getEnvVar('VITE_APP_VERSION', '1.0.0'),
    APP_DESCRIPTION: getEnvVar('VITE_APP_DESCRIPTION', 'Турнирная система для Zenless Zone Zero'),
    
    // Features Flags
    ENABLE_DEV_TOOLS: getEnvVar('VITE_ENABLE_DEV_TOOLS', 'true') === 'true',
    ENABLE_MOCK_API: getEnvVar('VITE_ENABLE_MOCK_API', 'false') === 'true',
    ENABLE_ANALYTICS: getEnvVar('VITE_ENABLE_ANALYTICS', 'false') === 'true',
    
    // External Services
    SENTRY_DSN: getEnvVar('VITE_SENTRY_DSN'),
    GOOGLE_ANALYTICS_ID: getEnvVar('VITE_GOOGLE_ANALYTICS_ID'),
    
    // Theme
    DEFAULT_THEME: getEnvVar('VITE_DEFAULT_THEME', 'dark'),
    
    // Pagination
    DEFAULT_PAGE_SIZE: parseInt(getEnvVar('VITE_DEFAULT_PAGE_SIZE', '20'), 10),
    MAX_PAGE_SIZE: parseInt(getEnvVar('VITE_MAX_PAGE_SIZE', '100'), 10),
    
    // Upload
    MAX_FILE_SIZE: parseInt(getEnvVar('VITE_MAX_FILE_SIZE', '10485760'), 10), // 10MB
    ALLOWED_FILE_TYPES: getEnvVar('VITE_ALLOWED_FILE_TYPES', 'image/jpeg,image/png,image/webp').split(','),
    
    // Rate Limiting (для UI feedback)
    API_RATE_LIMIT: parseInt(getEnvVar('VITE_API_RATE_LIMIT', '60'), 10),
    WS_RATE_LIMIT: parseInt(getEnvVar('VITE_WS_RATE_LIMIT', '10'), 10),
    
    // Cache
    CACHE_DURATION: parseInt(getEnvVar('VITE_CACHE_DURATION', '300000'), 10), // 5 minutes
    
    // Debug
    DEBUG_MODE: getEnvVar('VITE_DEBUG_MODE', 'false') === 'true',
    LOG_LEVEL: getEnvVar('VITE_LOG_LEVEL', 'info')
  }
  
  // Функции для работы с env
  export const isDev = () => env.DEV
  export const isProd = () => env.PROD
  export const isDebug = () => env.DEBUG_MODE || env.DEV
  
  // API URLs
  export const getApiUrl = (path = '') => {
    const baseUrl = env.API_BASE_URL.replace(/\/$/, '') // убираем слэш в конце
    const cleanPath = path.replace(/^\//, '') // убираем слэш в начале
    return cleanPath ? `${baseUrl}/${cleanPath}` : baseUrl
  }
  
  export const getWsUrl = (path = '') => {
    const baseUrl = env.WS_BASE_URL.replace(/\/$/, '')
    const cleanPath = path.replace(/^\//, '')
    return cleanPath ? `${baseUrl}/${cleanPath}` : baseUrl
  }
  
  // Логирование конфигурации (только в dev)
  if (isDev()) {
    console.group('🔧 Environment Configuration')
    console.log('Environment:', env.NODE_ENV)
    console.log('API Base URL:', env.API_BASE_URL)
    console.log('WebSocket URL:', env.WS_BASE_URL)
    console.log('Debug Mode:', env.DEBUG_MODE)
    console.log('Dev Tools:', env.ENABLE_DEV_TOOLS)
    console.groupEnd()
  }
  
  export default env