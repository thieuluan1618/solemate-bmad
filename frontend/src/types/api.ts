// API endpoint types for RTK Query

export interface AuthApiEndpoints {
  login: {
    request: { email: string; password: string; rememberMe?: boolean }
    response: { user: User; accessToken: string; refreshToken: string }
  }
  register: {
    request: { email: string; password: string; firstName: string; lastName: string; phone?: string }
    response: { user: User; accessToken: string; refreshToken: string }
  }
  refresh: {
    request: { refreshToken: string }
    response: { accessToken: string; refreshToken: string }
  }
  logout: {
    request: {}
    response: { message: string }
  }
  forgotPassword: {
    request: { email: string }
    response: { message: string }
  }
  resetPassword: {
    request: { token: string; password: string }
    response: { message: string }
  }
}

export interface UserApiEndpoints {
  getProfile: {
    request: {}
    response: User
  }
  updateProfile: {
    request: Partial<Pick<User, 'firstName' | 'lastName' | 'phone' | 'preferences'>>
    response: User
  }
  changePassword: {
    request: { currentPassword: string; newPassword: string }
    response: { message: string }
  }
  getAddresses: {
    request: {}
    response: Address[]
  }
  createAddress: {
    request: Omit<Address, 'id' | 'userId' | 'createdAt' | 'updatedAt'>
    response: Address
  }
  updateAddress: {
    request: { id: string; data: Partial<Omit<Address, 'id' | 'userId' | 'createdAt' | 'updatedAt'>> }
    response: Address
  }
  deleteAddress: {
    request: { id: string }
    response: { message: string }
  }
}

export interface ProductApiEndpoints {
  getProducts: {
    request: ProductSearchParams
    response: PaginatedResponse<Product>
  }
  getProduct: {
    request: { id: string }
    response: Product
  }
  searchProducts: {
    request: { query: string; limit?: number }
    response: { products: Product[]; suggestions: SearchSuggestion[] }
  }
  getCategories: {
    request: { parentId?: string }
    response: Category[]
  }
  getCategory: {
    request: { id: string }
    response: Category
  }
  getBrands: {
    request: {}
    response: Brand[]
  }
  getBrand: {
    request: { id: string }
    response: Brand
  }
  getFeaturedProducts: {
    request: { limit?: number }
    response: Product[]
  }
  getRelatedProducts: {
    request: { productId: string; limit?: number }
    response: Product[]
  }
}

export interface CartApiEndpoints {
  getCart: {
    request: {}
    response: Cart
  }
  addToCart: {
    request: AddToCartRequest
    response: Cart
  }
  updateCartItem: {
    request: { itemId: string; quantity: number }
    response: Cart
  }
  removeFromCart: {
    request: { itemId: string }
    response: Cart
  }
  clearCart: {
    request: {}
    response: { message: string }
  }
  syncCart: {
    request: { items: AddToCartRequest[] }
    response: Cart
  }
}

export interface OrderApiEndpoints {
  getOrders: {
    request: { page?: number; limit?: number; status?: OrderStatus }
    response: PaginatedResponse<Order>
  }
  getOrder: {
    request: { id: string }
    response: Order
  }
  createOrder: {
    request: CreateOrderRequest
    response: Order
  }
  cancelOrder: {
    request: { id: string; reason?: string }
    response: Order
  }
  trackOrder: {
    request: { id: string }
    response: { order: Order; tracking: TrackingInfo[] }
  }
}

export interface PaymentApiEndpoints {
  createPaymentIntent: {
    request: CreatePaymentIntentRequest
    response: { clientSecret: string; paymentIntentId: string }
  }
  confirmPayment: {
    request: { paymentIntentId: string; paymentMethodId: string }
    response: Payment
  }
  getPaymentMethods: {
    request: {}
    response: PaymentMethod[]
  }
  addPaymentMethod: {
    request: { paymentMethodId: string; isDefault?: boolean }
    response: PaymentMethod
  }
  updatePaymentMethod: {
    request: { id: string; isDefault: boolean }
    response: PaymentMethod
  }
  deletePaymentMethod: {
    request: { id: string }
    response: { message: string }
  }
}

export interface ReviewApiEndpoints {
  getProductReviews: {
    request: { productId: string; page?: number; limit?: number; sortBy?: string }
    response: PaginatedResponse<Review>
  }
  createReview: {
    request: CreateReviewRequest
    response: Review
  }
  updateReview: {
    request: { id: string; data: Partial<CreateReviewRequest> }
    response: Review
  }
  deleteReview: {
    request: { id: string }
    response: { message: string }
  }
  markReviewHelpful: {
    request: { id: string }
    response: Review
  }
}

export interface WishlistApiEndpoints {
  getWishlist: {
    request: {}
    response: Wishlist
  }
  addToWishlist: {
    request: { productId: string }
    response: Wishlist
  }
  removeFromWishlist: {
    request: { productId: string }
    response: Wishlist
  }
  clearWishlist: {
    request: {}
    response: { message: string }
  }
}

// Import types from main types file
import type {
  User,
  Address,
  Product,
  ProductSearchParams,
  PaginatedResponse,
  SearchSuggestion,
  Category,
  Brand,
  Cart,
  AddToCartRequest,
  Order,
  OrderStatus,
  CreateOrderRequest,
  TrackingInfo,
  Payment,
  CreatePaymentIntentRequest,
  PaymentMethod,
  Review,
  CreateReviewRequest,
  Wishlist,
} from './index'

// Additional types for tracking
export interface TrackingInfo {
  id: string
  status: string
  location: string
  timestamp: string
  description: string
}