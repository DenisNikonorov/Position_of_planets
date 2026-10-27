import type { FC } from 'react'
import { useState, useEffect } from 'react'
import { PLANETS_MOCK } from '../modules/planetsMock'

const PlanetsFeedPage: FC = () => {
  const [index, setIndex] = useState(0)
  const [likes, setLikes] = useState<Record<number, boolean>>({})

  const planet = PLANETS_MOCK[index]
  const liked = !!likes[planet.id]
  const likesCount = planet.likesCount + (liked ? 1 : 0)

  const goNext = () => setIndex((i) => (i + 1) % PLANETS_MOCK.length)
  const toggleLike = () => setLikes((state) => ({ ...state, [planet.id]: !state[planet.id] }))

  useEffect(() => {
    const onKey = (event: KeyboardEvent) => {
      if (event.key === 'ArrowDown') goNext()
      if (event.key === 'ArrowUp') setIndex((i) => (i - 1 + PLANETS_MOCK.length) % PLANETS_MOCK.length)
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [])

  return (
    <div className="feed-container">
      <video className="feed-media" autoPlay muted loop playsInline key={planet.id}>
        <source src={planet.urlVideo || '/media/default_video.mp4'} type="video/mp4" />
      </video>

      <div className="feed-sidebar">
        <div className="feed-like-wrap">
          <button type="button" className={`like-btn ${liked ? 'liked' : ''}`} onClick={toggleLike}>
            <svg viewBox="0 0 24 24">
              <path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z" />
            </svg>
          </button>
          <span className="feed-likes-count">{likesCount}</span>
        </div>
        <div className="feed-stats">
          <span>{planet.distanceFromSun} AU</span>
          <div className="feed-date">
            <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" strokeWidth="2">
              <rect x="3" y="4" width="18" height="18" rx="2" />
              <line x1="16" y1="2" x2="16" y2="6" />
              <line x1="8" y1="2" x2="8" y2="6" />
              <line x1="3" y1="10" x2="21" y2="10" />
            </svg>
            {planet.observationDate}
          </div>
        </div>
        <button type="button" className="next-btn" onClick={goNext} aria-label="Далее">
          <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
            <polyline points="6 9 12 15 18 9" />
          </svg>
        </button>
      </div>

      <div className="feed-description">
        <div className="planet-name">{planet.name}</div>
        <input type="checkbox" id="descToggle" className="desc-checkbox" />
        <div className="desc-text" id="descText">
          {planet.description}
        </div>
        <label htmlFor="descToggle" className="desc-toggle"></label>
      </div>
    </div>
  )
}

export default PlanetsFeedPage
