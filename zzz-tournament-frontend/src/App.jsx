import { useEffect } from 'react'
import { Routes, Route, Navigate } from 'react-router-dom'

// Stores
import { useAuthStore } from '@store/authStore'
import { useUIStore } from '@store/uiStore'

// Hooks
import { useAuth } from '@hooks/useAuth'

// Components
import { Layout } from '@components/layout'
import { PrivateRoute, PublicRoute } from '@/routes'

// Pages
import Home from '@pages/Home'
import Login from '@pages/Login'
import Register from '@pages/Register'
import Dashboard from '@pages/Dashboard'
import Rooms from '@pages/Rooms'
import RoomDetails from '@pages/RoomDetails'
import Tournament from '@pages/Tournament'
import Heroes from '@pages/Heroes'
import Leaderboard from '@pages/Leaderboard'
import Profile from '@pages/Profile'
import NotFound from '@pages/NotFound'

// Styles
import './App.css'

function App() {
  const { isAuthenticated, isLoading, user } = useAuthStore()
  const { theme, initializeTheme } = useUIStore()
  const { checkAuth } = useAuth()

  // Инициализация приложения
  useEffect(() => {
    const initApp = async () => {
      // Инициализируем тему
      initializeTheme()
      
      // Проверяем аутентификацию при загрузке
      await checkAuth()
    }

    initApp()
  }, [checkAuth, initializeTheme])

  // Применяем тему к документу
  useEffect(() => {
    document.documentElement.setAttribute('data-theme', theme)
    document.documentElement.className = `theme-${theme}`
  }, [theme])

  // Показываем лоадер пока идет инициализация
  if (isLoading) {
    return (
      <div className="app-loading">
        <div className="app-loading-spinner" />
        <p>Загружаем ZZZ Tournament...</p>
      </div>
    )
  }

  return (
    <div className="app" data-theme={theme}>
      <Routes>
        {/* Публичные маршруты */}
        <Route 
          path="/login" 
          element={
            <PublicRoute>
              <Login />
            </PublicRoute>
          } 
        />
        <Route 
          path="/register" 
          element={
            <PublicRoute>
              <Register />
            </PublicRoute>
          } 
        />

        {/* Защищенные маршруты с Layout */}
        <Route 
          path="/" 
          element={
            <PrivateRoute>
              <Layout />
            </PrivateRoute>
          }
        >
          {/* Главная страница */}
          <Route index element={<Home />} />
          
          {/* Дашборд */}
          <Route path="dashboard" element={<Dashboard />} />
          
          {/* Комнаты */}
          <Route path="rooms" element={<Rooms />} />
          <Route path="rooms/:id" element={<RoomDetails />} />
          
          {/* Турниры */}
          <Route path="tournament/:id" element={<Tournament />} />
          
          {/* Герои */}
          <Route path="heroes" element={<Heroes />} />
          
          {/* Лидерборд */}
          <Route path="leaderboard" element={<Leaderboard />} />
          
          {/* Профиль */}
          <Route path="profile" element={<Profile />} />
        </Route>

        {/* Редирект для неаутентифицированных пользователей */}
        <Route 
          path="/" 
          element={
            !isAuthenticated ? (
              <Navigate to="/login" replace />
            ) : (
              <Navigate to="/dashboard" replace />
            )
          } 
        />

        {/* 404 страница */}
        <Route path="*" element={<NotFound />} />
      </Routes>
    </div>
  )
}

export default App