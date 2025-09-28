'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import Link from 'next/link'
import { Button, Input } from '@/components/ui'
import { useGetCartQuery, useUpdateCartItemMutation, useRemoveFromCartMutation, useClearCartMutation, useApplyPromoCodeMutation, useRemovePromoCodeMutation } from '@/store/api/cartApi'
import { CartItem as CartItemType } from '@/types'

interface CartItemProps {
  item: CartItemType
  onUpdateQuantity: (itemId: string, quantity: number) => void
  onRemove: (itemId: string) => void
}

function CartItem({ item, onUpdateQuantity, onRemove }: CartItemProps) {
  const [quantity, setQuantity] = useState(item.quantity)

  const handleQuantityChange = (newQuantity: number) => {
    if (newQuantity >= 1 && newQuantity <= 99) {
      setQuantity(newQuantity)
      onUpdateQuantity(item.id, newQuantity)
    }
  }

  return (
    <div className="flex items-start space-x-4 py-6 border-b border-gray-200">
      {/* Product Image */}
      <div className="flex-shrink-0">
        <Link href={`/products/${item.product_id}`}>
          <div className="w-24 h-24 bg-gray-100 rounded-lg overflow-hidden">
            {item.image_url ? (
              <img
                src={item.image_url}
                alt={item.name}
                className="w-full h-full object-cover hover:scale-105 transition-transform"
              />
            ) : (
              <div className="w-full h-full flex items-center justify-center text-gray-400">
                <svg className="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
                </svg>
              </div>
            )}
          </div>
        </Link>
      </div>

      {/* Product Details */}
      <div className="flex-1 min-w-0">
        <div className="flex items-start justify-between">
          <div className="space-y-1">
            <Link href={`/products/${item.product_id}`}>
              <h3 className="text-sm font-medium text-gray-900 hover:text-blue-600">
                {item.name}
              </h3>
            </Link>
            <p className="text-sm text-gray-500">SKU: {item.sku}</p>

            {/* Variant Details */}
            <div className="flex items-center space-x-4 text-sm text-gray-600">
              {item.size && (
                <span>Size: {item.size}</span>
              )}
              {item.color && (
                <span>Color: {item.color}</span>
              )}
            </div>

            {/* Price */}
            <div className="flex items-center space-x-2">
              <span className="text-lg font-semibold text-gray-900">
                ${item.price.toFixed(2)}
              </span>
              {item.discount > 0 && (
                <span className="text-sm text-green-600">
                  -${item.discount.toFixed(2)}
                </span>
              )}
            </div>
          </div>

          {/* Item Total */}
          <div className="text-right">
            <p className="text-lg font-semibold text-gray-900">
              ${item.total_price.toFixed(2)}
            </p>
          </div>
        </div>

        {/* Quantity and Actions */}
        <div className="flex items-center justify-between mt-4">
          <div className="flex items-center space-x-2">
            <label htmlFor={`quantity-${item.id}`} className="text-sm text-gray-600">
              Qty:
            </label>
            <div className="flex items-center space-x-1">
              <button
                onClick={() => handleQuantityChange(quantity - 1)}
                disabled={quantity <= 1}
                className="w-8 h-8 flex items-center justify-center border border-gray-300 rounded-md hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                -
              </button>
              <Input
                id={`quantity-${item.id}`}
                type="number"
                min="1"
                max="99"
                value={quantity}
                onChange={(e) => handleQuantityChange(parseInt(e.target.value) || 1)}
                className="w-16 text-center"
              />
              <button
                onClick={() => handleQuantityChange(quantity + 1)}
                disabled={quantity >= 99}
                className="w-8 h-8 flex items-center justify-center border border-gray-300 rounded-md hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                +
              </button>
            </div>
          </div>

          <div className="flex items-center space-x-4">
            <button
              onClick={() => {
                // TODO: Implement add to wishlist
                console.log('Move to wishlist:', item.id)
              }}
              className="text-sm text-gray-500 hover:text-blue-600"
            >
              Save for later
            </button>
            <button
              onClick={() => onRemove(item.id)}
              className="text-sm text-red-600 hover:text-red-800"
            >
              Remove
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}

interface OrderSummaryProps {
  cart: any
  onApplyPromoCode: (code: string) => void
  onRemovePromoCode: () => void
  promoCodeLoading: boolean
}

function OrderSummary({ cart, onApplyPromoCode, onRemovePromoCode, promoCodeLoading }: OrderSummaryProps) {
  const [promoCode, setPromoCode] = useState('')
  const router = useRouter()

  const subtotal = cart.total_price || 0
  const shipping = subtotal > 50 ? 0 : 9.99 // Free shipping over $50
  const tax = subtotal * 0.08 // 8% tax
  const discount = cart.discountAmount || 0
  const total = subtotal + shipping + tax - discount

  const handleApplyPromoCode = (e: React.FormEvent) => {
    e.preventDefault()
    if (promoCode.trim()) {
      onApplyPromoCode(promoCode.trim())
    }
  }

  return (
    <div className="card p-6 sticky top-4">
      <h2 className="text-lg font-semibold text-gray-900 mb-4">Order Summary</h2>

      <div className="space-y-3 text-sm">
        <div className="flex justify-between">
          <span className="text-gray-600">Subtotal ({cart.total_items} items)</span>
          <span className="font-medium">${subtotal.toFixed(2)}</span>
        </div>

        <div className="flex justify-between">
          <span className="text-gray-600">Shipping</span>
          <span className="font-medium">
            {shipping === 0 ? (
              <span className="text-green-600">Free</span>
            ) : (
              `$${shipping.toFixed(2)}`
            )}
          </span>
        </div>

        <div className="flex justify-between">
          <span className="text-gray-600">Tax</span>
          <span className="font-medium">${tax.toFixed(2)}</span>
        </div>

        {discount > 0 && (
          <div className="flex justify-between">
            <span className="text-gray-600">Discount</span>
            <span className="font-medium text-green-600">-${discount.toFixed(2)}</span>
          </div>
        )}

        <div className="border-t border-gray-200 pt-3">
          <div className="flex justify-between">
            <span className="text-base font-semibold text-gray-900">Total</span>
            <span className="text-base font-semibold text-gray-900">${total.toFixed(2)}</span>
          </div>
        </div>
      </div>

      {/* Promo Code */}
      <div className="mt-6">
        {cart.promoCode ? (
          <div className="flex items-center justify-between p-3 bg-green-50 border border-green-200 rounded-md">
            <div>
              <span className="text-sm font-medium text-green-800">
                Promo code "{cart.promoCode}" applied
              </span>
              <p className="text-xs text-green-600">You saved ${discount.toFixed(2)}!</p>
            </div>
            <button
              onClick={onRemovePromoCode}
              className="text-green-600 hover:text-green-800"
            >
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
        ) : (
          <form onSubmit={handleApplyPromoCode} className="space-y-2">
            <Input
              placeholder="Enter promo code"
              value={promoCode}
              onChange={(e) => setPromoCode(e.target.value)}
            />
            <Button
              type="submit"
              variant="outline"
              className="w-full"
              disabled={!promoCode.trim() || promoCodeLoading}
            >
              {promoCodeLoading ? 'Applying...' : 'Apply Code'}
            </Button>
          </form>
        )}
      </div>

      {/* Free Shipping Notice */}
      {shipping > 0 && subtotal < 50 && (
        <div className="mt-4 p-3 bg-blue-50 border border-blue-200 rounded-md">
          <p className="text-sm text-blue-800">
            Add ${(50 - subtotal).toFixed(2)} more for free shipping!
          </p>
        </div>
      )}

      {/* Checkout Button */}
      <div className="mt-6 space-y-3">
        <Button
          onClick={() => router.push('/checkout')}
          variant="primary"
          className="w-full"
          disabled={cart.total_items === 0}
        >
          Proceed to Checkout
        </Button>

        <Link href="/products">
          <Button variant="outline" className="w-full">
            Continue Shopping
          </Button>
        </Link>
      </div>

      {/* Security Notice */}
      <div className="mt-4 flex items-center justify-center space-x-1 text-xs text-gray-500">
        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
        </svg>
        <span>Secure checkout</span>
      </div>
    </div>
  )
}

export default function CartPage() {
  const router = useRouter()

  // API hooks
  const { data: cart, isLoading, error } = useGetCartQuery()
  const [updateCartItem] = useUpdateCartItemMutation()
  const [removeFromCart] = useRemoveFromCartMutation()
  const [clearCart] = useClearCartMutation()
  const [applyPromoCode, { isLoading: promoCodeLoading }] = useApplyPromoCodeMutation()
  const [removePromoCode] = useRemovePromoCodeMutation()

  const handleUpdateQuantity = async (itemId: string, quantity: number) => {
    try {
      await updateCartItem({ itemId, quantity }).unwrap()
    } catch (error) {
      console.error('Failed to update quantity:', error)
    }
  }

  const handleRemoveItem = async (itemId: string) => {
    try {
      await removeFromCart({ itemId }).unwrap()
    } catch (error) {
      console.error('Failed to remove item:', error)
    }
  }

  const handleClearCart = async () => {
    if (window.confirm('Are you sure you want to clear your cart?')) {
      try {
        await clearCart().unwrap()
      } catch (error) {
        console.error('Failed to clear cart:', error)
      }
    }
  }

  const handleApplyPromoCode = async (code: string) => {
    try {
      await applyPromoCode({ code }).unwrap()
    } catch (error: any) {
      alert(error?.data?.message || 'Invalid promo code')
    }
  }

  const handleRemovePromoCode = async () => {
    try {
      await removePromoCode().unwrap()
    } catch (error) {
      console.error('Failed to remove promo code:', error)
    }
  }

  if (isLoading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
      </div>
    )
  }

  // Loading state
  if (isLoading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 mx-auto"></div>
          <p className="mt-2 text-gray-600">Loading your cart...</p>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <h3 className="text-lg font-medium text-gray-900">Error loading cart</h3>
          <p className="text-gray-600">Please try again later.</p>
          <Button onClick={() => window.location.reload()} className="mt-4">
            Retry
          </Button>
        </div>
      </div>
    )
  }

  // Empty cart state
  if (!cart || !cart.items || cart.items.length === 0) {
    return (
      <div className="min-h-screen bg-gray-50">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className="text-center py-16">
            <svg className="mx-auto h-24 w-24 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 11V7a4 4 0 00-8 0v4M5 9h14l-1 12H6L5 9z" />
            </svg>
            <h2 className="mt-4 text-2xl font-bold text-gray-900">Your cart is empty</h2>
            <p className="mt-2 text-gray-600">
              Looks like you haven't added anything to your cart yet. Start shopping to fill it up!
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
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <div className="bg-white shadow-sm">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <div className="flex items-center justify-between">
            <h1 className="text-2xl font-bold text-gray-900">Shopping Cart</h1>
            <div className="flex items-center space-x-4">
              <span className="text-sm text-gray-600">
                {cart.total_items} {cart.total_items === 1 ? 'item' : 'items'}
              </span>
              {cart.total_items > 0 && (
                <button
                  onClick={handleClearCart}
                  className="text-sm text-red-600 hover:text-red-800"
                >
                  Clear cart
                </button>
              )}
            </div>
          </div>
        </div>
      </div>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="lg:grid lg:grid-cols-3 lg:gap-8">
          {/* Cart Items */}
          <div className="lg:col-span-2">
            <div className="card p-6">
              <div className="space-y-0">
                {cart?.items?.map((item) => (
                  <CartItem
                    key={item.id}
                    item={item}
                    onUpdateQuantity={handleUpdateQuantity}
                    onRemove={handleRemoveItem}
                  />
                )) || []}
              </div>
            </div>
          </div>

          {/* Order Summary */}
          <div className="lg:col-span-1 mt-8 lg:mt-0">
            <OrderSummary
              cart={cart}
              onApplyPromoCode={handleApplyPromoCode}
              onRemovePromoCode={handleRemovePromoCode}
              promoCodeLoading={promoCodeLoading}
            />
          </div>
        </div>
      </div>
    </div>
  )
}