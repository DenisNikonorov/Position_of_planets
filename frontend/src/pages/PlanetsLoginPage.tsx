import type { FC } from 'react'
import { useState, type FormEvent } from 'react'
import { useNavigate } from 'react-router-dom'
import axios from 'axios'
import { setAuthToken } from '../api'
import PlanetsLoading from '../components/PlanetsLoading'
import { useAppDispatch } from '../hooks'
import { setToken } from '../lib/token'
import { login } from '../store/userSlice'

// Вход: ровно два поля и кнопка. Авторизация идёт ручным axios
// (не через кодогенерацию и не через thunk), затем plain action в Redux.
const PlanetsLoginPage: FC = () => {
  const dispatch = useAppDispatch()
  const navigate = useNavigate()

  const [loginValue, setLoginValue] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)

  const handleSubmit = async (event: FormEvent) => {
    event.preventDefault()
    setBusy(true)
    setError('')
    try {
      const response = await axios.post('/api/users/auth', {
        login: loginValue,
        password,
      })
      const token = response.data.access_token as string
      setToken(token)
      setAuthToken(token)
      dispatch(login({ login: loginValue }))
      navigate('/')
    } catch {
      setError('Неверный логин или пароль')
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="add-container">
      <form onSubmit={(event) => void handleSubmit(event)}>
        <div className="form-group">
          <label className="add-label" htmlFor="loginField">
            Логин
          </label>
          <input
            id="loginField"
            type="text"
            className="form-input"
            value={loginValue}
            onChange={(event) => setLoginValue(event.target.value)}
            required
          />
        </div>

        <div className="form-group">
          <label className="add-label" htmlFor="passwordField">
            Пароль
          </label>
          <input
            id="passwordField"
            type="password"
            className="form-input"
            value={password}
            onChange={(event) => setPassword(event.target.value)}
            required
          />
        </div>

        {error && <div className="form-error">{error}</div>}

        <button type="submit" className="step-btn" disabled={busy}>
          {busy ? 'Вход...' : 'Войти'}
        </button>
      </form>

      {busy && <PlanetsLoading />}
    </div>
  )
}

export default PlanetsLoginPage
