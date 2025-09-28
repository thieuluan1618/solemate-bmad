'use client'

import { useState } from 'react'
import { useParams, useRouter } from 'next/navigation'
import { Button, Input } from '@/components/ui'
import { useGetProductQuery, useGetRelatedProductsQuery, useGetProductReviewsQuery } from '@/store/api/productApi'
import { useAddToCartMutation } from '@/store/api/cartApi'
import { useAddToWishlistMutation, useRemoveFromWishlistMutation, useGetWishlistQuery } from '@/store/api/wishlistApi'
import { useAppSelector } from '@/hooks/redux'
import { selectIsAuthenticated } from '@/store/slices/authSlice'
import { Product, ProductSize, ProductColor } from '@/types'

interface ProductImageGalleryProps {
  images: any[]
  productName: string
}

function ProductImageGallery({ images, productName }: ProductImageGalleryProps) {
  const [selectedImageIndex, setSelectedImageIndex] = useState(0)

  if (!images || images.length === 0) {
    return (
      <div className="aspect-square bg-gray-100 rounded-lg flex items-center justify-center">
        <svg className="w-24 h-24 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
        </svg>
      </div>
    )
  }

  return (
    <div className="space-y-4">
      {/* Main Image */}
      <div className="aspect-square bg-gray-100 rounded-lg overflow-hidden">
        <img
          src={images[selectedImageIndex]?.url}
          alt={images[selectedImageIndex]?.altText || productName}
          className="w-full h-full object-cover"
        />
      </div>

      {/* Thumbnail Gallery */}
      {images.length > 1 && (
        <div className="grid grid-cols-4 gap-2">
          {images.map((image, index) => (
            <button
              key={image.id}
              onClick={() => setSelectedImageIndex(index)}
              className={`aspect-square bg-gray-100 rounded-lg overflow-hidden border-2 transition-colors ${
                index === selectedImageIndex ? 'border-primary-500' : 'border-transparent hover:border-gray-300'
              }`}
            >
              <img
                src={image.url}
                alt={image.altText || `${productName} view ${index + 1}`}
                className="w-full h-full object-cover"
              />
            </button>
          ))}
        </div>
      )}
    </div>
  )
}

interface AddToCartSectionProps {
  product: Product
}

