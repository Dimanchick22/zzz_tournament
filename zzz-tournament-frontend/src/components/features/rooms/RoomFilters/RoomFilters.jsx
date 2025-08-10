// RoomFilters.jsx - Компонент фильтрации комнат
import { useState, useEffect } from 'react'
import { useDebounce } from '@hooks/useDebounce'
import styles from './RoomFilters.module.css'

export const RoomFilters = ({ filters, onFilterChange, onSearch, loading }) => {
  const [searchTerm, setSearchTerm] = useState(filters.search || '')
  const debouncedSearchTerm = useDebounce(searchTerm, 500)

  // Handle debounced search
  useEffect(() => {
    if (debouncedSearchTerm !== filters.search) {
      onSearch(debouncedSearchTerm)
    }
  }, [debouncedSearchTerm, filters.search, onSearch])

  // Handle filter change
  const handleFilterChange = (key, value) => {
    onFilterChange({
      ...filters,
      [key]: value
    })
  }

  // Clear all filters
  const clearFilters = () => {
    setSearchTerm('')
    onFilterChange({
      search: '',
      status: 'all',
      maxPlayers: 'all'
    })
  }

  // Check if any filters are active
  const hasActiveFilters = 
    filters.search || 
    filters.status !== 'all' || 
    filters.maxPlayers !== 'all'

  return (
    <div className={styles.filters}>
      {/* Search */}
      <div className={styles.searchSection}>
        <div className={styles.searchField}>
          <div className={styles.searchInput}>
            <i className="fas fa-search" />
            <input
              type="text"
              placeholder="Поиск комнат..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              disabled={loading}
            />
            {searchTerm && (
              <button 
                className={styles.clearSearch}
                onClick={() => setSearchTerm('')}
                type="button"
              >
                <i className="fas fa-times" />
              </button>
            )}
          </div>
        </div>
      </div>

      {/* Filter Controls */}
      <div className={styles.filterControls}>
        {/* Status Filter */}
        <div className={styles.filterGroup}>
          <label>Статус</label>
          <select
            value={filters.status}
            onChange={(e) => handleFilterChange('status', e.target.value)}
            disabled={loading}
          >
            <option value="all">Все комнаты</option>
            <option value="waiting">Ожидают игроков</option>
            <option value="in_progress">В процессе</option>
            <option value="finished">Завершенные</option>
          </select>
        </div>

        {/* Max Players Filter */}
        <div className={styles.filterGroup}>
          <label>Размер</label>
          <select
            value={filters.maxPlayers}
            onChange={(e) => handleFilterChange('maxPlayers', e.target.value)}
            disabled={loading}
          >
            <option value="all">Любой размер</option>
            <option value="2">2 игрока</option>
            <option value="4">4 игрока</option>
            <option value="8">8 игроков</option>
            <option value="16">16 игроков</option>
          </select>
        </div>

        {/* Sort */}
        <div className={styles.filterGroup}>
          <label>Сортировка</label>
          <select
            value={filters.sortBy || 'created_at'}
            onChange={(e) => handleFilterChange('sortBy', e.target.value)}
            disabled={loading}
          >
            <option value="created_at">По дате создания</option>
            <option value="name">По названию</option>
            <option value="participants">По количеству игроков</option>
            <option value="max_players">По размеру комнаты</option>
          </select>
        </div>

        {/* Clear Filters */}
        {hasActiveFilters && (
          <button 
            className={styles.clearButton}
            onClick={clearFilters}
            disabled={loading}
          >
            <i className="fas fa-times" />
            Очистить фильтры
          </button>
        )}
      </div>

      {/* Active Filters Display */}
      {hasActiveFilters && (
        <div className={styles.activeFilters}>
          <span className={styles.activeFiltersLabel}>Активные фильтры:</span>
          
          {filters.search && (
            <div className={styles.filterTag}>
              <span>Поиск: "{filters.search}"</span>
              <button onClick={() => {
                setSearchTerm('')
                handleFilterChange('search', '')
              }}>
                <i className="fas fa-times" />
              </button>
            </div>
          )}
          
          {filters.status !== 'all' && (
            <div className={styles.filterTag}>
              <span>Статус: {getStatusLabel(filters.status)}</span>
              <button onClick={() => handleFilterChange('status', 'all')}>
                <i className="fas fa-times" />
              </button>
            </div>
          )}
          
          {filters.maxPlayers !== 'all' && (
            <div className={styles.filterTag}>
              <span>Размер: {filters.maxPlayers} игроков</span>
              <button onClick={() => handleFilterChange('maxPlayers', 'all')}>
                <i className="fas fa-times" />
              </button>
            </div>
          )}
        </div>
      )}
    </div>
  )
}

// Helper function to get status label
const getStatusLabel = (status) => {
  switch (status) {
    case 'waiting': return 'Ожидают игроков'
    case 'in_progress': return 'В процессе'
    case 'finished': return 'Завершенные'
    default: return status
  }
}