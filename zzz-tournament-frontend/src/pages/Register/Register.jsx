// Register Page
import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useAuth } from '@hooks/useAuth'
import styles from './Register.module.css'

export default function Register() {
  const [formData, setFormData] = useState({
    username: '',
    email: '',
    password: '',
    confirmPassword: ''
  })
  
  const [formErrors, setFormErrors] = useState({})
  const { register, isLoading, error } = useAuth()
  const navigate = useNavigate()

  const handleChange = (e) => {
    setFormData({
      ...formData,
      [e.target.name]: e.target.value
    })
    
    // Очищаем ошибку поля при изменении
    if (formErrors[e.target.name]) {
      setFormErrors({
        ...formErrors,
        [e.target.name]: ''
      })
    }
  }

  const validateForm = () => {
    const errors = {}
    
    // Валидация username
    if (!formData.username) {
      errors.username = 'Имя пользователя обязательно'
    } else if (formData.username.length < 3) {
      errors.username = 'Имя пользователя должно быть не менее 3 символов'
    } else if (!/^[a-zA-Z0-9_-]+$/.test(formData.username)) {
      errors.username = 'Имя пользователя может содержать только буквы, цифры, _ и -'
    }
    
    // Валидация email
    if (!formData.email) {
      errors.email = 'Email обязателен'
    } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(formData.email)) {
      errors.email = 'Введите корректный email'
    }
    
    // Валидация password
    if (!formData.password) {
      errors.password = 'Пароль обязателен'
    } else if (formData.password.length < 6) {
      errors.password = 'Пароль должен быть не менее 6 символов'
    }
    
    // Валидация confirmPassword
    if (!formData.confirmPassword) {
      errors.confirmPassword = 'Подтверждение пароля обязательно'
    } else if (formData.password !== formData.confirmPassword) {
      errors.confirmPassword = 'Пароли не совпадают'
    }
    
    return errors
  }

  const handleSubmit = async (e) => {
    e.preventDefault()
    
    const errors = validateForm()
    
    if (Object.keys(errors).length > 0) {
      setFormErrors(errors)
      return
    }
    
    setFormErrors({})
    
    const result = await register({
      username: formData.username,
      email: formData.email,
      password: formData.password
    })
    
    if (result.success) {
      navigate('/dashboard')
    }
  }

  return (
    <div className={styles.registerPage}>
      <div className={styles.registerContainer}>
        <div className={styles.registerCard}>
          <div className={styles.header}>
            <h1 className={styles.title}>ZZZ Tournament</h1>
            <p className={styles.subtitle}>Создание аккаунта</p>
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
                className={`${styles.input} ${formErrors.username ? styles.inputError : ''}`}
              />
              {formErrors.username && (
                <span className={styles.fieldError}>{formErrors.username}</span>
              )}
            </div>

            <div className={styles.field}>
              <label htmlFor="email">Email</label>
              <input
                type="email"
                id="email"
                name="email"
                value={formData.email}
                onChange={handleChange}
                required
                placeholder="Введите email"
                className={`${styles.input} ${formErrors.email ? styles.inputError : ''}`}
              />
              {formErrors.email && (
                <span className={styles.fieldError}>{formErrors.email}</span>
              )}
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
                className={`${styles.input} ${formErrors.password ? styles.inputError : ''}`}
              />
              {formErrors.password && (
                <span className={styles.fieldError}>{formErrors.password}</span>
              )}
            </div>

            <div className={styles.field}>
              <label htmlFor="confirmPassword">Подтверждение пароля</label>
              <input
                type="password"
                id="confirmPassword"
                name="confirmPassword"
                value={formData.confirmPassword}
                onChange={handleChange}
                required
                placeholder="Повторите пароль"
                className={`${styles.input} ${formErrors.confirmPassword ? styles.inputError : ''}`}
              />
              {formErrors.confirmPassword && (
                <span className={styles.fieldError}>{formErrors.confirmPassword}</span>
              )}
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
              {isLoading ? 'Создание аккаунта...' : 'Создать аккаунт'}
            </button>
          </form>

          <div className={styles.footer}>
            <p>
              Уже есть аккаунт?{' '}
              <Link to="/login" className={styles.link}>
                Войти
              </Link>
            </p>
          </div>
        </div>
      </div>
    </div>
  )
}