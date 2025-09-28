import { createSlice, PayloadAction } from '@reduxjs/toolkit'
import Cookies from 'js-cookie'

export interface User {
  id: string
  email: string
  firstName: string
  lastName: string
  role: 'customer' | 'admin'
  addresses?: Address[]
}

export interface Address {
  id: string
  type: 'shipping' | 'billing'
  street: string
  city: string
  state: string
  zipCode: string
  country: string
  isDefault: boolean
}

interface AuthState {
  user: User | null
  token: string | null
  isAuthenticated: boolean
  loading: boolean
}

const getInitialToken = () => {
  if (typeof window !== 'undefined') {
    return Cookies.get('accessToken') || null
  }
  return null
}

const initialToken = getInitialToken()

const initialState: AuthState = {
  user: null,
  token: initialToken,
  isAuthenticated: !!initialToken,
  loading: false,
}

const authSlice = createSlice({
  name: 'auth',
  initialState,
  reducers: {
    setCredentials: (
      state,
      action: PayloadAction<{ user: User; accessToken: string; refreshToken?: string }>
    ) => {
      const { user, accessToken, refreshToken } = action.payload
      state.user = user
      state.token = accessToken
      state.isAuthenticated = true
      state.loading = false

      // Store tokens in cookies
      Cookies.set('accessToken', accessToken, { expires: 1 }) // 1 day
      if (refreshToken) {
        Cookies.set('refreshToken', refreshToken, { expires: 7 }) // 7 days
      }
    },
    logout: (state) => {
      state.user = null
      state.token = null
      state.isAuthenticated = false
      state.loading = false

      // Remove tokens from cookies
      Cookies.remove('accessToken')
      Cookies.remove('refreshToken')
    },
    setLoading: (state, action: PayloadAction<boolean>) => {
      state.loading = action.payload
    },
    updateUser: (state, action: PayloadAction<Partial<User>>) => {
      if (state.user) {
        state.user = { ...state.user, ...action.payload }
      }
    },
  },
})

export const { setCredentials, logout, setLoading, updateUser } = authSlice.actions

export default authSlice.reducer

// Selectors
export const selectCurrentUser = (state: { auth: AuthState }) => state.auth.user
export const selectCurrentToken = (state: { auth: AuthState }) => state.auth.token
export const selectIsAuthenticated = (state: { auth: AuthState }) => state.auth.isAuthenticated
export const selectAuthLoading = (state: { auth: AuthState }) => state.auth.loading