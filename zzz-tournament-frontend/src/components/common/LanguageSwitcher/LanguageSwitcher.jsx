import { useState } from 'react'
import { useI18n } from '@hooks/useI18n'
import { useUIStore } from '@store/uiStore'
import styles from './LanguageSwitcher.module.css'

export const LanguageSwitcher = ({ variant = 'dropdown', size = 'base' }) => {
  const { currentLanguage, supportedLanguages, changeLanguage, getLanguageInfo } = useI18n()
  const { addNotification } = useUIStore()
  const [isOpen, setIsOpen] = useState(false)
  const [loading, setLoading] = useState(false)

  const currentLangInfo = getLanguageInfo(currentLanguage)

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
      addNotification({
        type: 'error',
        title: 'Error',
        message: 'Failed to change language'
      })
    } finally {
      setLoading(false)
    }
  }

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
          >
            <span className={styles.flag}>{lang.flag}</span>
            <span className={styles.code}>{lang.code.toUpperCase()}</span>
          </button>
        ))}
      </div>
    )
  }

  return (
    <div className={`${styles.dropdown} ${styles[size]}`}>
      <button
        className={`${styles.dropdownTrigger} ${isOpen ? styles.open : ''}`}
        onClick={() => setIsOpen(!isOpen)}
        disabled={loading}
        aria-expanded={isOpen}
        aria-haspopup="listbox"
      >
        {loading ? (
          <div className={styles.spinner} />
        ) : (
          <>
            <span className={styles.flag}>{currentLangInfo?.flag}</span>
            <span className={styles.langName}>{currentLangInfo?.nativeName}</span>
            <i className={`fas fa-chevron-down ${styles.chevron}`} />
          </>
        )}
      </button>

      {isOpen && (
        <>
          <div 
            className={styles.overlay}
            onClick={() => setIsOpen(false)}
          />
          <div className={styles.dropdownMenu} role="listbox">
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
                <span className={styles.flag}>{lang.flag}</span>
                <div className={styles.langInfo}>
                  <span className={styles.nativeName}>{lang.nativeName}</span>
                  <span className={styles.englishName}>{lang.name}</span>
                </div>
                {currentLanguage === lang.code && (
                  <i className={`fas fa-check ${styles.checkIcon}`} />
                )}
              </button>
            ))}
          </div>
        </>
      )}
    </div>
  )
}