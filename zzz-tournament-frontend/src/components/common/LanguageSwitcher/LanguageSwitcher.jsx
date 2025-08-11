import { useState, useEffect } from 'react'
import { useI18n } from '@hooks/useI18n'
import { useUIStore } from '@store/uiStore'
import styles from './LanguageSwitcher.module.css'

export const LanguageSwitcher = ({ variant = 'dropdown', size = 'base', compact = false }) => {
  const { currentLanguage, supportedLanguages, changeLanguage, getLanguageInfo } = useI18n()
  const { addNotification } = useUIStore()
  const [isOpen, setIsOpen] = useState(false)
  const [loading, setLoading] = useState(false)

  const currentLangInfo = getLanguageInfo(currentLanguage)

  // Закрываем выпадающее меню при клике вне его
  useEffect(() => {
    const handleClickOutside = (event) => {
      if (isOpen && !event.target.closest(`.${styles.dropdown}`)) {
        setIsOpen(false)
      }
    }

    if (isOpen) {
      document.addEventListener('click', handleClickOutside)
      return () => document.removeEventListener('click', handleClickOutside)
    }
  }, [isOpen])

  // Закрываем меню при нажатии Escape
  useEffect(() => {
    const handleEscape = (event) => {
      if (event.key === 'Escape' && isOpen) {
        setIsOpen(false)
      }
    }

    if (isOpen) {
      document.addEventListener('keydown', handleEscape)
      return () => document.removeEventListener('keydown', handleEscape)
    }
  }, [isOpen])

  const handleLanguageChange = async (languageCode) => {
    if (languageCode === currentLanguage) return

    setLoading(true)
    setIsOpen(false)

    try {
      const success = await changeLanguage(languageCode)
      
      if (success) {
        const newLangInfo = getLanguageInfo(languageCode)
        addNotification({
          type: 'success',
          title: 'Language Changed',
          message: `Language switched to ${newLangInfo.name}`
        })
      } else {
        addNotification({
          type: 'error',
          title: 'Error',
          message: 'Failed to change language'
        })
      }
    } catch (error) {
      console.error('Language change error:', error)
      addNotification({
        type: 'error',
        title: 'Error',
        message: 'Failed to change language'
      })
    } finally {
      setLoading(false)
    }
  }

  // Компактный переключатель (только флаги)
  if (variant === 'compact') {
    return (
      <div className={`${styles.compactSwitcher} ${styles[size]}`}>
        {supportedLanguages.map(lang => (
          <button
            key={lang.code}
            className={`${styles.compactButton} ${
              currentLanguage === lang.code ? styles.active : ''
            }`}
            onClick={() => handleLanguageChange(lang.code)}
            disabled={loading}
            title={lang.nativeName}
            aria-label={`Switch to ${lang.name}`}
          >
            <span className={styles.flag} role="img" aria-label={lang.name}>
              {lang.flag}
            </span>
          </button>
        ))}
        {loading && <div className={styles.loadingOverlay} />}
      </div>
    )
  }

  // Кнопочный переключатель
  if (variant === 'buttons') {
    return (
      <div className={`${styles.buttonGroup} ${styles[size]}`}>
        {supportedLanguages.map(lang => (
          <button
            key={lang.code}
            className={`${styles.langButton} ${
              currentLanguage === lang.code ? styles.active : ''
            }`}
            onClick={() => handleLanguageChange(lang.code)}
            disabled={loading}
            title={lang.nativeName}
            aria-label={`Switch to ${lang.name}`}
          >
            <span className={styles.flag} role="img" aria-label={lang.name}>
              {lang.flag}
            </span>
            {!compact && <span className={styles.code}>{lang.code.toUpperCase()}</span>}
          </button>
        ))}
        {loading && <div className={styles.loadingOverlay} />}
      </div>
    )
  }

  // Минималистичный переключатель (один флаг)
  if (variant === 'minimal') {
    return (
      <div className={`${styles.minimal} ${styles[size]}`}>
        <button
          className={styles.minimalButton}
          onClick={() => {
            const nextLang = supportedLanguages.find(lang => lang.code !== currentLanguage)
            if (nextLang) handleLanguageChange(nextLang.code)
          }}
          disabled={loading}
          title={`Switch to ${supportedLanguages.find(lang => lang.code !== currentLanguage)?.nativeName}`}
          aria-label={`Current language: ${currentLangInfo?.name}. Click to switch`}
        >
          {loading ? (
            <div className={styles.spinner} />
          ) : (
            <span className={styles.flag} role="img" aria-label={currentLangInfo?.name}>
              {currentLangInfo?.flag}
            </span>
          )}
        </button>
      </div>
    )
  }

  // Dropdown переключатель (по умолчанию)
  return (
    <div className={`${styles.dropdown} ${styles[size]}`}>
      <button
        className={`${styles.dropdownTrigger} ${isOpen ? styles.open : ''}`}
        onClick={() => setIsOpen(!isOpen)}
        disabled={loading}
        aria-expanded={isOpen}
        aria-haspopup="listbox"
        aria-label={`Language selector. Current: ${currentLangInfo?.name}`}
      >
        {loading ? (
          <div className={styles.spinner} />
        ) : (
          <>
            <span className={styles.flag} role="img" aria-label={currentLangInfo?.name}>
              {currentLangInfo?.flag}
            </span>
            {!compact && (
              <>
                <span className={styles.langName}>{currentLangInfo?.nativeName}</span>
                <i className={`fas fa-chevron-down ${styles.chevron}`} aria-hidden="true" />
              </>
            )}
          </>
        )}
      </button>

      {isOpen && (
        <div className={styles.dropdownMenu} role="listbox" aria-label="Language options">
          {supportedLanguages.map(lang => (
            <button
              key={lang.code}
              className={`${styles.langOption} ${
                currentLanguage === lang.code ? styles.selected : ''
              }`}
              onClick={() => handleLanguageChange(lang.code)}
              role="option"
              aria-selected={currentLanguage === lang.code}
            >
              <span className={styles.flag} role="img" aria-label={lang.name}>
                {lang.flag}
              </span>
              {!compact && (
                <div className={styles.langInfo}>
                  <span className={styles.nativeName}>{lang.nativeName}</span>
                  <span className={styles.englishName}>{lang.name}</span>
                </div>
              )}
              {currentLanguage === lang.code && (
                <i className={`fas fa-check ${styles.checkIcon}`} aria-hidden="true" />
              )}
            </button>
          ))}
        </div>
      )}
    </div>
  )
}