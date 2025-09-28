'use client'

import Link from 'next/link'
import { Button, Input } from '@/components/ui'

export default function Home() {
  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white shadow-sm">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-16">
            <div className="flex items-center">
              <h1 className="text-2xl font-bold bg-gradient-to-r from-blue-600 to-purple-600 bg-clip-text text-transparent">
                SoleMate
              </h1>
            </div>
            <div className="flex items-center space-x-4">
              <Link href="/login">
                <Button variant="outline">Login</Button>
              </Link>
              <Link href="/register">
                <Button variant="primary">Sign Up</Button>
              </Link>
            </div>
          </div>
        </div>
      </header>

      {/* Hero Section */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
        <div className="text-center">
          <h2 className="text-4xl font-bold text-gray-900 mb-4">
            Welcome to SoleMate
          </h2>
          <p className="text-xl text-gray-600 mb-8">
            Your premier destination for the latest in footwear fashion
          </p>

          {/* Call to Action */}
          <div className="max-w-md mx-auto space-y-4">
            <Link href="/products">
              <Button variant="primary" className="w-full text-lg py-4">
                Shop Now
              </Button>
            </Link>
            <p className="text-center text-gray-600">
              Discover our collection of premium footwear
            </p>
          </div>
        </div>
      </main>

      {/* Footer */}
      <footer className="bg-gray-900 text-white py-8 mt-12">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 text-center">
          <p>&copy; 2024 SoleMate. All rights reserved.</p>
        </div>
      </footer>
    </div>
  )
}
