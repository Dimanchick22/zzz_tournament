// RoomFilters.jsx - Компонент фильтрации комнат с переводами
import { useState, useEffect } from 'react'
import { useDebounce } from '@hooks/useDebounce'
import { useI18n } from '@hooks/useI18n'
import styles from './RoomFilters.module.css'

export const RoomFilters = ({ filters, onFilterChange, onSearch, loading }) => {
  const { t } = useI18n()
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
              placeholder={t('rooms.searchPlaceholder')}
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              disabled={loading}
            />
            {searchTerm && (
              <button 
                className={styles.clearSearch}
                onClick={() => setSearchTerm('')}
                type="button"
                title={t('common.clear')}
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
          <label>{t('rooms.filters.status')}</label>
          <select
            value={filters.status}
            onChange={(e) => handleFilterChange('status', e.target.value)}
            disabled={loading}
          >
            <option value="all">{t('rooms.filters.all')}</option>
            <option value="waiting">{t('rooms.filters.waiting')}</option>
            <option value="in_progress">{t('rooms.filters.inProgress')}</option>
            <option value="finished">{t('rooms.filters.finished')}</option>
          </select>
        </div>

        {/* Max Players Filter */}
        <div className={styles.filterGroup}>
          <label>{t('rooms.filters.size')}</label>
          <select
            value={filters.maxPlayers}
            onChange={(e) => handleFilterChange('maxPlayers', e.target.value)}
            disabled={loading}
          >
            <option value="all">{t('rooms.filters.anySize')}</option>
            <option value="2">{t('rooms.filters.players', { count: 2 })}</option>
            <option value="4">{t('rooms.filters.players', { count: 4 })}</option>
            <option value="8">{t('rooms.filters.players', { count: 8 })}</option>
            <option value="16">{t('rooms.filters.players', { count: 16 })}</option>
          </select>
        </div>

        {/* Sort */}
        <div className={styles.filterGroup}>
          <label>{t('common.sort')}</label>
          <select
            value={filters.sortBy || 'created_at'}
            onChange={(e) => handleFilterChange('sortBy', e.target.value)}
            disabled={loading}
          >
            <option value="created_at">{t('rooms.sort.byDate')}</option>
            <option value="name">{t('rooms.sort.byName')}</option>
            <option value="participants">{t('rooms.sort.byParticipants')}</option>
            <option value="max_players">{t('rooms.sort.bySize')}</option>
          </select>
        </div>

        {/* Clear Filters */}
        {hasActiveFilters && (
          <button 
            className={styles.clearButton}
            onClick={clearFilters}
            disabled={loading}
            title={t('rooms.filters.clearAll')}
          >
            <i className="fas fa-times" />
            {t('rooms.filters.clearAll')}
          </button>
        )}
      </div>

      {/* Active Filters Display */}
      {hasActiveFilters && (
        <div className={styles.activeFilters}>
          <span className={styles.activeFiltersLabel}>{t('rooms.filters.activeFilters')}:</span>
          
          {filters.search && (
            <div className={styles.filterTag}>
              <span>{t('rooms.filters.searchTag')}: "{filters.search}"</span>
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
              <span>{t('rooms.filters.status')}: {getStatusLabel(filters.status, t)}</span>
              <button onClick={() => handleFilterChange('status', 'all')}>
                <i className="fas fa-times" />
              </button>
            </div>
          )}
          
          {filters.maxPlayers !== 'all' && (
            <div className={styles.filterTag}>
              <span>{t('rooms.filters.size')}: {t('rooms.filters.players', { count: filters.maxPlayers })}</span>
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
const getStatusLabel = (status, t) => {
  switch (status) {
    case 'waiting': return t('rooms.filters.waiting')
    case 'in_progress': return t('rooms.filters.inProgress')
    case 'finished': return t('rooms.filters.finished')
    default: return status
  }
}