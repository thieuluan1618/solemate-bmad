import { createSlice, PayloadAction } from '@reduxjs/toolkit'

export interface CartItem {
  id: string
  productId: string
  name: string
  price: number
  quantity: number
  size?: string
  color?: string
  image: string
  maxQuantity: number
}

interface CartState {
  items: CartItem[]
  total: number
  itemCount: number
  loading: boolean
  error: string | null
}

const initialState: CartState = {
  items: [],
  total: 0,
  itemCount: 0,
  loading: false,
  error: null,
}

const cartSlice = createSlice({
  name: 'cart',
  initialState,
  reducers: {
    addItem: (state, action: PayloadAction<CartItem>) => {
      const existingItem = state.items.find(
        item =>
          item.productId === action.payload.productId &&
          item.size === action.payload.size &&
          item.color === action.payload.color
      )

      if (existingItem) {
        existingItem.quantity = Math.min(
          existingItem.quantity + action.payload.quantity,
          existingItem.maxQuantity
        )
      } else {
        state.items.push(action.payload)
      }

      cartSlice.caseReducers.calculateTotals(state)
    },
    removeItem: (state, action: PayloadAction<string>) => {
      state.items = state.items.filter(item => item.id !== action.payload)
      cartSlice.caseReducers.calculateTotals(state)
    },
    updateQuantity: (
      state,
      action: PayloadAction<{ id: string; quantity: number }>
    ) => {
      const item = state.items.find(item => item.id === action.payload.id)
      if (item) {
        item.quantity = Math.min(
          Math.max(action.payload.quantity, 1),
          item.maxQuantity
        )
      }
      cartSlice.caseReducers.calculateTotals(state)
    },
    clearCart: (state) => {
      state.items = []
      state.total = 0
      state.itemCount = 0
    },
    setLoading: (state, action: PayloadAction<boolean>) => {
      state.loading = action.payload
    },
    setError: (state, action: PayloadAction<string | null>) => {
      state.error = action.payload
    },
    calculateTotals: (state) => {
      state.total = state.items.reduce(
        (total, item) => total + item.price * item.quantity,
        0
      )
      state.itemCount = state.items.reduce(
        (count, item) => count + item.quantity,
        0
      )
    },
    syncCartItems: (state, action: PayloadAction<CartItem[]>) => {
      state.items = action.payload
      cartSlice.caseReducers.calculateTotals(state)
    },
  },
})

export const {
  addItem,
  removeItem,
  updateQuantity,
  clearCart,
  setLoading,
  setError,
  calculateTotals,
  syncCartItems,
} = cartSlice.actions

export default cartSlice.reducer

// Selectors
export const selectCartItems = (state: { cart: CartState }) => state.cart.items
export const selectCartTotal = (state: { cart: CartState }) => state.cart.total
export const selectCartItemCount = (state: { cart: CartState }) => state.cart.itemCount
export const selectCartLoading = (state: { cart: CartState }) => state.cart.loading
export const selectCartError = (state: { cart: CartState }) => state.cart.error