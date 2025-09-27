import Header from './Header'

interface LayoutProps {
  children: React.ReactNode
}

export default function Layout({ children }: LayoutProps) {
  return (
    <div className="min-h-screen bg-gray-50">
      <Header />
      <main>{children}</main>
      <footer className="bg-gray-900 text-white py-12 mt-16">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="grid grid-cols-1 md:grid-cols-4 gap-8">
            <div className="space-y-4">
              <h3 className="text-lg font-semibold">SoleMate</h3>
              <p className="text-gray-400">
                Your premier destination for the latest in footwear fashion.
              </p>
            </div>

            <div className="space-y-4">
              <h4 className="font-semibold">Shop</h4>
              <ul className="space-y-2 text-gray-400">
                <li><a href="/products" className="hover:text-white transition-colors">All Products</a></li>
                <li><a href="/products?category=mens" className="hover:text-white transition-colors">Men's Shoes</a></li>
                <li><a href="/products?category=womens" className="hover:text-white transition-colors">Women's Shoes</a></li>
                <li><a href="/products?featured=true" className="hover:text-white transition-colors">Featured</a></li>
              </ul>
            </div>

            <div className="space-y-4">
              <h4 className="font-semibold">Support</h4>
              <ul className="space-y-2 text-gray-400">
                <li><a href="/contact" className="hover:text-white transition-colors">Contact Us</a></li>
                <li><a href="/shipping" className="hover:text-white transition-colors">Shipping Info</a></li>
                <li><a href="/returns" className="hover:text-white transition-colors">Returns</a></li>
                <li><a href="/size-guide" className="hover:text-white transition-colors">Size Guide</a></li>
              </ul>
            </div>

            <div className="space-y-4">
              <h4 className="font-semibold">Company</h4>
              <ul className="space-y-2 text-gray-400">
                <li><a href="/about" className="hover:text-white transition-colors">About Us</a></li>
                <li><a href="/careers" className="hover:text-white transition-colors">Careers</a></li>
                <li><a href="/privacy" className="hover:text-white transition-colors">Privacy Policy</a></li>
                <li><a href="/terms" className="hover:text-white transition-colors">Terms of Service</a></li>
              </ul>
            </div>
          </div>

          <div className="border-t border-gray-800 mt-8 pt-8 text-center text-gray-400">
            <p>&copy; 2024 SoleMate. All rights reserved.</p>
          </div>
        </div>
      </footer>
    </div>
  )
}