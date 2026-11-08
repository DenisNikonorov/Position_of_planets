import type { FC } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import axios from 'axios'
import { setAuthToken } from '../api'
import { useAppDispatch, useAppSelector } from '../hooks'
import { clearToken, getToken } from '../lib/token'
import { getFeed, getPlanetsList, resetFilter } from '../store/planetsSlice'
import { logout } from '../store/userSlice'

const PlanetsNavbar: FC = () => {
  const dispatch = useAppDispatch()
  const navigate = useNavigate()
  const location = useLocation()
  const isAuthenticated = useAppSelector((state) => state.user.isAuthenticated)

  const isFeed = location.pathname === '/'
  const isTile = location.pathname === '/tile'
  const isAdd = location.pathname === '/add'
  const isLogin = location.pathname === '/login'

  // Выход: deauth через ручной axios (Authorization берётся из localStorage),
  // затем токен удаляется, Redux переводится в гостя, фильтр сбрасывается.
  const handleLogout = async () => {
    try {
      await axios.post('/api/users/deauth', null, {
        headers: { Authorization: `Bearer ${getToken()}` },
      })
    } catch (error) {
      console.warn('navbar: logout не прошёл, токен уже недействителен', error)
    }
    setAuthToken(null)
    clearToken()
    dispatch(logout())
    dispatch(resetFilter())
    // Список и лента перезапрашиваются уже как гость (без токена).
    dispatch(getPlanetsList())
    dispatch(getFeed())
    navigate('/')
  }

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
      <a
        href="/add"
        className={`nav-item ${isAdd ? 'active' : ''} ${isAuthenticated ? '' : 'disabled'}`}
        onClick={(event) => {
          event.preventDefault()
          if (isAuthenticated) navigate('/add')
        }}
      >
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
      {isAuthenticated ? (
        <a
          href="/logout"
          className="nav-item"
          onClick={(event) => {
            event.preventDefault()
            void handleLogout()
          }}
        >
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4" />
            <polyline points="16 17 21 12 16 7" />
            <line x1="21" y1="12" x2="9" y2="12" />
          </svg>
          Выход
        </a>
      ) : (
        <a
          href="/login"
          className={`nav-item ${isLogin ? 'active' : ''}`}
          onClick={(event) => {
            event.preventDefault()
            navigate('/login')
          }}
        >
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2" />
            <circle cx="12" cy="7" r="4" />
          </svg>
          Вход
        </a>
      )}
    </nav>
  )
}

export default PlanetsNavbar
