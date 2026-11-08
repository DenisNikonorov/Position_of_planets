import type { FC } from 'react'
import { useCallback, useEffect, useState } from 'react'
import { useLocation } from 'react-router-dom'
import PlanetsLoading from '../components/PlanetsLoading'
import { useAppDispatch, useAppSelector } from '../hooks'
import type { Planet } from '../modules/planets'
import { DEFAULT_VIDEO, feedVideoSrc } from '../modules/planets'
import { getFeed, likePlanet, setFeed } from '../store/planetsSlice'

// Видео ленты: меняет src при ошибке воспроизведения;
// key={feed.id} на месте вызова пересоздаёт компонент для новой планеты.
const FeedVideo: FC<{ planet: Planet }> = ({ planet }) => {
  const [src, setSrc] = useState(() => feedVideoSrc(planet))

  return (
    <video
      className="feed-media"
      autoPlay
      muted
      loop
      playsInline
      src={src}
      onError={() => {
        if (src !== DEFAULT_VIDEO) setSrc(DEFAULT_VIDEO)
      }}
    />
  )
}

const PlanetsFeedPage: FC = () => {
  const dispatch = useAppDispatch()
  const feed = useAppSelector((state) => state.planets.feed)
  const loading = useAppSelector((state) => state.planets.loading)
  const error = useAppSelector((state) => state.planets.error)
  const isAuthenticated = useAppSelector((state) => state.user.isAuthenticated)

  // Локальная история листания: назад отдаёт уже посещённые планеты
  // без лишних запросов (бэкенд умеет только next).
  const [history, setHistory] = useState<Planet[]>([])
  const [pos, setPos] = useState(-1)

  // Смена авторизации (вход/выход) сбрасывает локальную историю листания:
  // после logout лента уже перезапрошена как гость. Паттерн React —
  // корректировка состояния при рендере, без useEffect.
  const [prevAuth, setPrevAuth] = useState(isAuthenticated)
  if (prevAuth !== isAuthenticated) {
    setPrevAuth(isAuthenticated)
    setHistory([])
    setPos(-1)
  }

  const location = useLocation()

  useEffect(() => {
    // Клик по карточке ведёт на /feed?id=N — открываем конкретную планету;
    // без id (или при 404) бэкенд отдаёт первую.
    const rawId = new URLSearchParams(location.search).get('id')
    const feedId = rawId ? Number(rawId) : undefined
    let cancelled = false

    const load = async () => {
      const first = await dispatch(getFeed(feedId ? { id: feedId } : undefined))
      if (cancelled) return
      if (getFeed.fulfilled.match(first)) {
        setHistory([first.payload])
        setPos(0)
        return
      }
      if (feedId !== undefined) {
        const fallback = await dispatch(getFeed())
        if (cancelled) return
        if (getFeed.fulfilled.match(fallback)) {
          setHistory([fallback.payload])
          setPos(0)
        }
      }
    }

    void load()
    return () => {
      cancelled = true
    }
  }, [dispatch, location.search])

  useEffect(() => {
    if (feed) {
      window.scrollTo(0, 0)
    }
  }, [feed])

  const goNext = useCallback(async () => {
    if (loading || !feed) return
    if (pos < history.length - 1) {
      const nextPos = pos + 1
      setPos(nextPos)
      dispatch(setFeed(history[nextPos]))
      return
    }
    const action = await dispatch(getFeed({ id: feed.id, next: true }))
    if (getFeed.fulfilled.match(action)) {
      setHistory((items) => [...items, action.payload])
      setPos((current) => current + 1)
    }
  }, [loading, feed, pos, history, dispatch])

  const goBack = useCallback(() => {
    if (loading || pos <= 0) return
    const prevPos = pos - 1
    setPos(prevPos)
    dispatch(setFeed(history[prevPos]))
  }, [loading, pos, history, dispatch])

  const handleLike = useCallback(() => {
    if (!isAuthenticated || !feed) return
    const like = feed.liked === 1 ? 0 : 1
    dispatch(likePlanet({ id: feed.id, like })).then((action) => {
      if (likePlanet.fulfilled.match(action)) {
        const updated = action.payload
        setHistory((items) => items.map((item, index) => (index === pos ? updated : item)))
      }
    })
  }, [isAuthenticated, feed, pos, dispatch])

  useEffect(() => {
    const onKey = (event: KeyboardEvent) => {
      if (event.key === 'ArrowDown') void goNext()
      if (event.key === 'ArrowUp') goBack()
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [goNext, goBack])

  if (!feed) {
    if (error && !loading) {
      return (
        <div className="add-container">
          <div className="form-error">{error}</div>
        </div>
      )
    }
    return <PlanetsLoading />
  }

  const liked = feed.liked === 1

  return (
    <div className="feed-container">
      <FeedVideo key={feed.id} planet={feed} />

      <div className="feed-sidebar">
        <div className="feed-like-wrap">
          <button
            type="button"
            className={`like-btn ${liked ? 'liked' : ''}`}
            disabled={!isAuthenticated}
            onClick={handleLike}
          >
            <svg viewBox="0 0 24 24">
              <path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z" />
            </svg>
          </button>
          <span className="feed-likes-count">{feed.likes_count}</span>
        </div>
        <div className="feed-stats">
          <span>{feed.distance_from_sun} AU</span>
          <div className="feed-date">
            <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" strokeWidth="2">
              <rect x="3" y="4" width="18" height="18" rx="2" />
              <line x1="16" y1="2" x2="16" y2="6" />
              <line x1="8" y1="2" x2="8" y2="6" />
              <line x1="3" y1="10" x2="21" y2="10" />
            </svg>
            {feed.observation_date}
          </div>
        </div>
        <button
          type="button"
          className="next-btn"
          onClick={() => void goNext()}
          disabled={loading}
          aria-label="Далее"
        >
          <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
            <polyline points="6 9 12 15 18 9" />
          </svg>
        </button>
      </div>

      <div className="feed-description">
        <div className="planet-name">{feed.name}</div>
        <input type="checkbox" id="descToggle" className="desc-checkbox" />
        <div className="desc-text" id="descText">
          {feed.description}
        </div>
        <label htmlFor="descToggle" className="desc-toggle"></label>
      </div>

      {loading && <PlanetsLoading />}
    </div>
  )
}

export default PlanetsFeedPage
