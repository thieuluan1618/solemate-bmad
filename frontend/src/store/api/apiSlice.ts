import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react'
import type { RootState } from '../index'

const baseQuery = fetchBaseQuery({
  baseUrl: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1',
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
    // Try to refresh the token
    const refreshResult = await baseQuery(
      {
        url: '/auth/refresh',
        method: 'POST',
      },
      api,
      extraOptions
    )

    if (refreshResult.data) {
      // Store the new token
      api.dispatch(setCredentials(refreshResult.data))
      // Retry the original query
      result = await baseQuery(args, api, extraOptions)
    } else {
      // Refresh failed, logout user
      api.dispatch(logout())
    }
  }

  return result
}

export const apiSlice = createApi({
  reducerPath: 'api',
  baseQuery: baseQueryWithReauth,
  tagTypes: ['User', 'Product', 'Cart', 'Order', 'Category', 'Brand'],
  endpoints: () => ({}),
})

// Import auth actions for reauth logic
import { setCredentials, logout } from '../slices/authSlice'