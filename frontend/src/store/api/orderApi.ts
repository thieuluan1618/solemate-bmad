import { apiSlice } from './apiSlice'
import type {
  Order,
  OrderStatus,
  CreateOrderRequest,
  PaginatedResponse,
  ApiResponse,
} from '@/types'

export const orderApi = apiSlice.injectEndpoints({
  endpoints: (builder) => ({
    getOrders: builder.query<PaginatedResponse<Order>, { page?: number; limit?: number; status?: OrderStatus }>({
      query: ({ page = 1, limit = 10, status } = {}) => ({
        url: '/orders',
        params: {
          page,
          limit,
          ...(status && { status }),
        },
      }),
      providesTags: (result) =>
        result
          ? [
              ...result.data.map(({ id }) => ({ type: 'Order' as const, id })),
              { type: 'Order', id: 'LIST' },
            ]
          : [{ type: 'Order', id: 'LIST' }],
    }),

    getOrder: builder.query<Order, string>({
      query: (id) => {
        // Check if id is a UUID (contains hyphens) or order number (starts with 'order_')
        const isOrderNumber = id.startsWith('order_')
        return isOrderNumber ? `/orders/number/${id}` : `/orders/${id}`
      },
      providesTags: (result, error, id) => [{ type: 'Order', id }],
    }),

    createOrder: builder.mutation<Order, CreateOrderRequest>({
      query: (data) => ({
        url: '/orders',
        method: 'POST',
        body: data,
      }),
      invalidatesTags: [{ type: 'Order', id: 'LIST' }, 'Cart'],
    }),

    cancelOrder: builder.mutation<Order, { id: string; reason?: string }>({
      query: ({ id, reason }) => ({
        url: `/orders/${id}/cancel`,
        method: 'POST',
        body: { reason },
      }),
      invalidatesTags: (result, error, { id }) => [
        { type: 'Order', id },
        { type: 'Order', id: 'LIST' },
      ],
    }),

    trackOrder: builder.query<{ order: Order; tracking: any[] }, string>({
      query: (id) => `/orders/${id}/track`,
      providesTags: (result, error, id) => [{ type: 'Order', id }],
    }),

    reorderItems: builder.mutation<ApiResponse<{ addedToCart: number }>, { orderId: string }>({
      query: ({ orderId }) => ({
        url: `/orders/${orderId}/reorder`,
        method: 'POST',
      }),
      invalidatesTags: ['Cart'],
    }),

    downloadInvoice: builder.query<Blob, string>({
      query: (id) => ({
        url: `/orders/${id}/invoice`,
        responseHandler: (response) => response.blob(),
      }),
    }),

    getOrderStatuses: builder.query<{ status: OrderStatus; label: string; description: string }[], void>({
      query: () => '/orders/statuses',
    }),
  }),
})

export const {
  useGetOrdersQuery,
  useGetOrderQuery,
  useCreateOrderMutation,
  useCancelOrderMutation,
  useTrackOrderQuery,
  useReorderItemsMutation,
  useLazyDownloadInvoiceQuery,
  useGetOrderStatusesQuery,
} = orderApi