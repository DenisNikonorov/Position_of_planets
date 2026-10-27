import type { FC } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'

const PlanetsNavbar: FC = () => {
  const navigate = useNavigate()
  const location = useLocation()

  const isFeed = location.pathname === '/'
  const isTile = location.pathname === '/tile'

  return (
    <nav className="bottom-nav">
      <a
        href="/"
        className={`nav-item ${isFeed ? 'active' : ''}`}
        onClick={(event) => {
          event.preventDefault()
          navigate('/')
        }}
      >
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
          <circle cx="12" cy="12" r="10" />
          <circle cx="12" cy="12" r="3" />
          <line x1="12" y1="2" x2="12" y2="5" />
          <line x1="12" y1="19" x2="12" y2="22" />
          <line x1="2" y1="12" x2="5" y2="12" />
          <line x1="19" y1="12" x2="22" y2="12" />
        </svg>
        Лента
      </a>
      <a href="/add" className="nav-item disabled" onClick={(event) => event.preventDefault()}>
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
          <circle cx="12" cy="12" r="10" />
          <line x1="12" y1="8" x2="12" y2="16" />
          <line x1="8" y1="12" x2="16" y2="12" />
        </svg>
        Добавить
      </a>
      <a
        href="/tile"
        className={`nav-item ${isTile ? 'active' : ''}`}
        onClick={(event) => {
          event.preventDefault()
          navigate('/tile')
        }}
      >
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
          <rect x="3" y="3" width="7" height="7" rx="1" />
          <rect x="14" y="3" width="7" height="7" rx="1" />
          <rect x="3" y="14" width="7" height="7" rx="1" />
          <rect x="14" y="14" width="7" height="7" rx="1" />
        </svg>
        Плитка
      </a>
      <a href="/login" className="nav-item disabled" onClick={(event) => event.preventDefault()}>
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
          <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2" />
          <circle cx="12" cy="7" r="4" />
        </svg>
        Вход
      </a>
    </nav>
  )
}

export default PlanetsNavbar
