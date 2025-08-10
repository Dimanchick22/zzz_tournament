// Public Route Component
import { Navigate } from 'react-router-dom'
import { useAuthStore } from '@store/authStore'

export const PublicRoute = ({ children }) => {
  const { isAuthenticated, isLoading } = useAuthStore()

  // Показываем лоадер пока проверяем аутентификацию
  if (isLoading) {
    return (
      <div className="app-loading">
        <div className="app-loading-spinner" />
        <p>Загружаем...</p>
      </div>
    )
  }

  // Если уже аутентифицирован, редиректим на дашборд
  if (isAuthenticated) {
    return <Navigate to="/dashboard" replace />
  }

  // Если не аутентифицирован, показываем публичный контент
  return children
}