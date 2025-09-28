'use client'

import { useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { useSelector } from 'react-redux'
import { selectIsAuthenticated, selectAuthLoading } from '@/store/slices/authSlice'

interface AuthGuardProps {
  children: React.ReactNode
  fallback?: React.ReactNode
}

export default function AuthGuard({ children, fallback }: AuthGuardProps) {
  const router = useRouter()
  const isAuthenticated = useSelector(selectIsAuthenticated)
  const isLoading = useSelector(selectAuthLoading)

  useEffect(() => {
    if (!isLoading && !isAuthenticated) {
      // Redirect to login with current page as redirect parameter
      const currentPath = window.location.pathname + window.location.search
      const redirectUrl = `/login?redirect=${encodeURIComponent(currentPath)}`
      router.push(redirectUrl)
    }
  }, [isAuthenticated, isLoading, router])

  // Show loading state while checking auth
  if (isLoading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 mx-auto"></div>
          <p className="mt-2 text-gray-600">Loading...</p>
        </div>
      </div>
    )
  }

  // Show fallback or nothing while redirecting
  if (!isAuthenticated) {
    return fallback ? <>{fallback}</> : null
  }

  // User is authenticated, render children
  return <>{children}</>
}