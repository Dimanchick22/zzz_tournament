// Tournaments API methods
import { apiRequest } from './client'
import { API_ENDPOINTS } from '@config/api'
import { createPaginationParams } from '@config/api'

/**
 * Запуск турнира в комнате
 * @param {number} roomId - ID комнаты
 * @returns {Promise<Object>} результат запуска турнира
 */
export const startTournament = async (roomId) => {
  try {
    const response = await apiRequest.post(API_ENDPOINTS.TOURNAMENTS.START(roomId))

    return {
      success: true,
      tournament: response.data.data || response.data,
      message: response.data.message || 'Турнир запущен'
    }
  } catch (error) {
    return {
      success: false,
      error: error.message || 'Ошибка запуска турнира'
    }
  }
}

/**
 * Получение информации о турнире
 * @param {number} tournamentId - ID турнира
 * @returns {Promise<Object>} информация о турнире
 */
export const getTournament = async (tournamentId) => {
  try {
    const response = await apiRequest.get(API_ENDPOINTS.TOURNAMENTS.GET(tournamentId))

    return {
      success: true,
      tournament: response.data.data || response.data
    }
  } catch (error) {
    return {
      success: false,
      error: error.message || 'Ошибка получения информации о турнире'
    }
  }
}

/**
 * Отправка результата матча
 * @param {number} tournamentId - ID турнира
 * @param {number} matchId - ID матча
 * @param {number} winnerId - ID победителя
 * @returns {Promise<Object>} результат отправки
 */
export const submitMatchResult = async (tournamentId, matchId, winnerId) => {
  try {
    const response = await apiRequest.post(
      API_ENDPOINTS.TOURNAMENTS.SUBMIT_RESULT(tournamentId, matchId),
      {
        winner_id: winnerId
      }
    )

    return {
      success: true,
      match: response.data.data || response.data,
      message: response.data.message || 'Результат матча отправлен'
    }
  } catch (error) {
    return {
      success: false,
      error: error.message || 'Ошибка отправки результата матча'
    }
  }
}

/**
 * Получение списка турниров
 * @param {Object} params - параметры запроса
 * @param {number} [params.page=1] - номер страницы
 * @param {number} [params.limit=20] - количество записей
 * @param {string} [params.status] - фильтр по статусу
 * @param {string} [params.sortBy] - поле для сортировки
 * @returns {Promise<Object>} список турниров
 */
export const getTournaments = async (params = {}) => {
  try {
    const queryParams = {
      ...createPaginationParams(params.page, params.limit)
    }

    if (params.status) {
      queryParams.status = params.status
    }

    if (params.sortBy) {
      queryParams.sort_by = params.sortBy
    }

    const response = await apiRequest.get('/api/v1/tournaments', {
      params: queryParams
    })

    return {
      success: true,
      tournaments: response.data.data || response.data,
      pagination: response.data.pagination || null
    }
  } catch (error) {
    return {
      success: false,
      error: error.message || 'Ошибка получения списка турниров'
    }
  }
}

/**
 * Получение матчей турнира
 * @param {number} tournamentId - ID турнира
 * @returns {Promise<Object>} список матчей
 */
export const getTournamentMatches = async (tournamentId) => {
  try {
    const response = await apiRequest.get(`${API_ENDPOINTS.TOURNAMENTS.GET(tournamentId)}/matches`)

    return {
      success: true,
      matches: response.data.data || response.data
    }
  } catch (error) {
    return {
      success: false,
      error: error.message || 'Ошибка получения матчей турнира'
    }
  }
}

/**
 * Получение турнирной сетки
 * @param {number} tournamentId - ID турнира
 * @returns {Promise<Object>} турнирная сетка
 */
export const getTournamentBracket = async (tournamentId) => {
  try {
    const response = await apiRequest.get(`${API_ENDPOINTS.TOURNAMENTS.GET(tournamentId)}/bracket`)

    return {
      success: true,
      bracket: response.data.data || response.data
    }
  } catch (error) {
    return {
      success: false,
      error: error.message || 'Ошибка получения турнирной сетки'
    }
  }
}

/**
 * Получение участников турнира
 * @param {number} tournamentId - ID турнира
 * @returns {Promise<Object>} список участников
 */
export const getTournamentParticipants = async (tournamentId) => {
  try {
    const response = await apiRequest.get(`${API_ENDPOINTS.TOURNAMENTS.GET(tournamentId)}/participants`)

    return {
      success: true,
      participants: response.data.data || response.data
    }
  } catch (error) {
    return {
      success: false,
      error: error.message || 'Ошибка получения участников турнира'
    }
  }
}

/**
 * Получение результатов турнира
 * @param {number} tournamentId - ID турнира
 * @returns {Promise<Object>} результаты турнира
 */
export const getTournamentResults = async (tournamentId) => {
  try {
    const response = await apiRequest.get(`${API_ENDPOINTS.TOURNAMENTS.GET(tournamentId)}/results`)

    return {
      success: true,
      results: response.data.data || response.data
    }
  } catch (error) {
    return {
      success: false,
      error: error.message || 'Ошибка получения результатов турнира'
    }
  }
}

/**
 * Получение истории турниров пользователя
 * @param {number} [userId] - ID пользователя (если не указан, то текущий)
 * @param {Object} params - параметры запроса
 * @returns {Promise<Object>} история турниров
 */
export const getUserTournamentHistory = async (userId = null, params = {}) => {
  try {
    const endpoint = userId 
      ? `/api/v1/users/${userId}/tournaments`
      : '/api/v1/users/profile/tournaments'

    const queryParams = createPaginationParams(params.page, params.limit)

    const response = await apiRequest.get(endpoint, {
      params: queryParams
    })

    return {
      success: true,
      tournaments: response.data.data || response.data,
      pagination: response.data.pagination || null
    }
  } catch (error) {
    return {
      success: false,
      error: error.message || 'Ошибка получения истории турниров'
    }
  }
}

/**
 * Отмена турнира (только для админов)
 * @param {number} tournamentId - ID турнира
 * @param {string} reason - причина отмены
 * @returns {Promise<Object>} результат отмены
 */
export const cancelTournament = async (tournamentId, reason = '') => {
  try {
    const response = await apiRequest.post(`${API_ENDPOINTS.TOURNAMENTS.GET(tournamentId)}/cancel`, {
      reason
    })

    return {
      success: true,
      message: response.data.message || 'Турнир отменен'
    }
  } catch (error) {
    return {
      success: false,
      error: error.message || 'Ошибка отмены турнира'
    }
  }
}

/**
 * Получение статистики турниров
 * @param {Object} params - параметры запроса
 * @param {string} [params.period] - период (day, week, month, year)
 * @returns {Promise<Object>} статистика турниров
 */
export const getTournamentStats = async (params = {}) => {
  try {
    const queryParams = {}
    
    if (params.period) {
      queryParams.period = params.period
    }

    const response = await apiRequest.get('/api/v1/tournaments/stats', {
      params: queryParams
    })

    return {
      success: true,
      stats: response.data.data || response.data
    }
  } catch (error) {
    return {
      success: false,
      error: error.message || 'Ошибка получения статистики турниров'
    }
  }
}

// Экспорт всех методов
export const tournamentsAPI = {
  startTournament,
  getTournament,
  submitMatchResult,
  getTournaments,
  getTournamentMatches,
  getTournamentBracket,
  getTournamentParticipants,
  getTournamentResults,
  getUserTournamentHistory,
  cancelTournament,
  getTournamentStats
}

export default tournamentsAPI