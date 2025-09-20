# Use Case Diagrams & User Stories
## SoleMate E-Commerce Platform

### Document Information
- **Project Name:** SoleMate E-Commerce Platform  
- **Document Version:** 1.0
- **Date:** September 15, 2024
- **Prepared by:** Development Team

---

## 1. Use Case Overview

### 1.1 Actors
- **Customer:** End users who browse and purchase shoes
- **Admin:** System administrators managing the platform
- **Guest User:** Non-registered users browsing the platform
- **Payment Gateway:** External payment processing system
- **Inventory System:** External inventory management system
- **Email System:** External email notification service

### 1.2 Use Case Diagram Description

```
                    SoleMate E-Commerce Platform

    Customer                                                Admin
        |                                                     |
        |-- Browse Products                     Manage Products --|
        |-- Search Products                     Manage Orders ----|
        |-- View Product Details              Manage Inventory ---|
        |-- Add to Cart                      Manage Customers ----|
        |-- Manage Cart                       View Analytics -----|
        |-- Checkout                         Generate Reports ----|
        |-- Make Payment                                          |
        |-- Track Orders                                          |
        |-- Write Reviews                                         |
        |-- Manage Profile                                        |
        |-- View Order History                                    |
        |                                                         |
    Guest User                                                    |
        |                                                         |
        |-- Browse Products                                       |
        |-- Search Products                                       |
        |-- View Product Details                                  |
        |-- Register Account                                      |
        
    External Systems:
    - Payment Gateway (Stripe, PayPal, UPI)
    - Email Service (Notifications)
    - Shipping Providers (Tracking)
```

---

## 2. Detailed Use Cases

### 2.1 Customer Use Cases

#### UC-001: Browse Products
**Actor:** Customer, Guest User  
**Description:** Users can browse through the product catalog  
**Preconditions:** User accesses the website  
**Main Flow:**
1. User navigates to product catalog page
2. System displays products in grid/list format
3. User can filter by category, brand, price, size, color
4. User can sort by price, popularity, ratings, newest
5. System displays filtered and sorted results
6. User can paginate through results

**Alternative Flows:**
- A1: No products match filter criteria - display "No products found" message
- A2: Network error - display error message with retry option

#### UC-002: Search Products  
**Actor:** Customer, Guest User  
**Description:** Users can search for specific products  
**Preconditions:** User accesses the website  
**Main Flow:**
1. User enters search query in search bar
2. System provides autocomplete suggestions
3. User selects suggestion or presses enter
4. System displays search results
5. User can apply additional filters to results
6. User can sort search results

**Alternative Flows:**
- A1: No search results - display "No products found" with suggestions
- A2: Typo in search - display "Did you mean..." suggestions

#### UC-003: Add to Cart
**Actor:** Customer, Guest User  
**Description:** Users can add products to shopping cart  
**Preconditions:** User is viewing a product  
**Main Flow:**
1. User selects product size and color (if applicable)
2. User specifies quantity
3. User clicks "Add to Cart" button
4. System validates inventory availability
5. System adds item to cart
6. System displays cart summary
7. System shows success notification

**Alternative Flows:**
- A1: Out of stock - display "Out of Stock" message
- A2: Invalid selection - highlight required fields
- A3: Quantity exceeds available stock - display maximum available

#### UC-004: Checkout Process
**Actor:** Customer  
**Description:** Customer completes purchase  
**Preconditions:** Customer has items in cart and is logged in  
**Main Flow:**
1. User proceeds to checkout from cart
2. System displays order summary
3. User selects/enters shipping address
4. User selects shipping method
5. User applies discount codes (if any)
6. User selects payment method
7. User enters payment details
8. System processes payment through gateway
9. System creates order record
10. System sends confirmation email
11. System displays order confirmation page

**Alternative Flows:**
- A1: Payment failure - return to payment step with error message
- A2: Inventory changed - update cart and notify user
- A3: Invalid discount code - display error message

### 2.2 Admin Use Cases

#### UC-005: Manage Products
**Actor:** Admin  
**Description:** Admin manages product catalog  
**Preconditions:** Admin is authenticated  
**Main Flow:**
1. Admin navigates to product management section
2. System displays product list
3. Admin can add new product with details
4. Admin can edit existing product information
5. Admin can upload/manage product images
6. Admin can set product pricing and inventory
7. Admin can activate/deactivate products
8. System saves changes to database

