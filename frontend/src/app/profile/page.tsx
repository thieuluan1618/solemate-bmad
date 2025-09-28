'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { useForm } from 'react-hook-form'
import { yupResolver } from '@hookform/resolvers/yup'
import { Button, Input } from '@/components/ui'
import Layout from '@/components/layout/Layout'
import { useGetProfileQuery, useUpdateProfileMutation, useChangePasswordMutation, useGetAddressesQuery, useCreateAddressMutation, useUpdateAddressMutation, useDeleteAddressMutation, useSetDefaultAddressMutation } from '@/store/api/userApi'
import { useAppSelector } from '@/hooks/redux'
import { selectIsAuthenticated } from '@/store/slices/authSlice'
import { profileSchema, changePasswordSchema, addressSchema, type ProfileFormData, type ChangePasswordFormData, type AddressFormData } from '@/lib/validations'
import type { Address } from '@/types'

type TabType = 'profile' | 'addresses' | 'security'

interface ProfileFormProps {
  user: any
  onUpdate: (data: ProfileFormData) => void
  isLoading: boolean
}

function ProfileForm({ user, onUpdate, isLoading }: ProfileFormProps) {
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<ProfileFormData>({
    resolver: yupResolver(profileSchema),
    defaultValues: {
      firstName: user?.firstName || '',
      lastName: user?.lastName || '',
      phone: user?.phone || '',
      preferences: {
        emailNotifications: user?.preferences?.emailNotifications ?? true,
        smsNotifications: user?.preferences?.smsNotifications ?? false,
        marketingEmails: user?.preferences?.marketingEmails ?? true,
      },
    },
  })

  return (
    <div className="card p-6">
      <h2 className="text-xl font-semibold text-gray-900 mb-6">Personal Information</h2>

      <form onSubmit={handleSubmit(onUpdate)} className="space-y-4">
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
          label="Email"
          type="email"
          value={user?.email || ''}
          disabled
          className="bg-gray-50"
          helpText="Email cannot be changed. Contact support if you need to update your email."
        />

        <Input
          label="Phone Number"
          type="tel"
          error={errors.phone?.message}
          {...register('phone')}
        />

        <div className="space-y-3">
          <h3 className="text-lg font-medium text-gray-900">Notification Preferences</h3>

          <div className="space-y-3">
            <label className="flex items-center">
              <input
                type="checkbox"
                className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                {...register('preferences.emailNotifications')}
              />
              <span className="ml-2 text-sm text-gray-700">Email notifications for orders and updates</span>
            </label>

            <label className="flex items-center">
              <input
                type="checkbox"
                className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                {...register('preferences.smsNotifications')}
              />
              <span className="ml-2 text-sm text-gray-700">SMS notifications for order updates</span>
            </label>

            <label className="flex items-center">
              <input
                type="checkbox"
                className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                {...register('preferences.marketingEmails')}
              />
              <span className="ml-2 text-sm text-gray-700">Marketing emails and promotions</span>
            </label>
          </div>
        </div>

        <div className="pt-4">
          <Button type="submit" variant="primary" disabled={isLoading}>
            {isLoading ? 'Updating...' : 'Update Profile'}
          </Button>
        </div>
      </form>
    </div>
  )
}

interface SecurityFormProps {
  onChangePassword: (data: ChangePasswordFormData) => void
  isLoading: boolean
}

function SecurityForm({ onChangePassword, isLoading }: SecurityFormProps) {
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<ChangePasswordFormData>({
    resolver: yupResolver(changePasswordSchema),
  })

  const handlePasswordChange = (data: ChangePasswordFormData) => {
    onChangePassword(data)
    reset()
  }

  return (
    <div className="card p-6">
      <h2 className="text-xl font-semibold text-gray-900 mb-6">Security Settings</h2>

      <form onSubmit={handleSubmit(handlePasswordChange)} className="space-y-4">
        <Input
          label="Current Password"
          type="password"
          error={errors.currentPassword?.message}
          {...register('currentPassword')}
        />

        <Input
          label="New Password"
          type="password"
          error={errors.newPassword?.message}
          {...register('newPassword')}
        />

        <Input
          label="Confirm New Password"
          type="password"
          error={errors.confirmNewPassword?.message}
          {...register('confirmNewPassword')}
        />

        <div className="pt-4">
          <Button type="submit" variant="primary" disabled={isLoading}>
            {isLoading ? 'Changing Password...' : 'Change Password'}
          </Button>
        </div>
      </form>
    </div>
  )
}

interface AddressCardProps {
  address: Address
  onEdit: (address: Address) => void
  onDelete: (id: string) => void
  onSetDefault: (id: string) => void
  isLoading: boolean
}