function AddToCartSection({ product }: AddToCartSectionProps) {
  const [selectedSize, setSelectedSize] = useState<ProductSize | null>(null)
  const [selectedColor, setSelectedColor] = useState<ProductColor | null>(null)
  const [quantity, setQuantity] = useState(1)
  const router = useRouter()

  const [addToCart, { isLoading: addingToCart }] = useAddToCartMutation()

  const handleAddToCart = async () => {
    try {
      await addToCart({
        productId: product.id,
        quantity,
        sizeId: selectedSize?.id,
        colorId: selectedColor?.id,
      }).unwrap()

      // Show success message or redirect to cart
      const goToCart = window.confirm('Item added to cart! Would you like to view your cart?')
      if (goToCart) {
        router.push('/cart')
      }
    } catch (error: any) {
      alert(error?.data?.message || 'Failed to add item to cart. Please try again.')
    }
  }

  const canAddToCart = product.stockQuantity > 0 &&
    (!product.sizes?.length || selectedSize) &&
    (!product.colors?.length || selectedColor)

  return (
    <div className="space-y-6">
      {/* Size Selection */}
      {product.sizes && product.sizes.length > 0 && (
        <div>
          <h4 className="text-sm font-medium text-gray-900 mb-3">Size</h4>
          <div className="grid grid-cols-4 gap-2">
            {product.sizes.map((size) => (
              <button
                key={size.id}
                onClick={() => setSelectedSize(size)}
                disabled={size.stockQuantity === 0}
                className={`px-3 py-2 text-sm border rounded-md transition-colors ${
                  selectedSize?.id === size.id
                    ? 'border-primary-500 bg-primary-50 text-primary-700'
                    : size.stockQuantity === 0
                    ? 'border-gray-200 bg-gray-100 text-gray-400 cursor-not-allowed'
                    : 'border-gray-300 hover:border-gray-400'
                }`}
              >
                {size.name}
              </button>
            ))}
          </div>
        </div>
      )}

      {/* Color Selection */}
      {product.colors && product.colors.length > 0 && (
        <div>
          <h4 className="text-sm font-medium text-gray-900 mb-3">Color</h4>
          <div className="flex flex-wrap gap-2">
            {product.colors.map((color) => (
              <button
                key={color.id}
                onClick={() => setSelectedColor(color)}
                disabled={color.stockQuantity === 0}
                className={`w-8 h-8 rounded-full border-2 transition-all ${
                  selectedColor?.id === color.id
                    ? 'border-primary-500 ring-2 ring-primary-200'
                    : color.stockQuantity === 0
                    ? 'border-gray-200 opacity-50 cursor-not-allowed'
                    : 'border-gray-300 hover:border-gray-400'
                }`}
                style={{ backgroundColor: color.value }}
                title={color.name}
              />
            ))}
          </div>
          {selectedColor && (
            <p className="text-sm text-gray-600 mt-1">{selectedColor.name}</p>
          )}
        </div>
      )}

      {/* Quantity */}
      <div>
        <h4 className="text-sm font-medium text-gray-900 mb-3">Quantity</h4>
        <div className="flex items-center space-x-2">
          <button
            onClick={() => setQuantity(Math.max(1, quantity - 1))}
            className="w-8 h-8 flex items-center justify-center border border-gray-300 rounded-md hover:bg-gray-50"
          >
            -
          </button>
          <Input
            type="number"
            min="1"
            max={product.stockQuantity}
            value={quantity}
            onChange={(e) => setQuantity(Math.max(1, parseInt(e.target.value) || 1))}
            className="w-20 text-center"
          />
          <button
            onClick={() => setQuantity(Math.min(product.stockQuantity, quantity + 1))}
            className="w-8 h-8 flex items-center justify-center border border-gray-300 rounded-md hover:bg-gray-50"
          >
            +
          </button>
        </div>
      </div>

      {/* Add to Cart Button */}
      <div className="space-y-3">
        <Button
          onClick={handleAddToCart}
          disabled={!canAddToCart || addingToCart}
          className="w-full"
          variant="primary"
        >
          {addingToCart ? 'Adding to Cart...' : product.stockQuantity === 0 ? 'Out of Stock' : 'Add to Cart'}
        </Button>

        <WishlistButton productId={product.id} />
      </div>

      {/* Stock Status */}
      <div className="text-sm">
        {product.stockQuantity > 0 ? (
          <span className="text-green-600">✓ In stock ({product.stockQuantity} available)</span>
        ) : (
          <span className="text-red-600">Out of stock</span>
        )}
      </div>
    </div>
  )
}

interface ReviewSectionProps {
  productId: string
  averageRating: number
  reviewCount: number
}

function ReviewSection({ productId, averageRating, reviewCount }: ReviewSectionProps) {
  const { data: reviewsData, isLoading } = useGetProductReviewsQuery({ productId })

  return (
    <div className="space-y-6">
      {/* Rating Summary */}
      <div className="flex items-center space-x-4">
        <div className="flex items-center space-x-2">
          <span className="text-2xl font-bold">{(averageRating || 0).toFixed(1)}</span>
          <div className="flex items-center">
            {[...Array(5)].map((_, i) => (
              <svg
                key={i}
                className={`w-5 h-5 ${
                  i < Math.floor(averageRating || 0) ? 'text-yellow-400' : 'text-gray-300'
                }`}
                fill="currentColor"
                viewBox="0 0 24 24"
              >
                <path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z" />
              </svg>
            ))}
          </div>
          <span className="text-gray-600">({reviewCount} reviews)</span>
        </div>
      </div>

      {/* Reviews List */}
      {isLoading ? (
        <div className="space-y-4">
          {[...Array(3)].map((_, i) => (
            <div key={i} className="border-b border-gray-200 pb-4 animate-pulse">
              <div className="h-4 bg-gray-200 rounded w-1/4 mb-2"></div>
              <div className="h-4 bg-gray-200 rounded w-3/4 mb-2"></div>
              <div className="h-16 bg-gray-200 rounded"></div>
            </div>
          ))}
        </div>
      ) : reviewsData && reviewsData.data.length > 0 ? (
        <div className="space-y-6">
          {reviewsData.data.map((review: any) => (
            <div key={review.id} className="border-b border-gray-200 pb-6">
              <div className="flex items-start justify-between mb-2">
                <div>
                  <div className="flex items-center space-x-2 mb-1">
                    <span className="font-medium text-gray-900">
                      {review.user?.firstName} {review.user?.lastName}
                    </span>
                    {review.isVerifiedPurchase && (
                      <span className="text-xs bg-green-100 text-green-800 px-2 py-1 rounded">
                        Verified Purchase
                      </span>
                    )}
                  </div>
                  <div className="flex items-center space-x-2">
                    <div className="flex items-center">
                      {[...Array(5)].map((_, i) => (
                        <svg
                          key={i}
                          className={`w-4 h-4 ${
                            i < review.rating ? 'text-yellow-400' : 'text-gray-300'
                          }`}
                          fill="currentColor"
                          viewBox="0 0 24 24"
                        >
                          <path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z" />
                        </svg>
                      ))}
                    </div>
                    <span className="text-sm text-gray-500">
                      {new Date(review.createdAt).toLocaleDateString()}
                    </span>
                  </div>
                </div>
              </div>

              <h4 className="font-medium text-gray-900 mb-2">{review.title}</h4>
              <p className="text-gray-700 mb-3">{review.comment}</p>

              {review.isRecommended && (
                <p className="text-sm text-green-600">✓ Recommends this product</p>
              )}

              {review.helpfulCount > 0 && (
                <div className="flex items-center space-x-2 mt-3">
                  <button className="text-sm text-gray-500 hover:text-gray-700">
                    Helpful ({review.helpfulCount})
                  </button>
                </div>
              )}
            </div>
          ))}
        </div>
      ) : (
        <div className="text-center py-8">
          <p className="text-gray-500">No reviews yet. Be the first to review this product!</p>
        </div>
      )}
    </div>
  )
}

