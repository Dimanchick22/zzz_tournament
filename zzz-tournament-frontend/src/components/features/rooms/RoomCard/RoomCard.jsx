// RoomCard.jsx - Карточка комнаты
import { useState } from 'react'
import { Link } from 'react-router-dom'
import { formatDistanceToNow } from 'date-fns'
import { ru } from 'date-fns/locale'
import styles from './RoomCard.module.css'

export const RoomCard = ({ room, currentUser, onJoin }) => {
  const [showPasswordModal, setShowPasswordModal] = useState(false)
  const [password, setPassword] = useState('')
  const [joining, setJoining] = useState(false)

  // Check if user is in room
  const isInRoom = room.participants?.some(p => p.id === currentUser?.id)
  const isHost = room.host_id === currentUser?.id

  // Calculate room fill percentage
  const fillPercentage = (room.current_count / room.max_players) * 100

  // Get status info
  const getStatusInfo = () => {
    switch (room.status) {
      case 'waiting':
        return {
          text: 'Ожидание игроков',
          color: 'warning',
          icon: 'clock'
        }
      case 'in_progress':
        return {
          text: 'В процессе',
          color: 'success',
          icon: 'play'
        }
      case 'finished':
        return {
          text: 'Завершена',
          color: 'secondary',
          icon: 'check'
        }
      default:
        return {
          text: 'Неизвестно',
          color: 'secondary',
          icon: 'question'
        }
    }
  }

  const statusInfo = getStatusInfo()

  // Handle join
  const handleJoin = async () => {
    if (room.is_private && !showPasswordModal) {
      setShowPasswordModal(true)
      return
    }

    setJoining(true)
    try {
      await onJoin(room.id, password)
      setPassword('')
      setShowPasswordModal(false)
    } finally {
      setJoining(false)
    }
  }

  // Handle password submit
  const handlePasswordSubmit = (e) => {
    e.preventDefault()
    handleJoin()
  }

  return (
    <div className={`${styles.roomCard} ${isInRoom ? styles.inRoom : ''}`}>
      {/* Header */}
      <div className={styles.header}>
        <div className={styles.roomInfo}>
          <h3 className={styles.roomName}>
            {room.is_private && <i className="fas fa-lock" />}
            {room.name}
          </h3>
          <p className={styles.roomDescription}>{room.description}</p>
        </div>
        
        <div className={`${styles.status} ${styles[statusInfo.color]}`}>
          <i className={`fas fa-${statusInfo.icon}`} />
          <span>{statusInfo.text}</span>
        </div>
      </div>

      {/* Participants */}
      <div className={styles.participants}>
        <div className={styles.participantCount}>
          <i className="fas fa-users" />
          <span>{room.current_count}/{room.max_players}</span>
        </div>
        
        <div className={styles.progressBar}>
          <div 
            className={styles.progressFill}
            style={{ width: `${fillPercentage}%` }}
          />
        </div>
        
        {room.participants && room.participants.length > 0 && (
          <div className={styles.participantList}>
            {room.participants.slice(0, 3).map(participant => (
              <div key={participant.id} className={styles.participant}>
                <div className={styles.avatar}>
                  {participant.avatar ? (
                    <img src={participant.avatar} alt={participant.username} />
                  ) : (
                    <i className="fas fa-user" />
                  )}
                </div>
                <span className={styles.username}>
                  {participant.username}
                  {participant.id === room.host_id && (
                    <i className="fas fa-crown" title="Хост" />
                  )}
                </span>
                <span className={styles.rating}>
                  {participant.rating}
                </span>
              </div>
            ))}
            {room.participants.length > 3 && (
              <div className={styles.moreParticipants}>
                +{room.participants.length - 3}
              </div>
            )}
          </div>
        )}
      </div>

      {/* Footer */}
      <div className={styles.footer}>
        <div className={styles.roomMeta}>
          <span className={styles.createdAt}>
            <i className="fas fa-clock" />
            {formatDistanceToNow(new Date(room.created_at), {
              addSuffix: true,
              locale: ru
            })}
          </span>
          
          {room.host && (
            <span className={styles.host}>
              <i className="fas fa-crown" />
              {room.host.username}
            </span>
          )}
        </div>

        <div className={styles.actions}>
          {isInRoom ? (
            <Link 
              to={`/rooms/${room.id}`}
              className={`${styles.actionButton} ${styles.primary}`}
            >
              <i className="fas fa-door-open" />
              Войти
            </Link>
          ) : room.status === 'waiting' && room.current_count < room.max_players ? (
            <button 
              className={`${styles.actionButton} ${styles.primary}`}
              onClick={handleJoin}
              disabled={joining}
            >
              {joining ? (
                <>
                  <div className={styles.spinner} />
                  Присоединение...
                </>
              ) : (
                <>
                  <i className="fas fa-plus" />
                  Присоединиться
                </>
              )}
            </button>
          ) : room.status === 'in_progress' ? (
            <Link 
              to={`/rooms/${room.id}`}
              className={`${styles.actionButton} ${styles.secondary}`}
            >
              <i className="fas fa-eye" />
              Наблюдать
            </Link>
          ) : (
            <button 
              className={`${styles.actionButton} ${styles.disabled}`}
              disabled
            >
              <i className="fas fa-ban" />
              Недоступно
            </button>
          )}
        </div>
      </div>

      {/* Password Modal */}
      {showPasswordModal && (
        <div className={styles.modalOverlay}>
          <div className={styles.modal}>
            <div className={styles.modalHeader}>
              <h3>Введите пароль</h3>
              <button 
                className={styles.closeButton}
                onClick={() => setShowPasswordModal(false)}
              >
                <i className="fas fa-times" />
              </button>
            </div>
            
            <form onSubmit={handlePasswordSubmit} className={styles.modalForm}>
              <div className={styles.field}>
                <label htmlFor="password">Пароль комнаты</label>
                <input
                  type="password"
                  id="password"
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
                  className={styles.cancelButton}
                  onClick={() => setShowPasswordModal(false)}
                >
                  Отмена
                </button>
                <button 
                  type="submit"
                  className={styles.submitButton}
                  disabled={joining || !password.trim()}
                >
                  {joining ? 'Присоединение...' : 'Присоединиться'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}