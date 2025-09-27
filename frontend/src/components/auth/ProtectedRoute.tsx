'use client'

import { useEffect } from 'react'
import { useRouter, usePathname } from 'next/navigation'
import { useAuth } from '@/lib/auth'
import { routeUtils, permissionUtils } from '@/lib/auth'

interface ProtectedRouteProps {
  children: React.ReactNode
  requireAuth?: boolean
  requireAdmin?: boolean
  fallback?: React.ReactNode
}

export function ProtectedRoute({
  children,
  requireAuth = false,
  requireAdmin = false,
  fallback = <div>Loading...</div>,
}: ProtectedRouteProps) {
  const { user, isAuthenticated, isLoading } = useAuth()
  const router = useRouter()
  const pathname = usePathname()

  useEffect(() => {
    if (isLoading) return

    // Check for redirect requirements
    const redirectPath = routeUtils.getRedirectPath(user, pathname)
    if (redirectPath) {
      router.push(redirectPath)
      return
    }

    // Check auth requirements
    if (requireAuth && !isAuthenticated) {
      router.push(`/login?redirect=${encodeURIComponent(pathname)}`)
      return
    }

    // Check admin requirements
    if (requireAdmin && !permissionUtils.canAccessAdmin(user)) {
      router.push('/')
      return
    }
  }, [user, isAuthenticated, isLoading, requireAuth, requireAdmin, pathname, router])

  // Show loading fallback
  if (isLoading) {
    return <>{fallback}</>
  }

  // Check requirements before rendering
  if (requireAuth && !isAuthenticated) {
    return <>{fallback}</>
  }

  if (requireAdmin && !permissionUtils.canAccessAdmin(user)) {
    return <>{fallback}</>
  }

  return <>{children}</>
}

// Higher-order component for protecting pages
export function withAuth<P extends object>(
  Component: React.ComponentType<P>,
  options: {
    requireAuth?: boolean
    requireAdmin?: boolean
    fallback?: React.ReactNode
  } = {}
) {
  const WrappedComponent = (props: P) => {
    return (
      <ProtectedRoute {...options}>
        <Component {...props} />
      </ProtectedRoute>
    )
  }

  WrappedComponent.displayName = `withAuth(${Component.displayName || Component.name})`
  return WrappedComponent
}

// Route guard hook for programmatic protection
export function useRouteGuard(options: {
  requireAuth?: boolean
  requireAdmin?: boolean
  redirectTo?: string
} = {}) {
  const { user, isAuthenticated, isLoading } = useAuth()
  const router = useRouter()
  const pathname = usePathname()

  const checkAccess = () => {
    if (isLoading) return { canAccess: false, isLoading: true }

    if (options.requireAuth && !isAuthenticated) {
      return { canAccess: false, isLoading: false, shouldRedirect: true, redirectTo: '/login' }
    }

    if (options.requireAdmin && !permissionUtils.canAccessAdmin(user)) {
      return { canAccess: false, isLoading: false, shouldRedirect: true, redirectTo: '/' }
    }

    return { canAccess: true, isLoading: false }
  }

  const redirect = (path?: string) => {
    const redirectPath = path || options.redirectTo || '/login'
    router.push(`${redirectPath}?redirect=${encodeURIComponent(pathname)}`)
  }

  return {
    ...checkAccess(),
    redirect,
  }
}