interface WishlistButtonProps {
  productId: string
}

function WishlistButton({ productId }: WishlistButtonProps) {
  const router = useRouter()
  const isAuthenticated = useAppSelector(selectIsAuthenticated)
  const { data: wishlist } = useGetWishlistQuery(undefined, { skip: !isAuthenticated })
  const [addToWishlist, { isLoading: isAdding }] = useAddToWishlistMutation()
  const [removeFromWishlist, { isLoading: isRemoving }] = useRemoveFromWishlistMutation()

  const isInWishlist = wishlist?.items.some(item => item.product.id === productId)
  const isLoading = isAdding || isRemoving

  const handleWishlistToggle = async () => {
    if (!isAuthenticated) {
      router.push('/login?redirect=' + encodeURIComponent(window.location.pathname))
      return
    }

    try {
      if (isInWishlist) {
        await removeFromWishlist({ productId }).unwrap()
      } else {
        await addToWishlist({ productId }).unwrap()
      }
    } catch (error: any) {
      alert(error?.data?.message || 'Failed to update wishlist')
    }
  }

  return (
    <Button
      onClick={handleWishlistToggle}
      variant="outline"
      className="w-full"
      disabled={isLoading}
    >
      <svg
        className={`w-4 h-4 mr-2 ${isInWishlist ? 'fill-current text-red-500' : ''}`}
        fill={isInWishlist ? 'currentColor' : 'none'}
        stroke="currentColor"
        viewBox="0 0 24 24"
      >
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z" />
      </svg>
      {isLoading
        ? 'Updating...'
        : isInWishlist
        ? 'Remove from Wishlist'
        : 'Add to Wishlist'
      }
    </Button>
  )
}

