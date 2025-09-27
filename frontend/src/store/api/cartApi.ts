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
      providesTags: ['Cart'],
    }),

    addToCart: builder.mutation<Cart, AddToCartRequest>({
      query: (item) => ({
        url: '/cart/items',
        method: 'POST',
        body: item,
      }),
      invalidatesTags: ['Cart'],
    }),

    updateCartItem: builder.mutation<Cart, { itemId: string; quantity: number }>({
      query: ({ itemId, quantity }) => ({
        url: `/cart/items/${itemId}`,
        method: 'PATCH',
        body: { quantity },
      }),
      invalidatesTags: ['Cart'],
    }),

    removeFromCart: builder.mutation<Cart, { itemId: string }>({
      query: ({ itemId }) => ({
        url: `/cart/items/${itemId}`,
        method: 'DELETE',
      }),
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
      invalidatesTags: ['Cart'],
    }),

    applyPromoCode: builder.mutation<Cart, { code: string }>({
      query: (data) => ({
        url: '/cart/promo',
        method: 'POST',
        body: data,
      }),
      invalidatesTags: ['Cart'],
    }),

    removePromoCode: builder.mutation<Cart, void>({
      query: () => ({
        url: '/cart/promo',
        method: 'DELETE',
      }),
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