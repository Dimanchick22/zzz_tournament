// WebSocket Client
import { getWsUrl } from '@config/env'
import { WS_CONFIG } from '@config/api'

class WebSocketClient {
  constructor() {
    this.ws = null
    this.isConnected = false
    this.isConnecting = false
    this.reconnectAttempts = 0
    this.listeners = new Map()
    this.heartbeatInterval = null
    this.reconnectTimeout = null
    this.messageQueue = []
  }

  /**
   * Подключение к WebSocket серверу
   * @param {string} [token] - JWT токен для аутентификации
   * @returns {Promise<boolean>} успешность подключения
   */
  async connect(token = null) {
    if (this.isConnected || this.isConnecting) {
      return true
    }

    return new Promise((resolve, reject) => {
      try {
        this.isConnecting = true
        
        // Получаем токен из localStorage если не передан
        if (!token) {
          const authData = localStorage.getItem('auth-storage')
          if (authData) {
            try {
              const parsedAuth = JSON.parse(authData)
              token = parsedAuth.state?.token
            } catch (error) {
              console.warn('Failed to parse auth token for WebSocket:', error)
            }
          }
        }

        // Формируем URL с токеном
        const wsUrl = token 
          ? `${getWsUrl('/ws')}?token=${token}`
          : getWsUrl('/ws')

        this.ws = new WebSocket(wsUrl)

        // Таймаут подключения
        const connectionTimeout = setTimeout(() => {
          if (this.ws.readyState !== WebSocket.OPEN) {
            this.ws.close()
            this.isConnecting = false
            reject(new Error('WebSocket connection timeout'))
          }
        }, WS_CONFIG.CONNECTION_TIMEOUT)

        this.ws.onopen = () => {
          clearTimeout(connectionTimeout)
          this.isConnected = true
          this.isConnecting = false
          this.reconnectAttempts = 0
          
          console.log('✅ WebSocket connected')
          
          // Запускаем heartbeat
          this.startHeartbeat()
          
          // Отправляем сообщения из очереди
          this.flushMessageQueue()
          
          // Уведомляем слушателей о подключении
          this.emit('connected')
          
          resolve(true)
        }

        this.ws.onclose = (event) => {
          clearTimeout(connectionTimeout)
          this.isConnected = false
          this.isConnecting = false
          this.stopHeartbeat()
          
          console.log('❌ WebSocket disconnected:', event.code, event.reason)
          
          // Уведомляем слушателей об отключении
          this.emit('disconnected', { code: event.code, reason: event.reason })
          
          // Автоматическое переподключение если не было явного закрытия
          if (event.code !== 1000 && this.reconnectAttempts < WS_CONFIG.MAX_RECONNECT_ATTEMPTS) {
            this.scheduleReconnect(token)
          }
          
          if (this.isConnecting) {
            reject(new Error(`WebSocket connection failed: ${event.code} ${event.reason}`))
          }
        }

        this.ws.onerror = (error) => {
          console.error('❌ WebSocket error:', error)
          this.emit('error', error)
          
          if (this.isConnecting) {
            clearTimeout(connectionTimeout)
            this.isConnecting = false
            reject(error)
          }
        }

        this.ws.onmessage = (event) => {
          try {
            const message = JSON.parse(event.data)
            this.handleMessage(message)
          } catch (error) {
            console.error('Failed to parse WebSocket message:', error, event.data)
          }
        }

      } catch (error) {
        this.isConnecting = false
        reject(error)
      }
    })
  }

  /**
   * Отключение от WebSocket сервера
   */
  disconnect() {
    if (this.reconnectTimeout) {
      clearTimeout(this.reconnectTimeout)
      this.reconnectTimeout = null
    }

    this.stopHeartbeat()

    if (this.ws) {
      this.ws.close(1000, 'Client disconnect')
      this.ws = null
    }

    this.isConnected = false
    this.isConnecting = false
    this.reconnectAttempts = 0
    this.messageQueue = []
  }