**Alternative Flows:**
- A1: Validation errors - display error messages
- A2: Image upload fails - display error and allow retry

#### UC-006: Manage Orders
**Actor:** Admin  
**Description:** Admin processes and manages orders  
**Preconditions:** Admin is authenticated  
**Main Flow:**
1. Admin navigates to order management section
2. System displays order list with filters
3. Admin can view order details
4. Admin can update order status
5. Admin can process refunds/returns
6. Admin can generate shipping labels
7. System sends status update notifications to customers

**Alternative Flows:**
- A1: Order cannot be cancelled - display reason
- A2: Refund processing fails - display error message

---

## 3. User Stories

### 3.1 Epic 1: User Authentication & Profile Management

#### US-001: User Registration
**As a** new customer  
**I want to** create an account with my email and password  
**So that** I can save my preferences and track my orders  

**Acceptance Criteria:**
- User can register with valid email and password
- Password must meet security requirements (8+ chars, mixed case, numbers)
- Email verification is required before account activation
- User receives welcome email after successful registration
- Duplicate email registration is prevented

#### US-002: User Login
**As a** registered customer  
**I want to** log into my account  
**So that** I can access my saved information and order history  

**Acceptance Criteria:**
- User can login with valid email/password combination
- System remembers user session across browser tabs
- Failed login attempts are limited and tracked
- User can reset password via email if forgotten
- User can logout securely

#### US-003: Social Login
**As a** customer  
**I want to** login using my Google/Facebook account  
**So that** I don't need to remember another password  

**Acceptance Criteria:**
- User can authenticate via Google OAuth
- User can authenticate via Facebook OAuth
- System creates user profile from social media data
- User can link/unlink social accounts later
- Privacy controls are respected

### 3.2 Epic 2: Product Discovery & Browsing

#### US-004: Browse Product Catalog
**As a** customer  
**I want to** browse through available shoes  
**So that** I can discover products that interest me  

**Acceptance Criteria:**
- Products are displayed in an attractive grid layout
- Each product shows image, name, price, and rating
- Page loads within 2 seconds
- Pagination works smoothly with 20 products per page
- Mobile view is responsive and user-friendly

#### US-005: Filter Products
**As a** customer  
**I want to** filter products by category, brand, size, color, and price  
**So that** I can find shoes that match my specific needs  

**Acceptance Criteria:**
- Multiple filters can be applied simultaneously
- Filter results update immediately without page reload
- Active filters are clearly displayed and removable
- Filter counts show number of matching products
- Filters maintain state during session

#### US-006: Search Products
**As a** customer  
**I want to** search for shoes using keywords  
**So that** I can quickly find specific products  

**Acceptance Criteria:**
- Search bar is prominently displayed on all pages
- Autocomplete suggestions appear as user types
- Search handles typos and suggests corrections
- Results are ranked by relevance
- Search history is maintained for logged-in users

#### US-007: View Product Details
**As a** customer  
**I want to** see detailed information about a shoe  
**So that** I can make an informed purchase decision  

**Acceptance Criteria:**
- High-quality images with zoom functionality
- Complete product specifications (materials, sizing guide)
- Customer reviews and ratings are visible
- Stock availability is clearly indicated
- Related products are suggested

### 3.3 Epic 3: Shopping Cart & Checkout

#### US-008: Add Products to Cart
**As a** customer  
**I want to** add shoes to my shopping cart  
**So that** I can purchase multiple items together  

**Acceptance Criteria:**
- User can select size and color before adding to cart
- Cart icon updates with item count
- Success message confirms item was added
- Cart persists across browser sessions for logged-in users
- Inventory is checked before adding to cart

#### US-009: Manage Shopping Cart
**As a** customer  
**I want to** view and modify items in my cart  
**So that** I can review my selections before checkout  

**Acceptance Criteria:**
- Cart displays all added items with details
- User can modify quantities or remove items
- Total price updates automatically with changes
- Shipping costs are calculated and displayed
- Cart can be saved for later (wishlist functionality)

#### US-010: Secure Checkout
**As a** customer  
**I want to** complete my purchase securely  
**So that** I can buy the shoes I want with confidence  

