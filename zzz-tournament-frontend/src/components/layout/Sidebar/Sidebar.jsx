// Sidebar Component с поддержкой переводов
import { NavLink, useLocation } from 'react-router-dom'
import { useUIStore } from '@store/uiStore'
import { useI18n } from '@hooks/useI18n'
import { NAVIGATION } from '@config/routes'
import styles from './Sidebar.module.css'

export const Sidebar = () => {
  const { sidebarCollapsed, isMobile } = useUIStore()
  const { t } = useI18n()
  const location = useLocation()

  // Навигационные элементы с переводами
  const navigationItems = [
    {
      path: '/dashboard',
      label: t('navigation.dashboard'),
      icon: 'dashboard',
      exact: true
    },
    {
      path: '/rooms',
      label: t('navigation.rooms'),
      icon: 'users',
      badge: 'new'
    },
    {
      path: '/heroes',
      label: t('navigation.heroes'),
      icon: 'sword'
    },
    {
      path: '/leaderboard',
      label: t('navigation.leaderboard'),
      icon: 'trophy'
    }
  ]

  return (
    <aside 
      className={`${styles.sidebar} ${
        sidebarCollapsed ? styles.collapsed : ''
      } ${
        isMobile ? styles.mobile : ''
      }`}
    >
      <nav className={styles.nav}>
        <ul className={styles.navList}>
          {navigationItems.map((item) => (
            <li key={item.path} className={styles.navItem}>
              <NavLink
                to={item.path}
                className={({ isActive }) => 
                  `${styles.navLink} ${isActive ? styles.active : ''}`
                }
                title={item.label}
              >
                <div className={styles.navIcon}>
                  <i className={getIconClass(item.icon)} />
                </div>
                
                <span className={styles.navLabel}>
                  {item.label}
                </span>
                
                {item.badge && (
                  <span className={`${styles.badge} ${styles[`badge${item.badge}`]}`}>
                    {item.badge === 'new' ? t('common.new') : item.badge}
                  </span>
                )}
              </NavLink>
            </li>
          ))}
        </ul>
      </nav>
      
      {/* Footer */}
      <div className={styles.footer}>
        <div className={styles.footerContent}>
          <div className={styles.version}>
            <span>v1.0.0</span>
          </div>
        </div>
      </div>
    </aside>
  )
}

// Helper function для получения CSS класса иконки
const getIconClass = (iconName) => {
  const iconMap = {
    dashboard: 'fas fa-tachometer-alt',
    users: 'fas fa-users',
    sword: 'fas fa-sword',
    trophy: 'fas fa-trophy',
    user: 'fas fa-user',
    cog: 'fas fa-cog',
    home: 'fas fa-home',
    chart: 'fas fa-chart-bar',
    tournament: 'fas fa-trophy'
  }
  
  return iconMap[iconName] || 'fas fa-circle'
}