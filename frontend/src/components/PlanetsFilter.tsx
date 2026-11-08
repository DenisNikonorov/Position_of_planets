import type { FC } from 'react'
import { Button } from 'react-bootstrap'
import { useAppDispatch, useAppSelector } from '../hooks'
import { getPlanetsList, setAppliedMaxDist, setMaxDist } from '../store/planetsSlice'

interface Props {
  loading: boolean
}

// Значение фильтра хранится в Redux (planets.maxDist):
// ползунок диспатчит setMaxDist, «Найти» — getPlanetsList.
const PlanetsFilter: FC<Props> = ({ loading }) => {
  const dispatch = useAppDispatch()
  const maxDistance = useAppSelector((state) => state.planets.maxDist)

  return (
    <div className="filter-block">
      <div className="filter-header">
        <span className="filter-label">Макс. расстояние (AU)</span>
        <span className="filter-value">{maxDistance}</span>
      </div>
      <div className="filter-form">
        <input
          type="range"
          className="filter-slider"
          min={0}
          max={12}
          step={0.1}
          value={maxDistance}
          disabled={loading}
          onChange={(event) => dispatch(setMaxDist(Number(event.target.value)))}
        />
        <Button
          type="button"
          className="filter-btn"
          disabled={loading}
          onClick={() => {
            // Сначала фиксируем применённое значение в Redux,
            // затем getPlanetsList читает его из состояния.
            dispatch(setAppliedMaxDist(maxDistance))
            dispatch(getPlanetsList())
          }}
        >
          Найти
        </Button>
      </div>
      <div className="filter-marks">
        <span>0</span>
        <span>6</span>
        <span>12</span>
      </div>
    </div>
  )
}

export default PlanetsFilter
