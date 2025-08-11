// Login Page с улучшенным переключателем языков
import { useState } from 'react'
import { Link, useNavigate, useLocation } from 'react-router-dom'
import { useAuth } from '@hooks/useAuth'
import { useI18n } from '@hooks/useI18n'
import { LanguageSwitcher } from '@components/common/LanguageSwitcher'
import styles from './Login.module.css'

export default function Login() {
  const [formData, setFormData] = useState({
    username: '',
    password: ''
  })
  
  const { login, isLoading, error } = useAuth()
  const { t } = useI18n()
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
          {/* Language Switcher */}
          <div className={styles.languageSwitcher}>
            <LanguageSwitcher variant="buttons" size="small" />
          </div>

          <div className={styles.header}>
            <h1 className={styles.title}>ZZZ Tournament</h1>
            <p className={styles.subtitle}>{t('auth.loginTitle')}</p>
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
                className={styles.input}
                disabled={isLoading}
              />
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
                className={styles.input}
                disabled={isLoading}
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
              {isLoading ? (
                <>
                  <div className={styles.spinner} />
                  {t('auth.loggingIn')}
                </>
              ) : (
                t('auth.login')
              )}
            </button>
          </form>

          <div className={styles.footer}>
            <p>
              {t('auth.noAccount')}{' '}
              <Link to="/register" className={styles.link}>
                {t('auth.register')}
              </Link>
            </p>
            
            <Link to="/forgot-password" className={styles.forgotLink}>
              {t('auth.forgotPassword')}
            </Link>
          </div>
        </div>
      </div>
    </div>
  )
}