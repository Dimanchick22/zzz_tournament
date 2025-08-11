// src/hooks/useNotifications.js
import { useUIStore } from '@store/uiStore'
import { useI18n } from '@hooks/useI18n'

/**
 * Хук для работы с локализованными уведомлениями
 */
export const useNotifications = () => {
  const { addNotification: addUINotification } = useUIStore()
  const { t } = useI18n()

  const addNotification = (notification) => {
    // Если передан ключ перевода вместо готового текста
    const processedNotification = {
      ...notification,
      title: notification.titleKey ? t(notification.titleKey) : notification.title,
      message: notification.messageKey ? t(notification.messageKey, notification.messageParams) : notification.message
    }

    addUINotification(processedNotification)
  }

  // Готовые методы для часто используемых уведомлений
  const showSuccess = (messageKey, params = {}) => {
    addNotification({
      type: 'success',
      titleKey: 'common.success',
      messageKey,
      messageParams: params
    })
  }

  const showError = (messageKey, params = {}) => {
    addNotification({
      type: 'error',
      titleKey: 'common.error',
      messageKey,
      messageParams: params
    })
  }

  const showWarning = (messageKey, params = {}) => {
    addNotification({
      type: 'warning',
      titleKey: 'common.warning',
      messageKey,
      messageParams: params
    })
  }

  const showInfo = (messageKey, params = {}) => {
    addNotification({
      type: 'info',
      titleKey: 'common.info',
      messageKey,
      messageParams: params
    })
  }

  // Специализированные уведомления для игровых событий
  const showAuthSuccess = (action = 'login') => {
    const messageKey = action === 'login' ? 'auth.loginSuccess' : 'auth.registerSuccess'
    showSuccess(messageKey)
  }

  const showAuthError = (error = 'auth.invalidCredentials') => {
    showError(error)
  }

  const showRoomJoined = (roomName) => {
    addNotification({
      type: 'success',
      titleKey: 'common.success',
      messageKey: 'rooms.join.joinSuccess',
      messageParams: { roomName }
    })
  }

  const showTournamentStarted = () => {
    addNotification({
      type: 'info',
      titleKey: 'tournaments.tournamentStarted',
      messageKey: 'tournaments.tournamentStarted'
    })
  }

  return {
    addNotification,
    showSuccess,
    showError,
    showWarning,
    showInfo,
    showAuthSuccess,
    showAuthError,
    showRoomJoined,
    showTournamentStarted
  }
}