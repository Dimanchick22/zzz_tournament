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

// Функция для получения языка из localStorage
const getStoredLanguage = () => {
  try {
    const stored = localStorage.getItem('i18nextLng')
    if (stored && ['ru', 'en'].includes(stored)) {
      return stored
    }
  } catch (error) {
    console.warn('Error reading language from localStorage:', error)
  }
  return null
}

// Определяем язык по умолчанию
const getDefaultLanguage = () => {
  // Сначала проверяем localStorage
  const storedLang = getStoredLanguage()
  if (storedLang) {
    return storedLang
  }
  
  // Затем проверяем язык браузера
  const browserLang = navigator.language || navigator.languages?.[0]
  if (browserLang?.startsWith('ru')) {
    return 'ru'
  }
  
  // По умолчанию английский
  return 'en'
}

const defaultLanguage = getDefaultLanguage()

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
    
    // Устанавливаем определенный язык
    lng: defaultLanguage,
    
    // Отладка (включить в development)
    debug: import.meta.env.DEV,
    
    // Настройки интерполяции
    interpolation: {
      escapeValue: false, // React уже защищает от XSS
    },
    
    // Настройки определения языка
    detection: {
      // Порядок определения языка - localStorage в приоритете
      order: ['localStorage', 'navigator', 'htmlTag'],
      
      // Ключ для сохранения в localStorage
      lookupLocalStorage: 'i18nextLng',
      
      // Кэширование
      caches: ['localStorage'],
      
      // Исключения
      excludeCacheFor: ['cimode'],
      
      // Проверяем только поддерживаемые языки
      checkWhitelist: true,
    },
    
    // Настройки backend (если используете)
    backend: {
      loadPath: '/locales/{{lng}}.json',
    },
    
    // Поддерживаемые языки
    supportedLngs: ['en', 'ru'],
    nonExplicitSupportedLngs: true,
    
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

// Устанавливаем атрибут языка в HTML при инициализации
document.documentElement.lang = defaultLanguage

// Слушаем изменения языка для обновления HTML атрибута
i18n.on('languageChanged', (lng) => {
  document.documentElement.lang = lng
  // Также сохраняем в localStorage для надежности
  try {
    localStorage.setItem('i18nextLng', lng)
  } catch (error) {
    console.warn('Error saving language to localStorage:', error)
  }
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