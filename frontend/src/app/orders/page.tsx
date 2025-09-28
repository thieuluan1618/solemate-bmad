'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import Link from 'next/link'
import { Button } from '@/components/ui'
import Layout from '@/components/layout/Layout'
import { useGetOrdersQuery, useCancelOrderMutation, useReorderItemsMutation } from '@/store/api/orderApi'
import { useAppSelector } from '@/hooks/redux'
import { selectIsAuthenticated } from '@/store/slices/authSlice'
import type { Order, OrderStatus } from '@/types'

interface OrderCardProps {
  order: Order
  onCancelOrder: (orderId: string, reason?: string) => void
  onReorder: (orderId: string) => void
  isLoading: boolean
}

function OrderCard({ order, onCancelOrder, onReorder, isLoading }: OrderCardProps) {
  const router = useRouter()

  const getStatusColor = (status: OrderStatus) => {
    switch (status) {
      case 'pending':
        return 'bg-yellow-100 text-yellow-800'
      case 'confirmed':
        return 'bg-blue-100 text-blue-800'
      case 'processing':
        return 'bg-purple-100 text-purple-800'
      case 'shipped':
        return 'bg-indigo-100 text-indigo-800'
      case 'delivered':
        return 'bg-green-100 text-green-800'
      case 'cancelled':
        return 'bg-red-100 text-red-800'
      case 'refunded':
        return 'bg-gray-100 text-gray-800'
      default:
        return 'bg-gray-100 text-gray-800'
    }
  }

  const canCancel = ['pending', 'confirmed'].includes(order.status)
  const canReorder = ['delivered', 'cancelled'].includes(order.status)

  return (
    <div className="card p-6">
      <div className="flex items-start justify-between mb-4">
        <div>
          <div className="flex items-center space-x-3 mb-2">
            <h3 className="text-lg font-semibold text-gray-900">
              Order #{order.orderNumber}
            </h3>
            <span className={`px-2 py-1 text-xs font-medium rounded-full capitalize ${getStatusColor(order.status)}`}>
              {order.status.replace('_', ' ')}
            </span>
          </div>
          <div className="text-sm text-gray-600 space-y-1">
            <p>Placed on {new Date(order.createdAt).toLocaleDateString()}</p>
            <p>Total: ${order.total.toFixed(2)}</p>
            {order.trackingNumber && (
              <p>Tracking: {order.trackingNumber}</p>
            )}
            {order.estimatedDelivery && (
              <p>Estimated delivery: {new Date(order.estimatedDelivery).toLocaleDateString()}</p>
            )}
          </div>
        </div>

        <div className="flex items-center space-x-2">
          <Button
            variant="outline"
            onClick={() => router.push(`/orders/${order.id}`)}
            className="text-sm"
          >
            View Details
          </Button>

          {canReorder && (
            <Button
              variant="outline"
              onClick={() => onReorder(order.id)}
              disabled={isLoading}
              className="text-sm"
            >
              Reorder
            </Button>
          )}

          {canCancel && (
            <Button
              variant="outline"
              onClick={() => {
                const reason = prompt('Reason for cancellation (optional):')
                onCancelOrder(order.id, reason || undefined)
              }}
              disabled={isLoading}
              className="text-sm text-red-600 hover:text-red-800"
            >
              Cancel
            </Button>
          )}
        </div>
      </div>

      {/* Order Items Preview */}
      <div className="border-t border-gray-200 pt-4">
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          {order.items.slice(0, 3).map((item) => (
            <div key={item.id} className="flex items-center space-x-3">
              <div className="w-12 h-12 bg-gray-100 rounded-lg overflow-hidden flex-shrink-0">
                {item.product.images?.[0] ? (
                  <img
                    src={item.product.images[0].url}
                    alt={item.product.name}
                    className="w-full h-full object-cover"
                  />
                ) : (
                  <div className="w-full h-full flex items-center justify-center text-gray-400">
                    <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
                    </svg>
                  </div>
                )}
              </div>
              <div className="flex-1 min-w-0">
                <p className="text-sm font-medium text-gray-900 truncate">
                  {item.product.name}
                </p>
                <p className="text-sm text-gray-500">
                  Qty: {item.quantity} • ${item.unitPrice.toFixed(2)}
                </p>
              </div>
            </div>
          ))}
          {order.items.length > 3 && (
            <div className="flex items-center justify-center text-sm text-gray-500">
              +{order.items.length - 3} more items
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

export default function OrdersPage() {
  const router = useRouter()
  const isAuthenticated = useAppSelector(selectIsAuthenticated)
  const [statusFilter, setStatusFilter] = useState<OrderStatus | 'all'>('all')
  const [currentPage, setCurrentPage] = useState(1)

  const { data: ordersData, isLoading, error } = useGetOrdersQuery({
    page: currentPage,
    limit: 10,
    ...(statusFilter !== 'all' && { status: statusFilter }),
  })

  const [cancelOrder, { isLoading: isCancelling }] = useCancelOrderMutation()
  const [reorderItems, { isLoading: isReordering }] = useReorderItemsMutation()

  useEffect(() => {
    if (!isAuthenticated) {
      router.push('/login?redirect=/orders')
    }
  }, [isAuthenticated, router])

  const handleCancelOrder = async (orderId: string, reason?: string) => {
    if (window.confirm('Are you sure you want to cancel this order?')) {
      try {
        await cancelOrder({ id: orderId, reason }).unwrap()
        alert('Order cancelled successfully')
      } catch (error: any) {
        alert(error?.data?.message || 'Failed to cancel order')
      }
    }
  }

  const handleReorder = async (orderId: string) => {
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
        <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className="text-center">
            <h3 className="text-lg font-medium text-gray-900">Error loading orders</h3>
            <p className="text-gray-600">Please try again later.</p>
            <Button onClick={() => window.location.reload()} className="mt-4">
              Retry
            </Button>
          </div>
        </div>
      </Layout>
    )
  }

  const statusOptions = [
    { value: 'all', label: 'All Orders' },
    { value: 'pending', label: 'Pending' },
    { value: 'confirmed', label: 'Confirmed' },
    { value: 'processing', label: 'Processing' },
    { value: 'shipped', label: 'Shipped' },
    { value: 'delivered', label: 'Delivered' },
    { value: 'cancelled', label: 'Cancelled' },
  ]

  return (
    <Layout>
      <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="flex items-center justify-between mb-8">
          <div>
            <h1 className="text-2xl font-bold text-gray-900">Order History</h1>
            <p className="text-gray-600">Track and manage your orders</p>
          </div>

          <Link href="/products">
            <Button variant="primary">Continue Shopping</Button>
          </Link>
        </div>

        {/* Filter Bar */}
        <div className="flex items-center space-x-4 mb-6">
          <label className="text-sm font-medium text-gray-700">Filter by status:</label>
          <select
            value={statusFilter}
            onChange={(e) => {
              setStatusFilter(e.target.value as OrderStatus | 'all')
              setCurrentPage(1)
            }}
            className="border border-gray-300 rounded-md px-3 py-2 text-sm focus:ring-blue-500 focus:border-blue-500"
          >
            {statusOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>

          {ordersData && (
            <span className="text-sm text-gray-500">
              {ordersData.meta.total} total orders
            </span>
          )}
        </div>

        {/* Orders List */}
        {isLoading ? (
          <div className="space-y-4">
            {[...Array(3)].map((_, i) => (
              <div key={i} className="card p-6 animate-pulse">
                <div className="flex items-start justify-between mb-4">
                  <div className="space-y-2">
                    <div className="h-4 bg-gray-200 rounded w-48"></div>
                    <div className="h-3 bg-gray-200 rounded w-32"></div>
                    <div className="h-3 bg-gray-200 rounded w-24"></div>
                  </div>
                  <div className="h-8 bg-gray-200 rounded w-24"></div>
                </div>
                <div className="border-t border-gray-200 pt-4">
                  <div className="grid grid-cols-3 gap-4">
                    {[...Array(3)].map((_, j) => (
                      <div key={j} className="flex items-center space-x-3">
                        <div className="w-12 h-12 bg-gray-200 rounded-lg"></div>
                        <div className="space-y-1">
                          <div className="h-3 bg-gray-200 rounded w-24"></div>
                          <div className="h-3 bg-gray-200 rounded w-16"></div>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              </div>
            ))}
          </div>
        ) : ordersData && ordersData.data.length > 0 ? (
          <>
            <div className="space-y-6">
              {ordersData.data.map((order) => (
                <OrderCard
                  key={order.id}
                  order={order}
                  onCancelOrder={handleCancelOrder}
                  onReorder={handleReorder}
                  isLoading={isCancelling || isReordering}
                />
              ))}
            </div>

            {/* Pagination */}
            {ordersData.meta.totalPages > 1 && (
              <div className="mt-8 flex justify-center">
                <div className="flex items-center space-x-2">
                  <Button
                    variant="outline"
                    disabled={currentPage === 1}
                    onClick={() => setCurrentPage(currentPage - 1)}
                  >
                    Previous
                  </Button>

                  {[...Array(Math.min(5, ordersData.meta.totalPages))].map((_, i) => {
                    const page = i + 1
                    return (
                      <Button
                        key={page}
                        variant={page === currentPage ? 'primary' : 'outline'}
                        onClick={() => setCurrentPage(page)}
                      >
                        {page}
                      </Button>
                    )
                  })}

                  <Button
                    variant="outline"
                    disabled={currentPage === ordersData.meta.totalPages}
                    onClick={() => setCurrentPage(currentPage + 1)}
                  >
                    Next
                  </Button>
                </div>
              </div>
            )}
          </>
        ) : (
          <div className="text-center py-16">
            <svg className="mx-auto h-24 w-24 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 11V7a4 4 0 00-8 0v4M5 9h14l-1 12H6L5 9z" />
            </svg>
            <h2 className="mt-4 text-2xl font-bold text-gray-900">No orders found</h2>
            <p className="mt-2 text-gray-600">
              {statusFilter !== 'all'
                ? `No orders with status "${statusFilter}" found.`
                : "You haven't placed any orders yet. Start shopping to see your orders here!"
              }
            </p>
            <div className="mt-8 space-x-4">
              <Link href="/products">
                <Button variant="primary">
                  Start Shopping
                </Button>
              </Link>
              {statusFilter !== 'all' && (
                <Button
                  variant="outline"
                  onClick={() => setStatusFilter('all')}
                >
                  View All Orders
                </Button>
              )}
            </div>
          </div>
        )}
      </div>
    </Layout>
  )
}