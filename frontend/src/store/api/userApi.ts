import { apiSlice } from './apiSlice'
import type {
  User,
  Address,
  ApiResponse,
} from '@/types'

export const userApi = apiSlice.injectEndpoints({
  endpoints: (builder) => ({
    getProfile: builder.query<User, void>({
      query: () => '/users/profile',
      providesTags: ['User'],
    }),

    updateProfile: builder.mutation<User, Partial<Pick<User, 'firstName' | 'lastName' | 'phone' | 'preferences'>>>({
      query: (data) => ({
        url: '/users/profile',
        method: 'PUT',
        body: data,
      }),
      invalidatesTags: ['User'],
    }),

    changePassword: builder.mutation<ApiResponse<null>, { currentPassword: string; newPassword: string }>({
      query: (data) => ({
        url: '/users/change-password',
        method: 'PUT',
        body: data,
      }),
    }),

    getAddresses: builder.query<Address[], void>({
      query: () => '/users/addresses',
      providesTags: (result) =>
        result
          ? [
              ...result.map(({ id }) => ({ type: 'Address' as const, id })),
              { type: 'Address', id: 'LIST' },
            ]
          : [{ type: 'Address', id: 'LIST' }],
    }),

    createAddress: builder.mutation<Address, Omit<Address, 'id' | 'userId' | 'createdAt' | 'updatedAt'>>({
      query: (data) => ({
        url: '/users/addresses',
        method: 'POST',
        body: data,
      }),
      invalidatesTags: [{ type: 'Address', id: 'LIST' }],
    }),

    updateAddress: builder.mutation<Address, { id: string; data: Partial<Omit<Address, 'id' | 'userId' | 'createdAt' | 'updatedAt'>> }>({
      query: ({ id, data }) => ({
        url: `/users/addresses/${id}`,
        method: 'PUT',
        body: data,
      }),
      invalidatesTags: (result, error, { id }) => [
        { type: 'Address', id },
        { type: 'Address', id: 'LIST' },
      ],
    }),

    deleteAddress: builder.mutation<ApiResponse<null>, { id: string }>({
      query: ({ id }) => ({
        url: `/users/addresses/${id}`,
        method: 'DELETE',
      }),
      invalidatesTags: (result, error, { id }) => [
        { type: 'Address', id },
        { type: 'Address', id: 'LIST' },
      ],
    }),

    setDefaultAddress: builder.mutation<Address, { id: string }>({
      query: ({ id }) => ({
        url: `/users/addresses/${id}/default`,
        method: 'PUT',
      }),
      invalidatesTags: [{ type: 'Address', id: 'LIST' }],
    }),
  }),
})

export const {
  useGetProfileQuery,
  useUpdateProfileMutation,
  useChangePasswordMutation,
  useGetAddressesQuery,
  useCreateAddressMutation,
  useUpdateAddressMutation,
  useDeleteAddressMutation,
  useSetDefaultAddressMutation,
} = userApi