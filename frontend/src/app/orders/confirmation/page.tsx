'use client'

import { useEffect, useState } from 'react'
import { useRouter, useSearchParams } from 'next/navigation'
import Link from 'next/link'
import { Button } from '@/components/ui'
import Layout from '@/components/layout/Layout'
import { useGetOrderQuery } from '@/store/api/orderApi'
import { useAppSelector } from '@/hooks/redux'
import { selectIsAuthenticated } from '@/store/slices/authSlice'

export default function OrderConfirmationPage() {
  const router = useRouter()
  const searchParams = useSearchParams()
  const isAuthenticated = useAppSelector(selectIsAuthenticated)
  const [orderId, setOrderId] = useState<string | null>(null)

  // Get order ID from URL params or localStorage (fallback)
  useEffect(() => {
    const orderIdFromParams = searchParams.get('orderId')
    const orderIdFromStorage = typeof window !== 'undefined' ? localStorage.getItem('lastOrderId') : null

    if (orderIdFromParams) {
      setOrderId(orderIdFromParams)
      // Clear from localStorage if we got it from params
      if (typeof window !== 'undefined') {
        localStorage.removeItem('lastOrderId')
      }
    } else if (orderIdFromStorage) {
      setOrderId(orderIdFromStorage)
      localStorage.removeItem('lastOrderId')
    } else {
      // No order ID found, redirect to orders page
      router.push('/orders')
    }
  }, [searchParams, router])

  const { data: order, isLoading, error } = useGetOrderQuery(orderId!, {
    skip: !orderId,
  })

  useEffect(() => {
    if (!isAuthenticated) {
      router.push('/login?redirect=/orders/confirmation')
    }
  }, [isAuthenticated, router])

  if (!isAuthenticated) {
    return (
      <Layout>
        <div className="min-h-screen flex items-center justify-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
        </div>
      </Layout>
    )
  }

  if (isLoading) {
    return (
      <Layout>
        <div className="min-h-screen flex items-center justify-center">
          <div className="text-center">
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
            <p className="text-gray-600">Loading your order confirmation...</p>
          </div>
        </div>
      </Layout>
    )
  }

  if (error || !order) {
    return (
      <Layout>
        <div className="min-h-screen flex items-center justify-center">
          <div className="text-center">
            <svg className="mx-auto h-24 w-24 text-red-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.732-.833-2.5 0L4.268 15.5c-.77.833.192 2.5 1.732 2.5z" />
            </svg>
            <h2 className="mt-4 text-xl font-bold text-gray-900">Order not found</h2>
            <p className="mt-2 text-gray-600">
              We couldn't find your order confirmation. Please check your email or order history.
            </p>
            <div className="mt-6 space-x-4">
              <Link href="/orders">
                <Button variant="primary">View Orders</Button>
              </Link>
              <Link href="/products">
                <Button variant="outline">Continue Shopping</Button>
              </Link>
            </div>
          </div>
        </div>
      </Layout>
    )
  }

  return (
    <Layout>
      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Success Message */}
        <div className="text-center mb-8">
          <div className="mx-auto flex items-center justify-center h-16 w-16 rounded-full bg-green-100 mb-4">
            <svg className="h-8 w-8 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
            </svg>
          </div>
          <h1 className="text-3xl font-bold text-gray-900 mb-2">Order Confirmed!</h1>
          <p className="text-lg text-gray-600 mb-4">
            Thank you for your order. We've received it and will process it shortly.
          </p>
          <div className="inline-flex items-center px-4 py-2 bg-green-50 border border-green-200 rounded-lg">
            <span className="text-sm font-medium text-green-800">
              Order #{order.orderNumber}
            </span>
          </div>
        </div>

        {/* Order Details */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-8 mb-8">
          {/* Order Information */}
          <div className="card p-6">
            <h2 className="text-xl font-semibold text-gray-900 mb-4">Order Information</h2>
            <dl className="space-y-3 text-sm">
              <div className="flex justify-between">
                <dt className="text-gray-600">Order Number:</dt>
                <dd className="font-medium text-gray-900">#{order.orderNumber}</dd>
              </div>
              <div className="flex justify-between">
                <dt className="text-gray-600">Order Date:</dt>
                <dd className="font-medium text-gray-900">
                  {new Date(order.createdAt).toLocaleDateString()}
                </dd>
              </div>
              <div className="flex justify-between">
                <dt className="text-gray-600">Status:</dt>
                <dd>
                  <span className="inline-flex items-center px-2 py-1 text-xs font-medium bg-blue-100 text-blue-800 rounded-full capitalize">
                    {order.status}
                  </span>
                </dd>
              </div>
              {order.estimatedDelivery && (
                <div className="flex justify-between">
                  <dt className="text-gray-600">Estimated Delivery:</dt>
                  <dd className="font-medium text-gray-900">
                    {new Date(order.estimatedDelivery).toLocaleDateString()}
                  </dd>
                </div>
              )}
              <div className="flex justify-between">
                <dt className="text-gray-600">Total:</dt>
                <dd className="font-semibold text-lg text-gray-900">
                  ${order.total.toFixed(2)}
                </dd>
              </div>
            </dl>
          </div>

          {/* Shipping Address */}
          <div className="card p-6">
            <h2 className="text-xl font-semibold text-gray-900 mb-4">Shipping Address</h2>
            <div className="text-sm text-gray-700 space-y-1">
              <p className="font-medium text-gray-900">
                {order.shippingAddress.firstName} {order.shippingAddress.lastName}
              </p>
              {order.shippingAddress.company && (
                <p>{order.shippingAddress.company}</p>
              )}
              <p>{order.shippingAddress.street}</p>
              <p>
                {order.shippingAddress.city}, {order.shippingAddress.state} {order.shippingAddress.zipCode}
              </p>
              <p>{order.shippingAddress.country}</p>
            </div>
          </div>
        </div>

        {/* Order Items */}
        <div className="card p-6 mb-8">
          <h2 className="text-xl font-semibold text-gray-900 mb-4">Order Items</h2>
          <div className="space-y-4">
            {order.items.map((item: any) => (
              <div key={item.id} className="flex items-center space-x-4 py-4 border-b border-gray-200 last:border-b-0">
                <div className="w-16 h-16 bg-gray-100 rounded-lg overflow-hidden flex-shrink-0">
                  {item.product.images?.[0] ? (
                    <img
                      src={item.product.images[0].url}
                      alt={item.product.name}
                      className="w-full h-full object-cover"
                    />
                  ) : (
                    <div className="w-full h-full flex items-center justify-center text-gray-400">
                      <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
                      </svg>
                    </div>
                  )}
                </div>

                <div className="flex-1">
                  <h3 className="font-medium text-gray-900">{item.product.name}</h3>
                  <p className="text-sm text-gray-500">{item.product.brand?.name}</p>
                  <div className="flex items-center space-x-4 mt-1 text-sm text-gray-600">
                    {item.size && <span>Size: {item.size.name}</span>}
                    {item.color && (
                      <div className="flex items-center space-x-1">
                        <span>Color: {item.color.name}</span>
                        <div
                          className="w-4 h-4 rounded-full border border-gray-300"
                          style={{ backgroundColor: item.color.value }}
                        />
                      </div>
                    )}
                    <span>Qty: {item.quantity}</span>
                  </div>
                </div>

                <div className="text-right">
                  <p className="font-medium text-gray-900">
                    ${item.totalPrice.toFixed(2)}
                  </p>
                </div>
              </div>
            ))}
          </div>

          {/* Order Summary */}
          <div className="border-t border-gray-200 pt-4 mt-4">
            <div className="space-y-2 text-sm">
              <div className="flex justify-between">
                <span>Subtotal:</span>
                <span>${order.subtotal.toFixed(2)}</span>
              </div>
              <div className="flex justify-between">
                <span>Shipping:</span>
                <span>${order.shippingCost.toFixed(2)}</span>
              </div>
              <div className="flex justify-between">
                <span>Tax:</span>
                <span>${order.tax.toFixed(2)}</span>
              </div>
              {order.discountAmount > 0 && (
                <div className="flex justify-between text-green-600">
                  <span>Discount:</span>
                  <span>-${order.discountAmount.toFixed(2)}</span>
                </div>
              )}
              <div className="border-t border-gray-200 pt-2">
                <div className="flex justify-between font-semibold text-base">
                  <span>Total:</span>
                  <span>${order.total.toFixed(2)}</span>
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Next Steps */}
        <div className="card p-6 mb-8">
          <h2 className="text-xl font-semibold text-gray-900 mb-4">What's Next?</h2>
          <div className="space-y-3 text-sm text-gray-700">
            <div className="flex items-start space-x-3">
              <div className="flex-shrink-0 w-6 h-6 bg-blue-100 text-blue-600 rounded-full flex items-center justify-center text-xs font-medium">
                1
              </div>
              <div>
                <p className="font-medium text-gray-900">Order Confirmation Email</p>
                <p>We've sent a confirmation email with your order details to your registered email address.</p>
              </div>
            </div>
            <div className="flex items-start space-x-3">
              <div className="flex-shrink-0 w-6 h-6 bg-blue-100 text-blue-600 rounded-full flex items-center justify-center text-xs font-medium">
                2
              </div>
              <div>
                <p className="font-medium text-gray-900">Order Processing</p>
                <p>We'll prepare your order and send you tracking information once it ships.</p>
              </div>
            </div>
            <div className="flex items-start space-x-3">
              <div className="flex-shrink-0 w-6 h-6 bg-blue-100 text-blue-600 rounded-full flex items-center justify-center text-xs font-medium">
                3
              </div>
              <div>
                <p className="font-medium text-gray-900">Track Your Order</p>
                <p>You can track your order status in your account or with the tracking number we'll provide.</p>
              </div>
            </div>
          </div>
        </div>

        {/* Action Buttons */}
        <div className="flex flex-col sm:flex-row gap-4 justify-center">
          <Link href={`/orders/${order.id}`}>
            <Button variant="primary" className="w-full sm:w-auto">
              Track Your Order
            </Button>
          </Link>
          <Link href="/products">
            <Button variant="outline" className="w-full sm:w-auto">
              Continue Shopping
            </Button>
          </Link>
          <Link href="/orders">
            <Button variant="outline" className="w-full sm:w-auto">
              View All Orders
            </Button>
          </Link>
        </div>

        {/* Support Information */}
        <div className="text-center mt-8 pt-8 border-t border-gray-200">
          <p className="text-sm text-gray-600">
            Have questions about your order?{' '}
            <Link href="/contact" className="text-blue-600 hover:text-blue-800 font-medium">
              Contact our support team
            </Link>{' '}
            or email us at{' '}
            <a href="mailto:support@solemate.com" className="text-blue-600 hover:text-blue-800 font-medium">
              support@solemate.com
            </a>
          </p>
        </div>
      </div>
    </Layout>
  )
}