'use client'

import { useState, useEffect, useMemo } from 'react'
import { useRouter, useSearchParams } from 'next/navigation'
import { Button, Input } from '@/components/ui'
import { useGetProductsQuery, useGetCategoriesQuery, useGetBrandsQuery } from '@/store/api/productApi'
import { ProductSearchParams, ProductFilters } from '@/types'

interface ProductCardProps {
  product: any
}

function ProductCard({ product }: ProductCardProps) {
  const router = useRouter()

  return (
    <div
      className="card group cursor-pointer transform hover:scale-105 transition-all duration-200"
      onClick={() => router.push(`/products/${product.id}`)}
    >
      <div className="aspect-square overflow-hidden rounded-lg bg-gray-100 mb-4">
        {product.images?.[0] ? (
          <img
            src={product.images[0].url}
            alt={product.images[0].altText || product.name}
            className="w-full h-full object-cover group-hover:scale-110 transition-transform duration-200"
          />
        ) : (
          <div className="w-full h-full flex items-center justify-center text-gray-400">
            <svg className="w-12 h-12" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
            </svg>
          </div>
        )}
      </div>

      <div className="space-y-2">
        <div className="flex items-center justify-between">
          <span className="text-sm text-gray-500">{product.brand?.name}</span>
          <div className="flex items-center space-x-1">
            <span className="text-yellow-400 text-sm">★</span>
            <span className="text-sm text-gray-600">{product.rating}</span>
            <span className="text-sm text-gray-500">({product.reviewCount})</span>
          </div>
        </div>

        <h3 className="font-semibold text-gray-900 group-hover:text-primary-600 transition-colors">
          {product.name}
        </h3>

        <div className="flex items-center space-x-2">
          <span className="text-lg font-bold text-gray-900">
            ${product.price.toFixed(2)}
          </span>
          {product.originalPrice && product.originalPrice > product.price && (
            <>
              <span className="text-sm text-gray-500 line-through">
                ${product.originalPrice.toFixed(2)}
              </span>
              <span className="text-sm font-medium text-red-600">
                {product.discountPercentage}% off
              </span>
            </>
          )}
        </div>

        <div className="text-sm text-gray-600">{product.category?.name}</div>

        {product.stockQuantity === 0 && (
          <span className="text-sm text-red-600 font-medium">Out of Stock</span>
        )}
      </div>
    </div>
  )
}

function FilterSection({
  title,
  children,
  isOpen,
  onToggle
}: {
  title: string
  children: React.ReactNode
  isOpen: boolean
  onToggle: () => void
}) {
  return (
    <div className="border-b border-gray-200 pb-4">
      <button
        className="flex w-full items-center justify-between py-2 text-left"
        onClick={onToggle}
      >
        <span className="font-medium text-gray-900">{title}</span>
        <svg
          className={`h-5 w-5 transform transition-transform ${isOpen ? 'rotate-180' : ''}`}
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
        </svg>
      </button>
      {isOpen && <div className="mt-2 space-y-2">{children}</div>}
    </div>
  )
}

