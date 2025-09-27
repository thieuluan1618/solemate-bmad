import { createContext, useContext } from 'react'
import type { User } from '@/types'

// Auth context type
export interface AuthContextType {
  user: User | null
  isAuthenticated: boolean
  isLoading: boolean
  login: (email: string, password: string, rememberMe?: boolean) => Promise<void>
  register: (data: {
    email: string
    password: string
    firstName: string
    lastName: string
    phone?: string
  }) => Promise<void>
  logout: () => Promise<void>
  updateUser: (data: Partial<User>) => void
}

// Create auth context
export const AuthContext = createContext<AuthContextType | undefined>(undefined)

// Auth hook
export function useAuth() {
  const context = useContext(AuthContext)
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider')
  }
  return context
}

// JWT token utilities
export const tokenUtils = {
  getAccessToken: (): string | null => {
    if (typeof window === 'undefined') return null
    return localStorage.getItem('accessToken')
  },

  setAccessToken: (token: string): void => {
    if (typeof window === 'undefined') return
    localStorage.setItem('accessToken', token)
  },

  removeAccessToken: (): void => {
    if (typeof window === 'undefined') return
    localStorage.removeItem('accessToken')
  },

  getRefreshToken: (): string | null => {
    if (typeof window === 'undefined') return null
    return localStorage.getItem('refreshToken')
  },

  setRefreshToken: (token: string): void => {
    if (typeof window === 'undefined') return
    localStorage.setItem('refreshToken', token)
  },

  removeRefreshToken: (): void => {
    if (typeof window === 'undefined') return
    localStorage.removeItem('refreshToken')
  },

  clearTokens: (): void => {
    if (typeof window === 'undefined') return
    localStorage.removeItem('accessToken')
    localStorage.removeItem('refreshToken')
  },

  isTokenExpired: (token: string): boolean => {
    try {
      const payload = JSON.parse(atob(token.split('.')[1]))
      const currentTime = Date.now() / 1000
      return payload.exp < currentTime
    } catch {
      return true
    }
  },
}

// Route protection utilities
export const routeUtils = {
  isPublicRoute: (pathname: string): boolean => {
    const publicRoutes = ['/', '/login', '/register', '/forgot-password', '/reset-password', '/products', '/search']
    const publicPatterns = ['/products/', '/categories/', '/brands/']

    return publicRoutes.includes(pathname) ||
           publicPatterns.some(pattern => pathname.startsWith(pattern))
  },

  isAuthRoute: (pathname: string): boolean => {
    const authRoutes = ['/login', '/register', '/forgot-password', '/reset-password']
    return authRoutes.includes(pathname)
  },

  isProtectedRoute: (pathname: string): boolean => {
    const protectedRoutes = ['/profile', '/orders', '/cart', '/checkout', '/wishlist']
    const protectedPatterns = ['/profile/', '/orders/', '/checkout/']

    return protectedRoutes.includes(pathname) ||
           protectedPatterns.some(pattern => pathname.startsWith(pattern))
  },

  isAdminRoute: (pathname: string): boolean => {
    return pathname.startsWith('/admin')
  },

  getRedirectPath: (user: User | null, pathname: string): string | null => {
    // If user is logged in and trying to access auth pages, redirect to dashboard
    if (user && routeUtils.isAuthRoute(pathname)) {
      return user.role === 'admin' ? '/admin' : '/profile'
    }

    // If user is not logged in and trying to access protected pages, redirect to login
    if (!user && routeUtils.isProtectedRoute(pathname)) {
      return `/login?redirect=${encodeURIComponent(pathname)}`
    }

    // If user is not admin and trying to access admin pages, redirect to home
    if (user && user.role !== 'admin' && routeUtils.isAdminRoute(pathname)) {
      return '/'
    }

    return null
  },
}

// Permission utilities
export const permissionUtils = {
  canAccessAdmin: (user: User | null): boolean => {
    return user?.role === 'admin'
  },

  canEditProfile: (user: User | null, profileUserId: string): boolean => {
    return user?.id === profileUserId || user?.role === 'admin'
  },

  canViewOrder: (user: User | null, orderUserId: string): boolean => {
    return user?.id === orderUserId || user?.role === 'admin'
  },

  canCancelOrder: (user: User | null, orderUserId: string, orderStatus: string): boolean => {
    const canAccess = user?.id === orderUserId || user?.role === 'admin'
    const canCancel = ['pending', 'confirmed', 'processing'].includes(orderStatus)
    return canAccess && canCancel
  },

  canWriteReview: (user: User | null): boolean => {
    return !!user && user.isEmailVerified
  },

  canDeleteReview: (user: User | null, reviewUserId: string): boolean => {
    return user?.id === reviewUserId || user?.role === 'admin'
  },
}

// Session management
export const sessionUtils = {
  getSessionData: () => {
    if (typeof window === 'undefined') return null

    try {
      const sessionData = localStorage.getItem('sessionData')
      return sessionData ? JSON.parse(sessionData) : null
    } catch {
      return null
    }
  },

  setSessionData: (data: any) => {
    if (typeof window === 'undefined') return
    localStorage.setItem('sessionData', JSON.stringify(data))
  },

  clearSessionData: () => {
    if (typeof window === 'undefined') return
    localStorage.removeItem('sessionData')
  },

  getCartSessionId: (): string => {
    if (typeof window === 'undefined') return generateSessionId()

    let sessionId = localStorage.getItem('cartSessionId')
    if (!sessionId) {
      sessionId = generateSessionId()
      localStorage.setItem('cartSessionId', sessionId)
    }
    return sessionId
  },
}

// Helper function to generate session ID
function generateSessionId(): string {
  return 'sess_' + Math.random().toString(36).substring(2) + Date.now().toString(36)
}