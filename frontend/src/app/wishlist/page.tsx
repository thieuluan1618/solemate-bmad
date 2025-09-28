'use client'

import { useEffect } from 'react'
import { useRouter } from 'next/navigation'
import Link from 'next/link'
import { Button } from '@/components/ui'
import Layout from '@/components/layout/Layout'
import { useGetWishlistQuery, useRemoveFromWishlistMutation, useClearWishlistMutation, useMoveToCartMutation } from '@/store/api/wishlistApi'
import { useAppSelector } from '@/hooks/redux'
import { selectIsAuthenticated } from '@/store/slices/authSlice'
import type { WishlistItem } from '@/types'

interface WishlistItemCardProps {
  item: WishlistItem
  onRemove: (productId: string) => void
  onMoveToCart: (productId: string) => void
  isLoading: boolean
}

function WishlistItemCard({ item, onRemove, onMoveToCart, isLoading }: WishlistItemCardProps) {
  const router = useRouter()

  return (
    <div className="card group">
      <div className="aspect-square overflow-hidden rounded-t-lg bg-gray-100 relative">
        {item.product.images?.[0] ? (
          <img
            src={item.product.images[0].url}
            alt={item.product.images[0].altText || item.product.name}
            className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-200 cursor-pointer"
            onClick={() => router.push(`/products/${item.product.id}`)}
          />
        ) : (
          <div
            className="w-full h-full flex items-center justify-center text-gray-400 cursor-pointer"
            onClick={() => router.push(`/products/${item.product.id}`)}
          >
            <svg className="w-12 h-12" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
            </svg>
          </div>
        )}

        {/* Discount Badge */}
        {item.product.originalPrice && item.product.originalPrice > item.product.price && (
          <div className="absolute top-2 left-2 bg-red-600 text-white text-xs font-bold px-2 py-1 rounded">
            {item.product.discountPercentage}% OFF
          </div>
        )}

        {/* Stock Status */}
        {item.product.stockQuantity === 0 && (
          <div className="absolute top-2 right-2 bg-gray-900 text-white text-xs font-medium px-2 py-1 rounded">
            Out of Stock
          </div>
        )}

        {/* Remove Button */}
        <button
          onClick={() => onRemove(item.product.id)}
          disabled={isLoading}
          className="absolute top-2 right-2 w-8 h-8 bg-white bg-opacity-90 hover:bg-opacity-100 rounded-full flex items-center justify-center text-gray-600 hover:text-red-600 transition-colors opacity-0 group-hover:opacity-100 disabled:opacity-50"
          title="Remove from wishlist"
        >
          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <div className="p-4 space-y-3">
        {/* Product Info */}
        <div className="space-y-1">
          <div className="flex items-center justify-between">
            <span className="text-sm text-gray-500">{item.product.brand?.name}</span>
            <div className="flex items-center space-x-1">
              <span className="text-yellow-400 text-sm">★</span>
              <span className="text-sm text-gray-600">{item.product.rating}</span>
            </div>
          </div>

          <h3
            className="font-semibold text-gray-900 hover:text-blue-600 transition-colors cursor-pointer"
            onClick={() => router.push(`/products/${item.product.id}`)}
          >
            {item.product.name}
          </h3>

          <div className="flex items-center space-x-2">
            <span className="text-lg font-bold text-gray-900">
              ${item.product.price.toFixed(2)}
            </span>
            {item.product.originalPrice && item.product.originalPrice > item.product.price && (
              <span className="text-sm text-gray-500 line-through">
                ${item.product.originalPrice.toFixed(2)}
              </span>
            )}
          </div>

          <p className="text-sm text-gray-600">{item.product.category?.name}</p>
        </div>

        {/* Actions */}
        <div className="space-y-2">
          <Button
            onClick={() => onMoveToCart(item.product.id)}
            disabled={item.product.stockQuantity === 0 || isLoading}
            variant="primary"
            className="w-full"
          >
            {item.product.stockQuantity === 0 ? 'Out of Stock' : 'Add to Cart'}
          </Button>

          <Button
            onClick={() => router.push(`/products/${item.product.id}`)}
            variant="outline"
            className="w-full"
          >
            View Details
          </Button>
        </div>

        {/* Added Date */}
        <p className="text-xs text-gray-500 text-center">
          Added {new Date(item.addedAt).toLocaleDateString()}
        </p>
      </div>
    </div>
  )
}

