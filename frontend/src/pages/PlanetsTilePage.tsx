import type { FC } from 'react'
import { useState, useMemo } from 'react'
import { PLANETS_MOCK } from '../modules/planetsMock'
import PlanetsFilter from '../components/PlanetsFilter'
import PlanetsCard from '../components/PlanetsCard'

const PlanetsTilePage: FC = () => {
  const [maxDistance, setMaxDistance] = useState(12)
  const [appliedMaxDistance, setAppliedMaxDistance] = useState(1000)

  const handleApply = () => setAppliedMaxDistance(maxDistance)

  const filteredPlanets = useMemo(
    () => PLANETS_MOCK.filter((planet) => planet.distanceFromSun <= appliedMaxDistance),
    [appliedMaxDistance],
  )

  return (
    <div className="tile-container">
      <PlanetsFilter maxDistance={maxDistance} setMaxDistance={setMaxDistance} onApply={handleApply} />
      <div className="tiles-grid">
        {filteredPlanets.map((planet) => (
          <PlanetsCard key={planet.id} planet={planet} />
        ))}
      </div>
      {filteredPlanets.length === 0 && (
        <p style={{ color: '#888', textAlign: 'center', marginTop: 24 }}>Ничего не найдено</p>
      )}
    </div>
  )
}

export default PlanetsTilePage
