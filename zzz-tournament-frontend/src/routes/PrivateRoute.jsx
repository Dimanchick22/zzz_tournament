// Private Route Component
import { Navigate, useLocation } from 'react-router-dom'
import { useAuthStore } from '@store/authStore'

export const PrivateRoute = ({ children }) => {
  const { isAuthenticated, isLoading } = useAuthStore()
  const location = useLocation()

  // Показываем лоадер пока проверяем аутентификацию
  if (isLoading) {
    return (
      <div className="app-loading">
        <div className="app-loading-spinner" />
        <p>Проверяем авторизацию...</p>
      </div>
    )
  }

  // Если не аутентифицирован, редиректим на логин
  if (!isAuthenticated) {
    return (
      <Navigate 
        to="/login" 
        state={{ from: location }} 
        replace 
      />
    )
  }

  // Если всё ок, показываем дочерние компоненты
  return children
}