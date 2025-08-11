// src/hooks/useI18n.js
import { useTranslation } from 'react-i18next'
import { SUPPORTED_LANGUAGES, changeLanguage as i18nChangeLanguage } from '@config/i18n'

/**
 * Расширенный хук для работы с переводами
 */
export const useI18n = () => {
  const { t, i18n } = useTranslation()
  
  const currentLanguage = i18n.language
  const isRTL = false // Добавить поддержку RTL языков если нужно
  
  const changeLanguage = async (languageCode) => {
    try {
      await i18nChangeLanguage(languageCode)
      
      // i18next автоматически сохраняет в localStorage, но добавим явное сохранение для надежности
      localStorage.setItem('i18nextLng', languageCode)
      
      // Обновляем атрибут lang в HTML
      document.documentElement.lang = languageCode
      
      return true
    } catch (error) {
      console.error('Error changing language:', error)
      return false
    }
  }
  
  const getLanguageInfo = (code = currentLanguage) => {
    return SUPPORTED_LANGUAGES.find(lang => lang.code === code)
  }
  
  const formatDate = (date, options = {}) => {
    const locale = currentLanguage === 'ru' ? 'ru-RU' : 'en-US'
    return new Intl.DateTimeFormat(locale, options).format(new Date(date))
  }
  
  const formatNumber = (number, options = {}) => {
    const locale = currentLanguage === 'ru' ? 'ru-RU' : 'en-US'
    return new Intl.NumberFormat(locale, options).format(number)
  }
  
  const formatCurrency = (amount, currency = 'USD') => {
    const locale = currentLanguage === 'ru' ? 'ru-RU' : 'en-US'
    return new Intl.NumberFormat(locale, {
      style: 'currency',
      currency: currency
    }).format(amount)
  }
  
  const formatRelativeTime = (date) => {
    const now = new Date()
    const targetDate = new Date(date)
    const diffInSeconds = Math.floor((now - targetDate) / 1000)
    
    if (diffInSeconds < 60) {
      return t('time.now')
    }
    
    const diffInMinutes = Math.floor(diffInSeconds / 60)
    if (diffInMinutes < 60) {
      return t('time.minutesAgo', { count: diffInMinutes })
    }
    
    const diffInHours = Math.floor(diffInMinutes / 60)
    if (diffInHours < 24) {
      return t('time.hoursAgo', { count: diffInHours })
    }
    
    const diffInDays = Math.floor(diffInHours / 24)
    if (diffInDays < 7) {
      return t('time.daysAgo', { count: diffInDays })
    }
    
    const diffInWeeks = Math.floor(diffInDays / 7)
    if (diffInWeeks < 4) {
      return t('time.weeksAgo', { count: diffInWeeks })
    }
    
    const diffInMonths = Math.floor(diffInDays / 30)
    if (diffInMonths < 12) {
      return t('time.monthsAgo', { count: diffInMonths })
    }
    
    const diffInYears = Math.floor(diffInDays / 365)
    return t('time.yearsAgo', { count: diffInYears })
  }
  
  // Хелпер для плюрализации (русский язык имеет сложные правила)
  const pluralize = (count, options) => {
    if (currentLanguage === 'ru') {
      const lastDigit = count % 10
      const lastTwoDigits = count % 100
      
      if (lastTwoDigits >= 11 && lastTwoDigits <= 14) {
        return options.many || options.other
      }
      
      if (lastDigit === 1) {
        return options.one
      }
      
      if (lastDigit >= 2 && lastDigit <= 4) {
        return options.few || options.other
      }
      
      return options.many || options.other
    }
    
    // Для английского языка
    return count === 1 ? options.one : options.other
  }
  
  return {
    t,
    currentLanguage,
    supportedLanguages: SUPPORTED_LANGUAGES,
    isRTL,
    changeLanguage,
    getLanguageInfo,
    formatDate,
    formatNumber,
    formatCurrency,
    formatRelativeTime,
    pluralize
  }
}