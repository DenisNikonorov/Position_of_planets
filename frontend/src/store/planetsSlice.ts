import { createAsyncThunk, createSlice, type PayloadAction } from '@reduxjs/toolkit'
import { api } from '../api'
import type { Planet } from '../modules/planets'
import { PLANETS_MOCK } from '../modules/planetsMock'

export interface PlanetsState {
  // Значение слайдера фильтра (хранится в Redux).
  maxDist: number
  // Применённый фильтр: null — показываем все планеты.
  appliedMaxDist: number | null
  planets: Planet[]
  feed: Planet | null
  draft: Planet | null
  loading: boolean
  error: string | null
}

const initialState: PlanetsState = {
  maxDist: 12,
  appliedMaxDist: null,
  planets: [],
  feed: null,
  draft: null,
  // true до первого ответа: страница сразу показывает анимацию загрузки.
  loading: true,
  error: null,
}

interface ThunkConfig {
  state: { planets: PlanetsState; user: { isAuthenticated: boolean } }
}

const mockFiltered = (maxDist: number | null): Planet[] =>
  maxDist === null
    ? PLANETS_MOCK
    : PLANETS_MOCK.filter((planet) => Number(planet.distance_from_sun) <= maxDist)

// GET /api/planets — список с фильтром. Если бэкенд недоступен и это гость,
// возвращаем mock-коллекцию (см. лабу по Redux: rejected → mock + warning).
export const getPlanetsList = createAsyncThunk<Planet[], void, ThunkConfig>(
  'planets/getPlanetsList',
  async (_, { getState, rejectWithValue }) => {
    const { appliedMaxDist } = getState().planets
    try {
      const query = appliedMaxDist === null ? {} : { maxDist: appliedMaxDist }
      const response = await api.planets.planetsList(query)
      return response.data as Planet[]
    } catch (error) {
      if (!getState().user.isAuthenticated) {
        console.warn('planets: бэкенд недоступен, показаны mock-данные', error)
        return mockFiltered(appliedMaxDist)
      }
      return rejectWithValue('Не удалось получить список планет')
    }
  },
)

// GET /api/planets/feed — одна планета ленты; без аргументов — первая.
export const getFeed = createAsyncThunk<
  Planet,
  { id?: number; next?: boolean } | undefined,
  ThunkConfig
>('planets/getFeed', async (arg, { getState, rejectWithValue }) => {
  try {
    const response = await api.planets.feedList(
      arg?.id === undefined ? {} : { id: arg.id, next: arg.next ?? false },
    )
    return response.data as Planet
  } catch (error) {
    if (!getState().user.isAuthenticated) {
      console.warn('planets: бэкенд недоступен, показана mock-лента', error)
      let index = 0
      if (arg?.id !== undefined) {
        const current = PLANETS_MOCK.findIndex((planet) => planet.id === arg.id)
        if (current >= 0) {
          index = arg.next ? (current + 1) % PLANETS_MOCK.length : current
        }
      }
      return PLANETS_MOCK[index]
    }
    return rejectWithValue('Не удалось получить ленту')
  }
})

// POST /api/planets — черновик (шаг 1 публикации). Только авторизованный.
export const createPlanet = createAsyncThunk<
  Planet,
  { name: string; file?: File | null },
  ThunkConfig
>('planets/createPlanet', async ({ name, file }) => {
  const body: { name: string; image_url?: File; url_video?: File } = { name }
  if (file) {
    if (file.type.startsWith('image/')) {
      body.image_url = file
    } else {
      body.url_video = file
    }
  }
  const response = await api.planets.planetsCreate(body)
  return response.data as Planet
})

// GET /api/planets/draft — текущий черновик; 404 означает «черновика нет».
export const getDraft = createAsyncThunk<Planet | null, void, ThunkConfig>(
  'planets/getDraft',
  async () => {
    try {
      const response = await api.planets.draftList()
      return response.data as Planet
    } catch {
      return null
    }
  },
)

// PUT /api/planets/:id — публикация черновика (шаг 2).
export const publishPlanet = createAsyncThunk<
  Planet,
  { id: number; body: { description: string; distance_from_sun: number; perihelion: number; observation_date: string } },
  ThunkConfig