function AddressCard({ address, onEdit, onDelete, onSetDefault, isLoading }: AddressCardProps) {
  return (
    <div className={`card p-4 ${address.isDefault ? 'ring-2 ring-blue-500' : ''}`}>
      <div className="flex items-start justify-between">
        <div className="flex-1">
          <div className="flex items-center space-x-2 mb-2">
            <h3 className="font-medium text-gray-900">
              {address.firstName} {address.lastName}
            </h3>
            {address.isDefault && (
              <span className="px-2 py-1 text-xs bg-blue-100 text-blue-800 rounded-full">
                Default
              </span>
            )}
            <span className="px-2 py-1 text-xs bg-gray-100 text-gray-800 rounded-full capitalize">
              {address.type}
            </span>
          </div>

          <div className="text-sm text-gray-600 space-y-1">
            {address.company && <p>{address.company}</p>}
            <p>{address.street}</p>
            <p>{address.city}, {address.state} {address.zipCode}</p>
            <p>{address.country}</p>
          </div>
        </div>

        <div className="flex items-center space-x-2 ml-4">
          <button
            onClick={() => onEdit(address)}
            className="text-sm text-blue-600 hover:text-blue-800"
          >
            Edit
          </button>
          {!address.isDefault && (
            <button
              onClick={() => onSetDefault(address.id)}
              disabled={isLoading}
              className="text-sm text-gray-600 hover:text-gray-800 disabled:opacity-50"
            >
              Set Default
            </button>
          )}
          <button
            onClick={() => onDelete(address.id)}
            disabled={isLoading}
            className="text-sm text-red-600 hover:text-red-800 disabled:opacity-50"
          >
            Delete
          </button>
        </div>
      </div>
    </div>
  )
}

interface AddressFormProps {
  address?: Address
  onSave: (data: AddressFormData) => void
  onCancel: () => void
  isLoading: boolean
}

function AddressForm({ address, onSave, onCancel, isLoading }: AddressFormProps) {
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<AddressFormData>({
    resolver: yupResolver(addressSchema),
    defaultValues: address || { type: 'shipping' },
  })

  return (
    <div className="card p-6">
      <h3 className="text-lg font-semibold text-gray-900 mb-4">
        {address ? 'Edit Address' : 'Add New Address'}
      </h3>

      <form onSubmit={handleSubmit(onSave)} className="space-y-4">
        <div>
          <label className="text-sm font-medium text-gray-900 mb-2 block">Address Type</label>
          <div className="flex space-x-4">
            <label className="flex items-center">
              <input
                type="radio"
                value="shipping"
                className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300"
                {...register('type')}
              />
              <span className="ml-2 text-sm text-gray-700">Shipping</span>
            </label>
            <label className="flex items-center">
              <input
                type="radio"
                value="billing"
                className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300"
                {...register('type')}
              />
              <span className="ml-2 text-sm text-gray-700">Billing</span>
            </label>
          </div>
        </div>

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

        <div className="flex items-center">
          <input
            type="checkbox"
            className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
            {...register('isDefault')}
          />
          <span className="ml-2 text-sm text-gray-700">Set as default address</span>
        </div>

        <div className="flex space-x-4 pt-4">
          <Button type="button" variant="outline" onClick={onCancel}>
            Cancel
          </Button>
          <Button type="submit" variant="primary" disabled={isLoading}>
            {isLoading ? 'Saving...' : address ? 'Update Address' : 'Add Address'}
          </Button>
        </div>
      </form>
    </div>
  )
}

