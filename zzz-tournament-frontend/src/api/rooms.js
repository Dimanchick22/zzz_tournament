// Rooms API methods
import { apiRequest } from './client'
import { API_ENDPOINTS } from '@config/api'
import { createPaginationParams } from '@config/api'

/**
 * Получение списка комнат
 * @param {Object} params - параметры запроса
 * @param {number} [params.page=1] - номер страницы
 * @param {number} [params.limit=20] - количество записей
 * @param {string} [params.status] - фильтр по статусу
 * @param {string} [params.search] - поиск по названию
 * @returns {Promise<Object>} список комнат
 */
export const getRooms = async (params = {}) => {
  try {
    const queryParams = {
      ...createPaginationParams(params.page, params.limit)
    }

    if (params.status) {
      queryParams.status = params.status
    }

    if (params.search) {
      queryParams.search = params.search
    }

    const response = await apiRequest.get(API_ENDPOINTS.ROOMS.LIST, {
      params: queryParams
    })

    return {
      success: true,
      rooms: response.data.data || response.data,
      pagination: response.data.pagination || null
    }
  } catch (error) {
    return {
      success: false,
      error: error.message || 'Ошибка получения списка комнат'
    }
  }
}

/**
 * Получение информации о комнате
 * @param {number} roomId - ID комнаты
 * @returns {Promise<Object>} информация о комнате
 */
export const getRoom = async (roomId) => {
  try {
    const response = await apiRequest.get(API_ENDPOINTS.ROOMS.GET(roomId))

    return {
      success: true,
      room: response.data.data || response.data
    }
  } catch (error) {
    return {
      success: false,
      error: error.message || 'Ошибка получения информации о комнате'
    }
  }
}

/**
 * Создание новой комнаты
 * @param {Object} roomData - данные комнаты
 * @param {string} roomData.name - название комнаты
 * @param {string} [roomData.description] - описание комнаты
 * @param {number} roomData.max_players - максимальное количество игроков
 * @param {boolean} [roomData.is_private] - приватная ли комната
 * @param {string} [roomData.password] - пароль для приватной комнаты
 * @returns {Promise<Object>} результат создания
 */
export const createRoom = async (roomData) => {
  try {
    const response = await apiRequest.post(API_ENDPOINTS.ROOMS.CREATE, {
      name: roomData.name,
      description: roomData.description || '',
      max_players: roomData.max_players,
      is_private: roomData.is_private || false,
      password: roomData.password || ''
    })

    return {
      success: true,
      room: response.data.data || response.data,
      message: response.data.message || 'Комната создана успешно'
    }
  } catch (error) {
    return {
      success: false,
      error: error.message || 'Ошибка создания комнаты',
      details: error.details || []
    }
  }
}

/**
 * Обновление комнаты
 * @param {number} roomId - ID комнаты
 * @param {Object} roomData - данные для обновления
 * @returns {Promise<Object>} результат обновления
 */
export const updateRoom = async (roomId, roomData) => {
  try {
    const response = await apiRequest.put(API_ENDPOINTS.ROOMS.UPDATE(roomId), roomData)

    return {
      success: true,
      room: response.data.data || response.data,
      message: response.data.message || 'Комната обновлена'
    }
  } catch (error) {
    return {
      success: false,
      error: error.message || 'Ошибка обновления комнаты',
      details: error.details || []
    }
  }
}

/**
 * Удаление комнаты
 * @param {number} roomId - ID комнаты
 * @returns {Promise<Object>} результат удаления
 */
export const deleteRoom = async (roomId) => {
  try {
    const response = await apiRequest.delete(API_ENDPOINTS.ROOMS.DELETE(roomId))

    return {
      success: true,
      message: response.data.message || 'Комната удалена'
    }
  } catch (error) {
    return {
      success: false,
      error: error.message || 'Ошибка удаления комнаты'
    }
  }
}

/**
 * Присоединение к комнате
 * @param {number} roomId - ID комнаты
 * @param {string} [password] - пароль для приватной комнаты
 * @returns {Promise<Object>} результат присоединения
 */
export const joinRoom = async (roomId, password = '') => {
  try {
    const response = await apiRequest.post(API_ENDPOINTS.ROOMS.JOIN(roomId), {
      password
    })

    return {
      success: true,
      room: response.data.data || response.data,
      message: response.data.message || 'Вы присоединились к комнате'
    }
  } catch (error) {
    return {
      success: false,
      error: error.message || 'Ошибка присоединения к комнате'
    }
  }
}

/**
 * Покидание комнаты
 * @param {number} roomId - ID комнаты
 * @returns {Promise<Object>} результат покидания
 */
export const leaveRoom = async (roomId) => {
  try {
    const response = await apiRequest.post(API_ENDPOINTS.ROOMS.LEAVE(roomId))

    return {
      success: true,
      message: response.data.message || 'Вы покинули комнату'
    }
  } catch (error) {
    return {
      success: false,
      error: error.message || 'Ошибка покидания комнаты'
    }
  }
}

/**
 * Получение участников комнаты
 * @param {number} roomId - ID комнаты
 * @returns {Promise<Object>} список участников
 */
export const getRoomParticipants = async (roomId) => {
  try {
    const response = await apiRequest.get(`${API_ENDPOINTS.ROOMS.GET(roomId)}/participants`)

    return {
      success: true,
      participants: response.data.data || response.data
    }
  } catch (error) {
    return {
      success: false,
      error: error.message || 'Ошибка получения списка участников'
    }
  }
}

/**
 * Кик игрока из комнаты (только для хоста)
 * @param {number} roomId - ID комнаты
 * @param {number} userId - ID игрока для кика
 * @returns {Promise<Object>} результат кика
 */
export const kickPlayer = async (roomId, userId) => {
  try {
    const response = await apiRequest.post(`${API_ENDPOINTS.ROOMS.GET(roomId)}/kick`, {
      user_id: userId
    })

    return {
      success: true,
      message: response.data.message || 'Игрок исключен из комнаты'
    }
  } catch (error) {
    return {
      success: false,
      error: error.message || 'Ошибка исключения игрока'
    }
  }
}

/**
 * Передача права хоста другому игроку
 * @param {number} roomId - ID комнаты
 * @param {number} userId - ID нового хоста
 * @returns {Promise<Object>} результат передачи
 */
export const transferHost = async (roomId, userId) => {
  try {
    const response = await apiRequest.post(`${API_ENDPOINTS.ROOMS.GET(roomId)}/transfer-host`, {
      user_id: userId
    })

    return {
      success: true,
      message: response.data.message || 'Права хоста переданы'
    }
  } catch (error) {
    return {
      success: false,
      error: error.message || 'Ошибка передачи прав хоста'
    }
  }
}

// Экспорт всех методов
export const roomsAPI = {
  getRooms,
  getRoom,
  createRoom,
  updateRoom,
  deleteRoom,
  joinRoom,
  leaveRoom,
  getRoomParticipants,
  kickPlayer,
  transferHost
}

export default roomsAPI