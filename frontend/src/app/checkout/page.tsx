'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { useForm } from 'react-hook-form'
import { yupResolver } from '@hookform/resolvers/yup'
import { Button, Input } from '@/components/ui'
import { useGetCartQuery } from '@/store/api/cartApi'
import { useAppSelector } from '@/hooks/redux'
import { selectIsAuthenticated } from '@/store/slices/authSlice'
import { addressSchema, type AddressFormData } from '@/lib/validations'

type CheckoutStep = 'shipping' | 'payment' | 'review'

interface ShippingFormProps {
  onNext: (data: AddressFormData) => void
  initialData?: Partial<AddressFormData>
}

function ShippingForm({ onNext, initialData }: ShippingFormProps) {
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<AddressFormData>({
    resolver: yupResolver(addressSchema),
    defaultValues: {
      type: 'shipping',
      ...initialData,
    },
  })

  return (
    <div className="card p-6">
      <h2 className="text-xl font-semibold text-gray-900 mb-6">Shipping Address</h2>

      <form onSubmit={handleSubmit(onNext)} className="space-y-4">
        <div className="grid grid-cols-2 gap-4">
          <Input
            label="First Name"
            error={errors.firstName?.message}
            {...register('firstName')}
          />
          <Input
            label="Last Name"
            error={errors.lastName?.message}
            {...register('lastName')}
          />
        </div>

        <Input
          label="Company (optional)"
          error={errors.company?.message}
          {...register('company')}
        />

        <Input
          label="Street Address"
          error={errors.street?.message}
          {...register('street')}
        />

        <div className="grid grid-cols-2 gap-4">
          <Input
            label="City"
            error={errors.city?.message}
            {...register('city')}
          />
          <Input
            label="State/Province"
            error={errors.state?.message}
            {...register('state')}
          />
        </div>

        <div className="grid grid-cols-2 gap-4">
          <Input
            label="ZIP Code"
            error={errors.zipCode?.message}
            {...register('zipCode')}
          />
          <Input
            label="Country"
            error={errors.country?.message}
            {...register('country')}
          />
        </div>

        <div className="pt-4">
          <Button type="submit" variant="primary" className="w-full">
            Continue to Payment
          </Button>
        </div>
      </form>
    </div>
  )
}

interface PaymentFormProps {
  onNext: (data: any) => void
  onBack: () => void
}

