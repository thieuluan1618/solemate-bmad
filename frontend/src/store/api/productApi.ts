import { apiSlice } from './apiSlice'
import type {
  Product,
  Category,
  Brand,
  ProductSearchParams,
  PaginatedResponse,
  SearchSuggestion,
} from '@/types'

export const productApi = apiSlice.injectEndpoints({
  endpoints: (builder) => ({
    getProducts: builder.query<PaginatedResponse<Product>, ProductSearchParams>({
      query: (params) => ({
        url: '/products',
        params: {
          ...params,
          // Convert arrays to comma-separated strings for URL params
          sizes: params.sizes?.join(','),
          colors: params.colors?.join(','),
        },
      }),
      transformResponse: (response: { data: Product[]; pagination: any; success: boolean; message: string }) => ({
        data: response.data,
        meta: response.pagination,
        success: response.success,
        message: response.message,
      }),
      providesTags: (result) =>
        result
          ? [
              ...result.data.map(({ id }) => ({ type: 'Product' as const, id })),
              { type: 'Product', id: 'LIST' },
            ]
          : [{ type: 'Product', id: 'LIST' }],
    }),

    getProduct: builder.query<Product, string>({
      query: (id) => `/products/${id}`,
      transformResponse: (response: { data: Product; success: boolean; message: string }) => response.data,
      providesTags: (result, error, id) => [{ type: 'Product', id }],
    }),

    searchProducts: builder.query<{ products: Product[]; suggestions: SearchSuggestion[] }, { query: string; limit?: number }>({
      query: ({ query, limit = 10 }) => ({
        url: '/products/search',
        params: { query, limit },
      }),
      providesTags: ['Product'],
    }),

    getCategories: builder.query<Category[], { parentId?: string }>({
      query: ({ parentId } = {}) => ({
        url: '/categories',
        params: parentId ? { parentId } : {},
      }),
      providesTags: (result) =>
        result
          ? [
              ...result.map(({ id }) => ({ type: 'Category' as const, id })),
              { type: 'Category', id: 'LIST' },
            ]
          : [{ type: 'Category', id: 'LIST' }],
    }),

    getCategory: builder.query<Category, string>({
      query: (id) => `/categories/${id}`,
      providesTags: (result, error, id) => [{ type: 'Category', id }],
    }),

    getBrands: builder.query<Brand[], void>({
      query: () => '/brands',
      providesTags: (result) =>
        result
          ? [
              ...result.map(({ id }) => ({ type: 'Brand' as const, id })),
              { type: 'Brand', id: 'LIST' },
            ]
          : [{ type: 'Brand', id: 'LIST' }],
    }),

    getBrand: builder.query<Brand, string>({
      query: (id) => `/brands/${id}`,
      providesTags: (result, error, id) => [{ type: 'Brand', id }],
    }),

    getFeaturedProducts: builder.query<Product[], { limit?: number }>({
      query: ({ limit = 8 } = {}) => ({
        url: '/products/featured',
        params: { limit },
      }),
      providesTags: ['Product'],
    }),

    getRelatedProducts: builder.query<Product[], { productId: string; limit?: number }>({
      query: ({ productId, limit = 4 }) => ({
        url: `/products/${productId}/related`,
        params: { limit },
      }),
      providesTags: ['Product'],
    }),

    getProductReviews: builder.query<PaginatedResponse<any>, { productId: string; page?: number; limit?: number; sortBy?: string }>({
      query: ({ productId, page = 1, limit = 10, sortBy = 'newest' }) => ({
        url: `/products/${productId}/reviews`,
        params: { page, limit, sortBy },
      }),
      providesTags: (result, error, { productId }) => [
        { type: 'Review', id: productId },
        { type: 'Review', id: 'LIST' },
      ],
    }),
  }),
})

export const {
  useGetProductsQuery,
  useGetProductQuery,
  useSearchProductsQuery,
  useGetCategoriesQuery,
  useGetCategoryQuery,
  useGetBrandsQuery,
  useGetBrandQuery,
  useGetFeaturedProductsQuery,
  useGetRelatedProductsQuery,
  useGetProductReviewsQuery,
} = productApi