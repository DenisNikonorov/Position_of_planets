/* eslint-disable */
/* tslint:disable */
// @ts-nocheck
/*
 * ---------------------------------------------------------------
 * ## THIS FILE WAS GENERATED VIA SWAGGER-TYPESCRIPT-API        ##
 * ##                                                           ##
 * ## AUTHOR: acacode                                           ##
 * ## SOURCE: https://github.com/acacode/swagger-typescript-api ##
 * ---------------------------------------------------------------
 */

export type AuthCreateData = DsTokenResponse;

export type DeauthCreateData = string;

export type DraftListData = DsPlanetSerializer;

export interface DsLikeRequest {
  like: 0 | 1;
}

export interface DsLoginUserRequest {
  login: string;
  password: string;
}

export interface DsPlanetSerializer {
  created_at?: string;
  description?: string;
  distance_from_sun?: string;
  id?: number;
  image_url?: string;
  is_owner?: number;
  liked?: number;
  likes_count?: number;
  name?: string;
  observation_date?: string;
  perihelion?: string;
  updated_at?: string;
  url_video?: string;
}

export interface DsPublishPlanetRequest {
  description: string;
  distance_from_sun: number;
  observation_date: string;
  perihelion: number;
}

export interface DsRegisterUserRequest {
  /** @minLength 3 */
  login: string;
  /** @minLength 4 */
  password: string;
}

export interface DsTokenResponse {
  access_token?: string;
  expires_in?: number;
  token_type?: string;
}

export interface DsUserSerializer {
  id?: number;
  login?: string;
}

export type FeedListData = DsPlanetSerializer;

export type LikeCreateData = DsPlanetSerializer;

export type PlanetsCreateData = DsPlanetSerializer;

export type PlanetsDeleteData = DsPlanetSerializer;

export type PlanetsListData = DsPlanetSerializer[];

export type PlanetsUpdateData = DsPlanetSerializer;

export type UsersCreateData = DsUserSerializer;

import type {
  AxiosInstance,
  AxiosRequestConfig,
  AxiosResponse,
  HeadersDefaults,
  ResponseType,
} from "axios";
import axios from "axios";

export type QueryParamsType = Record<string | number, any>;

export interface FullRequestParams
  extends Omit<AxiosRequestConfig, "data" | "params" | "url" | "responseType"> {
  /** set parameter to `true` for call `securityWorker` for this request */
  secure?: boolean;
  /** request path */
  path: string;
  /** content type of request body */
  type?: ContentType;
  /** query params */
  query?: QueryParamsType;
  /** format of response (i.e. response.json() -> format: "json") */
  format?: ResponseType;
  /** request body */
  body?: unknown;
}

export type RequestParams = Omit<
  FullRequestParams,
  "body" | "method" | "query" | "path"
>;

export interface ApiConfig<SecurityDataType = unknown>
  extends Omit<AxiosRequestConfig, "data" | "cancelToken"> {
  securityWorker?: (
    securityData: SecurityDataType | null,
  ) => Promise<AxiosRequestConfig | void> | AxiosRequestConfig | void;
  secure?: boolean;
  format?: ResponseType;
}

export enum ContentType {
  Json = "application/json",
  JsonApi = "application/vnd.api+json",
  FormData = "multipart/form-data",
  UrlEncoded = "application/x-www-form-urlencoded",
  Text = "text/plain",
}

export class HttpClient<SecurityDataType = unknown> {
  public instance: AxiosInstance;
  private securityData: SecurityDataType | null = null;
  private securityWorker?: ApiConfig<SecurityDataType>["securityWorker"];
  private secure?: boolean;
  private format?: ResponseType;

  constructor({
    securityWorker,
    secure,
    format,
    ...axiosConfig
  }: ApiConfig<SecurityDataType> = {}) {
    this.instance = axios.create({
      ...axiosConfig,
      baseURL: axiosConfig.baseURL || "/api",
    });
    this.secure = secure;
    this.format = format;
    this.securityWorker = securityWorker;
  }