function AddressesSection() {
  const [editingAddress, setEditingAddress] = useState<Address | null>(null)
  const [showAddForm, setShowAddForm] = useState(false)

  const { data: addresses, isLoading } = useGetAddressesQuery()
  const [createAddress, { isLoading: isCreating }] = useCreateAddressMutation()
  const [updateAddress, { isLoading: isUpdating }] = useUpdateAddressMutation()
  const [deleteAddress, { isLoading: isDeleting }] = useDeleteAddressMutation()
  const [setDefaultAddress, { isLoading: isSettingDefault }] = useSetDefaultAddressMutation()

  const handleSaveAddress = async (data: AddressFormData) => {
    try {
      if (editingAddress) {
        await updateAddress({ id: editingAddress.id, data }).unwrap()
        setEditingAddress(null)
      } else {
        await createAddress(data).unwrap()
        setShowAddForm(false)
      }
    } catch (error: any) {
      alert(error?.data?.message || 'Failed to save address')
    }
  }

  const handleDeleteAddress = async (id: string) => {
    if (window.confirm('Are you sure you want to delete this address?')) {
      try {
        await deleteAddress({ id }).unwrap()
      } catch (error: any) {
        alert(error?.data?.message || 'Failed to delete address')
      }
    }
  }

  const handleSetDefault = async (id: string) => {
    try {
      await setDefaultAddress({ id }).unwrap()
    } catch (error: any) {
      alert(error?.data?.message || 'Failed to set default address')
    }
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-8">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
      </div>
    )
  }

  if (editingAddress || showAddForm) {
    return (
      <AddressForm
        address={editingAddress || undefined}
        onSave={handleSaveAddress}
        onCancel={() => {
          setEditingAddress(null)
          setShowAddForm(false)
        }}
        isLoading={isCreating || isUpdating}
      />
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-semibold text-gray-900">Saved Addresses</h2>
        <Button onClick={() => setShowAddForm(true)} variant="primary">
          Add Address
        </Button>
      </div>

      {addresses && addresses.length > 0 ? (
        <div className="grid gap-4">
          {addresses.map((address) => (
            <AddressCard
              key={address.id}
              address={address}
              onEdit={setEditingAddress}
              onDelete={handleDeleteAddress}
              onSetDefault={handleSetDefault}
              isLoading={isDeleting || isSettingDefault}
            />
          ))}
        </div>
      ) : (
        <div className="text-center py-8">
          <svg className="mx-auto h-12 w-12 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 11a3 3 0 11-6 0 3 3 0 016 0z" />
          </svg>
          <h3 className="mt-2 text-sm font-medium text-gray-900">No addresses</h3>
          <p className="mt-1 text-sm text-gray-500">Add an address to get started.</p>
          <div className="mt-6">
            <Button onClick={() => setShowAddForm(true)} variant="primary">
              Add Address
            </Button>
          </div>
        </div>
      )}
    </div>
  )
}

export default function ProfilePage() {
  const router = useRouter()
  const isAuthenticated = useAppSelector(selectIsAuthenticated)
  const [activeTab, setActiveTab] = useState<TabType>('profile')

  const { data: user, isLoading: userLoading } = useGetProfileQuery()
  const [updateProfile, { isLoading: isUpdatingProfile }] = useUpdateProfileMutation()
  const [changePassword, { isLoading: isChangingPassword }] = useChangePasswordMutation()

  useEffect(() => {
    if (!userLoading && !isAuthenticated) {
      router.push('/login?redirect=/profile')
    }
  }, [isAuthenticated, userLoading, router])

  const handleUpdateProfile = async (data: ProfileFormData) => {
    try {
      await updateProfile(data).unwrap()
      alert('Profile updated successfully!')
    } catch (error: any) {
      alert(error?.data?.message || 'Failed to update profile')
    }
  }

  const handleChangePassword = async (data: ChangePasswordFormData) => {
    try {
      await changePassword(data).unwrap()
      alert('Password changed successfully!')
    } catch (error: any) {
      alert(error?.data?.message || 'Failed to change password')
    }
  }

  if (userLoading || !isAuthenticated) {
    return (
      <Layout>
        <div className="min-h-screen flex items-center justify-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
        </div>
      </Layout>
    )
  }

  const tabs = [
    { id: 'profile', label: 'Profile', icon: '👤' },
    { id: 'addresses', label: 'Addresses', icon: '📍' },
    { id: 'security', label: 'Security', icon: '🔒' },
  ]

  return (
    <Layout>
      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="mb-8">
          <h1 className="text-2xl font-bold text-gray-900">Account Settings</h1>
          <p className="text-gray-600">Manage your account information and preferences</p>
        </div>

        {/* Tabs */}
        <div className="border-b border-gray-200 mb-8">
          <nav className="flex space-x-8">
            {tabs.map((tab) => (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id as TabType)}
                className={`py-2 px-1 border-b-2 font-medium text-sm transition-colors ${
                  activeTab === tab.id
                    ? 'border-blue-500 text-blue-600'
                    : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
                }`}
              >
                <span className="mr-2">{tab.icon}</span>
                {tab.label}
              </button>
            ))}
          </nav>
        </div>

        {/* Tab Content */}
        <div>
          {activeTab === 'profile' && (
            <ProfileForm
              user={user}
              onUpdate={handleUpdateProfile}
              isLoading={isUpdatingProfile}
            />
          )}

          {activeTab === 'addresses' && <AddressesSection />}

          {activeTab === 'security' && (
            <SecurityForm
              onChangePassword={handleChangePassword}
              isLoading={isChangingPassword}
            />
          )}
        </div>
      </div>
    </Layout>
  )
}