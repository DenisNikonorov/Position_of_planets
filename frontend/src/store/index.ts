import { configureStore } from '@reduxjs/toolkit'
import planetsReducer from './planetsSlice'
import userReducer from './userSlice'

// Redux DevTools подключены по умолчанию (для показа состояний в лабораторной).
export const store = configureStore({
  reducer: {
    planets: planetsReducer,
    user: userReducer,
  },
})

export type RootState = ReturnType<typeof store.getState>
export type AppDispatch = typeof store.dispatch
