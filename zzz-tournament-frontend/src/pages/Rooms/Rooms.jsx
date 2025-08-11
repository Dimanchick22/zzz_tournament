// Rooms Page - Полная реализация с переводами
import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { useAuthStore } from '@store/authStore'
import { useUIStore } from '@store/uiStore'
import { useI18n } from '@hooks/useI18n'
import { roomsAPI } from '@api/rooms'
import { CreateRoomModal } from '@components/features/rooms/CreateRoomModal'
import { RoomCard } from '@components/features/rooms/RoomCard'
import { RoomFilters } from '@components/features/rooms/RoomFilters'
import styles from './Rooms.module.css'

export default function Rooms() {
  const { user } = useAuthStore()
  const { addNotification } = useUIStore()
  const { t } = useI18n()
  
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
        const roomsArray = Array.isArray(result.rooms) ? result.rooms : []
        setRooms(roomsArray)
        
        setPagination(prev => ({
          ...prev,
          page,
          total: result.pagination?.total || 0,
          totalPages: result.pagination?.totalPages || 1
        }))
      } else {
        setError(result.error)
        setRooms([])
        addNotification({
          type: 'error',
          title: t('errors.loadingError'),
          message: result.error
        })
      }
    } catch (err) {
      setError(t('rooms.errorLoading'))
      setRooms([])
      addNotification({
        type: 'error',
        title: t('common.error'),
        message: t('rooms.errorLoading')
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
          title: t('common.success'),
          message: result.message || t('rooms.join.joinSuccess')
        })
        // Refresh rooms list
        loadRooms(pagination.page)
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

  // Create room success
  const handleRoomCreated = (newRoom) => {
    addNotification({
      type: 'success',
      title: t('rooms.create.createSuccess'),
      message: t('rooms.create.createSuccess')
    })
    setShowCreateModal(false)
    loadRooms(pagination.page)
  }

  // Filter rooms by max players
  const filteredRooms = Array.isArray(rooms) ? rooms.filter(room => {
    if (filters.maxPlayers === 'all') return true
    const maxPlayers = parseInt(filters.maxPlayers)
    return room.max_players === maxPlayers
  }) : []

  const waitingRoomsCount = Array.isArray(rooms) ? rooms.filter(r => r.status === 'waiting').length : 0
  const inProgressRoomsCount = Array.isArray(rooms) ? rooms.filter(r => r.status === 'in_progress').length : 0

  return (
    <div className={styles.roomsPage}>
      {/* Header */}
      <div className={styles.header}>
        <div className={styles.headerContent}>
          <div className={styles.headerText}>
            <h1>{t('rooms.title')}</h1>
            <p>{t('rooms.subtitle')}</p>
          </div>
          <button 
            className={styles.createButton}
            onClick={() => setShowCreateModal(true)}
          >
            <i className="fas fa-plus" />
            {t('rooms.createRoom')}
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
          <span className={styles.statLabel}>{t('rooms.filters.all')}</span>
        </div>
        <div className={styles.stat}>
          <span className={styles.statNumber}>{waitingRoomsCount}</span>
          <span className={styles.statLabel}>{t('rooms.filters.waiting')}</span>
        </div>
        <div className={styles.stat}>
          <span className={styles.statNumber}>{inProgressRoomsCount}</span>
          <span className={styles.statLabel}>{t('rooms.filters.inProgress')}</span>
        </div>
      </div>

      {/* Content */}
      <div className={styles.content}>
        {loading && (
          <div className={styles.loading}>
            <div className={styles.spinner} />
            <p>{t('rooms.loadingRooms')}</p>
          </div>
        )}

        {error && !loading && (
          <div className={styles.error}>
            <i className="fas fa-exclamation-triangle" />
            <h3>{t('errors.loadingError')}</h3>
            <p>{error}</p>
            <button 
              className={styles.retryButton}
              onClick={() => loadRooms(pagination.page)}
            >
              {t('common.retry')}
            </button>
          </div>
        )}

        {!loading && !error && filteredRooms.length === 0 && (
          <div className={styles.empty}>
            <i className="fas fa-inbox" />
            <h3>{t('rooms.noRooms')}</h3>
            <p>
              {filters.search ? 
                t('rooms.noRoomsWithFilters') : 
                t('rooms.noRoomsAvailable')
              }
            </p>
            <button 
              className={styles.createButton}
              onClick={() => setShowCreateModal(true)}
            >
              {t('rooms.createRoom')}
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
                  {t('common.previous')}
                </button>
                
                <div className={styles.paginationInfo}>
                  <span>
                    {t('pagination.pageOf', {
                      current: pagination.page,
                      total: pagination.totalPages
                    })}
                  </span>
                  <span className={styles.paginationTotal}>
                    ({t('pagination.totalItems', { count: pagination.total })})
                  </span>
                </div>
                
                <button 
                  className={styles.paginationButton}
                  onClick={() => handlePageChange(pagination.page + 1)}
                  disabled={pagination.page === pagination.totalPages}
                >
                  {t('common.next')}
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