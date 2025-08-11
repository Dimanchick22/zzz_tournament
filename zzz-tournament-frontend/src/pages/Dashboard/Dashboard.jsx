// src/pages/Dashboard/Dashboard.jsx - обновленная версия с переводами
import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { useAuthStore } from '@store/authStore'
import { useI18n } from '@hooks/useI18n'
import styles from './Dashboard.module.css'

export default function Dashboard() {
  const { user } = useAuthStore()
  const { t, formatRelativeTime } = useI18n()
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
        {/* Global Stats */}
        <div className={styles.section}>
          <h2>{t('dashboard.serverStats')}</h2>
          <div className={styles.globalStats}>
            <div className={styles.globalStatCard}>
              <h3>{stats.activeRooms}</h3>
              <p>{t('dashboard.stats.activeRooms')}</p>
              <span>{t('common.of')} {stats.totalRooms} {t('dashboard.stats.totalRooms').toLowerCase()}</span>
            </div>
            <div className={styles.globalStatCard}>
              <h3>{stats.activeTournaments}</h3>
              <p>{t('dashboard.stats.activeTournaments')}</p>
              <span>{t('common.of')} {stats.totalTournaments} {t('common.total').toLowerCase()}</span>
            </div>
            <div className={styles.globalStatCard}>
              <h3>{stats.onlinePlayers}</h3>
              <p>{t('dashboard.stats.onlinePlayers')}</p>
              <span>{t('common.of')} {stats.totalPlayers} {t('common.total').toLowerCase()}</span>
            </div>
          </div>
        </div>

        <div className={styles.row}>
          {/* Recent Matches */}
          <div className={styles.section}>
            <div className={styles.sectionHeader}>
              <h2>{t('dashboard.recentMatches')}</h2>
              <Link to="/leaderboard" className={styles.sectionLink}>
                {t('common.viewAll')}
              </Link>
            </div>
            
            <div className={styles.matchList}>
              {recentMatches.length > 0 ? recentMatches.map(match => (
                <div key={match.id} className={styles.matchCard}>
                  <div className={styles.matchPlayers}>
                    <span className={match.winner === match.player1 ? styles.winner : ''}>
                      {match.player1}
                    </span>
                    <span className={styles.vs}>{t('tournaments.match.vs')}</span>
                    <span className={match.winner === match.player2 ? styles.winner : ''}>
                      {match.player2}
                    </span>
                  </div>
                  <div className={styles.matchInfo}>
                    <span className={styles.tournament}>{match.tournament}</span>
                    <span className={styles.time}>{formatRelativeTime(match.date)}</span>
                  </div>
                </div>
              )) : (
                <p className={styles.emptyState}>{t('dashboard.noRecentMatches')}</p>
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
              {activeRooms.length > 0 ? activeRooms.map(room => (
                <div key={room.id} className={styles.roomCard}>
                  <div className={styles.roomInfo}>
                    <h3>{room.name}</h3>
                    <div className={styles.roomMeta}>
                      <span className={styles.participants}>
                        <i className="fas fa-users" />
                        {room.participants}/{room.maxParticipants}
                      </span>
                      <span className={`${styles.status} ${styles[room.status]}`}>
                        {t(`rooms.roomStatus.${room.status}`)}
                      </span>
                    </div>
                  </div>
                  <Link 
                    to={`/rooms/${room.id}`} 
                    className={styles.joinButton}
                  >
                    {room.status === 'waiting' ? t('rooms.joinRoom') : t('common.view')}
                  </Link>
                </div>
              )) : (
                <p className={styles.emptyState}>{t('dashboard.noActiveRooms')}</p>
              )}
            </div>
          </div>
        </div>

        {/* Quick Actions */}
        <div className={styles.section}>
          <h2>{t('dashboard.quickActions')}</h2>
          <div className={styles.quickActions}>
            <Link to="/rooms/create" className={styles.actionCard}>
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