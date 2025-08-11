// src/components/layout/Header/Header.jsx - обновленная версия с переводами
import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuthStore } from '@store/authStore'
import { useUIStore } from '@store/uiStore'
import { useAuth } from '@hooks/useAuth'
import { useI18n } from '@hooks/useI18n'
import { LanguageSwitcher } from '@components/common/LanguageSwitcher'
import styles from './Header.module.css'

export const Header = () => {
  const { user } = useAuthStore()
  const { theme, toggleTheme, toggleSidebar, isMobile } = useUIStore()
  const { logout } = useAuth()
  const { t } = useI18n()
  const navigate = useNavigate()
  const [userMenuOpen, setUserMenuOpen] = useState(false)

  const handleLogout = () => {
    logout()
    setUserMenuOpen(false)
  }

  const navigateToProfile = () => {
    navigate('/profile')
    setUserMenuOpen(false)
  }

  const navigateToSettings = () => {
    // TODO: Создать страницу настроек
    console.log('Navigate to settings')
    setUserMenuOpen(false)
  }

  return (
    <header className={styles.header}>
      <div className={styles.left}>
        {/* Sidebar Toggle */}
        <button 
          className={styles.sidebarToggle}
          onClick={toggleSidebar}
          aria-label={t('common.toggleSidebar')}
          title={t('common.toggleSidebar')}
        >
          <i className="fas fa-bars" />
        </button>
        
        {/* Logo/Title */}
        <div className={styles.logo}>
          <h1>ZZZ Tournament</h1>
        </div>
      </div>

      <div className={styles.right}>
        {/* Language Switcher */}
        <LanguageSwitcher variant="dropdown" size="small" />

        {/* Theme Toggle */}
        <button 
          className={styles.themeToggle}
          onClick={toggleTheme}
          aria-label={t('common.toggleTheme')}
          title={t('common.toggleTheme')}
        >
          <i className={`fas ${theme === 'dark' ? 'fa-sun' : 'fa-moon'}`} />
        </button>

        {/* User Menu */}
        <div className={styles.userMenu}>
          <button 
            className={styles.userButton}
            onClick={() => setUserMenuOpen(!userMenuOpen)}
            aria-label={t('common.userMenu')}
          >
            <div className={styles.userAvatar}>
              {user?.avatar ? (
                <img src={user.avatar} alt={user.username} />
              ) : (
                <i className="fas fa-user" />
              )}
            </div>
            <div className={styles.userInfo}>
              <span className={styles.username}>{user?.username}</span>
              <span className={styles.rating}>
                <i className="fas fa-star" />
                {user?.rating || 0}
              </span>
            </div>
            <i className={`fas fa-chevron-down ${userMenuOpen ? styles.chevronUp : ''}`} />
          </button>

          {userMenuOpen && (
            <div className={styles.userDropdown}>
              <div className={styles.userStats}>
                <div className={styles.stat}>
                  <span className={styles.statLabel}>{t('dashboard.stats.wins')}</span>
                  <span className={styles.statValue}>{user?.wins || 0}</span>
                </div>
                <div className={styles.stat}>
                  <span className={styles.statLabel}>{t('dashboard.stats.losses')}</span>
                  <span className={styles.statValue}>{user?.losses || 0}</span>
                </div>
                <div className={styles.stat}>
                  <span className={styles.statLabel}>{t('dashboard.stats.winrate')}</span>
                  <span className={styles.statValue}>
                    {user?.wins && user?.losses 
                      ? Math.round((user.wins / (user.wins + user.losses)) * 100) 
                      : 0}%
                  </span>
                </div>
              </div>
              
              <div className={styles.menuDivider} />
              
              <button 
                className={styles.menuItem}
                onClick={navigateToProfile}
              >
                <i className="fas fa-user" />
                {t('navigation.profile')}
              </button>
              
              <button 
                className={styles.menuItem}
                onClick={navigateToSettings}
              >
                <i className="fas fa-cog" />
                {t('navigation.settings')}
              </button>
              
              <div className={styles.menuDivider} />
              
              <button 
                className={`${styles.menuItem} ${styles.logoutItem}`}
                onClick={handleLogout}
              >
                <i className="fas fa-sign-out-alt" />
                {t('auth.logout')}
              </button>
            </div>
          )}
        </div>
      </div>

      {/* Click outside to close menu */}
      {userMenuOpen && (
        <div 
          className={styles.overlay}
          onClick={() => setUserMenuOpen(false)}
        />
      )}
    </header>
  )
}