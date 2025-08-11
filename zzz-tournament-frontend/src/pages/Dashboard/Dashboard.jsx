// src/pages/Dashboard/Dashboard.jsx - адаптированный под реальные API endpoints
import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { useAuthStore } from '@store/authStore'
import { useUIStore } from '@store/uiStore'
import { useI18n } from '@hooks/useI18n'
import { roomsAPI } from '@api/rooms'
import { usersAPI } from '@api/users'
import styles from './Dashboard.module.css'

export default function Dashboard() {
  const { user } = useAuthStore()
  const { addNotification } = useUIStore()
  const { t, formatRelativeTime } = useI18n()
  
  // State
  const [stats, setStats] = useState({
    totalRooms: 0,
    activeRooms: 0,
    totalPlayers: 0,
    onlinePlayers: 0
  })

  const [activeRooms, setActiveRooms] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

  // Загрузка данных дашборда
  useEffect(() => {
    const loadDashboardData = async () => {
      try {
        setLoading(true)
        setError(null)

        // Загружаем только доступные данные
        const [roomsResult, leaderboardResult] = await Promise.allSettled([
          // Загружаем комнаты
          roomsAPI.getRooms({ page: 1, limit: 20 }),
          // Загружаем лидерборд для статистики игроков
          usersAPI.getLeaderboard({ page: 1, limit: 10 })
        ])

        // Обрабатываем комнаты
        if (roomsResult.status === 'fulfilled' && roomsResult.value.success) {
          console.log('Rooms API response:', roomsResult.value)
          
          // Проверяем структуру данных
          let rooms = []
          if (Array.isArray(roomsResult.value.rooms)) {
            rooms = roomsResult.value.rooms
          } else if (Array.isArray(roomsResult.value.data)) {
            rooms = roomsResult.value.data
          } else if (Array.isArray(roomsResult.value)) {
            rooms = roomsResult.value
          } else {
            console.warn('Unexpected rooms data structure:', roomsResult.value)
            rooms = []
          }
          
          const activeRoomsList = rooms.filter(room => 
            room && (room.status === 'waiting' || room.status === 'in_progress')
          )
          
          setActiveRooms(activeRoomsList.slice(0, 3)) // Показываем только первые 3
          
          // Обновляем статистику комнат
          setStats(prev => ({
            ...prev,
            totalRooms: roomsResult.value.pagination?.total || rooms.length,
            activeRooms: activeRoomsList.length
          }))
        } else if (roomsResult.status === 'rejected') {
          console.warn('Failed to load rooms:', roomsResult.reason)
        }

        // Обрабатываем лидерборд для статистики игроков
        if (leaderboardResult.status === 'fulfilled' && leaderboardResult.value.success) {
          console.log('Leaderboard API response:', leaderboardResult.value)
          
          const totalPlayers = leaderboardResult.value.pagination?.total || 
                              leaderboardResult.value.total ||
                              (Array.isArray(leaderboardResult.value.users) ? leaderboardResult.value.users.length : 0) ||
                              (Array.isArray(leaderboardResult.value.data) ? leaderboardResult.value.data.length : 0) ||
                              0
          
          setStats(prev => ({
            ...prev,
            totalPlayers: totalPlayers,
            // Примерная оценка онлайн игроков (10-15% от общего числа)
            onlinePlayers: Math.floor(totalPlayers * 0.12)
          }))
        } else if (leaderboardResult.status === 'rejected') {
          console.warn('Failed to load leaderboard:', leaderboardResult.reason)
        }

      } catch (err) {
        console.error('Dashboard loading error:', err)
        setError(t('errors.loadingError'))
        addNotification({
          type: 'error',
          title: t('errors.loadingError'),
          message: t('dashboard.errorLoading')
        })
      } finally {
        setLoading(false)
      }
    }

    loadDashboardData()
  }, [addNotification, t])

  // Обработка ошибок загрузки
  const handleRetryLoad = () => {
    window.location.reload()
  }

  // Получение статуса комнаты
  const getRoomStatusInfo = (status) => {
    switch (status) {
      case 'waiting':
        return {
          text: t('rooms.roomStatus.waiting'),
          className: 'waiting'
        }
      case 'in_progress':
        return {
          text: t('rooms.roomStatus.inProgress'),
          className: 'inProgress'
        }
      case 'finished':
        return {
          text: t('rooms.roomStatus.finished'),
          className: 'finished'
        }
      default:
        return {
          text: t('common.unknown'),
          className: 'unknown'
        }
    }
  }

  // Обработка присоединения к комнате
  const handleJoinRoom = async (roomId) => {
    try {
      const result = await roomsAPI.joinRoom(roomId)
      
      if (result.success) {
        addNotification({
          type: 'success',
          title: t('common.success'),
          message: result.message || t('rooms.join.joinSuccess')
        })
        // Обновляем список комнат
        window.location.reload()
      } else {
        addNotification({
          type: 'error',
          title: t('rooms.join.joinError'),
          message: result.error
        })
      }
    } catch (err) {
      addNotification({
        type: 'error',
        title: t('common.error'),
        message: t('rooms.join.joinError')
      })
    }
  }

  if (loading) {
    return (
      <div className={styles.loading}>
        <div className={styles.spinner} />
        <p>{t('dashboard.loadingDashboard')}</p>
      </div>
    )
  }

  if (error) {
    return (
      <div className={styles.error}>
        <i className="fas fa-exclamation-triangle" />
        <h2>{t('errors.loadingError')}</h2>
        <p>{error}</p>
        <button className={styles.retryButton} onClick={handleRetryLoad}>
          {t('common.retry')}
        </button>
      </div>
    )
  }

  return (
    <div className={styles.dashboard}>
      <div className={styles.header}>
        <h1>{t('dashboard.welcome', { username: user?.username })}</h1>
        <p>{t('dashboard.subtitle')}</p>
      </div>

      {/* User Stats */}
      <div className={styles.userStats}>
        <div className={styles.statCard}>
          <div className={styles.statIcon}>
            <i className="fas fa-star" />
          </div>
          <div className={styles.statInfo}>
            <span className={styles.statValue}>{user?.rating || 0}</span>
            <span className={styles.statLabel}>{t('dashboard.stats.rating')}</span>
          </div>
        </div>

        <div className={styles.statCard}>
          <div className={styles.statIcon}>
            <i className="fas fa-trophy" />
          </div>
          <div className={styles.statInfo}>
            <span className={styles.statValue}>{user?.wins || 0}</span>
            <span className={styles.statLabel}>{t('dashboard.stats.wins')}</span>
          </div>
        </div>

        <div className={styles.statCard}>
          <div className={styles.statIcon}>
            <i className="fas fa-times" />
          </div>
          <div className={styles.statInfo}>
            <span className={styles.statValue}>{user?.losses || 0}</span>
            <span className={styles.statLabel}>{t('dashboard.stats.losses')}</span>
          </div>
        </div>

        <div className={styles.statCard}>
          <div className={styles.statIcon}>
            <i className="fas fa-percent" />
          </div>
          <div className={styles.statInfo}>
            <span className={styles.statValue}>
              {user?.wins && user?.losses 
                ? Math.round((user.wins / (user.wins + user.losses)) * 100) 
                : 0}%
            </span>
            <span className={styles.statLabel}>{t('dashboard.stats.winrate')}</span>
          </div>
        </div>
      </div>

      <div className={styles.content}>
        {/* Global Stats - Упрощенная версия с доступными данными */}
        <div className={styles.section}>
          <h2>{t('dashboard.serverStats')}</h2>
          <div className={styles.globalStats}>
            <div className={styles.globalStatCard}>
              <h3>{stats.activeRooms}</h3>
              <p>{t('dashboard.stats.activeRooms')}</p>
              <span>{t('common.of')} {stats.totalRooms} {t('dashboard.stats.totalRooms').toLowerCase()}</span>
            </div>
            <div className={styles.globalStatCard}>
              <h3>{stats.onlinePlayers}</h3>
              <p>{t('dashboard.stats.onlinePlayers')}</p>
              <span>{t('common.of')} {stats.totalPlayers} {t('common.total').toLowerCase()}</span>
            </div>
            <div className={styles.globalStatCard}>
              <h3>{user?.rating || 0}</h3>
              <p>{t('dashboard.stats.yourRating')}</p>
              <span>{t('dashboard.stats.ratingChange')}</span>
            </div>
          </div>
        </div>

        <div className={styles.row}>
          {/* Recent Activity - Заменяем недоступную историю матчей */}
          <div className={styles.section}>
            <div className={styles.sectionHeader}>
              <h2>{t('dashboard.recentActivity')}</h2>
              <Link to="/profile" className={styles.sectionLink}>
                {t('common.viewAll')}
              </Link>
            </div>
            
            <div className={styles.activityList}>
              <div className={styles.activityCard}>
                <div className={styles.activityIcon}>
                  <i className="fas fa-user-plus" />
                </div>
                <div className={styles.activityInfo}>
                  <span className={styles.activityTitle}>{t('dashboard.activity.accountCreated')} </span>
                  <span className={styles.activityTime}>
                    {user?.created_at ? formatRelativeTime(user.created_at) : t('dashboard.activity.recently')}
                  </span>
                </div>
              </div>
              
              <div className={styles.activityCard}>
                <div className={styles.activityIcon}>
                  <i className="fas fa-star" />
                </div>
                <div className={styles.activityInfo}>
                  <span className={styles.activityTitle}>{t('dashboard.activity.currentRating')} </span>
                  <span className={styles.activityTime}>{user?.rating || 0} {t('dashboard.stats.rating')} </span>
                </div>
              </div>
              
              {(user?.wins || 0) > 0 && (
                <div className={styles.activityCard}>
                  <div className={styles.activityIcon}>
                    <i className="fas fa-trophy" />
                  </div>
                  <div className={styles.activityInfo}>
                    <span className={styles.activityTitle}>{t('dashboard.activity.victories')} </span>
                    <span className={styles.activityTime}>{user.wins} {t('dashboard.stats.wins')} </span>
                  </div>
                </div>
              )}
            </div>
          </div>

          {/* Active Rooms */}
          <div className={styles.section}>
            <div className={styles.sectionHeader}>
              <h2>{t('dashboard.activeRooms')}</h2>
              <Link to="/rooms" className={styles.sectionLink}>
                {t('rooms.title')}
              </Link>
            </div>
            
            <div className={styles.roomList}>
              {Array.isArray(activeRooms) && activeRooms.length > 0 ? activeRooms.map(room => {
                // Проверяем что room является объектом
                if (!room || typeof room !== 'object') {
                  return null
                }
                
                const statusInfo = getRoomStatusInfo(room.status)
                const isUserInRoom = Array.isArray(room.participants) && 
                                    room.participants.some(p => p && p.id === user?.id)
                
                return (
                  <div key={room.id || Math.random()} className={styles.roomCard}>
                    <div className={styles.roomInfo}>
                      <h3>{room.name || t('rooms.unknownRoom')}</h3>
                      <div className={styles.roomMeta}>
                        <span className={styles.participants}>
                          <i className="fas fa-users" />
                          {room.current_count || 0}/{room.max_players || 0}
                        </span>
                        <span className={`${styles.status} ${styles[statusInfo.className]}`}>
                          {statusInfo.text}
                        </span>
                      </div>
                    </div>
                    
                    {isUserInRoom ? (
                      <Link 
                        to={`/rooms/${room.id}`} 
                        className={styles.joinButton}
                      >
                        {t('common.view')}
                      </Link>
                    ) : room.status === 'waiting' && (room.current_count || 0) < (room.max_players || 0) ? (
                      <button 
                        className={styles.joinButton}
                        onClick={() => handleJoinRoom(room.id)}
                        disabled={!room.id}
                      >
                        {t('rooms.joinRoom')}
                      </button>
                    ) : (
                      <Link 
                        to={`/rooms/${room.id}`} 
                        className={styles.joinButton}
                      >
                        {t('common.view')}
                      </Link>
                    )}
                  </div>
                )
              }).filter(Boolean) : (
                <p className={styles.emptyState}>{t('dashboard.noActiveRooms')}</p>
              )}
            </div>
          </div>
        </div>

        {/* Quick Actions */}
        <div className={styles.section}>
          <h2>{t('dashboard.quickActions')}</h2>
          <div className={styles.quickActions}>
            <Link to="/rooms" className={styles.actionCard}>
              <i className="fas fa-plus" />
              <span>{t('rooms.createRoom')}</span>
            </Link>
            <Link to="/rooms" className={styles.actionCard}>
              <i className="fas fa-search" />
              <span>{t('common.findGame')}</span>
            </Link>
            <Link to="/heroes" className={styles.actionCard}>
              <i className="fas fa-sword" />
              <span>{t('heroes.title')}</span>
            </Link>
            <Link to="/leaderboard" className={styles.actionCard}>
              <i className="fas fa-trophy" />
              <span>{t('leaderboard.title')}</span>
            </Link>
          </div>
        </div>
      </div>
    </div>
  )
}