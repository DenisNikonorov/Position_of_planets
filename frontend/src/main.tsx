import React from 'react'
import ReactDOM from 'react-dom/client'
import { createBrowserRouter, RouterProvider } from 'react-router-dom'
import 'bootstrap/dist/css/bootstrap.min.css'
import './planets.css'
import App from './App.tsx'
import PlanetsFeedPage from './pages/PlanetsFeedPage.tsx'
import PlanetsTilePage from './pages/PlanetsTilePage.tsx'

const router = createBrowserRouter([
  {
    path: '/',
    element: <App />,
    children: [
      {
        index: true,
        element: <PlanetsFeedPage />,
      },
      {
        path: 'tile',
        element: <PlanetsTilePage />,
      },
    ],
  },
])

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <RouterProvider router={router} />
  </React.StrictMode>,
)
