import { Api } from './Api'

// Токен подставляется во ВСЕ запросы (включая публичные list/feed:
// бэкенд считает по нему liked/is_owner через OptionalAuthCheck).
// После F5 модуль перезагружается — authToken пуст, хотя JWT остаётся
// в localStorage (его видно в Application → Local Storage).
let authToken: string | null = null

export function setAuthToken(token: string | null): void {
  authToken = token
}

// Клиент, сгенерированный из swagger-спецификации бэкенда (npm run generate-api).
// baseURL "/api" уходит в Vite-прокси на http://localhost:8080.
export const api = new Api({
  baseURL: '/api',
})

api.instance.interceptors.request.use((config) => {
  if (authToken && !config.headers.Authorization) {
    config.headers.Authorization = `Bearer ${authToken}`
  }
  return config
})