export default function WishlistPage() {
  const router = useRouter()
  const isAuthenticated = useAppSelector(selectIsAuthenticated)

  const { data: wishlist, isLoading, error } = useGetWishlistQuery()
  const [removeFromWishlist, { isLoading: isRemoving }] = useRemoveFromWishlistMutation()
  const [clearWishlist, { isLoading: isClearing }] = useClearWishlistMutation()
  const [moveToCart, { isLoading: isMoving }] = useMoveToCartMutation()

  useEffect(() => {
    if (!isAuthenticated) {
      router.push('/login?redirect=/wishlist')
    }
  }, [isAuthenticated, router])

  const handleRemoveItem = async (productId: string) => {
    try {
      await removeFromWishlist({ productId }).unwrap()
    } catch (error: any) {
      alert(error?.data?.message || 'Failed to remove item from wishlist')
    }
  }

  const handleMoveToCart = async (productId: string) => {
    try {
      const result = await moveToCart({ productId, quantity: 1 }).unwrap()
      const goToCart = window.confirm('Item added to cart! Would you like to view your cart?')
      if (goToCart) {
        router.push('/cart')
      }
    } catch (error: any) {
      alert(error?.data?.message || 'Failed to add item to cart')
    }
  }

  const handleClearWishlist = async () => {
    if (window.confirm('Are you sure you want to clear your entire wishlist?')) {
      try {
        await clearWishlist().unwrap()
      } catch (error: any) {
        alert(error?.data?.message || 'Failed to clear wishlist')
      }
    }
  }

  if (!isAuthenticated) {
    return (
      <Layout>
        <div className="min-h-screen flex items-center justify-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
        </div>
      </Layout>
    )
  }

  if (error) {
    return (
      <Layout>
        <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className="text-center">
            <h3 className="text-lg font-medium text-gray-900">Error loading wishlist</h3>
            <p className="text-gray-600">Please try again later.</p>
            <Button onClick={() => window.location.reload()} className="mt-4">
              Retry
            </Button>
          </div>
        </div>
      </Layout>
    )
  }

  if (isLoading) {
    return (
      <Layout>
        <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className="mb-8">
            <div className="h-8 bg-gray-200 rounded w-1/4 mb-2 animate-pulse"></div>
            <div className="h-4 bg-gray-200 rounded w-1/3 animate-pulse"></div>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
            {[...Array(8)].map((_, i) => (
              <div key={i} className="card animate-pulse">
                <div className="aspect-square bg-gray-200 rounded-t-lg"></div>
                <div className="p-4 space-y-3">
                  <div className="h-4 bg-gray-200 rounded w-3/4"></div>
                  <div className="h-4 bg-gray-200 rounded w-1/2"></div>
                  <div className="h-8 bg-gray-200 rounded"></div>
                  <div className="h-8 bg-gray-200 rounded"></div>
                </div>
              </div>
            ))}
          </div>
        </div>
      </Layout>
    )
  }

  // Empty wishlist state
  if (!wishlist || wishlist.items.length === 0) {
    return (
      <Layout>
        <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className="text-center py-16">
            <svg className="mx-auto h-24 w-24 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z" />
            </svg>
            <h2 className="mt-4 text-2xl font-bold text-gray-900">Your wishlist is empty</h2>
            <p className="mt-2 text-gray-600">
              Save items you love to your wishlist so you can easily find them later!
            </p>
            <div className="mt-8">
              <Link href="/products">
                <Button variant="primary" className="mr-4">
                  Start Shopping
                </Button>
              </Link>
              <Link href="/">
                <Button variant="outline">
                  Back to Home
                </Button>
              </Link>
            </div>
          </div>
        </div>
      </Layout>
    )
  }

  return (
    <Layout>
      <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Header */}
        <div className="flex items-center justify-between mb-8">
          <div>
            <h1 className="text-2xl font-bold text-gray-900">My Wishlist</h1>
            <p className="text-gray-600">
              {wishlist.items.length} {wishlist.items.length === 1 ? 'item' : 'items'} saved
            </p>
          </div>

          <div className="flex items-center space-x-4">
            {wishlist.items.length > 0 && (
              <button
                onClick={handleClearWishlist}
                disabled={isClearing}
                className="text-sm text-red-600 hover:text-red-800 disabled:opacity-50"
              >
                {isClearing ? 'Clearing...' : 'Clear All'}
              </button>
            )}
            <Link href="/products">
              <Button variant="primary">Continue Shopping</Button>
            </Link>
          </div>
        </div>

        {/* Wishlist Items Grid */}
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
          {wishlist.items.map((item) => (
            <WishlistItemCard
              key={item.id}
              item={item}
              onRemove={handleRemoveItem}
              onMoveToCart={handleMoveToCart}
              isLoading={isRemoving || isMoving}
            />
          ))}
        </div>

        {/* Additional Actions */}
        <div className="mt-12 text-center">
          <div className="inline-flex items-center space-x-1 text-sm text-gray-500">
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            <span>Items in your wishlist are saved across devices</span>
          </div>
        </div>
      </div>
    </Layout>
  )
}