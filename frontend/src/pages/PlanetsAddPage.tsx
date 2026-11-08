import type { FC, FormEvent } from 'react'
import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import PlanetsLoading from '../components/PlanetsLoading'
import { useAppDispatch, useAppSelector } from '../hooks'
import { createPlanet, getDraft, publishPlanet } from '../store/planetsSlice'

// Страница «Добавление» — два шага:
// 1) разметка шаблона telescope.html (создание черновика, multipart);
// 2) разметка шаблона planet_publish.tmpl 1:1 (публикация, PUT).
const PlanetsAddPage: FC = () => {
  const dispatch = useAppDispatch()
  const navigate = useNavigate()

  const draft = useAppSelector((state) => state.planets.draft)
  const loading = useAppSelector((state) => state.planets.loading)
  const error = useAppSelector((state) => state.planets.error)
  const isAuthenticated = useAppSelector((state) => state.user.isAuthenticated)

  const [ready, setReady] = useState(false)
  const [name, setName] = useState('')
  const [file, setFile] = useState<File | null>(null)
  const [previewUrl, setPreviewUrl] = useState('')
  const [dragOver, setDragOver] = useState(false)
  const [description, setDescription] = useState('')
  const [distance, setDistance] = useState('')
  const [observationDate, setObservationDate] = useState('')

  useEffect(() => {
    dispatch(getDraft()).then((action) => {
      // Поля шага 2 заполняются из черновика, как в шаблоне.
      if (getDraft.fulfilled.match(action) && action.payload) {
        const existing = action.payload
        setDescription(existing.description || '')
        setDistance(existing.distance_from_sun || '')
        setObservationDate(existing.observation_date || '')
      }
      setReady(true)
    })
  }, [dispatch])

  useEffect(() => () => URL.revokeObjectURL(previewUrl), [previewUrl])

  const handleFile = (next: File | null | undefined) => {
    if (!next) return
    setFile(next)
    setPreviewUrl(URL.createObjectURL(next))
  }

  const handleStep1 = async (event: FormEvent) => {
    event.preventDefault()
    await dispatch(createPlanet({ name, file }))
  }

  const handleStep2 = async (event: FormEvent) => {
    event.preventDefault()
    if (!draft) return
    const parsed = Number(distance)
    if (Number.isNaN(parsed)) return
    const action = await dispatch(
      publishPlanet({
        id: draft.id,
        body: {
          description,
          distance_from_sun: parsed,
          perihelion: parsed,
          observation_date: observationDate,
        },
      }),
    )
    if (publishPlanet.fulfilled.match(action)) {
      navigate('/')
    }
  }

  // «Добавление» доступно только авторизованному: после F5 (Redux — гость)
  // или выхода уводим на страницу входа.
  useEffect(() => {
    if (ready && !isAuthenticated) {
      navigate('/login')
    }
  }, [ready, isAuthenticated, navigate])

  if (!ready || !isAuthenticated) {
    return <PlanetsLoading />
  }

  if (!draft) {
    const isVideo = file ? file.type === 'video/mp4' || /\.mp4$/i.test(file.name) : false

    return (
      <div className="add-container">
        <h1 className="add-title">Добавление данных</h1>
        <p className="add-subtitle">Покажите новую планету — шаг 1 из 2</p>

        <form onSubmit={(event) => void handleStep1(event)}>
          <label className="add-label">Медиафайл (фото/видео)</label>
          <label
            className={`media-dropzone ${dragOver ? 'dragover' : ''}`}
            onDragEnter={(event) => {
              event.preventDefault()
              setDragOver(true)
            }}
            onDragOver={(event) => {
              event.preventDefault()
              setDragOver(true)
            }}
            onDragLeave={() => setDragOver(false)}
            onDrop={(event) => {
              event.preventDefault()
              setDragOver(false)
              handleFile(event.dataTransfer.files?.[0])
            }}
          >
            <div className="media-dropzone-icons">
              <svg viewBox="0 0 24 24">
                <rect x="3" y="3" width="18" height="18" rx="2" ry="2" />
                <circle cx="8.5" cy="8.5" r="1.5" />
                <polyline points="21 15 16 10 5 21" />
              </svg>
              <svg viewBox="0 0 24 24">
                <polygon points="23 7 16 12 23 17 23 7" />
                <rect x="1" y="5" width="15" height="14" rx="2" ry="2" />
              </svg>
            </div>
            <p className="media-dropzone-text">
              {file ? file.name : 'Перетащите видео или сделайте снимок'}
            </p>
            <div className="media-preview" hidden={!file}>
              {file &&
                (isVideo ? (
                  <video className="preview-media" src={previewUrl} muted loop autoPlay controls />
                ) : (
                  <img className="preview-media" src={previewUrl} alt={file.name} />
                ))}
            </div>
            <input
              type="file"
              name="media"
              accept="image/*,video/mp4"
              onChange={(event) => handleFile(event.target.files?.[0])}
            />
          </label>

          <div className="form-group">
            <label className="add-label">Название планеты</label>
            <input
              type="text"
              name="name"
              className="form-input"
              placeholder="Например: Марс"
              required
              value={name}
              onChange={(event) => setName(event.target.value)}
            />
          </div>

          {error && <div className="form-error">{error}</div>}

          <button type="submit" className="step-btn" disabled={loading}>
            Далее
          </button>
        </form>

        {loading && <PlanetsLoading />}
      </div>
    )
  }

  const realImage = draft.image_url && !draft.image_url.includes('default_photo')
  const realVideo = draft.url_video && !draft.url_video.includes('default_video')

  return (
    <div className="add-container">
      <h1 className="add-title">Публикация планеты</h1>
      <p className="add-subtitle">Добавьте описание и данные орбиты — шаг 2 из 2</p>

      <div className="publish-preview">
        {realImage || (!realVideo && draft.image_url) ? (
          <img className="publish-media" src={draft.image_url} alt={draft.name} />
        ) : realVideo || draft.url_video ? (
          <video className="publish-media" autoPlay muted loop playsInline>
            <source src={draft.url_video} type="video/mp4" />
          </video>
        ) : (
          <div className="publish-media media-placeholder">Нет медиа</div>
        )}
        <div className="publish-name">{draft.name}</div>
      </div>

      <form onSubmit={(event) => void handleStep2(event)}>
        <div className="form-group">
          <label className="add-label">Описание</label>
          <textarea
            name="description"
            className="form-input"
            placeholder="Описание планеты..."
            required
            value={description}
            onChange={(event) => setDescription(event.target.value)}
          />
        </div>

        <div className="form-row">
          <div className="form-group">
            <label className="add-label">Расстояние (AU)</label>
            <input
              type="text"
              name="distance"
              className="form-input"
              placeholder="0.00"
              required
              value={distance}
              onChange={(event) => setDistance(event.target.value)}
            />
          </div>
          <div className="form-group">
            <label className="add-label">Дата наблюдения</label>
            <input
              type="date"
              name="observation_date"
              className="form-input"
              required
              value={observationDate}
              onChange={(event) => setObservationDate(event.target.value)}
            />
          </div>
        </div>

        {error && <div className="form-error">{error}</div>}

        <button type="submit" className="step-btn" disabled={loading}>
          Опубликовать
        </button>
      </form>

      {loading && <PlanetsLoading />}
    </div>
  )
}

export default PlanetsAddPage
