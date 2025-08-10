// CreateRoomModal.jsx - Модал создания комнаты
import { useState } from 'react'
import { roomsAPI } from '@api/rooms'
import { useUIStore } from '@store/uiStore'
import styles from './CreateRoomModal.module.css'

export const CreateRoomModal = ({ onClose, onSuccess }) => {
  const { addNotification } = useUIStore()
  
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    max_players: 8,
    is_private: false,
    password: ''
  })
  
  const [errors, setErrors] = useState({})
  const [loading, setLoading] = useState(false)

  // Handle form change
  const handleChange = (e) => {
    const { name, value, type, checked } = e.target
    
    setFormData(prev => ({
      ...prev,
      [name]: type === 'checkbox' ? checked : value
    }))
    
    // Clear error when user starts typing
    if (errors[name]) {
      setErrors(prev => ({
        ...prev,
        [name]: ''
      }))
    }
  }

  // Validate form
  const validateForm = () => {
    const newErrors = {}
    
    if (!formData.name.trim()) {
      newErrors.name = 'Название комнаты обязательно'
    } else if (formData.name.length < 3) {
      newErrors.name = 'Название должно быть не менее 3 символов'
    } else if (formData.name.length > 50) {
      newErrors.name = 'Название не должно превышать 50 символов'
    }
    
    if (formData.description.length > 200) {
      newErrors.description = 'Описание не должно превышать 200 символов'
    }
    
    if (formData.max_players < 2) {
      newErrors.max_players = 'Минимум 2 игрока'
    } else if (formData.max_players > 16) {
      newErrors.max_players = 'Максимум 16 игроков'
    }
    
    if (formData.is_private && !formData.password.trim()) {
      newErrors.password = 'Пароль обязателен для приватной комнаты'
    } else if (formData.password && formData.password.length < 4) {
      newErrors.password = 'Пароль должен быть не менее 4 символов'
    }
    
    return newErrors
  }

  // Handle submit
  const handleSubmit = async (e) => {
    e.preventDefault()
    
    const validationErrors = validateForm()
    if (Object.keys(validationErrors).length > 0) {
      setErrors(validationErrors)
      return
    }
    
    setLoading(true)
    
    try {
      const result = await roomsAPI.createRoom({
        name: formData.name.trim(),
        description: formData.description.trim(),
        max_players: parseInt(formData.max_players),
        is_private: formData.is_private,
        password: formData.is_private ? formData.password : ''
      })
      
      if (result.success) {
        onSuccess(result.room)
      } else {
        addNotification({
          type: 'error',
          title: 'Ошибка создания',
          message: result.error
        })
        
        // Handle validation errors from server
        if (result.details) {
          const serverErrors = {}
          result.details.forEach(detail => {
            serverErrors[detail.field] = detail.message
          })
          setErrors(serverErrors)
        }
      }
    } catch (err) {
      addNotification({
        type: 'error',
        title: 'Ошибка',
        message: 'Не удалось создать комнату'
      })
    } finally {
      setLoading(false)
    }
  }

  // Handle close
  const handleClose = () => {
    if (!loading) {
      onClose()
    }
  }

  // Handle overlay click
  const handleOverlayClick = (e) => {
    if (e.target === e.currentTarget) {
      handleClose()
    }
  }

  return (
    <div className={styles.overlay} onClick={handleOverlayClick}>
      <div className={styles.modal}>
        {/* Header */}
        <div className={styles.header}>
          <h2>Создать комнату</h2>
          <button 
            className={styles.closeButton}
            onClick={handleClose}
            disabled={loading}
          >
            <i className="fas fa-times" />
          </button>
        </div>

        {/* Form */}
        <form onSubmit={handleSubmit} className={styles.form}>
          {/* Room Name */}
          <div className={styles.field}>
            <label htmlFor="name">
              Название комнаты *
            </label>
            <input
              type="text"
              id="name"
              name="name"
              value={formData.name}
              onChange={handleChange}
              placeholder="Введите название комнаты"
              className={errors.name ? styles.error : ''}
              disabled={loading}
              maxLength={50}
            />
            {errors.name && (
              <span className={styles.errorText}>{errors.name}</span>
            )}
          </div>

          {/* Description */}
          <div className={styles.field}>
            <label htmlFor="description">
              Описание
            </label>
            <textarea
              id="description"
              name="description"
              value={formData.description}
              onChange={handleChange}
              placeholder="Описание комнаты (необязательно)"
              className={errors.description ? styles.error : ''}
              disabled={loading}
              rows={3}
              maxLength={200}
            />
            <div className={styles.charCount}>
              {formData.description.length}/200
            </div>
            {errors.description && (
              <span className={styles.errorText}>{errors.description}</span>
            )}
          </div>

          {/* Max Players */}
          <div className={styles.field}>
            <label htmlFor="max_players">
              Максимум игроков
            </label>
            <select
              id="max_players"
              name="max_players"
              value={formData.max_players}
              onChange={handleChange}
              className={errors.max_players ? styles.error : ''}
              disabled={loading}
            >
              {Array.from({ length: 15 }, (_, i) => i + 2).map(num => (
                <option key={num} value={num}>
                  {num} игроков
                </option>
              ))}
            </select>
            {errors.max_players && (
              <span className={styles.errorText}>{errors.max_players}</span>
            )}
          </div>

          {/* Private Room */}
          <div className={styles.checkboxField}>
            <label className={styles.checkboxLabel}>
              <input
                type="checkbox"
                name="is_private"
                checked={formData.is_private}
                onChange={handleChange}
                disabled={loading}
              />
              <span className={styles.checkmark}></span>
              Приватная комната
            </label>
            <p className={styles.fieldHint}>
              Приватные комнаты требуют пароль для входа
            </p>
          </div>

          {/* Password */}
          {formData.is_private && (
            <div className={styles.field}>
              <label htmlFor="password">
                Пароль *
              </label>
              <input
                type="password"
                id="password"
                name="password"
                value={formData.password}
                onChange={handleChange}
                placeholder="Введите пароль"
                className={errors.password ? styles.error : ''}
                disabled={loading}
                minLength={4}
              />
              {errors.password && (
                <span className={styles.errorText}>{errors.password}</span>
              )}
            </div>
          )}

          {/* Actions */}
          <div className={styles.actions}>
            <button
              type="button"
              className={styles.cancelButton}
              onClick={handleClose}
              disabled={loading}
            >
              Отмена
            </button>
            
            <button
              type="submit"
              className={styles.submitButton}
              disabled={loading}
            >
              {loading ? (
                <>
                  <div className={styles.spinner} />
                  Создание...
                </>
              ) : (
                <>
                  <i className="fas fa-plus" />
                  Создать комнату
                </>
              )}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}