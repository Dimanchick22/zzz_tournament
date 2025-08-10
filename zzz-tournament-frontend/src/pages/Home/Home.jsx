// Home Page
import { Link } from 'react-router-dom'
import styles from './Home.module.css'

export default function Home() {
  return (
    <div className={styles.home}>
      <div className={styles.hero}>
        <div className={styles.heroContent}>
          <h1 className={styles.title}>
            ZZZ Tournament
          </h1>
          <p className={styles.subtitle}>
            Турнирная система для Zenless Zone Zero
          </p>
          <p className={styles.description}>
            Соревнуйтесь с игроками со всего мира, участвуйте в турнирах 
            и поднимайтесь в рейтинге!
          </p>
          
          <div className={styles.actions}>
            <Link to="/register" className={styles.primaryButton}>
              Начать играть
            </Link>
            <Link to="/login" className={styles.secondaryButton}>
              Войти
            </Link>
          </div>
        </div>
        
        <div className={styles.heroImage}>
          <div className={styles.placeholder}>
            <i className="fas fa-gamepad" />
            <span>ZZZ</span>
          </div>
        </div>
      </div>
      
      <div className={styles.features}>
        <div className={styles.feature}>
          <i className="fas fa-trophy" />
          <h3>Турниры</h3>
          <p>Участвуйте в турнирах на выбывание с другими игроками</p>
        </div>
        <div className={styles.feature}>
          <i className="fas fa-star" />
          <h3>Рейтинг</h3>
          <p>Зарабатывайте рейтинговые очки и поднимайтесь в лидерборде</p>
        </div>
        <div className={styles.feature}>
          <i className="fas fa-users" />
          <h3>Сообщество</h3>
          <p>Общайтесь с другими игроками в чате комнат</p>
        </div>
      </div>
    </div>
  )
}