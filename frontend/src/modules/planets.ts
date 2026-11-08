import type { DsPlanetSerializer } from '../api/Api'

// Форма планеты совпадает с PlanetSerializer бэкенда:
// liked и is_owner — числа 0/1, расстояния приходят строками.
// Swagger помечает поля необязательными, но сериализатор бэкенда
// гарантирует их наличие — поэтому Required<>.
export type Planet = Required<DsPlanetSerializer>

export const DEFAULT_PHOTO = '/media/default_photo.png'
export const DEFAULT_VIDEO = '/media/default_video.mp4'

const isDefault = (url: string, name: string) => !url || url.includes(name)

// Выбор медиа для плитки: настоящее фото → настоящее видео → дефолтное фото.
// У опубликованной планеты бэкенд всегда подставляет дефолтные файлы, поэтому
// отличаем реальные загруженные файлы от дефолтных по имени.
export function tileMedia(planet: Planet): { type: 'video'; src: string } | { type: 'image'; src: string } {
  if (!isDefault(planet.image_url, 'default_photo.png')) {
    return { type: 'image', src: planet.image_url }
  }
  if (!isDefault(planet.url_video, 'default_video.mp4')) {
    return { type: 'video', src: planet.url_video }
  }
  return { type: 'image', src: planet.image_url || DEFAULT_PHOTO }
}

// Видео для ленты: настоящее видео планеты, иначе дефолтное.
export function feedVideoSrc(planet: Planet): string {
  if (isDefault(planet.url_video, 'default_video.mp4')) {
    return DEFAULT_VIDEO
  }
  return planet.url_video
}
