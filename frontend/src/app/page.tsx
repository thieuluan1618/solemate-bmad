'use client'

import { Button, Input } from '@/components/ui'

export default function Home() {
  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white shadow-sm">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-16">
            <div className="flex items-center">
              <h1 className="text-2xl font-bold gradient-primary bg-clip-text text-transparent">
                SoleMate
              </h1>
            </div>
            <div className="flex items-center space-x-4">
              <Button variant="outline">Login</Button>
              <Button variant="primary">Sign Up</Button>
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

          {/* Test our components */}
          <div className="max-w-md mx-auto space-y-4">
            <div className="card p-6">
              <h3 className="text-lg font-semibold mb-4">Test Our Components</h3>

              <div className="space-y-4">
                <Input
                  label="Email"
                  type="email"
                  placeholder="Enter your email"
                />

                <Input
                  label="Password"
                  type="password"
                  placeholder="Enter your password"
                />

                <div className="flex space-x-2">
                  <Button variant="primary" className="flex-1">
                    Primary Button
                  </Button>
                  <Button variant="secondary" className="flex-1">
                    Secondary Button
                  </Button>
                </div>

                <Button variant="outline" className="w-full">
                  Outline Button
                </Button>

                <Button variant="ghost" className="w-full">
                  Ghost Button
                </Button>
              </div>
            </div>
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