  public setSecurityData = (data: SecurityDataType | null) => {
    this.securityData = data;
  };

  protected mergeRequestParams(
    params1: AxiosRequestConfig,
    params2?: AxiosRequestConfig,
  ): AxiosRequestConfig {
    const method = params1.method || (params2 && params2.method);

    return {
      ...this.instance.defaults,
      ...params1,
      ...(params2 || {}),
      headers: {
        ...((method &&
          this.instance.defaults.headers[
            method.toLowerCase() as keyof HeadersDefaults
          ]) ||
          {}),
        ...(params1.headers || {}),
        ...((params2 && params2.headers) || {}),
      },
    };
  }

  protected stringifyFormItem(formItem: unknown) {
    if (typeof formItem === "object" && formItem !== null) {
      return JSON.stringify(formItem);
    } else {
      return `${formItem}`;
    }
  }

  protected createFormData(input: Record<string, unknown>): FormData {
    if (input instanceof FormData) {
      return input;
    }
    return Object.keys(input || {}).reduce((formData, key) => {
      const property = input[key];
      const propertyContent: any[] =
        property instanceof Array ? property : [property];

      for (const formItem of propertyContent) {
        const isFileType = formItem instanceof Blob || formItem instanceof File;
        formData.append(
          key,
          isFileType ? formItem : this.stringifyFormItem(formItem),
        );
      }

      return formData;
    }, new FormData());
  }

  public request = async <T = any, _E = any>({
    secure,
    path,
    type,
    query,
    format,
    body,
    ...params
  }: FullRequestParams): Promise<AxiosResponse<T>> => {
    const secureParams =
      ((typeof secure === "boolean" ? secure : this.secure) &&
        this.securityWorker &&
        (await this.securityWorker(this.securityData))) ||
      {};
    const requestParams = this.mergeRequestParams(params, secureParams);
    const responseFormat = format || this.format || undefined;

    if (
      type === ContentType.FormData &&
      body &&
      body !== null &&
      typeof body === "object"
    ) {
      body = this.createFormData(body as Record<string, unknown>);
    }

    if (
      type === ContentType.Text &&
      body &&
      body !== null &&
      typeof body !== "string"
    ) {
      body = JSON.stringify(body);
    }

    return this.instance.request({
      ...requestParams,
      headers: {
        ...(requestParams.headers || {}),
        ...(type ? { "Content-Type": type } : {}),
      },
      params: query,
      responseType: responseFormat,
      data: body,
      url: path,
    });
  };
}

/**
 * @title Planets Position API
 * @version 1.0
 * @baseUrl /api
 * @contact
 *
 * Бэкенд SPA с планетами. Аутентификация по JWT: access-токен выдаёт POST /api/users/auth и проверяется middleware, logout кладёт токен в блеклист Redis.
 */
export class Api<
  SecurityDataType extends unknown,
