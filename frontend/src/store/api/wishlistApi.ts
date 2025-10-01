import { apiSlice } from './apiSlice'
import type {
  Wishlist,
  ApiResponse,
} from '@/types'

export const wishlistApi = apiSlice.injectEndpoints({
  endpoints: (builder) => ({
    getWishlist: builder.query<Wishlist, void>({
      query: () => '/wishlist',
      transformResponse: (response: ApiResponse<Wishlist>) => response.data,
      providesTags: ['Wishlist'],
    }),

    addToWishlist: builder.mutation<Wishlist, { productId: string }>({
      query: (data) => ({
        url: '/wishlist/items',
        method: 'POST',
        body: { product_id: data.productId },
      }),
      transformResponse: (response: ApiResponse<Wishlist>) => response.data,
      invalidatesTags: ['Wishlist'],
    }),

    removeFromWishlist: builder.mutation<Wishlist, { productId: string }>({
      query: ({ productId }) => ({
        url: `/wishlist/items/${productId}`,
        method: 'DELETE',
      }),
      transformResponse: (response: ApiResponse<Wishlist>) => response.data,
      invalidatesTags: ['Wishlist'],
    }),

    clearWishlist: builder.mutation<ApiResponse<null>, void>({
      query: () => ({
        url: '/wishlist',
        method: 'DELETE',
      }),
      invalidatesTags: ['Wishlist'],
    }),

    moveToCart: builder.mutation<ApiResponse<{ addedToCart: number }>, { productId: string; quantity?: number }>({
      query: (data) => ({
        url: '/wishlist/move-to-cart',
        method: 'POST',
        body: data,
      }),
      invalidatesTags: ['Wishlist', 'Cart'],
    }),
  }),
})

export const {
  useGetWishlistQuery,
  useAddToWishlistMutation,
  useRemoveFromWishlistMutation,
  useClearWishlistMutation,
  useMoveToCartMutation,
} = wishlistApi