**Acceptance Criteria:**
- Checkout process is simple and intuitive
- Multiple payment methods are supported (cards, PayPal, UPI)
- SSL encryption protects payment information
- Order confirmation is displayed immediately
- Confirmation email is sent within 5 minutes

### 3.4 Epic 4: Order Management

#### US-011: Track Orders
**As a** customer  
**I want to** track the status of my orders  
**So that** I know when to expect delivery  

**Acceptance Criteria:**
- Order status is updated in real-time
- Tracking information is provided when available
- Email notifications are sent for status changes
- Estimated delivery dates are accurate
- User can view detailed order history

#### US-012: Order History
**As a** customer  
**I want to** view my past orders  
**So that** I can reorder items or check previous purchases  

**Acceptance Criteria:**
- Complete order history is accessible from user profile
- Orders can be filtered by date range and status
- Invoice/receipt can be downloaded for each order
- Reorder functionality is available for past purchases
- Return/exchange options are clearly presented

### 3.5 Epic 5: Product Reviews

#### US-013: Write Product Reviews
**As a** customer  
**I want to** write reviews for shoes I've purchased  
**So that** I can help other customers make decisions  

**Acceptance Criteria:**
- Only verified purchasers can write reviews
- Reviews include 5-star rating and written feedback
- Photos can be uploaded with reviews
- Reviews are moderated before publication
- User can edit/delete their own reviews

#### US-014: Read Product Reviews
**As a** customer  
**I want to** read reviews from other customers  
**So that** I can make informed purchase decisions  

**Acceptance Criteria:**
- Reviews are displayed prominently on product pages
- Average rating is calculated and displayed
- Reviews can be sorted by date, rating, helpfulness
- Most helpful reviews are highlighted
- Review summary shows rating distribution

### 3.6 Epic 6: Admin Management

#### US-015: Product Management (Admin)
**As an** admin  
**I want to** manage the product catalog  
**So that** customers have access to current inventory  

**Acceptance Criteria:**
- Admin can add new products with all details
- Bulk operations are supported for efficiency
- Product images can be uploaded and managed
- Inventory levels can be set and tracked
- Products can be featured or hidden

#### US-016: Order Management (Admin)
**As an** admin  
**I want to** process and fulfill customer orders  
**So that** customers receive their purchases promptly  

**Acceptance Criteria:**
- All orders are visible in admin dashboard
- Order status can be updated with tracking information
- Shipping labels can be generated
- Returns and refunds can be processed
- Customer communication is tracked

#### US-017: Analytics Dashboard (Admin)
**As an** admin  
**I want to** view sales and performance analytics  
**So that** I can make informed business decisions  

**Acceptance Criteria:**
- Sales metrics are updated in real-time
- Popular products and trends are identified
- Customer behavior insights are provided
- Revenue reports can be generated
- Performance benchmarks are tracked

---

## 4. User Story Mapping

### Phase 1 (MVP - 3 months)
- User Registration & Authentication
- Product Browsing & Search
- Basic Shopping Cart
- Simple Checkout Process
- Order Confirmation

### Phase 2 (Extended Features - 2 months)  
- Advanced Filtering
- Product Reviews & Ratings
- Order Tracking
- User Profile Management
- Admin Dashboard

### Phase 3 (Advanced Features - 1 month)
- Promotional Features
- Advanced Analytics
- Mobile App
- Social Features
- API Integrations

---

## 5. Acceptance Testing Scenarios

### Scenario 1: New Customer Journey
1. New user visits the website
2. Browses product catalog
3. Searches for specific shoe type
4. Views product details and reviews
5. Creates account during checkout
6. Completes purchase with card payment
7. Receives order confirmation
8. Tracks order status
9. Receives product and writes review

### Scenario 2: Returning Customer Journey
1. Existing customer logs in
2. Checks order history
3. Reorders previous purchase
4. Applies discount code
5. Uses saved payment method
6. Completes checkout quickly
7. Updates delivery preferences

### Scenario 3: Admin Management Scenario
1. Admin logs into dashboard
2. Adds new product with images
3. Updates inventory levels
4. Processes pending orders
5. Generates sales report
6. Responds to customer inquiries

---

**Document Status:** Ready for Review  
**Next Steps:** Requirements Traceability Matrix (RTM) creation