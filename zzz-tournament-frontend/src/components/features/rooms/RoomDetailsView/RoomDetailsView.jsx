// src/components/features/rooms/RoomDetailsView/RoomDetailsView.jsx
import { useState, useEffect } from 'react'
import { useParams, useNavigate, Link } from 'react-router-dom'
import { formatDistanceToNow } from 'date-fns'
import { ru } from 'date-fns/locale'
import { useAuthStore } from '@store/authStore'
import { useUIStore } from '@store/uiStore'
import { roomsAPI } from '@api/rooms'
import { tournamentsAPI } from '@api/tournaments'
import { wsClient } from '@api/websocket'
import styles from './RoomDetailsView.module.css'

export const RoomDetailsView = () => {
  const { id } = useParams()
  const navigate = useNavigate()
  const { user } = useAuthStore()
  const { addNotification } = useUIStore()

  // State
  const [room, setRoom] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [actionLoading, setActionLoading] = useState(false)
  const [showPasswordModal, setShowPasswordModal] = useState(false)
  const [password, setPassword] = useState('')

  // WebSocket connection status
  const [wsConnected, setWsConnected] = useState(false)

  // Load room data
  const loadRoom = async () => {
    try {
      setLoading(true)
      setError(null)

      const result = await roomsAPI.getRoom(id)

      if (result.success) {
        setRoom(result.room)
      } else {
        setError(result.error)
        addNotification({
          type: 'error',
          title: 'Ошибка загрузки',
          message: result.error
        })
      }
    } catch (err) {
      setError('Не удалось загрузить комнату')
      addNotification({
        type: 'error',
        title: 'Ошибка',
        message: 'Не удалось загрузить комнату'
      })
    } finally {
      setLoading(false)
    }
  }

  // Initial load and WebSocket setup
  useEffect(() => {
    loadRoom()

    // Connect to WebSocket
    const connectWS = async () => {
      try {
        await wsClient.connect()
        setWsConnected(true)
        
        // Join room for real-time updates
        wsClient.joinRoom(parseInt(id))

        // Listen for room updates
        wsClient.on('room_updated', (data) => {
          if (data.room_id === parseInt(id)) {
            setRoom(prev => ({ ...prev, ...data.room }))
          }
        })

        wsClient.on('user_joined', (data) => {
          if (data.room_id === parseInt(id)) {
            setRoom(prev => ({
              ...prev,
              participants: [...(prev.participants || []), data.user],
              current_count: (prev.current_count || 0) + 1
            }))
            
            addNotification({
              type: 'info',
              title: 'Игрок присоединился',
              message: `${data.user.username} присоединился к комнате`
            })
          }
        })

        wsClient.on('user_left', (data) => {
          if (data.room_id === parseInt(id)) {
            setRoom(prev => ({
              ...prev,
              participants: (prev.participants || []).filter(p => p.id !== data.user_id),
              current_count: Math.max((prev.current_count || 0) - 1, 0)
            }))
            
            addNotification({
              type: 'info',
              title: 'Игрок покинул комнату',
              message: `${data.username} покинул комнату`
            })
          }
        })

        wsClient.on('tournament_started', (data) => {
          if (data.room_id === parseInt(id)) {
            setRoom(prev => ({ ...prev, status: 'in_progress' }))
            
            addNotification({
              type: 'success',
              title: 'Турнир начался!',
              message: 'Турнир в этой комнате был запущен'
            })
            
            // Redirect to tournament page
            navigate(`/tournament/${data.tournament_id}`)
          }
        })

      } catch (error) {
        console.error('WebSocket connection failed:', error)
        setWsConnected(false)
      }
    }

    connectWS()

    // Cleanup
    return () => {
      if (wsConnected) {
        wsClient.leaveRoom(parseInt(id))
      }
    }
  }, [id, navigate, addNotification])

  // Check if user is in room
  const isInRoom = room?.participants?.some(p => p.id === user?.id)
  const isHost = room?.host_id === user?.id
  const canStartTournament = isHost && room?.current_count >= 2 && room?.status === 'waiting'

  // Join room
  const handleJoinRoom = async () => {
    if (room?.is_private && !showPasswordModal) {
      setShowPasswordModal(true)
      return
    }

    setActionLoading(true)
    try {
      const result = await roomsAPI.joinRoom(id, password)
      
      if (result.success) {
        setPassword('')
        setShowPasswordModal(false)
        loadRoom() // Reload room data
        
        addNotification({
          type: 'success',
          title: 'Успешно!',
          message: result.message
        })
      } else {
        addNotification({
          type: 'error',
          title: 'Ошибка присоединения',
          message: result.error
        })
      }
    } catch (err) {
      addNotification({
        type: 'error',
        title: 'Ошибка',
        message: 'Не удалось присоединиться к комнате'
      })
    } finally {
      setActionLoading(false)
    }
  }

  // Leave room
  const handleLeaveRoom = async () => {
    setActionLoading(true)
    try {
      const result = await roomsAPI.leaveRoom(id)
      
      if (result.success) {
        addNotification({
          type: 'info',
          title: 'Покинули комнату',
          message: result.message
        })
        
        navigate('/rooms')
      } else {
        addNotification({
          type: 'error',
          title: 'Ошибка',
          message: result.error
        })
      }
    } catch (err) {
      addNotification({
        type: 'error',
        title: 'Ошибка',
        message: 'Не удалось покинуть комнату'
      })
    } finally {
      setActionLoading(false)
    }
  }

  // Start tournament
  const handleStartTournament = async () => {
    setActionLoading(true)
    try {
      const result = await tournamentsAPI.startTournament(id)
      
      if (result.success) {
        addNotification({
          type: 'success',
          title: 'Турнир запущен!',
          message: 'Турнир успешно создан'
        })
        
        navigate(`/tournament/${result.tournament.id}`)
      } else {
        addNotification({
          type: 'error',
          title: 'Ошибка запуска',
          message: result.error
        })
      }
    } catch (err) {
      addNotification({
        type: 'error',
        title: 'Ошибка',
        message: 'Не удалось запустить турнир'
      })
    } finally {
      setActionLoading(false)
    }
  }

  // Kick player (host only)
  const handleKickPlayer = async (playerId) => {
    try {
      const result = await roomsAPI.kickPlayer(id, playerId)
      
      if (result.success) {
        loadRoom()
        addNotification({
          type: 'info',
          title: 'Игрок исключен',
          message: result.message
        })
      } else {
        addNotification({
          type: 'error',
          title: 'Ошибка',
          message: result.error
        })
      }
    } catch (err) {
      addNotification({
        type: 'error',
        title: 'Ошибка',
        message: 'Не удалось исключить игрока'
      })
    }
  }

  if (loading) {
    return (
      <div className={styles.loading}>
        <div className={styles.spinner} />
        <p>Загружаем комнату...</p>
      </div>
    )
  }

  if (error || !room) {
    return (
      <div className={styles.error}>
        <i className="fas fa-exclamation-triangle" />
        <h2>Комната не найдена</h2>
        <p>{error || 'Запрашиваемая комната не существует'}</p>
        <Link to="/rooms" className={styles.backButton}>
          <i className="fas fa-arrow-left" />
          Вернуться к комнатам
        </Link>
      </div>
    )
  }

  return (
    <div className={styles.roomDetails}>
      {/* Header */}
      <div className={styles.header}>
        <div className={styles.breadcrumb}>
          <Link to="/rooms">Комнаты</Link>
          <i className="fas fa-chevron-right" />
          <span>{room.name}</span>
        </div>

        <div className={styles.connectionStatus}>
          <span className={`${styles.statusDot} ${wsConnected ? styles.connected : styles.disconnected}`} />
          {wsConnected ? 'Подключено' : 'Не подключено'}
        </div>
      </div>

      {/* Room Info */}
      <div className={styles.roomInfo}>
        <div className={styles.roomHeader}>
          <div className={styles.roomTitle}>
            <h1>
              {room.is_private && <i className="fas fa-lock" />}
              {room.name}
            </h1>
            <div className={`${styles.status} ${styles[room.status]}`}>
              <i className={`fas fa-${getStatusIcon(room.status)}`} />
              {getStatusLabel(room.status)}
            </div>
          </div>

          {room.description && (
            <p className={styles.roomDescription}>{room.description}</p>
          )}

          <div className={styles.roomMeta}>
            <div className={styles.metaItem}>
              <i className="fas fa-users" />
              <span>{room.current_count}/{room.max_players} игроков</span>
            </div>
            <div className={styles.metaItem}>
              <i className="fas fa-clock" />
              <span>
                Создана {formatDistanceToNow(new Date(room.created_at), {
                  addSuffix: true,
                  locale: ru
                })}
              </span>
            </div>
            {room.host && (
              <div className={styles.metaItem}>
                <i className="fas fa-crown" />
                <span>Хост: {room.host.username}</span>
              </div>
            )}
          </div>

          <div className={styles.progressContainer}>
            <div className={styles.progressBar}>
              <div 
                className={styles.progressFill}
                style={{ width: `${(room.current_count / room.max_players) * 100}%` }}
              />
            </div>
          </div>
        </div>

        <div className={styles.roomActions}>
          {!isInRoom ? (
            <button 
              className={styles.joinButton}
              onClick={handleJoinRoom}
              disabled={actionLoading || room.status !== 'waiting' || room.current_count >= room.max_players}
            >
              {actionLoading ? (
                <>
                  <div className={styles.actionSpinner} />
                  Присоединение...
                </>
              ) : (
                <>
                  <i className="fas fa-plus" />
                  Присоединиться
                </>
              )}
            </button>
          ) : (
            <div className={styles.participantActions}>
              <button 
                className={styles.leaveButton}
                onClick={handleLeaveRoom}
                disabled={actionLoading}
              >
                {actionLoading ? (
                  <>
                    <div className={styles.actionSpinner} />
                    Выход...
                  </>
                ) : (
                  <>
                    <i className="fas fa-sign-out-alt" />
                    Покинуть комнату
                  </>
                )}
              </button>

              {canStartTournament && (
                <button 
                  className={styles.startTournamentButton}
                  onClick={handleStartTournament}
                  disabled={actionLoading}
                >
                  {actionLoading ? (
                    <>
                      <div className={styles.actionSpinner} />
                      Запуск...
                    </>
                  ) : (
                    <>
                      <i className="fas fa-play" />
                      Запустить турнир
                    </>
                  )}
                </button>
              )}
            </div>
          )}
        </div>
      </div>

      {/* Participants */}
      <div className={styles.participants}>
        <h2>
          <i className="fas fa-users" />
          Участники ({room.current_count}/{room.max_players})
        </h2>

        {room.participants && room.participants.length > 0 ? (
          <div className={styles.participantsList}>
            {room.participants.map(participant => (
              <div key={participant.id} className={styles.participantCard}>
                <div className={styles.participantInfo}>
                  <div className={styles.participantAvatar}>
                    {participant.avatar ? (
                      <img src={participant.avatar} alt={participant.username} />
                    ) : (
                      <i className="fas fa-user" />
                    )}
                  </div>
                  
                  <div className={styles.participantDetails}>
                    <div className={styles.participantName}>
                      {participant.username}
                      {participant.id === room.host_id && (
                        <span className={styles.hostBadge}>
                          <i className="fas fa-crown" />
                          Хост
                        </span>
                      )}
                    </div>
                    <div className={styles.participantStats}>
                      <span className={styles.rating}>
                        <i className="fas fa-star" />
                        {participant.rating || 0}
                      </span>
                      <span className={styles.winRate}>
                        Винрейт: {participant.wins && participant.losses 
                          ? Math.round((participant.wins / (participant.wins + participant.losses)) * 100) 
                          : 0}%
                      </span>
                    </div>
                  </div>
                </div>

                {isHost && participant.id !== user?.id && (
                  <div className={styles.participantActions}>
                    <button 
                      className={styles.kickButton}
                      onClick={() => handleKickPlayer(participant.id)}
                      title="Исключить игрока"
                    >
                      <i className="fas fa-times" />
                    </button>
                  </div>
                )}
              </div>
            ))}
          </div>
        ) : (
          <div className={styles.emptyParticipants}>
            <i className="fas fa-user-plus" />
            <p>Пока нет участников</p>
          </div>
        )}
      </div>

      {/* Tournament Status */}
      {room.status === 'in_progress' && room.tournament_id && (
        <div className={styles.tournamentStatus}>
          <div className={styles.tournamentInfo}>
            <i className="fas fa-trophy" />
            <div>
              <h3>Турнир в процессе</h3>
              <p>В этой комнате идет турнир</p>
            </div>
          </div>
          <Link 
            to={`/tournament/${room.tournament_id}`}
            className={styles.viewTournamentButton}
          >
            <i className="fas fa-eye" />
            Смотреть турнир
          </Link>
        </div>
      )}

      {/* Password Modal */}
      {showPasswordModal && (
        <div className={styles.modalOverlay}>
          <div className={styles.modal}>
            <div className={styles.modalHeader}>
              <h3>Введите пароль</h3>
              <button 
                className={styles.modalClose}
                onClick={() => setShowPasswordModal(false)}
              >
                <i className="fas fa-times" />
              </button>
            </div>
            
            <form onSubmit={(e) => { e.preventDefault(); handleJoinRoom(); }} className={styles.modalForm}>
              <div className={styles.modalField}>
                <label htmlFor="roomPassword">Пароль комнаты</label>
                <input
                  type="password"
                  id="roomPassword"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  placeholder="Введите пароль"
                  autoFocus
                  required
                />
              </div>
              
              <div className={styles.modalActions}>
                <button 
                  type="button"
                  className={styles.modalCancel}
                  onClick={() => setShowPasswordModal(false)}
                >
                  Отмена
                </button>
                <button 
                  type="submit"
                  className={styles.modalSubmit}
                  disabled={actionLoading || !password.trim()}
                >
                  {actionLoading ? 'Присоединение...' : 'Присоединиться'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}

// Helper functions
const getStatusIcon = (status) => {
  switch (status) {
    case 'waiting': return 'clock'
    case 'in_progress': return 'play'
    case 'finished': return 'check'
    default: return 'question'
  }
}

const getStatusLabel = (status) => {
  switch (status) {
    case 'waiting': return 'Ожидание игроков'
    case 'in_progress': return 'В процессе'
    case 'finished': return 'Завершена'
    default: return 'Неизвестно'
  }
}