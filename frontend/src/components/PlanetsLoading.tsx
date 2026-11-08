import type { FC } from 'react'

// Затемняющая анимация загрузки: блокирует все кнопки страницы,
// пока идёт запрос (pointer-events перехватывает клики).
const PlanetsLoading: FC = () => (
  <div className="page-loading" role="status" aria-label="Загрузка">
    <div className="page-spinner" />
  </div>
)

export default PlanetsLoading
