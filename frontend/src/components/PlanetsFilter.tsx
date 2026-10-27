import type { FC } from 'react'
import { Button } from 'react-bootstrap'

interface Props {
  maxDistance: number
  setMaxDistance: (value: number) => void
  onApply: () => void
}

const PlanetsFilter: FC<Props> = ({ maxDistance, setMaxDistance, onApply }) => (
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
        onChange={(event) => setMaxDistance(Number(event.target.value))}
      />
      <Button type="button" className="filter-btn" onClick={onApply}>
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

export default PlanetsFilter
