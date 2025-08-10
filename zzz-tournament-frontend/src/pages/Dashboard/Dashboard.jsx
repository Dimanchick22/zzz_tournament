// Dashboard Page
import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { useAuthStore } from '@store/authStore'
import styles from './Dashboard.module.css'

export default function Dashboard() {
  const { user } = useAuthStore()
  const [stats, setStats] = useState({
    totalRooms: 0,
    activeRooms: 0,
    totalTournaments: 0,
    activeTournaments: 0,
    totalPlayers: 0,
    onlinePlayers: 0
  })

  const [recentMatches, setRecentMatches] = useState([])
  const [activeRooms, setActiveRooms] = useState([])

  // Имитация загрузки данных
  useEffect(() => {
    const loadDashboardData = async () => {
      // Имитация API запроса
      await new Promise(resolve => setTimeout(resolve, 1000))
      
      setStats({
        totalRooms: 24,
        activeRooms: 8,
        totalTournaments: 156,
        activeTournaments: 3,
        totalPlayers: 1247,
        onlinePlayers: 89
      })

      setRecentMatches([
        {
          id: 1,
          player1: 'Player123',
          player2: 'ProGamer',
          winner: 'Player123',
          tournament: 'Summer Championship',
          date: new Date(Date.now() - 2 * 60 * 60 * 1000)
        },
        {
          id: 2,
          player1: 'ZZZMaster',
          player2: 'ElitePlayer',
          winner: 'ElitePlayer',
          tournament: 'Weekly Cup',
          date: new Date(Date.now() - 5 * 60 * 60 * 1000)
        },
        {
          id: 3,
          player1: 'NinjaGamer',
          player2: 'SkillShot',
          winner: 'NinjaGamer',
          tournament: 'Rookie League',
          date: new Date(Date.now() - 8 * 60 * 60 * 1000)
        }
      ])

      setActiveRooms([
        {
          id: 1,
          name: 'Pro Tournament Room',
          participants: 8,
          maxParticipants: 16,
          status: 'waiting'
        },
        {
          id: 2,
          name: 'Casual Matches',
          participants: 4,
          maxParticipants: 8,
          status: 'in_progress'
        },
        {
          id: 3,
          name: 'Beginners Welcome',
          participants: 12,
          maxParticipants: 16,
          status: 'waiting'
        }
      ])
    }

    loadDashboardData()
  }, [])

  const formatTimeAgo = (date) => {
    const now = new Date()
    const diffInHours = Math.floor((now - date) / (1000 * 60 * 60))
    
    if (diffInHours < 1) {
      const diffInMinutes = Math.floor((now - date) / (1000 * 60))
      return `${diffInMinutes} мин. назад`
    }
    
    if (diffInHours < 24) {
      return `${diffInHours} ч. назад`
    }
    
    const diffInDays = Math.floor(diffInHours / 24)
    return `${diffInDays} дн. назад`
  }

  return (
    <div className={styles.dashboard}>
      <div className={styles.header}>
        <h1>Добро пожаловать, {user?.username}!</h1>
        <p>Вот что происходит в ZZZ Tournament сегодня</p>
      </div>

      {/* User Stats */}
      <div className={styles.userStats}>
        <div className={styles.statCard}>
          <div className={styles.statIcon}>
            <i className="fas fa-star" />
          </div>
          <div className={styles.statInfo}>
            <span className={styles.statValue}>{user?.rating || 0}</span>
            <span className={styles.statLabel}>Рейтинг</span>
          </div>
        </div>

        <div className={styles.statCard}>
          <div className={styles.statIcon}>
            <i className="fas fa-trophy" />
          </div>
          <div className={styles.statInfo}>
            <span className={styles.statValue}>{user?.wins || 0}</span>
            <span className={styles.statLabel}>Побед</span>
          </div>
        </div>

        <div className={styles.statCard}>
          <div className={styles.statIcon}>
            <i className="fas fa-times" />
          </div>
          <div className={styles.statInfo}>
            <span className={styles.statValue}>{user?.losses || 0}</span>
            <span className={styles.statLabel}>Поражений</span>
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
            <span className={styles.statLabel}>Винрейт</span>
          </div>
        </div>
      </div>

      <div className={styles.content}>
        {/* Global Stats */}
        <div className={styles.section}>
          <h2>Статистика сервера</h2>
          <div className={styles.globalStats}>
            <div className={styles.globalStatCard}>
              <h3>{stats.activeRooms}</h3>
              <p>Активных комнат</p>
              <span>из {stats.totalRooms} всего</span>
            </div>
            <div className={styles.globalStatCard}>
              <h3>{stats.activeTournaments}</h3>
              <p>Активных турниров</p>
              <span>из {stats.totalTournaments} всего</span>
            </div>
            <div className={styles.globalStatCard}>
              <h3>{stats.onlinePlayers}</h3>
              <p>Игроков онлайн</p>
              <span>из {stats.totalPlayers} всего</span>
            </div>
          </div>
        </div>

        <div className={styles.row}>
          {/* Recent Matches */}
          <div className={styles.section}>
            <div className={styles.sectionHeader}>
              <h2>Последние матчи</h2>
              <Link to="/leaderboard" className={styles.sectionLink}>
                Все результаты
              </Link>
            </div>
            
            <div className={styles.matchList}>
              {recentMatches.map(match => (
                <div key={match.id} className={styles.matchCard}>
                  <div className={styles.matchPlayers}>
                    <span className={match.winner === match.player1 ? styles.winner : ''}>
                      {match.player1}
                    </span>
                    <span className={styles.vs}>vs</span>
                    <span className={match.winner === match.player2 ? styles.winner : ''}>
                      {match.player2}
                    </span>
                  </div>
                  <div className={styles.matchInfo}>
                    <span className={styles.tournament}>{match.tournament}</span>
                    <span className={styles.time}>{formatTimeAgo(match.date)}</span>
                  </div>
                </div>
              ))}
            </div>
          </div>

          {/* Active Rooms */}
          <div className={styles.section}>
            <div className={styles.sectionHeader}>
              <h2>Активные комнаты</h2>
              <Link to="/rooms" className={styles.sectionLink}>
                Все комнаты
              </Link>
            </div>
            
            <div className={styles.roomList}>
              {activeRooms.map(room => (
                <div key={room.id} className={styles.roomCard}>
                  <div className={styles.roomInfo}>
                    <h3>{room.name}</h3>
                    <div className={styles.roomMeta}>
                      <span className={styles.participants}>
                        <i className="fas fa-users" />
                        {room.participants}/{room.maxParticipants}
                      </span>
                      <span className={`${styles.status} ${styles[room.status]}`}>
                        {room.status === 'waiting' ? 'Ожидание' : 'В процессе'}
                      </span>
                    </div>
                  </div>
                  <Link 
                    to={`/rooms/${room.id}`} 
                    className={styles.joinButton}
                  >
                    {room.status === 'waiting' ? 'Присоединиться' : 'Смотреть'}
                  </Link>
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* Quick Actions */}
        <div className={styles.section}>
          <h2>Быстрые действия</h2>
          <div className={styles.quickActions}>
            <Link to="/rooms/create" className={styles.actionCard}>
              <i className="fas fa-plus" />
              <span>Создать комнату</span>
            </Link>
            <Link to="/rooms" className={styles.actionCard}>
              <i className="fas fa-search" />
              <span>Найти игру</span>
            </Link>
            <Link to="/heroes" className={styles.actionCard}>
              <i className="fas fa-sword" />
              <span>Изучить героев</span>
            </Link>
            <Link to="/leaderboard" className={styles.actionCard}>
              <i className="fas fa-trophy" />
              <span>Рейтинг</span>
            </Link>
          </div>
        </div>
      </div>
    </div>
  )
}