>('planets/publishPlanet', async ({ id, body }) => {
  const response = await api.planets.planetsUpdate(id, body)
  return response.data as Planet
})

// POST /api/planets/:id/like — 1 ставит лайк, 0 снимает.
export const likePlanet = createAsyncThunk<
  Planet,
  { id: number; like: 0 | 1 },
  ThunkConfig
>('planets/likePlanet', async ({ id, like }) => {
  const response = await api.planets.likeCreate(id, { like })
  return response.data as Planet
})

// DELETE /api/planets/:id — мягкое удаление своей планеты.
export const deletePlanet = createAsyncThunk<number, number, ThunkConfig>(
  'planets/deletePlanet',
  async (id) => {
    await api.planets.planetsDelete(id)
    return id
  },
)

const planetsSlice = createSlice({
  name: 'planets',
  initialState,
  reducers: {
    setMaxDist(state, action: PayloadAction<number>) {
      state.maxDist = action.payload
    },
    setAppliedMaxDist(state, action: PayloadAction<number>) {
      state.appliedMaxDist = action.payload
    },
    resetFilter(state) {
      state.maxDist = 12
      state.appliedMaxDist = null
    },
    setFeed(state, action: PayloadAction<Planet>) {
      state.feed = action.payload
    },
    clearError(state) {
      state.error = null
    },
  },
  extraReducers: (builder) => {
    builder
      // Любой запрос включает анимацию, блокирующую кнопки пользователя.
      .addCase(getPlanetsList.pending, (state) => {
        state.loading = true
      })
      .addCase(getFeed.pending, (state) => {
        state.loading = true
      })
      .addCase(createPlanet.pending, (state) => {
        state.loading = true
      })
      .addCase(getDraft.pending, (state) => {
        state.loading = true
      })
      .addCase(publishPlanet.pending, (state) => {
        state.loading = true
      })
      .addCase(likePlanet.pending, (state) => {
        state.loading = true
      })
      .addCase(deletePlanet.pending, (state) => {
        state.loading = true
      })

      .addCase(getPlanetsList.fulfilled, (state, action) => {
        state.loading = false
        state.planets = action.payload
        state.error = null
      })
      .addCase(getPlanetsList.rejected, (state, action) => {
        state.loading = false
        state.error = (action.payload as string) ?? 'Ошибка загрузки'
      })
      .addCase(getFeed.fulfilled, (state, action) => {
        state.loading = false
        state.feed = action.payload
        state.error = null
      })
      .addCase(getFeed.rejected, (state, action) => {
        state.loading = false
        state.error = (action.payload as string) ?? 'Ошибка загрузки ленты'
      })
      .addCase(createPlanet.fulfilled, (state, action) => {
        state.loading = false
        state.draft = action.payload
        state.error = null
      })
      .addCase(createPlanet.rejected, (state, action) => {
        state.loading = false
        state.error = (action.payload as string) ?? 'Не удалось создать черновик'
      })
      .addCase(getDraft.fulfilled, (state, action) => {
        state.loading = false
        state.draft = action.payload
      })
      .addCase(publishPlanet.fulfilled, (state) => {
        state.loading = false
        state.draft = null
        state.error = null
      })
      .addCase(publishPlanet.rejected, (state, action) => {
        state.loading = false
        state.error = (action.payload as string) ?? 'Не удалось опубликовать планету'
      })
      .addCase(likePlanet.fulfilled, (state, action) => {
        state.loading = false
        const updated = action.payload
        state.planets = state.planets.map((planet) =>
          planet.id === updated.id ? updated : planet,
        )
        if (state.feed?.id === updated.id) {
          state.feed = updated
        }
      })
      .addCase(likePlanet.rejected, (state, action) => {
        state.loading = false
        console.error('planets: не удалось изменить лайк', action.error)
      })
      .addCase(deletePlanet.fulfilled, (state, action) => {
        state.loading = false
        state.planets = state.planets.filter((planet) => planet.id !== action.payload)
      })
      .addCase(deletePlanet.rejected, (state, action) => {
        state.loading = false
        console.error('planets: не удалось удалить планету', action.error)
      })
  },
})

export const { setMaxDist, setAppliedMaxDist, resetFilter, setFeed, clearError } =
  planetsSlice.actions
export default planetsSlice.reducer