export default function ProductsPage() {
  const router = useRouter()
  const searchParams = useSearchParams()

  // Parse URL parameters
  const initialParams = useMemo(() => {
    const params: ProductSearchParams = {
      page: parseInt(searchParams.get('page') || '1'),
      limit: parseInt(searchParams.get('limit') || '12'),
      sortBy: (searchParams.get('sortBy') as ProductSearchParams['sortBy']) || 'newest',
    }

    if (searchParams.get('query')) params.query = searchParams.get('query')!
    if (searchParams.get('category')) params.category = searchParams.get('category')!
    if (searchParams.get('brand')) params.brand = searchParams.get('brand')!
    if (searchParams.get('minPrice')) params.minPrice = parseFloat(searchParams.get('minPrice')!)
    if (searchParams.get('maxPrice')) params.maxPrice = parseFloat(searchParams.get('maxPrice')!)
    if (searchParams.get('rating')) params.rating = parseFloat(searchParams.get('rating')!)
    if (searchParams.get('inStock')) params.inStock = searchParams.get('inStock') === 'true'

    const sizes = searchParams.get('sizes')
    if (sizes) params.sizes = sizes.split(',')

    const colors = searchParams.get('colors')
    if (colors) params.colors = colors.split(',')

    return params
  }, [searchParams])

  const [filters, setFilters] = useState<ProductSearchParams>(initialParams)
  const [openFilters, setOpenFilters] = useState<Record<string, boolean>>({
    category: true,
    brand: true,
    price: true,
    rating: false,
    size: false,
    color: false,
  })
  const [searchQuery, setSearchQuery] = useState(initialParams.query || '')

  // Fetch data
  const { data: productsData, isLoading: productsLoading, error: productsError } = useGetProductsQuery(filters)
  const { data: categories } = useGetCategoriesQuery({})
  const { data: brands } = useGetBrandsQuery()

  // Update URL when filters change
  useEffect(() => {
    const params = new URLSearchParams()

    Object.entries(filters).forEach(([key, value]) => {
      if (value !== undefined && value !== null && value !== '') {
        if (Array.isArray(value)) {
          if (value.length > 0) params.set(key, value.join(','))
        } else {
          params.set(key, value.toString())
        }
      }
    })

    router.replace(`/products?${params.toString()}`)
  }, [filters, router])

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault()
    setFilters(prev => ({ ...prev, query: searchQuery, page: 1 }))
  }

  const handleFilterChange = (newFilters: Partial<ProductSearchParams>) => {
    setFilters(prev => ({ ...prev, ...newFilters, page: 1 }))
  }

  const clearFilters = () => {
    setFilters({ page: 1, limit: 12, sortBy: 'newest' })
    setSearchQuery('')
  }

  const toggleFilter = (filterName: string) => {
    setOpenFilters(prev => ({ ...prev, [filterName]: !prev[filterName] }))
  }

  if (productsError) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <h3 className="text-lg font-medium text-gray-900">Error loading products</h3>
          <p className="text-gray-600">Please try again later.</p>
          <Button onClick={() => window.location.reload()} className="mt-4">
            Retry
          </Button>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <div className="bg-white shadow-sm">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <div className="flex items-center justify-between">
            <h1 className="text-2xl font-bold text-gray-900">Products</h1>
            <div className="flex items-center space-x-4">
              <form onSubmit={handleSearch} className="flex space-x-2">
                <Input
                  type="search"
                  placeholder="Search products..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="w-64"
                />
                <Button type="submit" variant="primary">
                  Search
                </Button>
              </form>
            </div>
          </div>
        </div>
      </div>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="lg:grid lg:grid-cols-4 lg:gap-8">
          {/* Filters Sidebar */}
          <div className="lg:col-span-1">
            <div className="card p-6 sticky top-4">
              <div className="flex items-center justify-between mb-4">
                <h3 className="text-lg font-medium text-gray-900">Filters</h3>
                <Button variant="ghost" onClick={clearFilters} className="text-sm">
                  Clear all
                </Button>
              </div>

              <div className="space-y-4">
                {/* Categories */}
                <FilterSection
                  title="Categories"
                  isOpen={openFilters.category}
                  onToggle={() => toggleFilter('category')}
                >
                  <div className="space-y-2">
                    {categories?.map((category) => (
                      <label key={category.id} className="flex items-center space-x-2">
                        <input
                          type="radio"
                          name="category"
                          value={category.id}
                          checked={filters.category === category.id}
                          onChange={(e) => handleFilterChange({ category: e.target.value })}
                          className="h-4 w-4 text-primary-600 focus:ring-primary-500 border-gray-300"
                        />
                        <span className="text-sm text-gray-700">{category.name}</span>
                      </label>
                    ))}
                  </div>
                </FilterSection>

                {/* Brands */}
                <FilterSection
                  title="Brands"
                  isOpen={openFilters.brand}
                  onToggle={() => toggleFilter('brand')}
                >
                  <div className="space-y-2">
                    {brands?.map((brand) => (
                      <label key={brand.id} className="flex items-center space-x-2">
                        <input
                          type="radio"
                          name="brand"
                          value={brand.id}
                          checked={filters.brand === brand.id}
                          onChange={(e) => handleFilterChange({ brand: e.target.value })}
                          className="h-4 w-4 text-primary-600 focus:ring-primary-500 border-gray-300"
                        />
                        <span className="text-sm text-gray-700">{brand.name}</span>
                      </label>
                    ))}
                  </div>
                </FilterSection>

                {/* Price Range */}
                <FilterSection
                  title="Price Range"
                  isOpen={openFilters.price}
                  onToggle={() => toggleFilter('price')}
                >
                  <div className="grid grid-cols-2 gap-2">
                    <Input
                      type="number"
                      placeholder="Min"
                      value={filters.minPrice || ''}
                      onChange={(e) => handleFilterChange({ minPrice: parseFloat(e.target.value) || undefined })}
                      className="text-sm"
                    />
                    <Input
                      type="number"
                      placeholder="Max"
                      value={filters.maxPrice || ''}
                      onChange={(e) => handleFilterChange({ maxPrice: parseFloat(e.target.value) || undefined })}
                      className="text-sm"
                    />
                  </div>
                </FilterSection>

                {/* Rating */}
                <FilterSection
                  title="Rating"
                  isOpen={openFilters.rating}
                  onToggle={() => toggleFilter('rating')}
                >
                  <div className="space-y-2">
                    {[4, 3, 2, 1].map((rating) => (
                      <label key={rating} className="flex items-center space-x-2">
                        <input
                          type="radio"
                          name="rating"
                          value={rating}
                          checked={filters.rating === rating}
                          onChange={(e) => handleFilterChange({ rating: parseFloat(e.target.value) })}
                          className="h-4 w-4 text-primary-600 focus:ring-primary-500 border-gray-300"
                        />
                        <span className="flex items-center space-x-1 text-sm text-gray-700">
                          <span>{rating}</span>
                          <span className="text-yellow-400">★</span>
                          <span>& up</span>
                        </span>
                      </label>
                    ))}
                  </div>
                </FilterSection>

                {/* In Stock */}
                <div className="pt-4">
                  <label className="flex items-center space-x-2">
                    <input
                      type="checkbox"
                      checked={filters.inStock || false}
                      onChange={(e) => handleFilterChange({ inStock: e.target.checked })}
                      className="h-4 w-4 text-primary-600 focus:ring-primary-500 border-gray-300 rounded"
                    />
                    <span className="text-sm text-gray-700">In stock only</span>
                  </label>
                </div>
              </div>
            </div>
          </div>

          {/* Products Grid */}
          <div className="lg:col-span-3">
            {/* Toolbar */}
            <div className="flex items-center justify-between mb-6">
              <div className="text-sm text-gray-600">
                {productsData && productsData.pagination && (
                  <>
                    Showing {((productsData.pagination.page - 1) * productsData.pagination.limit) + 1}-
                    {Math.min(productsData.pagination.page * productsData.pagination.limit, productsData.pagination.total)} of{' '}
                    {productsData.pagination.total} results
                  </>
                )}
              </div>

              <select
                value={filters.sortBy || 'newest'}
                onChange={(e) => handleFilterChange({ sortBy: e.target.value as ProductSearchParams['sortBy'] })}
                className="border border-gray-300 rounded-md px-3 py-1 text-sm focus:ring-primary-500 focus:border-primary-500"
              >
                <option value="newest">Newest</option>
                <option value="price_asc">Price: Low to High</option>
                <option value="price_desc">Price: High to Low</option>
                <option value="name_asc">Name: A to Z</option>
                <option value="name_desc">Name: Z to A</option>
                <option value="rating">Highest Rated</option>
              </select>
            </div>

            {/* Products Grid */}
            {productsLoading ? (
              <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
                {[...Array(12)].map((_, i) => (
                  <div key={i} className="card animate-pulse">
                    <div className="aspect-square bg-gray-200 rounded-lg mb-4"></div>
                    <div className="space-y-2">
                      <div className="h-4 bg-gray-200 rounded w-3/4"></div>
                      <div className="h-4 bg-gray-200 rounded w-1/2"></div>
                      <div className="h-4 bg-gray-200 rounded w-1/4"></div>
                    </div>
                  </div>
                ))}
              </div>
            ) : productsData?.data.length === 0 ? (
              <div className="text-center py-12">
                <svg className="mx-auto h-12 w-12 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9.172 16.172a4 4 0 015.656 0M9 12h6m-6-4h6m2 5.291A7.962 7.962 0 0112 15c-2.34 0-4.44-.824-6.115-2.21A8 8 0 1112 3c-2.08 0-3.97.79-5.39 2.1A8 8 0 003 12a7.962 7.962 0 002.291 5.5z" />
                </svg>
                <h3 className="mt-2 text-sm font-medium text-gray-900">No products found</h3>
                <p className="mt-1 text-sm text-gray-500">Try adjusting your search or filter criteria.</p>
                <Button onClick={clearFilters} className="mt-4">
                  Clear filters
                </Button>
              </div>
            ) : (
              <>
                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
                  {productsData?.data.map((product) => (
                    <ProductCard key={product.id} product={product} />
                  ))}
                </div>

                {/* Pagination */}
                {productsData && productsData.pagination && productsData.pagination.total_pages > 1 && (
                  <div className="mt-8 flex justify-center">
                    <div className="flex items-center space-x-2">
                      <Button
                        variant="outline"
                        disabled={productsData.pagination.page === 1}
                        onClick={() => handleFilterChange({ page: productsData.pagination.page - 1 })}
                      >
                        Previous
                      </Button>

                      {[...Array(Math.min(5, productsData.pagination.total_pages))].map((_, i) => {
                        const page = i + 1
                        return (
                          <Button
                            key={page}
                            variant={page === productsData.pagination.page ? 'primary' : 'outline'}
                            onClick={() => handleFilterChange({ page })}
                          >
                            {page}
                          </Button>
                        )
                      })}

                      <Button
                        variant="outline"
                        disabled={productsData.pagination.page === productsData.pagination.total_pages}
                        onClick={() => handleFilterChange({ page: productsData.pagination.page + 1 })}
                      >
                        Next
                      </Button>
                    </div>
                  </div>
                )}
              </>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}