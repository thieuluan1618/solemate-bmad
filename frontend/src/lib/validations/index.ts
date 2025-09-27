import * as yup from 'yup'

// Authentication schemas
export const loginSchema = yup.object({
  email: yup
    .string()
    .email('Please enter a valid email address')
    .required('Email is required'),
  password: yup
    .string()
    .min(6, 'Password must be at least 6 characters')
    .required('Password is required'),
  rememberMe: yup.boolean().optional(),
})

export const registerSchema = yup.object({
  firstName: yup
    .string()
    .min(2, 'First name must be at least 2 characters')
    .required('First name is required'),
  lastName: yup
    .string()
    .min(2, 'Last name must be at least 2 characters')
    .required('Last name is required'),
  email: yup
    .string()
    .email('Please enter a valid email address')
    .required('Email is required'),
  password: yup
    .string()
    .min(8, 'Password must be at least 8 characters')
    .matches(
      /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)/,
      'Password must contain at least one uppercase letter, one lowercase letter, and one number'
    )
    .required('Password is required'),
  confirmPassword: yup
    .string()
    .oneOf([yup.ref('password')], 'Passwords must match')
    .required('Please confirm your password'),
  phone: yup
    .string()
    .matches(/^[\+]?[1-9][\d]{0,15}$/, 'Please enter a valid phone number')
    .optional(),
  acceptTerms: yup
    .boolean()
    .oneOf([true], 'You must accept the terms and conditions')
    .required(),
})

export const forgotPasswordSchema = yup.object({
  email: yup
    .string()
    .email('Please enter a valid email address')
    .required('Email is required'),
})

export const resetPasswordSchema = yup.object({
  password: yup
    .string()
    .min(8, 'Password must be at least 8 characters')
    .matches(
      /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)/,
      'Password must contain at least one uppercase letter, one lowercase letter, and one number'
    )
    .required('Password is required'),
  confirmPassword: yup
    .string()
    .oneOf([yup.ref('password')], 'Passwords must match')
    .required('Please confirm your password'),
})

export const changePasswordSchema = yup.object({
  currentPassword: yup
    .string()
    .required('Current password is required'),
  newPassword: yup
    .string()
    .min(8, 'Password must be at least 8 characters')
    .matches(
      /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)/,
      'Password must contain at least one uppercase letter, one lowercase letter, and one number'
    )
    .notOneOf([yup.ref('currentPassword')], 'New password must be different from current password')
    .required('New password is required'),
  confirmNewPassword: yup
    .string()
    .oneOf([yup.ref('newPassword')], 'Passwords must match')
    .required('Please confirm your new password'),
})

// Profile schemas
export const profileSchema = yup.object({
  firstName: yup
    .string()
    .min(2, 'First name must be at least 2 characters')
    .required('First name is required'),
  lastName: yup
    .string()
    .min(2, 'Last name must be at least 2 characters')
    .required('Last name is required'),
  phone: yup
    .string()
    .matches(/^[\+]?[1-9][\d]{0,15}$/, 'Please enter a valid phone number')
    .optional(),
  preferences: yup.object({
    emailNotifications: yup.boolean().optional(),
    smsNotifications: yup.boolean().optional(),
    marketingEmails: yup.boolean().optional(),
  }).optional(),
})

// Address schemas
export const addressSchema = yup.object({
  type: yup
    .string()
    .oneOf(['shipping', 'billing'], 'Please select address type')
    .required('Address type is required'),
  firstName: yup
    .string()
    .min(2, 'First name must be at least 2 characters')
    .required('First name is required'),
  lastName: yup
    .string()
    .min(2, 'Last name must be at least 2 characters')
    .required('Last name is required'),
  company: yup.string().optional(),
  street: yup
    .string()
    .min(5, 'Street address must be at least 5 characters')
    .required('Street address is required'),
  city: yup
    .string()
    .min(2, 'City must be at least 2 characters')
    .required('City is required'),
  state: yup
    .string()
    .min(2, 'State must be at least 2 characters')
    .required('State is required'),
  zipCode: yup
    .string()
    .matches(/^\d{5}(-\d{4})?$/, 'Please enter a valid ZIP code')
    .required('ZIP code is required'),
  country: yup
    .string()
    .min(2, 'Country must be at least 2 characters')
    .required('Country is required'),
  isDefault: yup.boolean().optional(),
})

// Review schemas
export const reviewSchema = yup.object({
  rating: yup
    .number()
    .min(1, 'Rating must be at least 1 star')
    .max(5, 'Rating cannot exceed 5 stars')
    .required('Rating is required'),
  title: yup
    .string()
    .min(5, 'Title must be at least 5 characters')
    .max(100, 'Title cannot exceed 100 characters')
    .required('Title is required'),
  comment: yup
    .string()
    .min(20, 'Comment must be at least 20 characters')
    .max(1000, 'Comment cannot exceed 1000 characters')
    .required('Comment is required'),
  isRecommended: yup.boolean().required('Recommendation is required'),
})

// Contact form schemas
export const contactSchema = yup.object({
  name: yup
    .string()
    .min(2, 'Name must be at least 2 characters')
    .required('Name is required'),
  email: yup
    .string()
    .email('Please enter a valid email address')
    .required('Email is required'),
  subject: yup
    .string()
    .min(5, 'Subject must be at least 5 characters')
    .required('Subject is required'),
  message: yup
    .string()
    .min(20, 'Message must be at least 20 characters')
    .required('Message is required'),
})

export const newsletterSchema = yup.object({
  email: yup
    .string()
    .email('Please enter a valid email address')
    .required('Email is required'),
})

// Search schemas
export const searchSchema = yup.object({
  query: yup
    .string()
    .min(2, 'Search query must be at least 2 characters')
    .required('Search query is required'),
})

// Filter schemas
export const filterSchema = yup.object({
  categories: yup.array().of(yup.string()).optional(),
  brands: yup.array().of(yup.string()).optional(),
  priceMin: yup.number().min(0, 'Minimum price cannot be negative').optional(),
  priceMax: yup
    .number()
    .min(0, 'Maximum price cannot be negative')
    .when('priceMin', (priceMin, schema) =>
      priceMin ? schema.min(priceMin, 'Maximum price must be greater than minimum price') : schema
    )
    .optional(),
  sizes: yup.array().of(yup.string()).optional(),
  colors: yup.array().of(yup.string()).optional(),
  rating: yup.number().min(1).max(5).optional(),
  inStock: yup.boolean().optional(),
  featured: yup.boolean().optional(),
})

// Type definitions for forms
export type LoginFormData = yup.InferType<typeof loginSchema>
export type RegisterFormData = yup.InferType<typeof registerSchema>
export type ForgotPasswordFormData = yup.InferType<typeof forgotPasswordSchema>
export type ResetPasswordFormData = yup.InferType<typeof resetPasswordSchema>
export type ChangePasswordFormData = yup.InferType<typeof changePasswordSchema>
export type ProfileFormData = yup.InferType<typeof profileSchema>
export type AddressFormData = yup.InferType<typeof addressSchema>
export type ReviewFormData = yup.InferType<typeof reviewSchema>
export type ContactFormData = yup.InferType<typeof contactSchema>
export type NewsletterFormData = yup.InferType<typeof newsletterSchema>
export type SearchFormData = yup.InferType<typeof searchSchema>
export type FilterFormData = yup.InferType<typeof filterSchema>