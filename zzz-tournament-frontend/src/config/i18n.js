// src/config/i18n.js
import i18n from 'i18next'
import { initReactI18next } from 'react-i18next'
import LanguageDetector from 'i18next-browser-languagedetector'
import Backend from 'i18next-http-backend'

// Импорт переводов
import enTranslations from '@/locales/en.json'
import ruTranslations from '@/locales/ru.json'

const resources = {
  en: {
    translation: enTranslations
  },
  ru: {
    translation: ruTranslations
  }
}

i18n
  // Подключаем плагины
  .use(Backend) // Для загрузки переводов с сервера (опционально)
  .use(LanguageDetector) // Автоопределение языка
  .use(initReactI18next) // Интеграция с React
  
  // Инициализация
  .init({
    resources,
    
    // Язык по умолчанию
    fallbackLng: 'en',
    
    // Язык по умолчанию для русскоязычных пользователей
    lng: 'ru',
    
    // Отладка (включить в development)
    debug: process.env.NODE_ENV === 'development',
    
    // Настройки интерполяции
    interpolation: {
      escapeValue: false, // React уже защищает от XSS
    },
    
    // Настройки определения языка
    detection: {
      // Порядок определения языка
      order: ['localStorage', 'navigator', 'htmlTag'],
      
      // Ключ для сохранения в localStorage
      lookupLocalStorage: 'i18nextLng',
      
      // Кэширование
      caches: ['localStorage'],
      
      // Исключения
      excludeCacheFor: ['cimode'],
    },
    
    // Настройки backend (если используете)
    backend: {
      loadPath: '/locales/{{lng}}.json',
    },
    
    // Поддерживаемые языки
    supportedLngs: ['en', 'ru'],
    
    // Не загружать язык по умолчанию дважды
    load: 'languageOnly',
    
    // Настройки пространств имен
    defaultNS: 'translation',
    ns: ['translation'],
    
    // Реакция на изменение языка
    react: {
      useSuspense: false, // Отключить suspense для избежания проблем
      bindI18n: 'languageChanged loaded',
      bindI18nStore: 'added removed',
    },
  })

export default i18n

// Хелперы для работы с языками
export const SUPPORTED_LANGUAGES = [
  {
    code: 'ru',
    name: 'Русский',
    nativeName: 'Русский',
    flag: '🇷🇺'
  },
  {
    code: 'en',
    name: 'English',
    nativeName: 'English',
    flag: '🇺🇸'
  }
]

export const getCurrentLanguage = () => i18n.language

export const changeLanguage = (lng) => {
  return i18n.changeLanguage(lng)
}

export const getLanguageInfo = (code) => {
  return SUPPORTED_LANGUAGES.find(lang => lang.code === code)
}