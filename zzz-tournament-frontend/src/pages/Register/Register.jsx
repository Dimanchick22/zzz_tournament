// Register Page с поддержкой переводов
import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useAuth } from '@hooks/useAuth'
import { useI18n } from '@hooks/useI18n'
import { LanguageSwitcher } from '@components/common/LanguageSwitcher'
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
  const { t } = useI18n()
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
      errors.username = t('validation.required')
    } else if (formData.username.length < 3) {
      errors.username = t('validation.minLength', { count: 3 })
    } else if (!/^[a-zA-Z0-9_-]+$/.test(formData.username)) {
      errors.username = t('validation.invalidUsername')
    }
    
    // Валидация email
    if (!formData.email) {
      errors.email = t('validation.required')
    } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(formData.email)) {
      errors.email = t('validation.email')
    }
    
    // Валидация password
    if (!formData.password) {
      errors.password = t('validation.required')
    } else if (formData.password.length < 6) {
      errors.password = t('validation.minLength', { count: 6 })
    }
    
    // Валидация confirmPassword
    if (!formData.confirmPassword) {
      errors.confirmPassword = t('validation.required')
    } else if (formData.password !== formData.confirmPassword) {
      errors.confirmPassword = t('validation.passwordMismatch')
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
          {/* Language Switcher */}
          <div className={styles.languageSwitcher}>
            <LanguageSwitcher variant="buttons" size="small" />
          </div>

          <div className={styles.header}>
            <h1 className={styles.title}>ZZZ Tournament</h1>
            <p className={styles.subtitle}>{t('auth.registerTitle')}</p>
          </div>

          <form onSubmit={handleSubmit} className={styles.form}>
            <div className={styles.field}>
              <label htmlFor="username">{t('auth.username')}</label>
              <input
                type="text"
                id="username"
                name="username"
                value={formData.username}
                onChange={handleChange}
                required
                placeholder={t('auth.username')}
                className={`${styles.input} ${formErrors.username ? styles.inputError : ''}`}
                disabled={isLoading}
              />
              <small className={styles.hint}>{t('auth.usernameHint')}</small>
              {formErrors.username && (
                <span className={styles.fieldError}>{formErrors.username}</span>
              )}
            </div>

            <div className={styles.field}>
              <label htmlFor="email">{t('auth.email')}</label>
              <input
                type="email"
                id="email"
                name="email"
                value={formData.email}
                onChange={handleChange}
                required
                placeholder={t('auth.email')}
                className={`${styles.input} ${formErrors.email ? styles.inputError : ''}`}
                disabled={isLoading}
              />
              {formErrors.email && (
                <span className={styles.fieldError}>{formErrors.email}</span>
              )}
            </div>

            <div className={styles.field}>
              <label htmlFor="password">{t('auth.password')}</label>
              <input
                type="password"
                id="password"
                name="password"
                value={formData.password}
                onChange={handleChange}
                required
                placeholder={t('auth.password')}
                className={`${styles.input} ${formErrors.password ? styles.inputError : ''}`}
                disabled={isLoading}
              />
              <small className={styles.hint}>{t('auth.passwordHint')}</small>
              {formErrors.password && (
                <span className={styles.fieldError}>{formErrors.password}</span>
              )}
            </div>

            <div className={styles.field}>
              <label htmlFor="confirmPassword">{t('auth.confirmPassword')}</label>
              <input
                type="password"
                id="confirmPassword"
                name="confirmPassword"
                value={formData.confirmPassword}
                onChange={handleChange}
                required
                placeholder={t('auth.confirmPassword')}
                className={`${styles.input} ${formErrors.confirmPassword ? styles.inputError : ''}`}
                disabled={isLoading}
              />
              {formErrors.confirmPassword && (
                <span className={styles.fieldError}>{formErrors.confirmPassword}</span>
              )}
            </div>

            {/* Terms acceptance */}
            <div className={styles.termsField}>
              <label className={styles.checkboxLabel}>
                <input
                  type="checkbox"
                  required
                  disabled={isLoading}
                />
                <span className={styles.checkmark}></span>
                <span className={styles.termsText}>
                  {t('auth.termsAccept')}
                </span>
              </label>
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
              {isLoading ? (
                <>
                  <div className={styles.spinner} />
                  {t('auth.registering')}
                </>
              ) : (
                t('auth.createAccount')
              )}
            </button>
          </form>

          <div className={styles.footer}>
            <p>
              {t('auth.haveAccount')}{' '}
              <Link to="/login" className={styles.link}>
                {t('auth.login')}
              </Link>
            </p>
          </div>
        </div>
      </div>
    </div>
  )
}