export default function ProductDetailPage() {
  const params = useParams()
  const router = useRouter()
  const productId = params.id as string

  const { data: product, isLoading, error } = useGetProductQuery(productId)
  const { data: relatedProducts } = useGetRelatedProductsQuery({ productId })

  if (isLoading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600"></div>
      </div>
    )
  }

  if (error || !product) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <h3 className="text-lg font-medium text-gray-900">Product not found</h3>
          <p className="text-gray-600">The product you're looking for doesn't exist.</p>
          <Button onClick={() => router.push('/products')} className="mt-4">
            Browse Products
          </Button>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Breadcrumb */}
      <div className="bg-white border-b border-gray-200">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <nav className="flex" aria-label="Breadcrumb">
            <ol className="flex items-center space-x-2 text-sm">
              <li>
                <button onClick={() => router.push('/')} className="text-gray-500 hover:text-gray-700">
                  Home
                </button>
              </li>
              <li className="text-gray-500">/</li>
              <li>
                <button onClick={() => router.push('/products')} className="text-gray-500 hover:text-gray-700">
                  Products
                </button>
              </li>
              <li className="text-gray-500">/</li>
              {product.category && (
                <li>
                  <button onClick={() => router.push(`/products?category=${product.category.id}`)} className="text-gray-500 hover:text-gray-700">
                    {product.category.name}
                  </button>
                </li>
              )}
              {product.category && <li className="text-gray-500">/</li>}
              <li className="text-gray-900 font-medium">{product.name}</li>
            </ol>
          </nav>
        </div>
      </div>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="lg:grid lg:grid-cols-2 lg:gap-12">
          {/* Product Images */}
          <div>
            <ProductImageGallery images={product.images} productName={product.name} />
          </div>

          {/* Product Information */}
          <div className="mt-8 lg:mt-0">
            <div className="space-y-6">
              {/* Header */}
              <div>
                <div className="flex items-center justify-between mb-2">
                  <span className="text-sm text-gray-500">{product.brand?.name}</span>
                  <div className="flex items-center space-x-1">
                    <span className="text-yellow-400">★</span>
                    <span className="text-sm text-gray-600">{(product.rating || 0).toFixed(1)}</span>
                    <span className="text-sm text-gray-500">({product.reviewCount} reviews)</span>
                  </div>
                </div>

                <h1 className="text-2xl font-bold text-gray-900">{product.name}</h1>

                <div className="flex items-center space-x-3 mt-2">
                  <span className="text-3xl font-bold text-gray-900">
                    ${product.price.toFixed(2)}
                  </span>
                  {product.originalPrice && product.originalPrice > product.price && (
                    <>
                      <span className="text-xl text-gray-500 line-through">
                        ${product.originalPrice.toFixed(2)}
                      </span>
                      <span className="text-lg font-medium text-red-600">
                        {product.discountPercentage}% off
                      </span>
                    </>
                  )}
                </div>
              </div>

              {/* Description */}
              <div>
                <h3 className="text-lg font-medium text-gray-900 mb-2">Description</h3>
                <p className="text-gray-700">{product.description}</p>
              </div>

              {/* Add to Cart Section */}
              <AddToCartSection product={product} />

              {/* Specifications */}
              {product.specifications && product.specifications.length > 0 && (
                <div>
                  <h3 className="text-lg font-medium text-gray-900 mb-4">Specifications</h3>
                  <dl className="grid grid-cols-1 gap-x-4 gap-y-2 sm:grid-cols-2">
                    {product.specifications.map((spec) => (
                      <div key={spec.id} className="flex">
                        <dt className="text-sm font-medium text-gray-500 w-1/2">{spec.name}:</dt>
                        <dd className="text-sm text-gray-900 w-1/2">{spec.value}</dd>
                      </div>
                    ))}
                  </dl>
                </div>
              )}
            </div>
          </div>
        </div>

        {/* Reviews Section */}
        <div className="mt-16">
          <div className="card p-8">
            <h2 className="text-2xl font-bold text-gray-900 mb-6">Customer Reviews</h2>
            <ReviewSection
              productId={product.id}
              averageRating={product.rating}
              reviewCount={product.reviewCount}
            />
          </div>
        </div>

        {/* Related Products */}
        {relatedProducts && relatedProducts.length > 0 && (
          <div className="mt-16">
            <h2 className="text-2xl font-bold text-gray-900 mb-6">Related Products</h2>
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
              {relatedProducts.map((relatedProduct) => (
                <div
                  key={relatedProduct.id}
                  className="card group cursor-pointer"
                  onClick={() => router.push(`/products/${relatedProduct.id}`)}
                >
                  <div className="aspect-square overflow-hidden rounded-lg bg-gray-100 mb-4">
                    {relatedProduct.images?.[0] ? (
                      <img
                        src={relatedProduct.images[0].url}
                        alt={relatedProduct.images[0].altText || relatedProduct.name}
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
                    <h3 className="font-semibold text-gray-900 group-hover:text-primary-600 transition-colors">
                      {relatedProduct.name}
                    </h3>
                    <p className="text-lg font-bold text-gray-900">
                      ${relatedProduct.price.toFixed(2)}
                    </p>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}
      </div>
    </div>
  )
}