import { createSlice, type PayloadAction } from '@reduxjs/toolkit'

export interface UserState {
  login: string | null
  isAuthenticated: boolean
}

const initialState: UserState = {
  login: null,
  isAuthenticated: false,
}

// Redux хранит только факт авторизации: после F5 состояние сбрасывается
// в гостя, а токен остаётся в localStorage (см. lib/token.ts).
const userSlice = createSlice({
  name: 'user',
  initialState,
  reducers: {
    login(state, action: PayloadAction<{ login: string }>) {
      state.login = action.payload.login
      state.isAuthenticated = true
    },
    logout(state) {
      state.login = null
      state.isAuthenticated = false
    },
  },
})

export const { login, logout } = userSlice.actions
export default userSlice.reducer
