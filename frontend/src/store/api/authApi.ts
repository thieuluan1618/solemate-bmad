import { apiSlice } from './apiSlice'
import type {
  User,
  AuthResponse,
  LoginRequest,
  RegisterRequest,
  ApiResponse,
} from '@/types'

export const authApi = apiSlice.injectEndpoints({
  endpoints: (builder) => ({
    login: builder.mutation<AuthResponse, LoginRequest>({
      query: (credentials) => ({
        url: '/auth/login',
        method: 'POST',
        body: credentials,
      }),
      invalidatesTags: ['User'],
    }),

    register: builder.mutation<AuthResponse, RegisterRequest>({
      query: (userData) => ({
        url: '/auth/register',
        method: 'POST',
        body: userData,
      }),
      invalidatesTags: ['User'],
    }),

    logout: builder.mutation<ApiResponse<null>, void>({
      query: () => ({
        url: '/auth/logout',
        method: 'POST',
      }),
      invalidatesTags: ['User', 'Cart', 'Order'],
    }),

    refreshToken: builder.mutation<{ accessToken: string; refreshToken: string }, void>({
      query: () => ({
        url: '/auth/refresh',
        method: 'POST',
      }),
    }),

    forgotPassword: builder.mutation<ApiResponse<null>, { email: string }>({
      query: (data) => ({
        url: '/auth/forgot-password',
        method: 'POST',
        body: data,
      }),
    }),

    resetPassword: builder.mutation<ApiResponse<null>, { token: string; password: string }>({
      query: (data) => ({
        url: '/auth/reset-password',
        method: 'POST',
        body: data,
      }),
    }),

    verifyEmail: builder.mutation<ApiResponse<null>, { token: string }>({
      query: (data) => ({
        url: '/auth/verify-email',
        method: 'POST',
        body: data,
      }),
      invalidatesTags: ['User'],
    }),

    resendVerification: builder.mutation<ApiResponse<null>, { email: string }>({
      query: (data) => ({
        url: '/auth/resend-verification',
        method: 'POST',
        body: data,
      }),
    }),

    getCurrentUser: builder.query<User, void>({
      query: () => '/auth/me',
      providesTags: ['User'],
    }),
  }),
})

export const {
  useLoginMutation,
  useRegisterMutation,
  useLogoutMutation,
  useRefreshTokenMutation,
  useForgotPasswordMutation,
  useResetPasswordMutation,
  useVerifyEmailMutation,
  useResendVerificationMutation,
  useGetCurrentUserQuery,
} = authApi