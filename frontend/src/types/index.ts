// Common types
export interface ApiResponse<T> {
  data: T
  message: string
  success: boolean
}

export interface PaginationMeta {
  page: number
  limit: number
  total: number
  totalPages: number
}

export interface PaginatedResponse<T> {
  data: T[]
  meta: PaginationMeta
  message: string
  success: boolean
}

// User & Authentication types
export interface User {
  id: string
  email: string
  firstName: string
  lastName: string
  phone?: string
  role: 'customer' | 'admin'
  isEmailVerified: boolean
  addresses?: Address[]
  preferences?: UserPreferences
  createdAt: string
  updatedAt: string
}

export interface Address {
  id: string
  userId: string
  type: 'shipping' | 'billing'
  firstName: string
  lastName: string
  company?: string
  street: string
  city: string
  state: string
  zipCode: string
  country: string
  isDefault: boolean
  createdAt: string
  updatedAt: string
}

export interface UserPreferences {
  emailNotifications: boolean
  smsNotifications: boolean
  marketingEmails: boolean
}

export interface LoginRequest {
  email: string
  password: string
  rememberMe?: boolean
}

export interface RegisterRequest {
  email: string
  password: string
  firstName: string
  lastName: string
  phone?: string
}

export interface AuthResponse {
  user: User
  accessToken: string
  refreshToken: string
}

// Product types
export interface Product {
  id: string
  name: string
  description: string
  shortDescription?: string
  price: number
  originalPrice?: number
  discountPercentage?: number
  sku: string
  category?: Category
  brand?: Brand
  sizes: ProductSize[]
  colors: ProductColor[]
  images: ProductImage[]
  specifications: ProductSpecification[]
  stockQuantity: number
  isActive: boolean
  isFeatured: boolean
  rating: number
  reviewCount: number
  tags: string[]
  seoTitle?: string
  seoDescription?: string
  createdAt: string
  updatedAt: string
}

export interface Category {
  id: string
  name: string
  slug: string
  description?: string
  image?: string
  parentId?: string
  isActive: boolean
  createdAt: string
  updatedAt: string
}

export interface Brand {
  id: string
  name: string
  slug: string
  description?: string
  logo?: string
  website?: string
  isActive: boolean
  createdAt: string
  updatedAt: string
}

export interface ProductSize {
  id: string
  name: string // e.g., "US 8", "EU 42"
  value: string // e.g., "8", "42"
  stockQuantity: number
}

export interface ProductColor {
  id: string
  name: string
  value: string // hex color code
  stockQuantity: number
}

export interface ProductImage {
  id: string
  url: string
  altText: string
  isPrimary: boolean
  sortOrder: number
}

export interface ProductSpecification {
  id: string
  name: string
  value: string
}

export interface ProductFilters {
  categories?: string[]
  brands?: string[]
  priceMin?: number
  priceMax?: number
  sizes?: string[]
  colors?: string[]
  rating?: number
  inStock?: boolean
  featured?: boolean
  search?: string
}

export interface ProductSearchParams {
  query?: string
  category?: string
  brand?: string
  minPrice?: number
  maxPrice?: number
  sizes?: string[]
  colors?: string[]
  rating?: number
  inStock?: boolean
  sortBy?: 'price_asc' | 'price_desc' | 'name_asc' | 'name_desc' | 'rating' | 'newest'
  page?: number
  limit?: number
}

// Cart types
export interface CartItem {
  id: string
  productId: string
  product: Product
  quantity: number
  size?: ProductSize
  color?: ProductColor
  addedAt: string
}

export interface Cart {
  id: string
  userId?: string
  items: CartItem[]
  total: number
  itemCount: number
  updatedAt: string
}

export interface AddToCartRequest {
  productId: string
  quantity: number
  sizeId?: string
  colorId?: string
}

// Order types
export interface Order {
  id: string
  orderNumber: string
  userId: string
  user?: User
  status: OrderStatus
  items: OrderItem[]
  shippingAddress: Address
  billingAddress: Address
  paymentMethod: PaymentMethod
  subtotal: number
  shippingCost: number
  tax: number
  discountAmount: number
  total: number
  notes?: string
  trackingNumber?: string
  estimatedDelivery?: string
  actualDelivery?: string
  createdAt: string
  updatedAt: string
}

export interface OrderItem {
  id: string
  orderId: string
  productId: string
  product: Product
  quantity: number
  size?: ProductSize
  color?: ProductColor
  unitPrice: number
  totalPrice: number
}

export type OrderStatus =
  | 'pending'
  | 'confirmed'
  | 'processing'
  | 'shipped'
  | 'delivered'
  | 'cancelled'
  | 'refunded'

export interface CreateOrderRequest {
  items: Array<{
    productId: string
    quantity: number
    sizeId?: string
    colorId?: string
  }>
  shippingAddressId: string
  billingAddressId: string
  paymentMethodId: string
  notes?: string
  discountCode?: string
}

// Payment types
export interface PaymentMethod {
  id: string
  type: 'card' | 'paypal' | 'stripe' | 'apple_pay' | 'google_pay'
  last4?: string
  brand?: string
  expiryMonth?: number
  expiryYear?: number
  isDefault: boolean
  createdAt: string
}

export interface Payment {
  id: string
  orderId: string
  paymentMethodId: string
  amount: number
  currency: string
  status: PaymentStatus
  paymentIntentId?: string
  transactionId?: string
  failureReason?: string
  refundedAmount: number
  createdAt: string
  updatedAt: string
}

export type PaymentStatus =
  | 'pending'
  | 'processing'
  | 'succeeded'
  | 'failed'
  | 'cancelled'
  | 'refunded'
  | 'partially_refunded'

export interface CreatePaymentIntentRequest {
  orderId: string
  paymentMethodId: string
  returnUrl?: string
}

// Review types
export interface Review {
  id: string
  productId: string
  userId: string
  user: User
  rating: number
  title: string
  comment: string
  isVerifiedPurchase: boolean
  isRecommended: boolean
  helpfulCount: number
  images?: ReviewImage[]
  createdAt: string
  updatedAt: string
}

export interface ReviewImage {
  id: string
  url: string
  altText: string
}

export interface CreateReviewRequest {
  productId: string
  rating: number
  title: string
  comment: string
  isRecommended: boolean
  images?: File[]
}

// Wishlist types
export interface WishlistItem {
  id: string
  productId: string
  product: Product
  addedAt: string
}

export interface Wishlist {
  id: string
  userId: string
  items: WishlistItem[]
  updatedAt: string
}

// Form types
export interface ContactForm {
  name: string
  email: string
  subject: string
  message: string
}

export interface NewsletterSubscription {
  email: string
}

// Error types
export interface ApiError {
  message: string
  code?: string
  field?: string
  details?: Record<string, any>
}

// UI State types
export interface LoadingState {
  [key: string]: boolean
}

export interface ErrorState {
  [key: string]: string | null
}

// Search & Filter types
export interface SearchSuggestion {
  type: 'product' | 'category' | 'brand'
  id: string
  name: string
  image?: string
}

export interface FilterOption {
  label: string
  value: string
  count?: number
}

export interface PriceRange {
  min: number
  max: number
}

// Analytics types
export interface AnalyticsEvent {
  event: string
  properties: Record<string, any>
  timestamp: string
}