function PaymentForm({ onNext, onBack }: PaymentFormProps) {
  const [paymentMethod, setPaymentMethod] = useState<'card' | 'paypal'>('card')
  const [cardData, setCardData] = useState({
    cardNumber: '',
    expiryDate: '',
    cvv: '',
    nameOnCard: '',
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    onNext({ paymentMethod, ...cardData })
  }

  return (
    <div className="card p-6">
      <h2 className="text-xl font-semibold text-gray-900 mb-6">Payment Information</h2>

      <form onSubmit={handleSubmit} className="space-y-6">
        {/* Payment Method Selection */}
        <div>
          <label className="text-sm font-medium text-gray-900 mb-3 block">
            Payment Method
          </label>
          <div className="grid grid-cols-2 gap-4">
            <label className={`border-2 rounded-lg p-4 cursor-pointer transition-colors ${
              paymentMethod === 'card' ? 'border-primary-500 bg-primary-50' : 'border-gray-200 hover:border-gray-300'
            }`}>
              <input
                type="radio"
                name="paymentMethod"
                value="card"
                checked={paymentMethod === 'card'}
                onChange={(e) => setPaymentMethod(e.target.value as 'card')}
                className="sr-only"
              />
              <div className="flex items-center space-x-3">
                <div className="flex-shrink-0">
                  <svg className="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 10h18M7 15h1m4 0h1m-7 4h12a3 3 0 003-3V8a3 3 0 00-3-3H6a3 3 0 00-3 3v8a3 3 0 003 3z" />
                  </svg>
                </div>
                <div>
                  <p className="font-medium">Credit Card</p>
                  <p className="text-sm text-gray-500">Visa, Mastercard, Amex</p>
                </div>
              </div>
            </label>

            <label className={`border-2 rounded-lg p-4 cursor-pointer transition-colors ${
              paymentMethod === 'paypal' ? 'border-primary-500 bg-primary-50' : 'border-gray-200 hover:border-gray-300'
            }`}>
              <input
                type="radio"
                name="paymentMethod"
                value="paypal"
                checked={paymentMethod === 'paypal'}
                onChange={(e) => setPaymentMethod(e.target.value as 'paypal')}
                className="sr-only"
              />
              <div className="flex items-center space-x-3">
                <div className="flex-shrink-0">
                  <svg className="w-8 h-8 text-blue-600" viewBox="0 0 24 24" fill="currentColor">
                    <path d="M7.076 21.337H2.47a.641.641 0 0 1-.633-.74L4.944.901C5.026.382 5.474 0 5.998 0h7.46c2.57 0 4.578.543 5.69 1.81 1.01 1.15 1.304 2.42 1.012 4.287-.023.143-.047.288-.077.437-.983 5.05-4.349 6.797-8.647 6.797h-2.19c-.524 0-.968.382-1.05.9l-1.12 7.106zm14.146-14.42a3.35 3.35 0 0 0-.4-.41c-.4-.4-.91-.77-1.52-1.09-.61-.32-1.34-.57-2.17-.74-.83-.17-1.76-.26-2.77-.26H9.29c-.13 0-.26.05-.35.15-.09.1-.13.23-.1.36l1.02 6.49h2.19c3.85 0 6.7-1.56 7.57-6.06.02-.11.04-.22.05-.33.01-.11.02-.22.02-.33 0-.44-.05-.85-.15-1.24z"/>
                  </svg>
                </div>
                <div>
                  <p className="font-medium">PayPal</p>
                  <p className="text-sm text-gray-500">Pay with your PayPal account</p>
                </div>
              </div>
            </label>
          </div>
        </div>

        {/* Card Details */}
        {paymentMethod === 'card' && (
          <div className="space-y-4">
            <Input
              label="Name on Card"
              value={cardData.nameOnCard}
              onChange={(e) => setCardData({ ...cardData, nameOnCard: e.target.value })}
              required
            />

            <Input
              label="Card Number"
              placeholder="1234 5678 9012 3456"
              value={cardData.cardNumber}
              onChange={(e) => {
                const value = e.target.value.replace(/\s/g, '').replace(/(.{4})/g, '$1 ').trim()
                setCardData({ ...cardData, cardNumber: value })
              }}
              maxLength={19}
              required
            />

            <div className="grid grid-cols-2 gap-4">
              <Input
                label="Expiry Date"
                placeholder="MM/YY"
                value={cardData.expiryDate}
                onChange={(e) => {
                  let value = e.target.value.replace(/\D/g, '')
                  if (value.length >= 2) {
                    value = value.substring(0, 2) + '/' + value.substring(2, 4)
                  }
                  setCardData({ ...cardData, expiryDate: value })
                }}
                maxLength={5}
                required
              />

              <Input
                label="CVV"
                placeholder="123"
                value={cardData.cvv}
                onChange={(e) => setCardData({ ...cardData, cvv: e.target.value.replace(/\D/g, '') })}
                maxLength={4}
                required
              />
            </div>
          </div>
        )}

        <div className="flex space-x-4 pt-4">
          <Button type="button" variant="outline" className="flex-1" onClick={onBack}>
            Back to Shipping
          </Button>
          <Button type="submit" variant="primary" className="flex-1">
            Review Order
          </Button>
        </div>
      </form>
    </div>
  )
}

interface OrderReviewProps {
  cart: any
  shippingData: AddressFormData
  paymentData: any
  onBack: () => void
  onPlaceOrder: () => void
  isLoading: boolean
}

function OrderReview({ cart, shippingData, paymentData, onBack, onPlaceOrder, isLoading }: OrderReviewProps) {
  const subtotal = cart.total || 0
  const shipping = subtotal > 50 ? 0 : 9.99
  const tax = subtotal * 0.08
  const total = subtotal + shipping + tax

  return (
    <div className="space-y-6">
      {/* Order Items */}
      <div className="card p-6">
        <h2 className="text-xl font-semibold text-gray-900 mb-4">Order Items</h2>
        <div className="space-y-4">
          {cart.items?.map((item: any) => (
            <div key={item.id} className="flex items-center space-x-4">
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
                <p className="text-sm text-gray-500">
                  {item.size && `Size: ${item.size.name}`}
                  {item.size && item.color && ' • '}
                  {item.color && `Color: ${item.color.name}`}
                </p>
                <p className="text-sm text-gray-500">Qty: {item.quantity}</p>
              </div>
              <div className="text-right">
                <p className="font-medium text-gray-900">
                  ${(item.product.price * item.quantity).toFixed(2)}
                </p>
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Shipping Address */}
      <div className="card p-6">
        <h2 className="text-xl font-semibold text-gray-900 mb-4">Shipping Address</h2>
        <div className="text-gray-700">
          <p className="font-medium">{shippingData.firstName} {shippingData.lastName}</p>
          {shippingData.company && <p>{shippingData.company}</p>}
          <p>{shippingData.street}</p>
          <p>{shippingData.city}, {shippingData.state} {shippingData.zipCode}</p>
          <p>{shippingData.country}</p>
        </div>
      </div>

      {/* Payment Method */}
      <div className="card p-6">
        <h2 className="text-xl font-semibold text-gray-900 mb-4">Payment Method</h2>
        <div className="text-gray-700">
          {paymentData.paymentMethod === 'card' ? (
            <div>
              <p className="font-medium">Credit Card</p>
              <p>**** **** **** {paymentData.cardNumber?.slice(-4)}</p>
              <p>{paymentData.nameOnCard}</p>
            </div>
          ) : (
            <div>
              <p className="font-medium">PayPal</p>
              <p>Pay with your PayPal account</p>
            </div>
          )}
        </div>
      </div>

      {/* Order Summary */}
      <div className="card p-6">
        <h2 className="text-xl font-semibold text-gray-900 mb-4">Order Summary</h2>
        <div className="space-y-2 text-sm">
          <div className="flex justify-between">
            <span>Subtotal ({cart.itemCount} items)</span>
            <span>${subtotal.toFixed(2)}</span>
          </div>
          <div className="flex justify-between">
            <span>Shipping</span>
            <span>{shipping === 0 ? 'Free' : `$${shipping.toFixed(2)}`}</span>
          </div>
          <div className="flex justify-between">
            <span>Tax</span>
            <span>${tax.toFixed(2)}</span>
          </div>
          <div className="border-t border-gray-200 pt-2">
            <div className="flex justify-between font-semibold text-base">
              <span>Total</span>
              <span>${total.toFixed(2)}</span>
            </div>
          </div>
        </div>
      </div>

      {/* Actions */}
      <div className="flex space-x-4">
        <Button variant="outline" className="flex-1" onClick={onBack}>
          Back to Payment
        </Button>
        <Button
          variant="primary"
          className="flex-1"
          onClick={onPlaceOrder}
          disabled={isLoading}
        >
          {isLoading ? 'Placing Order...' : 'Place Order'}
        </Button>
      </div>
    </div>
  )
}

export default function CheckoutPage() {
  const router = useRouter()
  const isAuthenticated = useAppSelector(selectIsAuthenticated)
  const { data: cart, isLoading: cartLoading } = useGetCartQuery()

  const [currentStep, setCurrentStep] = useState<CheckoutStep>('shipping')
  const [shippingData, setShippingData] = useState<AddressFormData | null>(null)
  const [paymentData, setPaymentData] = useState<any>(null)
  const [isPlacingOrder, setIsPlacingOrder] = useState(false)

  // Redirect to login if not authenticated
  useEffect(() => {
    if (!cartLoading && !isAuthenticated) {
      router.push('/login?redirect=/checkout')
    }
  }, [isAuthenticated, cartLoading, router])

  // Redirect to cart if empty
  useEffect(() => {
    if (cart && cart.itemCount === 0) {
      router.push('/cart')
    }
  }, [cart, router])

  const handleShippingNext = (data: AddressFormData) => {
    setShippingData(data)
    setCurrentStep('payment')
  }

  const handlePaymentNext = (data: any) => {
    setPaymentData(data)
    setCurrentStep('review')
  }

  const handlePlaceOrder = async () => {
    setIsPlacingOrder(true)
    try {
      // TODO: Implement actual order creation with real API
      // For now, simulate successful order creation
      await new Promise(resolve => setTimeout(resolve, 2000))

      // In a real implementation, you would:
      // 1. Create order with the API
      // 2. Get the order ID from the response
      // 3. Store it for the confirmation page

      // Simulate order ID for demo
      const simulatedOrderId = 'order_' + Date.now()
      localStorage.setItem('lastOrderId', simulatedOrderId)

      // Redirect to order confirmation
      router.push('/orders/confirmation')
    } catch (error) {
      alert('Failed to place order. Please try again.')
    } finally {
      setIsPlacingOrder(false)
    }
  }

  if (cartLoading || !isAuthenticated) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600"></div>
      </div>
    )
  }

  if (!cart || cart.itemCount === 0) {
    return null // Will redirect
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <div className="bg-white shadow-sm">
        <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
          <h1 className="text-2xl font-bold text-gray-900">Checkout</h1>

          {/* Progress Steps */}
          <div className="mt-6">
            <nav aria-label="Progress">
              <ol className="flex items-center">
                {[
                  { name: 'Shipping', step: 'shipping' },
                  { name: 'Payment', step: 'payment' },
                  { name: 'Review', step: 'review' },
                ].map((item, index) => (
                  <li key={item.name} className={`${index !== 0 ? 'ml-8' : ''} relative`}>
                    {index !== 0 && (
                      <div className="absolute inset-0 flex items-center" aria-hidden="true">
                        <div className="h-0.5 w-full bg-gray-200" />
                      </div>
                    )}
                    <div className="relative flex items-center justify-center">
                      <div
                        className={`h-8 w-8 rounded-full flex items-center justify-center text-sm font-medium ${
                          currentStep === item.step
                            ? 'bg-primary-600 text-white'
                            : currentStep === 'payment' && item.step === 'shipping' ||
                              currentStep === 'review' && (item.step === 'shipping' || item.step === 'payment')
                            ? 'bg-primary-600 text-white'
                            : 'bg-gray-300 text-gray-500'
                        }`}
                      >
                        {currentStep === 'payment' && item.step === 'shipping' ||
                         currentStep === 'review' && (item.step === 'shipping' || item.step === 'payment') ? (
                          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                          </svg>
                        ) : (
                          index + 1
                        )}
                      </div>
                      <span className={`ml-3 text-sm font-medium ${
                        currentStep === item.step ? 'text-primary-600' : 'text-gray-500'
                      }`}>
                        {item.name}
                      </span>
                    </div>
                  </li>
                ))}
              </ol>
            </nav>
          </div>
        </div>
      </div>

      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {currentStep === 'shipping' && (
          <ShippingForm onNext={handleShippingNext} initialData={shippingData || undefined} />
        )}

        {currentStep === 'payment' && (
          <PaymentForm
            onNext={handlePaymentNext}
            onBack={() => setCurrentStep('shipping')}
          />
        )}

        {currentStep === 'review' && shippingData && paymentData && (
          <OrderReview
            cart={cart}
            shippingData={shippingData}
            paymentData={paymentData}
            onBack={() => setCurrentStep('payment')}
            onPlaceOrder={handlePlaceOrder}
            isLoading={isPlacingOrder}
          />
        )}
      </div>
    </div>
  )
}