  /**
   * Отправка сообщения
   * @param {string} type - тип сообщения
   * @param {Object} data - данные сообщения
   */
  send(type, data = {}) {
    const message = {
      type,
      data,
      timestamp: Date.now()
    }

    if (this.isConnected && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(message))
    } else {
      // Добавляем в очередь если не подключены
      this.messageQueue.push(message)
      
      // Пытаемся подключиться если еще не подключены
      if (!this.isConnecting && !this.isConnected) {
        this.connect()
      }
    }
  }

  /**
   * Подписка на события
   * @param {string} event - название события
   * @param {Function} callback - коллбек функция
   */
  on(event, callback) {
    if (!this.listeners.has(event)) {
      this.listeners.set(event, [])
    }
    this.listeners.get(event).push(callback)
  }

  /**
   * Отписка от событий
   * @param {string} event - название события
   * @param {Function} [callback] - конкретный коллбек (если не указан, удаляются все)
   */
  off(event, callback = null) {
    if (!this.listeners.has(event)) return

    if (callback) {
      const callbacks = this.listeners.get(event)
      const index = callbacks.indexOf(callback)
      if (index > -1) {
        callbacks.splice(index, 1)
      }
    } else {
      this.listeners.set(event, [])
    }
  }

  /**
   * Эмит события для слушателей
   * @param {string} event - название события
   * @param {*} data - данные события
   */
  emit(event, data = null) {
    if (this.listeners.has(event)) {
      this.listeners.get(event).forEach(callback => {
        try {
          callback(data)
        } catch (error) {
          console.error(`Error in WebSocket event handler for ${event}:`, error)
        }
      })
    }
  }

  /**
   * Обработка входящих сообщений
   * @param {Object} message - сообщение от сервера
   */
  handleMessage(message) {
    const { type, data } = message

    // Обрабатываем системные сообщения
    if (type === WS_CONFIG.MESSAGE_TYPES.HEARTBEAT) {
      return // Игнорируем heartbeat ответы
    }

    // Эмитим событие для конкретного типа сообщения
    this.emit(type, data)
    
    // Эмитим общее событие 'message'
    this.emit('message', message)
  }

  /**
   * Запуск heartbeat для поддержания соединения
   */
  startHeartbeat() {
    this.stopHeartbeat()
    
    this.heartbeatInterval = setInterval(() => {
      if (this.isConnected) {
        this.send(WS_CONFIG.MESSAGE_TYPES.HEARTBEAT)
      }
    }, WS_CONFIG.HEARTBEAT_INTERVAL)
  }

  /**
   * Остановка heartbeat
   */
  stopHeartbeat() {
    if (this.heartbeatInterval) {
      clearInterval(this.heartbeatInterval)
      this.heartbeatInterval = null
    }
  }

  /**
   * Планирование переподключения
   * @param {string} token - JWT токен
   */
  scheduleReconnect(token) {
    if (this.reconnectTimeout) {
      clearTimeout(this.reconnectTimeout)
    }

    this.reconnectAttempts++
    const delay = Math.min(1000 * Math.pow(2, this.reconnectAttempts - 1), 30000) // Exponential backoff, max 30s

    console.log(`🔄 Scheduling WebSocket reconnect in ${delay}ms (attempt ${this.reconnectAttempts}/${WS_CONFIG.MAX_RECONNECT_ATTEMPTS})`)

    this.reconnectTimeout = setTimeout(() => {
      this.connect(token).catch(error => {
        console.error('WebSocket reconnect failed:', error)
      })
    }, delay)
  }

  /**
   * Отправка сообщений из очереди
   */
  flushMessageQueue() {
    while (this.messageQueue.length > 0 && this.isConnected) {
      const message = this.messageQueue.shift()
      this.ws.send(JSON.stringify(message))
    }
  }

  /**
   * Проверка состояния подключения
   */
  getConnectionState() {
    return {
      isConnected: this.isConnected,
      isConnecting: this.isConnecting,
      reconnectAttempts: this.reconnectAttempts,
      queuedMessages: this.messageQueue.length
    }
  }

  // === Удобные методы для отправки типичных сообщений ===

  /**
   * Присоединение к комнате
   * @param {number} roomId - ID комнаты
   */
  joinRoom(roomId) {
    this.send(WS_CONFIG.MESSAGE_TYPES.JOIN_ROOM, { room_id: roomId })
  }

  /**
   * Покидание комнаты
   * @param {number} roomId - ID комнаты
   */
  leaveRoom(roomId) {
    this.send(WS_CONFIG.MESSAGE_TYPES.LEAVE_ROOM, { room_id: roomId })
  }

  /**
   * Отправка сообщения в чат
   * @param {number} roomId - ID комнаты
   * @param {string} content - текст сообщения
   */
  sendChatMessage(roomId, content) {
    this.send(WS_CONFIG.MESSAGE_TYPES.CHAT_MESSAGE, {
      room_id: roomId,
      content: content
    })
  }
}

// Создаем глобальный экземпляр WebSocket клиента
const wsClient = new WebSocketClient()

// Экспортируем клиент и константы
export { wsClient, WS_CONFIG }
export default wsClient