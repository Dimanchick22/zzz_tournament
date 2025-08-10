// Heroes API methods
import { apiRequest } from './client'
import { API_ENDPOINTS } from '@config/api'
import { createPaginationParams } from '@config/api'

/**
 * Получение списка героев
 * @param {Object} params - параметры запроса
 * @param {number} [params.page=1] - номер страницы
 * @param {number} [params.limit=50] - количество записей
 * @param {string} [params.element] - фильтр по элементу
 * @param {string} [params.rarity] - фильтр по редкости
 * @param {string} [params.role] - фильтр по роли
 * @param {string} [params.search] - поиск по имени
 * @returns {Promise<Object>} список героев
 */
export const getHeroes = async (params = {}) => {
  try {
    const queryParams = {
      ...createPaginationParams(params.page, params.limit || 50)
    }

    // Добавляем фильтры если они есть
    if (params.element) queryParams.element = params.element
    if (params.rarity) queryParams.rarity = params.rarity
    if (params.role) queryParams.role = params.role
    if (params.search) queryParams.search = params.search

    const response = await apiRequest.get(API_ENDPOINTS.HEROES.LIST, {
      params: queryParams
    })

    return {
      success: true,
      heroes: response.data.data || response.data,
      pagination: response.data.pagination || null
    }
  } catch (error) {
    return {
      success: false,
      error: error.message || 'Ошибка получения списка героев'
    }
  }
}

/**
 * Получение информации о герое
 * @param {number} heroId - ID героя
 * @returns {Promise<Object>} информация о герое
 */
export const getHero = async (heroId) => {
  try {
    const response = await apiRequest.get(API_ENDPOINTS.HEROES.GET(heroId))

    return {
      success: true,
      hero: response.data.data || response.data
    }
  } catch (error) {
    return {
      success: false,
      error: error.message || 'Ошибка получения информации о герое'
    }
  }
}

/**
 * Создание нового героя (только для админов)
 * @param {Object} heroData - данные героя
 * @param {string} heroData.name - имя героя
 * @param {string} heroData.element - элемент (Physical, Fire, Ice, Electric, Ether)
 * @param {string} heroData.rarity - редкость (A, S)
 * @param {string} heroData.role - роль (Attack, Stun, Anomaly, Support, Defense)
 * @param {string} [heroData.description] - описание
 * @param {string} [heroData.image_url] - URL изображения
 * @returns {Promise<Object>} результат создания
 */
export const createHero = async (heroData) => {
  try {
    const response = await apiRequest.post(API_ENDPOINTS.HEROES.CREATE, {
      name: heroData.name,
      element: heroData.element,
      rarity: heroData.rarity,
      role: heroData.role,
      description: heroData.description || '',
      image_url: heroData.image_url || ''
    })

    return {
      success: true,
      hero: response.data.data || response.data,
      message: response.data.message || 'Герой создан успешно'
    }
  } catch (error) {
    return {
      success: false,
      error: error.message || 'Ошибка создания героя',
      details: error.details || []
    }
  }
}

/**
 * Обновление героя (только для админов)
 * @param {number} heroId - ID героя
 * @param {Object} heroData - данные для обновления
 * @returns {Promise<Object>} результат обновления
 */
export const updateHero = async (heroId, heroData) => {
  try {
    const response = await apiRequest.put(API_ENDPOINTS.HEROES.UPDATE(heroId), heroData)

    return {
      success: true,
      hero: response.data.data || response.data,
      message: response.data.message || 'Герой обновлен'
    }
  } catch (error) {
    return {
      success: false,
      error: error.message || 'Ошибка обновления героя',
      details: error.details || []
    }
  }
}

/**
 * Удаление героя (только для админов)
 * @param {number} heroId - ID героя
 * @returns {Promise<Object>} результат удаления
 */
export const deleteHero = async (heroId) => {
  try {
    const response = await apiRequest.delete(API_ENDPOINTS.HEROES.DELETE(heroId))

    return {
      success: true,
      message: response.data.message || 'Герой удален'
    }
  } catch (error) {
    return {
      success: false,
      error: error.message || 'Ошибка удаления героя'
    }
  }
}

/**
 * Получение статистики использования героев
 * @param {Object} params - параметры запроса
 * @param {string} [params.period] - период (day, week, month, year)
 * @returns {Promise<Object>} статистика героев
 */
export const getHeroStats = async (params = {}) => {
  try {
    const queryParams = {}
    
    if (params.period) {
      queryParams.period = params.period
    }

    const response = await apiRequest.get('/api/v1/heroes/stats', {
      params: queryParams
    })

    return {
      success: true,
      stats: response.data.data || response.data
    }
  } catch (error) {
    return {
      success: false,
      error: error.message || 'Ошибка получения статистики героев'
    }
  }
}

/**
 * Получение рейтинга героев (по винрейту, популярности)
 * @param {Object} params - параметры запроса
 * @param {string} [params.sortBy='winrate'] - сортировка (winrate, pickrate, banrate)
 * @param {number} [params.limit=20] - количество героев
 * @returns {Promise<Object>} рейтинг героев
 */
export const getHeroRankings = async (params = {}) => {
  try {
    const queryParams = {
      sort_by: params.sortBy || 'winrate',
      limit: params.limit || 20
    }

    const response = await apiRequest.get('/api/v1/heroes/rankings', {
      params: queryParams
    })

    return {
      success: true,
      rankings: response.data.data || response.data
    }
  } catch (error) {
    return {
      success: false,
      error: error.message || 'Ошибка получения рейтинга героев'
    }
  }
}

/**
 * Получение популярных сборок для героя
 * @param {number} heroId - ID героя
 * @returns {Promise<Object>} популярные сборки
 */
export const getHeroBuilds = async (heroId) => {
  try {
    const response = await apiRequest.get(`${API_ENDPOINTS.HEROES.GET(heroId)}/builds`)

    return {
      success: true,
      builds: response.data.data || response.data
    }
  } catch (error) {
    return {
      success: false,
      error: error.message || 'Ошибка получения сборок героя'
    }
  }
}

/**
 * Получение матчапов героя (против каких героев сильнее/слабее)
 * @param {number} heroId - ID героя
 * @returns {Promise<Object>} матчапы героя
 */
export const getHeroMatchups = async (heroId) => {
  try {
    const response = await apiRequest.get(`${API_ENDPOINTS.HEROES.GET(heroId)}/matchups`)

    return {
      success: true,
      matchups: response.data.data || response.data
    }
  } catch (error) {
    return {
      success: false,
      error: error.message || 'Ошибка получения матчапов героя'
    }
  }
}

// Константы для фильтрации
export const HERO_ELEMENTS = ['Physical', 'Fire', 'Ice', 'Electric', 'Ether']
export const HERO_RARITIES = ['A', 'S']
export const HERO_ROLES = ['Attack', 'Stun', 'Anomaly', 'Support', 'Defense']

// Экспорт всех методов
export const heroesAPI = {
  getHeroes,
  getHero,
  createHero,
  updateHero,
  deleteHero,
  getHeroStats,
  getHeroRankings,
  getHeroBuilds,
  getHeroMatchups,
  
  // Константы
  ELEMENTS: HERO_ELEMENTS,
  RARITIES: HERO_RARITIES,
  ROLES: HERO_ROLES
}

export default heroesAPI