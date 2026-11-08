import type { FC } from 'react'
import { useNavigate } from 'react-router-dom'
import type { Planet } from '../modules/planets'
import { tileMedia } from '../modules/planets'
import { deletePlanet, likePlanet } from '../store/planetsSlice'
import { useAppDispatch, useAppSelector } from '../hooks'

interface Props {
  planet: Planet
}

const PlanetsCard: FC<Props> = ({ planet }) => {
  const dispatch = useAppDispatch()
  const navigate = useNavigate()
  const isAuthenticated = useAppSelector((state) => state.user.isAuthenticated)
  const media = tileMedia(planet)
  const liked = planet.liked === 1

  const handleLike = (event: React.MouseEvent) => {
    event.preventDefault()
    event.stopPropagation()
    if (!isAuthenticated) return
    dispatch(likePlanet({ id: planet.id, like: liked ? 0 : 1 }))
  }

  const handleDelete = (event: React.MouseEvent) => {
    event.preventDefault()
    event.stopPropagation()
    dispatch(deletePlanet(planet.id))
  }

  return (
    <div className="tile-card">
      {planet.is_owner === 1 && (
        <div className="tile-delete">
          <button type="button" className="filter-btn" onClick={handleDelete}>
            Удалить
          </button>
        </div>
      )}
      <a
        href={`/feed?id=${planet.id}`}
        className="tile-card-link"
        onClick={(event) => {
          // SPA-переход без перезагрузки: иначе Redux сбросился бы в гостя.
          event.preventDefault()
          navigate(`/feed?id=${planet.id}`)
        }}
      >
        <div className="tile-card-media">
          {media.type === 'video' ? (
            <video autoPlay muted loop playsInline>
              <source src={media.src} type="video/mp4" />
            </video>
          ) : (
            <img src={media.src} alt={planet.name} />
          )}
        </div>
        <div className="tile-card-info">
          <div className="tile-card-name">{planet.name}</div>
          {planet.observation_date && <div className="tile-card-date">{planet.observation_date}</div>}
          <div className="tile-card-meta">
            <button
              type="button"
              className={`tile-card-likes ${liked ? 'liked' : ''}`}
              disabled={!isAuthenticated}
              onClick={handleLike}
              aria-label={liked ? 'Снять лайк' : 'Поставить лайк'}
            >
              <svg viewBox="0 0 24 24">
                <path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z" />
              </svg>
              {planet.likes_count}
            </button>
            <span className="tile-card-distance">{planet.distance_from_sun} AU</span>
          </div>
        </div>
      </a>
    </div>
  )
}

export default PlanetsCard
