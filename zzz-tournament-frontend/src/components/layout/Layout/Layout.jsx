// Main Layout Component
import { Outlet } from 'react-router-dom'
import { useUIStore } from '@store/uiStore'
import { Header } from '../Header'
import { Sidebar } from '../Sidebar'
import styles from './Layout.module.css'

export const Layout = () => {
  const { sidebarCollapsed, isMobile } = useUIStore()

  return (
    <div className={styles.layout}>
      {/* Header */}
      <Header />
      
      {/* Main Content Area */}
      <div className={styles.container}>
        {/* Sidebar */}
        <Sidebar />
        
        {/* Main Content */}
        <main 
          className={`${styles.main} ${
            sidebarCollapsed ? styles.mainExpanded : ''
          } ${
            isMobile ? styles.mainMobile : ''
          }`}
        >
          <div className={styles.content}>
            <Outlet />
          </div>
        </main>
      </div>
      
      {/* Mobile Overlay для закрытия sidebar */}
      {isMobile && !sidebarCollapsed && (
        <div 
          className={styles.overlay}
          onClick={() => useUIStore.getState().setSidebarCollapsed(true)}
        />
      )}
    </div>
  )
}