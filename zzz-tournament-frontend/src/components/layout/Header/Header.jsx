// Header Component
import { useState } from 'react'
import { useAuthStore } from '@store/authStore'
import { useUIStore } from '@store/uiStore'
import { useAuth } from '@hooks/useAuth'
import styles from './Header.module.css'

export const Header = () => {
  const { user } = useAuthStore()
  const { theme, toggleTheme, toggleSidebar, isMobile } = useUIStore()
  const { logout } = useAuth()
  const [userMenuOpen, setUserMenuOpen] = useState(false)

  const handleLogout = () => {
    logout()
    setUserMenuOpen(false)
  }

  return (
    <header className={styles.header}>
      <div className={styles.left}>
        {/* Sidebar Toggle */}
        <button 
          className={styles.sidebarToggle}
          onClick={toggleSidebar}
          aria-label="Toggle sidebar"
        >
          <i className="fas fa-bars" />
        </button>
        
        {/* Logo/Title */}
        <div className={styles.logo}>
          <h1>ZZZ Tournament</h1>
        </div>
      </div>

      <div className={styles.right}>
        {/* Theme Toggle */}
        <button 
          className={styles.themeToggle}
          onClick={toggleTheme}
          aria-label="Toggle theme"
        >
          <i className={`fas ${theme === 'dark' ? 'fa-sun' : 'fa-moon'}`} />
        </button>

        {/* User Menu */}
        <div className={styles.userMenu}>
          <button 
            className={styles.userButton}
            onClick={() => setUserMenuOpen(!userMenuOpen)}
            aria-label="User menu"
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
                  <span className={styles.statLabel}>Побед</span>
                  <span className={styles.statValue}>{user?.wins || 0}</span>
                </div>
                <div className={styles.stat}>
                  <span className={styles.statLabel}>Поражений</span>
                  <span className={styles.statValue}>{user?.losses || 0}</span>
                </div>
                <div className={styles.stat}>
                  <span className={styles.statLabel}>Винрейт</span>
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
                onClick={() => {
                  setUserMenuOpen(false)
                  // Navigate to profile
                }}
              >
                <i className="fas fa-user" />
                Профиль
              </button>
              
              <button 
                className={styles.menuItem}
                onClick={() => {
                  setUserMenuOpen(false)
                  // Navigate to settings
                }}
              >
                <i className="fas fa-cog" />
                Настройки
              </button>
              
              <div className={styles.menuDivider} />
              
              <button 
                className={`${styles.menuItem} ${styles.logoutItem}`}
                onClick={handleLogout}
              >
                <i className="fas fa-sign-out-alt" />
                Выйти
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