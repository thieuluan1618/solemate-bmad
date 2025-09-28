import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react'
import Cookies from 'js-cookie'
import type { RootState } from '../index'

const baseQuery = fetchBaseQuery({
  baseUrl: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8000/api/v1',
  prepareHeaders: (headers, { getState }) => {
    const token = (getState() as RootState).auth.token
    if (token) {
      headers.set('authorization', `Bearer ${token}`)
    }
    headers.set('content-type', 'application/json')
    return headers
  },
})

const baseQueryWithReauth = async (args: any, api: any, extraOptions: any) => {
  let result = await baseQuery(args, api, extraOptions)

  if (result.error && result.error.status === 401) {
    // Get refresh token from cookies
    const refreshToken = Cookies.get('refreshToken')

    if (refreshToken) {
      // Try to refresh the token
      const refreshResult = await baseQuery(
        {
          url: '/auth/refresh',
          method: 'POST',
          body: { refresh_token: refreshToken },
        },
        api,
        extraOptions
      )

      if (refreshResult.data) {
        // Transform and store the new tokens
        const responseData = refreshResult.data as any
        api.dispatch(setCredentials({
          user: (api.getState() as RootState).auth.user!, // Keep existing user data
          accessToken: responseData.data.access_token,
          refreshToken: responseData.data.refresh_token,
        }))
        // Retry the original query
        result = await baseQuery(args, api, extraOptions)
      } else {
        // Refresh failed, logout user and redirect to login
        api.dispatch(logout())

        // Redirect to login page with current page as redirect parameter
        if (typeof window !== 'undefined') {
          const currentPath = window.location.pathname + window.location.search
          const redirectUrl = currentPath === '/login' ? '/login' : `/login?redirect=${encodeURIComponent(currentPath)}`
          window.location.href = redirectUrl
        }
      }
    } else {
      // No refresh token, logout user and redirect to login
      api.dispatch(logout())

      // Redirect to login page
      if (typeof window !== 'undefined') {
        const currentPath = window.location.pathname + window.location.search
        const redirectUrl = currentPath === '/login' ? '/login' : `/login?redirect=${encodeURIComponent(currentPath)}`
        window.location.href = redirectUrl
      }
    }
  }

  return result
}

export const apiSlice = createApi({
  reducerPath: 'api',
  baseQuery: baseQueryWithReauth,
  tagTypes: ['User', 'Product', 'Cart', 'Order', 'Category', 'Brand', 'Review'],
  endpoints: () => ({}),
})

// Import auth actions for reauth logic
import { setCredentials, logout } from '../slices/authSlice'