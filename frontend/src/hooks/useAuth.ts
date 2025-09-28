import { useSelector } from 'react-redux'
import { selectCurrentUser, selectCurrentToken, selectIsAuthenticated, selectAuthLoading } from '@/store/slices/authSlice'

export const useAuth = () => {
  const user = useSelector(selectCurrentUser)
  const token = useSelector(selectCurrentToken)
  const isAuthenticated = useSelector(selectIsAuthenticated)
  const isLoading = useSelector(selectAuthLoading)

  return {
    user,
    token,
    isAuthenticated,
    isLoading,
    isAdmin: user?.role === 'admin',
    isCustomer: user?.role === 'customer',
  }
}