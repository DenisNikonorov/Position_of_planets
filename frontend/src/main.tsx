import React from 'react'
import ReactDOM from 'react-dom/client'
import { Provider } from 'react-redux'
import { createBrowserRouter, RouterProvider } from 'react-router-dom'
import 'bootstrap/dist/css/bootstrap.min.css'
import './planets.css'
import App from './App.tsx'
import { store } from './store'
import PlanetsFeedPage from './pages/PlanetsFeedPage.tsx'
import PlanetsTilePage from './pages/PlanetsTilePage.tsx'
import PlanetsAddPage from './pages/PlanetsAddPage.tsx'
import PlanetsLoginPage from './pages/PlanetsLoginPage.tsx'

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
      {
        path: 'feed',
        element: <PlanetsFeedPage />,
      },
      {
        path: 'add',
        element: <PlanetsAddPage />,
      },
      {
        path: 'login',
        element: <PlanetsLoginPage />,
      },
    ],
  },
])

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <Provider store={store}>
      <RouterProvider router={router} />
    </Provider>
  </React.StrictMode>,
)
