'use client'

import { useEffect } from 'react'
import { useParams, useRouter } from 'next/navigation'
import Link from 'next/link'
import { Button } from '@/components/ui'
import Layout from '@/components/layout/Layout'
import { useGetOrderQuery, useTrackOrderQuery, useCancelOrderMutation, useReorderItemsMutation, useLazyDownloadInvoiceQuery } from '@/store/api/orderApi'
import { useAppSelector } from '@/hooks/redux'
import { selectIsAuthenticated } from '@/store/slices/authSlice'
import type { OrderStatus } from '@/types'

function OrderTimeline({ order }: { order: any }) {
  const getStatusSteps = (status: OrderStatus) => {
    const allSteps = [
      { key: 'pending', label: 'Order Placed', description: 'We have received your order' },
      { key: 'confirmed', label: 'Order Confirmed', description: 'Your order has been confirmed' },
      { key: 'processing', label: 'Processing', description: 'Your order is being prepared' },
      { key: 'shipped', label: 'Shipped', description: 'Your order is on its way' },
      { key: 'delivered', label: 'Delivered', description: 'Your order has been delivered' },
    ]

    const statusIndex = allSteps.findIndex(step => step.key === status)

    if (status === 'cancelled') {
      return [
        allSteps[0],
        { key: 'cancelled', label: 'Cancelled', description: 'Your order has been cancelled' },
      ]
    }

    if (status === 'refunded') {
      return [
        ...allSteps.slice(0, statusIndex + 1),
        { key: 'refunded', label: 'Refunded', description: 'Your order has been refunded' },
      ]
    }

    return allSteps.slice(0, statusIndex + 1)
  }

  const steps = getStatusSteps(order.status)

  return (
    <div className="card p-6">
      <h2 className="text-xl font-semibold text-gray-900 mb-6">Order Status</h2>

      <div className="space-y-6">
        {steps.map((step, index) => {
          const isCompleted = index < steps.length - 1 || order.status === step.key
          const isCurrent = order.status === step.key
          const isCancelled = step.key === 'cancelled'
          const isRefunded = step.key === 'refunded'

          return (
            <div key={step.key} className="flex items-start">
              <div className="flex-shrink-0">
                <div
                  className={`w-8 h-8 rounded-full flex items-center justify-center ${
                    isCancelled || isRefunded
                      ? 'bg-red-100 text-red-600'
                      : isCompleted
                      ? 'bg-green-100 text-green-600'
                      : isCurrent
                      ? 'bg-blue-100 text-blue-600'
                      : 'bg-gray-100 text-gray-400'
                  }`}
                >
                  {isCompleted && !isCancelled && !isRefunded ? (
                    <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                    </svg>
                  ) : isCancelled || isRefunded ? (
                    <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                    </svg>
                  ) : (
                    <span className="text-sm font-medium">{index + 1}</span>
                  )}
                </div>
              </div>

              <div className="ml-4 flex-1">
                <h3
                  className={`text-sm font-medium ${
                    isCancelled || isRefunded
                      ? 'text-red-900'
                      : isCompleted
                      ? 'text-green-900'
                      : isCurrent
                      ? 'text-blue-900'
                      : 'text-gray-500'
                  }`}
                >
                  {step.label}
                </h3>
                <p
                  className={`text-sm ${
                    isCancelled || isRefunded
                      ? 'text-red-600'
                      : isCompleted
                      ? 'text-green-600'
                      : isCurrent
                      ? 'text-blue-600'
                      : 'text-gray-400'
                  }`}
                >
                  {step.description}
                </p>

                {isCurrent && order.estimatedDelivery && (
                  <p className="text-sm text-gray-500 mt-1">
                    Estimated delivery: {new Date(order.estimatedDelivery).toLocaleDateString()}
                  </p>
                )}
              </div>

              {isCurrent && order.trackingNumber && (
                <div className="ml-4">
                  <span className="text-sm text-gray-500">
                    Tracking: {order.trackingNumber}
                  </span>
                </div>
              )}
            </div>
          )
        })}
      </div>
    </div>
  )
}

