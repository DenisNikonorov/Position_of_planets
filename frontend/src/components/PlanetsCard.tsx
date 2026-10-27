import type { FC } from 'react'
import type { Planet } from '../modules/planets'

interface Props {
  planet: Planet
}

const PlanetsCard: FC<Props> = ({ planet }) => (
  <div className="tile-card">
    <a href={`/feed?id=${planet.id}`} className="tile-card-link">
      <div className="tile-card-media">
        {planet.hasVideo ? (
          <video autoPlay muted loop playsInline>
            <source src={planet.urlVideo || '/media/default_video.mp4'} type="video/mp4" />
          </video>
        ) : (
          <img src={planet.imageURL || '/media/default_photo.png'} alt={planet.name} />
        )}
      </div>
      <div className="tile-card-info">
        <div className="tile-card-name">{planet.name}</div>
        {planet.observationDate && <div className="tile-card-date">{planet.observationDate}</div>}
        <div className="tile-card-meta">
          <span className="tile-card-likes">
            <svg viewBox="0 0 24 24">
              <path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z" />
            </svg>
            {planet.likesCount}
          </span>
          <span className="tile-card-distance">{planet.distanceFromSun} AU</span>
        </div>
      </div>
    </a>
  </div>
)

export default PlanetsCard