> extends HttpClient<SecurityDataType> {
  planets = {
    /**
     * @description Опубликованные планеты с фильтрами. Работает и без токена: у гостя liked и is_owner равны 0.
     *
     * @tags planets
     * @name PlanetsList
     * @summary Список планет
     * @request GET:/planets
     */
    planetsList: (
      query?: {
        /** Максимальное расстояние от Солнца */
        maxDist?: number;
        /** Поиск по названию (ILIKE) */
        search?: string;
      },
      params: RequestParams = {},
    ) =>
      this.request<PlanetsListData, string>({
        path: `/planets`,
        method: "GET",
        query: query,
        format: "json",
        ...params,
      }),

    /**
     * @description Первый этап публикации. Владелец берётся из токена, поэтому подделать user_id нельзя. Отсутствующие файлы заменяются дефолтными.
     *
     * @tags planets
     * @name PlanetsCreate
     * @summary Создать планету (черновик)
     * @request POST:/planets
     * @secure
     */
    planetsCreate: (
      data: {
        /**
         * Изображение (иначе дефолтное)
         * @format binary
         */
        image_url?: File;
        /** Название планеты */
        name: string;
        /**
         * Видео (иначе дефолтное)
         * @format binary
         */
        url_video?: File;
      },
      params: RequestParams = {},
    ) =>
      this.request<PlanetsCreateData, string>({
        path: `/planets`,
        method: "POST",
        body: data,
        secure: true,
        type: ContentType.FormData,
        format: "json",
        ...params,
      }),

    /**
     * @description Последний черновик пользователя из токена. Требуется Authorization: Bearer.
     *
     * @tags planets
     * @name DraftList
     * @summary Поля черновика
     * @request GET:/planets/draft
     * @secure
     */
    draftList: (params: RequestParams = {}) =>
      this.request<DraftListData, string>({
        path: `/planets/draft`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Одна планета из ленты. Без id — первая; с id и next=true — следующая после id (по кругу).
     *
     * @tags planets
     * @name FeedList
     * @summary Лента планет
     * @request GET:/planets/feed
     */
    feedList: (
      query?: {
        /** Идентификатор текущей планеты */
        id?: number;
        /** Взять следующую после id */
        next?: boolean;
      },
      params: RequestParams = {},
    ) =>
      this.request<FeedListData, string>({
        path: `/planets/feed`,
        method: "GET",
        query: query,
        format: "json",
        ...params,
      }),

    /**
     * @description Второй этап публикации: заполняет бизнес-поля черновика и переводит его в published. Изменить чужую запись нельзя.
     *
     * @tags planets
     * @name PlanetsUpdate
     * @summary Опубликовать планету
     * @request PUT:/planets/{id}
     * @secure
     */
    planetsUpdate: (
      id: number,
      body: DsPublishPlanetRequest,
      params: RequestParams = {},
    ) =>
      this.request<PlanetsUpdateData, string>({
        path: `/planets/${id}`,
        method: "PUT",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Мягкое удаление: запись остаётся в БД, но исчезает из выдачи. Удалить можно только свою.
     *
     * @tags planets
     * @name PlanetsDelete
     * @summary Удалить планету
     * @request DELETE:/planets/{id}
     * @secure
     */
    planetsDelete: (id: number, params: RequestParams = {}) =>
      this.request<PlanetsDeleteData, string>({
        path: `/planets/${id}`,
        method: "DELETE",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Лайк ставит авторизованный пользователь на опубликованную планету.
     *
     * @tags planets
     * @name LikeCreate
     * @summary Поставить или снять лайк
     * @request POST:/planets/{id}/like
     * @secure
     */
    likeCreate: (id: number, body: DsLikeRequest, params: RequestParams = {}) =>
      this.request<LikeCreateData, string>({
        path: `/planets/${id}/like`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),
  };
  users = {
    /**
     * @description Создаёт пользователя с bcrypt-хешем пароля. Логин должен быть свободен.
     *
     * @tags users
     * @name UsersCreate
     * @summary Регистрация пользователя
     * @request POST:/users
     */
    usersCreate: (body: DsRegisterUserRequest, params: RequestParams = {}) =>
      this.request<UsersCreateData, string>({
        path: `/users`,
        method: "POST",
        body: body,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Возвращает access-токен. Неверный логин или пароль — 401. Токен действует JWT_EXPIRES_IN, после logout попадает в блеклист Redis.
     *
     * @tags users
     * @name AuthCreate
     * @summary Аутентификация (вход)
     * @request POST:/users/auth
     */
    authCreate: (body: DsLoginUserRequest, params: RequestParams = {}) =>
      this.request<AuthCreateData, string>({
        path: `/users/auth`,
        method: "POST",
        body: body,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Отзывает текущий токен: записывает его в блеклист Redis на срок жизни до exp. Повторное использование токена — 403.
     *
     * @tags users
     * @name DeauthCreate
     * @summary Выход (logout)
     * @request POST:/users/deauth
     * @secure
     */
    deauthCreate: (params: RequestParams = {}) =>
      this.request<DeauthCreateData, string>({
        path: `/users/deauth`,
        method: "POST",
        secure: true,
        format: "json",
        ...params,
      }),
  };
}