function OrderItems({ order }: { order: any }) {
  return (
    <div className="card p-6">
      <h2 className="text-xl font-semibold text-gray-900 mb-6">Order Items</h2>

      <div className="space-y-4">
        {order.items.map((item: any) => (
          <div key={item.id} className="flex items-start space-x-4 py-4 border-b border-gray-200 last:border-b-0">
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
              <div className="flex items-start justify-between">
                <div>
                  <Link href={`/products/${item.product.id}`}>
                    <h3 className="font-medium text-gray-900 hover:text-primary-600">
                      {item.product.name}
                    </h3>
                  </Link>
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
                  <p className="text-sm text-gray-500">
                    ${item.unitPrice.toFixed(2)} each
                  </p>
                </div>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}

function OrderSummary({ order }: { order: any }) {
  return (
    <div className="space-y-6">
      {/* Order Summary */}
      <div className="card p-6">
        <h2 className="text-xl font-semibold text-gray-900 mb-4">Order Summary</h2>

        <div className="space-y-2 text-sm">
          <div className="flex justify-between">
            <span>Subtotal</span>
            <span>${order.subtotal.toFixed(2)}</span>
          </div>
          <div className="flex justify-between">
            <span>Shipping</span>
            <span>${order.shippingCost.toFixed(2)}</span>
          </div>
          <div className="flex justify-between">
            <span>Tax</span>
            <span>${order.tax.toFixed(2)}</span>
          </div>
          {order.discountAmount > 0 && (
            <div className="flex justify-between text-green-600">
              <span>Discount</span>
              <span>-${order.discountAmount.toFixed(2)}</span>
            </div>
          )}
          <div className="border-t border-gray-200 pt-2">
            <div className="flex justify-between font-semibold">
              <span>Total</span>
              <span>${order.total.toFixed(2)}</span>
            </div>
          </div>
        </div>
      </div>

      {/* Shipping Address */}
      <div className="card p-6">
        <h3 className="font-semibold text-gray-900 mb-3">Shipping Address</h3>
        <div className="text-sm text-gray-700 space-y-1">
          <p className="font-medium">
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

      {/* Payment Method */}
      <div className="card p-6">
        <h3 className="font-semibold text-gray-900 mb-3">Payment Method</h3>
        <div className="text-sm text-gray-700">
          <p className="font-medium">{order.paymentMethod.type}</p>
          {order.paymentMethod.last4 && (
            <p>**** **** **** {order.paymentMethod.last4}</p>
          )}
        </div>
      </div>
    </div>
  )
}

export default function OrderDetailPage() {
  const params = useParams()
  const router = useRouter()
  const isAuthenticated = useAppSelector(selectIsAuthenticated)
  const orderId = params.id as string

  const { data: order, isLoading, error } = useGetOrderQuery(orderId)
  const { data: trackingData } = useTrackOrderQuery(orderId, { skip: !order })
  const [cancelOrder, { isLoading: isCancelling }] = useCancelOrderMutation()
  const [reorderItems, { isLoading: isReordering }] = useReorderItemsMutation()
  const [downloadInvoice] = useLazyDownloadInvoiceQuery()

  useEffect(() => {
    if (!isAuthenticated) {
      router.push('/login?redirect=/orders')
    }
  }, [isAuthenticated, router])

  const handleCancelOrder = async () => {
    if (window.confirm('Are you sure you want to cancel this order?')) {
      const reason = prompt('Reason for cancellation (optional):')
      try {
        await cancelOrder({ id: orderId, reason: reason || undefined }).unwrap()
        alert('Order cancelled successfully')
      } catch (error: any) {
        alert(error?.data?.message || 'Failed to cancel order')
      }
    }
  }

  const handleReorder = async () => {
    try {
      const result = await reorderItems({ orderId }).unwrap()
      const goToCart = window.confirm(
        `${result.data.addedToCart} items added to cart! Would you like to view your cart?`
      )
      if (goToCart) {
        router.push('/cart')
      }
    } catch (error: any) {
      alert(error?.data?.message || 'Failed to reorder items')
    }
  }

  const handleDownloadInvoice = async () => {
    try {
      const result = await downloadInvoice(orderId).unwrap()
      const url = window.URL.createObjectURL(result)
      const a = document.createElement('a')
      a.href = url
      a.download = `invoice-${order?.orderNumber}.pdf`
      document.body.appendChild(a)
      a.click()
      window.URL.revokeObjectURL(url)
      document.body.removeChild(a)
    } catch (error) {
      alert('Failed to download invoice')
    }
  }

  if (!isAuthenticated) {
    return (
      <Layout>
        <div className="min-h-screen flex items-center justify-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600"></div>
        </div>
      </Layout>
    )
  }

  if (isLoading) {
    return (
      <Layout>
        <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className="animate-pulse space-y-6">
            <div className="h-8 bg-gray-200 rounded w-1/3"></div>
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
              <div className="lg:col-span-2 space-y-6">
                <div className="h-64 bg-gray-200 rounded"></div>
                <div className="h-96 bg-gray-200 rounded"></div>
              </div>
              <div className="space-y-6">
                <div className="h-48 bg-gray-200 rounded"></div>
                <div className="h-32 bg-gray-200 rounded"></div>
              </div>
            </div>
          </div>
        </div>
      </Layout>
    )
  }

  if (error || !order) {
    return (
      <Layout>
        <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className="text-center">
            <h3 className="text-lg font-medium text-gray-900">Order not found</h3>
            <p className="text-gray-600">The order you're looking for doesn't exist or you don't have access to it.</p>
            <div className="mt-6 space-x-4">
              <Link href="/orders">
                <Button variant="primary">View All Orders</Button>
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

  const canCancel = ['pending', 'confirmed'].includes(order.status)
  const canReorder = ['delivered', 'cancelled'].includes(order.status)

  return (
    <Layout>
      <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Header */}
        <div className="flex items-center justify-between mb-8">
          <div>
            <div className="flex items-center space-x-2 mb-2">
              <Link href="/orders" className="text-gray-500 hover:text-gray-700">
                Orders
              </Link>
              <span className="text-gray-500">/</span>
              <span className="text-gray-900">#{order.orderNumber}</span>
            </div>
            <h1 className="text-2xl font-bold text-gray-900">
              Order #{order.orderNumber}
            </h1>
            <p className="text-gray-600">
              Placed on {new Date(order.createdAt).toLocaleDateString()}
            </p>
          </div>

          <div className="flex items-center space-x-3">
            <Button
              variant="outline"
              onClick={handleDownloadInvoice}
              className="text-sm"
            >
              Download Invoice
            </Button>

            {canReorder && (
              <Button
                variant="outline"
                onClick={handleReorder}
                disabled={isReordering}
                className="text-sm"
              >
                {isReordering ? 'Reordering...' : 'Reorder'}
              </Button>
            )}

            {canCancel && (
              <Button
                variant="outline"
                onClick={handleCancelOrder}
                disabled={isCancelling}
                className="text-sm text-red-600 hover:text-red-800"
              >
                {isCancelling ? 'Cancelling...' : 'Cancel Order'}
              </Button>
            )}
          </div>
        </div>

        {/* Content */}
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          <div className="lg:col-span-2 space-y-6">
            <OrderTimeline order={order} />
            <OrderItems order={order} />
          </div>

          <div className="lg:col-span-1">
            <OrderSummary order={order} />
          </div>
        </div>
      </div>
    </Layout>
  )
}