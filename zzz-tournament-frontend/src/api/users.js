// Users API methods
import { apiRequest } from './client'
import { API_ENDPOINTS } from '@config/api'
import { createPaginationParams } from '@config/api'

/**
 * Получение профиля текущего пользователя
 * @returns {Promise<Object>} профиль пользователя
 */
export const getUserProfile = async () => {
  try {
    const response = await apiRequest.get(API_ENDPOINTS.USERS.PROFILE)

    return {
      success: true,
      user: response.data.data || response.data.user || response.data
    }
  } catch (error) {
    return {
      success: false,
      error: error.message || 'Ошибка получения профиля'
    }
  }
}

/**
 * Обновление профиля пользователя
 * @param {Object} profileData - данные профиля для обновления
 * @param {string} [profileData.username] - новое имя пользователя
 * @param {string} [profileData.email] - новый email
 * @param {string} [profileData.avatar] - URL аватара
 * @returns {Promise<Object>} результат обновления
 */
export const updateUserProfile = async (profileData) => {
  try {
    const response = await apiRequest.put(API_ENDPOINTS.USERS.UPDATE_PROFILE, profileData)

    return {
      success: true,
      user: response.data.data || response.data.user || response.data,
      message: response.data.message || 'Профиль обновлен'
    }
  } catch (error) {
    return {
      success: false,
      error: error.message || 'Ошибка обновления профиля',
      details: error.details || []
    }
  }
}

/**
 * Получение лидерборда
 * @param {Object} params - параметры запроса
 * @param {number} [params.page=1] - номер страницы
 * @param {number} [params.limit=20] - количество записей на странице
 * @param {string} [params.sortBy='rating'] - поле для сортировки
 * @param {string} [params.order='desc'] - порядок сортировки
 * @returns {Promise<Object>} список лидеров
 */
export const getLeaderboard = async (params = {}) => {
  try {
    const queryParams = {
      ...createPaginationParams(params.page, params.limit),
      sort_by: params.sortBy || 'rating',
      order: params.order || 'desc'
    }

    const response = await apiRequest.get(API_ENDPOINTS.USERS.LEADERBOARD, {
      params: queryParams
    })

    return {
      success: true,
      users: response.data.data || response.data,
      pagination: response.data.pagination || null
    }
  } catch (error) {
    return {
      success: false,
      error: error.message || 'Ошибка получения лидерборда'
    }
  }
}

/**
 * Поиск пользователей
 * @param {Object} searchParams - параметры поиска
 * @param {string} searchParams.query - поисковый запрос
 * @param {number} [searchParams.page=1] - номер страницы
 * @param {number} [searchParams.limit=20] - количество записей
 * @returns {Promise<Object>} результаты поиска
 */
export const searchUsers = async (searchParams) => {
  try {
    const queryParams = {
      q: searchParams.query,
      ...createPaginationParams(searchParams.page, searchParams.limit)
    }

    const response = await apiRequest.get('/api/v1/users/search', {
      params: queryParams
    })

    return {
      success: true,
      users: response.data.data || response.data,
      pagination: response.data.pagination || null
    }
  } catch (error) {
    return {
      success: false,
      error: error.message || 'Ошибка поиска пользователей'
    }
  }
}

/**
 * Получение статистики пользователя
 * @param {number} [userId] - ID пользователя (если не указан, то текущий)
 * @returns {Promise<Object>} статистика пользователя
 */
export const getUserStats = async (userId = null) => {
  try {
    const endpoint = userId 
      ? `/api/v1/users/${userId}/stats`
      : '/api/v1/users/profile/stats'

    const response = await apiRequest.get(endpoint)

    return {
      success: true,
      stats: response.data.data || response.data
    }
  } catch (error) {
    return {
      success: false,
      error: error.message || 'Ошибка получения статистики'
    }
  }
}

/**
 * Получение истории матчей пользователя
 * @param {Object} params - параметры запроса
 * @param {number} [params.userId] - ID пользователя
 * @param {number} [params.page=1] - номер страницы
 * @param {number} [params.limit=20] - количество записей
 * @returns {Promise<Object>} история матчей
 */
export const getUserMatchHistory = async (params = {}) => {
  try {
    const endpoint = params.userId 
      ? `/api/v1/users/${params.userId}/matches`
      : '/api/v1/users/profile/matches'

    const queryParams = createPaginationParams(params.page, params.limit)

    const response = await apiRequest.get(endpoint, {
      params: queryParams
    })

    return {
      success: true,
      matches: response.data.data || response.data,
      pagination: response.data.pagination || null
    }
  } catch (error) {
    return {
      success: false,
      error: error.message || 'Ошибка получения истории матчей'
    }
  }
}

/**
 * Загрузка аватара пользователя
 * @param {File} file - файл изображения
 * @param {Function} [onProgress] - коллбек для отслеживания прогресса
 * @returns {Promise<Object>} результат загрузки
 */
export const uploadAvatar = async (file, onProgress) => {
  try {
    const formData = new FormData()
    formData.append('avatar', file)

    const response = await apiRequest.post('/api/v1/users/avatar', formData, {
      headers: {
        'Content-Type': 'multipart/form-data'
      },
      timeout: 60000, // 60 секунд для загрузки
      onUploadProgress: (progressEvent) => {
        if (onProgress) {
          const percentCompleted = Math.round(
            (progressEvent.loaded * 100) / progressEvent.total
          )
          onProgress(percentCompleted)
        }
      }
    })

    return {
      success: true,
      avatarUrl: response.data.avatar_url || response.data.data?.avatar_url,
      message: response.data.message || 'Аватар загружен'
    }
  } catch (error) {
    return {
      success: false,
      error: error.message || 'Ошибка загрузки аватара'
    }
  }
}

/**
 * Удаление аватара пользователя
 * @returns {Promise<Object>} результат удаления
 */
export const deleteAvatar = async () => {
  try {
    const response = await apiRequest.delete('/api/v1/users/avatar')

    return {
      success: true,
      message: response.data.message || 'Аватар удален'
    }
  } catch (error) {
    return {
      success: false,
      error: error.message || 'Ошибка удаления аватара'
    }
  }
}

/**
 * Получение публичного профиля пользователя
 * @param {number} userId - ID пользователя
 * @returns {Promise<Object>} публичный профиль
 */
export const getPublicProfile = async (userId) => {
  try {
    const response = await apiRequest.get(`/api/v1/users/${userId}/profile`)

    return {
      success: true,
      user: response.data.data || response.data
    }
  } catch (error) {
    return {
      success: false,
      error: error.message || 'Ошибка получения профиля пользователя'
    }
  }
}

// Экспорт всех методов
export const usersAPI = {
  getProfile: getUserProfile,
  updateProfile: updateUserProfile,
  getLeaderboard,
  searchUsers,
  getStats: getUserStats,
  getMatchHistory: getUserMatchHistory,
  uploadAvatar,
  deleteAvatar,
  getPublicProfile
}

export default usersAPI