import type { ServiceLoginRequest } from './common.types'

export type AuthOidcCallbackGetParams = { code: string }

export type AuthOidcCallbackGetRequest = Record<string, never>

export type AuthOidcCallbackGetResponse = Record<string, unknown>

export type AuthOidcLoginGetParams = Record<string, never>

export type AuthOidcLoginGetRequest = Record<string, never>

export type AuthOidcLoginGetResponse = Record<string, unknown>

export type UserLoginPostParams = Record<string, never>

export type UserLoginPostRequest = ServiceLoginRequest

export type UserLoginPostResponse = Record<string, unknown>

export type UserLogoutPostParams = Record<string, never>

export type UserLogoutPostRequest = Record<string, never>

export type UserLogoutPostResponse = Record<string, unknown>
