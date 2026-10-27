import type { FC } from 'react'
import { Outlet } from 'react-router-dom'
import PlanetsNavbar from './components/PlanetsNavbar'

const App: FC = () => (
  <>
    <Outlet />
    <PlanetsNavbar />
  </>
)

export default App
