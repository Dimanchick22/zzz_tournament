// Rooms Page - Полная реализация с API интеграцией
import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { useAuthStore } from '@store/authStore'
import { useUIStore } from '@store/uiStore'
import { roomsAPI } from '@api/rooms'
import { CreateRoomModal } from '@components/features/rooms/CreateRoomModal'
import { RoomCard } from '@components/features/rooms/RoomCard'
import { RoomFilters } from '@components/features/rooms/RoomFilters'
import styles from './Rooms.module.css'

export default function Rooms() {
  const { user } = useAuthStore()
  const { addNotification } = useUIStore()
  
  // State
  const [rooms, setRooms] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [filters, setFilters] = useState({
    search: '',
    status: 'all', // all, waiting, in_progress
    maxPlayers: 'all'
  })
  const [pagination, setPagination] = useState({
    page: 1,
    limit: 12,
    total: 0,
    totalPages: 0
  })

  // Load rooms
  const loadRooms = async (page = 1, newFilters = filters) => {
    try {
      setLoading(true)
      setError(null)

      const params = {
        page,
        limit: pagination.limit,
        ...(newFilters.search && { search: newFilters.search }),
        ...(newFilters.status !== 'all' && { status: newFilters.status })
      }

      const result = await roomsAPI.getRooms(params)

      if (result.success) {
        setRooms(result.rooms || [])
        setPagination(prev => ({
          ...prev,
          page,
          total: result.pagination?.total || 0,
          totalPages: result.pagination?.totalPages || 1
        }))
      } else {
        setError(result.error)
        addNotification({
          type: 'error',
          title: 'Ошибка загрузки',
          message: result.error
        })
      }
    } catch (err) {
      setError('Не удалось загрузить комнаты')
      addNotification({
        type: 'error',
        title: 'Ошибка',
        message: 'Не удалось загрузить комнаты'
      })
    } finally {
      setLoading(false)
    }
  }

  // Initial load
  useEffect(() => {
    loadRooms()
  }, [])

  // Handle filter changes
  const handleFilterChange = (newFilters) => {
    setFilters(newFilters)
    loadRooms(1, newFilters)
  }

  // Handle search
  const handleSearch = (searchTerm) => {
    const newFilters = { ...filters, search: searchTerm }
    handleFilterChange(newFilters)
  }

  // Handle page change
  const handlePageChange = (page) => {
    loadRooms(page)
  }

  // Join room
  const handleJoinRoom = async (roomId, password = '') => {
    try {
      const result = await roomsAPI.joinRoom(roomId, password)
      
      if (result.success) {
        addNotification({
          type: 'success',
          title: 'Успешно!',
          message: result.message
        })
        // Refresh rooms list
        loadRooms(pagination.page)
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
    }
  }

  // Create room success
  const handleRoomCreated = (newRoom) => {
    addNotification({
      type: 'success',
      title: 'Комната создана!',
      message: `Комната "${newRoom.name}" успешно создана`
    })
    setShowCreateModal(false)
    loadRooms(pagination.page)
  }

  // Filter rooms by max players
  const filteredRooms = rooms.filter(room => {
    if (filters.maxPlayers === 'all') return true
    const maxPlayers = parseInt(filters.maxPlayers)
    return room.max_players === maxPlayers
  })

  return (
    <div className={styles.roomsPage}>
      {/* Header */}
      <div className={styles.header}>
        <div className={styles.headerContent}>
          <div className={styles.headerText}>
            <h1>Игровые комнаты</h1>
            <p>Присоединяйтесь к существующим комнатам или создайте свою</p>
          </div>
          <button 
            className={styles.createButton}
            onClick={() => setShowCreateModal(true)}
          >
            <i className="fas fa-plus" />
            Создать комнату
          </button>
        </div>
      </div>

      {/* Filters */}
      <RoomFilters 
        filters={filters}
        onFilterChange={handleFilterChange}
        onSearch={handleSearch}
        loading={loading}
      />

      {/* Stats */}
      <div className={styles.stats}>
        <div className={styles.stat}>
          <span className={styles.statNumber}>{pagination.total}</span>
          <span className={styles.statLabel}>Всего комнат</span>
        </div>
        <div className={styles.stat}>
          <span className={styles.statNumber}>
            {rooms.filter(r => r.status === 'waiting').length}
          </span>
          <span className={styles.statLabel}>Ожидают игроков</span>
        </div>
        <div className={styles.stat}>
          <span className={styles.statNumber}>
            {rooms.filter(r => r.status === 'in_progress').length}
          </span>
          <span className={styles.statLabel}>В процессе</span>
        </div>
      </div>

      {/* Content */}
      <div className={styles.content}>
        {loading && (
          <div className={styles.loading}>
            <div className={styles.spinner} />
            <p>Загружаем комнаты...</p>
          </div>
        )}

        {error && !loading && (
          <div className={styles.error}>
            <i className="fas fa-exclamation-triangle" />
            <h3>Ошибка загрузки</h3>
            <p>{error}</p>
            <button 
              className={styles.retryButton}
              onClick={() => loadRooms(pagination.page)}
            >
              Попробовать снова
            </button>
          </div>
        )}

        {!loading && !error && filteredRooms.length === 0 && (
          <div className={styles.empty}>
            <i className="fas fa-inbox" />
            <h3>Комнаты не найдены</h3>
            <p>
              {filters.search ? 
                'Попробуйте изменить параметры поиска' : 
                'Пока нет доступных комнат. Создайте первую!'
              }
            </p>
            <button 
              className={styles.createButton}
              onClick={() => setShowCreateModal(true)}
            >
              Создать комнату
            </button>
          </div>
        )}

        {!loading && !error && filteredRooms.length > 0 && (
          <>
            {/* Rooms Grid */}
            <div className={styles.roomsGrid}>
              {filteredRooms.map(room => (
                <RoomCard
                  key={room.id}
                  room={room}
                  currentUser={user}
                  onJoin={handleJoinRoom}
                />
              ))}
            </div>

            {/* Pagination */}
            {pagination.totalPages > 1 && (
              <div className={styles.pagination}>
                <button 
                  className={styles.paginationButton}
                  onClick={() => handlePageChange(pagination.page - 1)}
                  disabled={pagination.page === 1}
                >
                  <i className="fas fa-chevron-left" />
                  Назад
                </button>
                
                <div className={styles.paginationInfo}>
                  <span>
                    Страница {pagination.page} из {pagination.totalPages}
                  </span>
                  <span className={styles.paginationTotal}>
                    ({pagination.total} комнат)
                  </span>
                </div>
                
                <button 
                  className={styles.paginationButton}
                  onClick={() => handlePageChange(pagination.page + 1)}
                  disabled={pagination.page === pagination.totalPages}
                >
                  Вперед
                  <i className="fas fa-chevron-right" />
                </button>
              </div>
            )}
          </>
        )}
      </div>

      {/* Create Room Modal */}
      {showCreateModal && (
        <CreateRoomModal
          onClose={() => setShowCreateModal(false)}
          onSuccess={handleRoomCreated}
        />
      )}
    </div>
  )
}