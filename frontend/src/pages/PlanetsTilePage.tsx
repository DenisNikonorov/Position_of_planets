import type { FC } from 'react'
import { useEffect } from 'react'
import PlanetsCard from '../components/PlanetsCard'
import PlanetsFilter from '../components/PlanetsFilter'
import PlanetsLoading from '../components/PlanetsLoading'
import { useAppDispatch, useAppSelector } from '../hooks'
import { getPlanetsList } from '../store/planetsSlice'

const PlanetsTilePage: FC = () => {
  const dispatch = useAppDispatch()
  const planets = useAppSelector((state) => state.planets.planets)
  const loading = useAppSelector((state) => state.planets.loading)
  const error = useAppSelector((state) => state.planets.error)

  useEffect(() => {
    dispatch(getPlanetsList())
  }, [dispatch])

  return (
    <div className="tile-container">
      <PlanetsFilter loading={loading} />
      <div className="tiles-grid">
        {planets.map((planet) => (
          <PlanetsCard key={planet.id} planet={planet} />
        ))}
      </div>
      {planets.length === 0 && !loading && !error && (
        <p style={{ color: '#888', textAlign: 'center', marginTop: 24 }}>Ничего не найдено</p>
      )}
      {error && !loading && (
        <p className="form-error" style={{ marginTop: 24 }}>
          {error}
        </p>
      )}
      {loading && <PlanetsLoading />}
    </div>
  )
}

export default PlanetsTilePage
