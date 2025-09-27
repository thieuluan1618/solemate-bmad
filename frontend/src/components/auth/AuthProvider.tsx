'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { AuthContext, tokenUtils } from '@/lib/auth'
import { useAppDispatch, useAppSelector } from '@/hooks/redux'
import { setCredentials, logout, setLoading, selectCurrentUser, selectIsAuthenticated, selectAuthLoading } from '@/store/slices/authSlice'
import { useLoginMutation, useRegisterMutation, useLogoutMutation, useGetCurrentUserQuery } from '@/store/api/authApi'
import type { User } from '@/types'

interface AuthProviderProps {
  children: React.ReactNode
}

export function AuthProvider({ children }: AuthProviderProps) {
  const dispatch = useAppDispatch()
  const router = useRouter()
  const user = useAppSelector(selectCurrentUser)
  const isAuthenticated = useAppSelector(selectIsAuthenticated)
  const isLoading = useAppSelector(selectAuthLoading)

  const [login] = useLoginMutation()
  const [register] = useRegisterMutation()
  const [logoutMutation] = useLogoutMutation()

  // Get current user if token exists
  const {
    data: currentUser,
    isLoading: isUserLoading,
    error: userError,
  } = useGetCurrentUserQuery(undefined, {
    skip: !tokenUtils.getAccessToken() || isAuthenticated,
  })

  // Initialize auth state on mount
  useEffect(() => {
    const token = tokenUtils.getAccessToken()
    if (token && !tokenUtils.isTokenExpired(token) && currentUser) {
      dispatch(setCredentials({
        user: currentUser,
        accessToken: token,
      }))
    } else if (token && tokenUtils.isTokenExpired(token)) {
      // Token is expired, clear it
      tokenUtils.clearTokens()
      dispatch(logout())
    }
  }, [currentUser, dispatch])

  // Handle authentication loading state
  useEffect(() => {
    dispatch(setLoading(isUserLoading))
  }, [isUserLoading, dispatch])

  // Handle user error (token invalid, etc.)
  useEffect(() => {
    if (userError) {
      tokenUtils.clearTokens()
      dispatch(logout())
    }
  }, [userError, dispatch])

  const handleLogin = async (email: string, password: string, rememberMe = false) => {
    try {
      dispatch(setLoading(true))
      const result = await login({ email, password, rememberMe }).unwrap()

      // Store tokens
      tokenUtils.setAccessToken(result.accessToken)
      tokenUtils.setRefreshToken(result.refreshToken)

      // Update Redux state
      dispatch(setCredentials({
        user: result.user,
        accessToken: result.accessToken,
        refreshToken: result.refreshToken,
      }))

      // Redirect based on user role or intended destination
      const redirect = new URLSearchParams(window.location.search).get('redirect')
      if (redirect) {
        router.push(redirect)
      } else {
        router.push(result.user.role === 'admin' ? '/admin' : '/profile')
      }
    } catch (error) {
      dispatch(setLoading(false))
      throw error
    }
  }

  const handleRegister = async (data: {
    email: string
    password: string
    firstName: string
    lastName: string
    phone?: string
  }) => {
    try {
      dispatch(setLoading(true))
      const result = await register(data).unwrap()

      // Store tokens
      tokenUtils.setAccessToken(result.accessToken)
      tokenUtils.setRefreshToken(result.refreshToken)

      // Update Redux state
      dispatch(setCredentials({
        user: result.user,
        accessToken: result.accessToken,
        refreshToken: result.refreshToken,
      }))

      // Redirect to email verification or profile
      router.push('/profile?welcome=true')
    } catch (error) {
      dispatch(setLoading(false))
      throw error
    }
  }

  const handleLogout = async () => {
    try {
      dispatch(setLoading(true))

      // Call logout API
      await logoutMutation().unwrap()

      // Clear tokens and state
      tokenUtils.clearTokens()
      dispatch(logout())

      // Redirect to home
      router.push('/')
    } catch (error) {
      // Even if API call fails, clear local state
      tokenUtils.clearTokens()
      dispatch(logout())
      router.push('/')
    }
  }

  const updateUser = (userData: Partial<User>) => {
    if (user) {
      dispatch(setCredentials({
        user: { ...user, ...userData },
        accessToken: tokenUtils.getAccessToken() || '',
      }))
    }
  }

  const contextValue = {
    user,
    isAuthenticated,
    isLoading,
    login: handleLogin,
    register: handleRegister,
    logout: handleLogout,
    updateUser,
  }

  return (
    <AuthContext.Provider value={contextValue}>
      {children}
    </AuthContext.Provider>
  )
}