import { apiSlice } from './apiSlice'
import type {
  Cart,
  AddToCartRequest,
  ApiResponse,
} from '@/types'

export const cartApi = apiSlice.injectEndpoints({
  endpoints: (builder) => ({
    getCart: builder.query<Cart, void>({
      query: () => '/cart',
      transformResponse: (response: ApiResponse<Cart>) => response.data,
      providesTags: ['Cart'],
    }),

    addToCart: builder.mutation<Cart, AddToCartRequest>({
      query: (item) => ({
        url: '/cart/items',
        method: 'POST',
        body: item,
      }),
      transformResponse: (response: ApiResponse<Cart>) => response.data,
      invalidatesTags: ['Cart'],
    }),

    updateCartItem: builder.mutation<Cart, { itemId: string; quantity: number }>({
      query: ({ itemId, quantity }) => ({
        url: `/cart/items/${itemId}`,
        method: 'PATCH',
        body: { quantity },
      }),
      transformResponse: (response: ApiResponse<Cart>) => response.data,
      invalidatesTags: ['Cart'],
    }),

    removeFromCart: builder.mutation<Cart, { itemId: string }>({
      query: ({ itemId }) => ({
        url: `/cart/items/${itemId}`,
        method: 'DELETE',
      }),
      transformResponse: (response: ApiResponse<Cart>) => response.data,
      invalidatesTags: ['Cart'],
    }),

    clearCart: builder.mutation<ApiResponse<null>, void>({
      query: () => ({
        url: '/cart',
        method: 'DELETE',
      }),
      invalidatesTags: ['Cart'],
    }),

    syncCart: builder.mutation<Cart, { items: AddToCartRequest[] }>({
      query: (data) => ({
        url: '/cart/sync',
        method: 'POST',
        body: data,
      }),
      transformResponse: (response: ApiResponse<Cart>) => response.data,
      invalidatesTags: ['Cart'],
    }),

    applyPromoCode: builder.mutation<Cart, { code: string }>({
      query: (data) => ({
        url: '/cart/promo',
        method: 'POST',
        body: data,
      }),
      transformResponse: (response: ApiResponse<Cart>) => response.data,
      invalidatesTags: ['Cart'],
    }),

    removePromoCode: builder.mutation<Cart, void>({
      query: () => ({
        url: '/cart/promo',
        method: 'DELETE',
      }),
      transformResponse: (response: ApiResponse<Cart>) => response.data,
      invalidatesTags: ['Cart'],
    }),
  }),
})

export const {
  useGetCartQuery,
  useAddToCartMutation,
  useUpdateCartItemMutation,
  useRemoveFromCartMutation,
  useClearCartMutation,
  useSyncCartMutation,
  useApplyPromoCodeMutation,
  useRemovePromoCodeMutation,
} = cartApi