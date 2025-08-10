// Login Page
import { useState } from 'react'
import { Link, useNavigate, useLocation } from 'react-router-dom'
import { useAuth } from '@hooks/useAuth'
import styles from './Login.module.css'

export default function Login() {
  const [formData, setFormData] = useState({
    username: '',
    password: ''
  })
  
  const { login, isLoading, error } = useAuth()
  const navigate = useNavigate()
  const location = useLocation()
  
  const from = location.state?.from?.pathname || '/dashboard'

  const handleChange = (e) => {
    setFormData({
      ...formData,
      [e.target.name]: e.target.value
    })
  }

  const handleSubmit = async (e) => {
    e.preventDefault()
    
    const result = await login(formData)
    
    if (result.success) {
      navigate(from, { replace: true })
    }
  }

  return (
    <div className={styles.loginPage}>
      <div className={styles.loginContainer}>
        <div className={styles.loginCard}>
          <div className={styles.header}>
            <h1 className={styles.title}>ZZZ Tournament</h1>
            <p className={styles.subtitle}>Вход в систему</p>
          </div>

          <form onSubmit={handleSubmit} className={styles.form}>
            <div className={styles.field}>
              <label htmlFor="username">Имя пользователя</label>
              <input
                type="text"
                id="username"
                name="username"
                value={formData.username}
                onChange={handleChange}
                required
                placeholder="Введите имя пользователя"
                className={styles.input}
              />
            </div>

            <div className={styles.field}>
              <label htmlFor="password">Пароль</label>
              <input
                type="password"
                id="password"
                name="password"
                value={formData.password}
                onChange={handleChange}
                required
                placeholder="Введите пароль"
                className={styles.input}
              />
            </div>

            {error && (
              <div className={styles.error}>
                {error}
              </div>
            )}

            <button 
              type="submit" 
              disabled={isLoading}
              className={styles.submitButton}
            >
              {isLoading ? 'Вход...' : 'Войти'}
            </button>
          </form>

          <div className={styles.footer}>
            <p>
              Нет аккаунта?{' '}
              <Link to="/register" className={styles.link}>
                Зарегистрироваться
              </Link>
            </p>
          </div>
        </div>
      </div>
    </div>
  )
}