# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

SoleMate is an e-commerce platform for shoe retail currently in the planning and early development phase. The project follows a microservices architecture with separate frontend, backend, and database components.

## Project Architecture

The codebase is structured with a clear separation of concerns:

```
mock_project/
├── src/
│   ├── frontend/     # React.js + Next.js frontend application
│   ├── backend/      # Node.js + Express.js backend services
│   └── database/     # PostgreSQL database schemas and migrations
├── tests/            # Test files for all components
├── docs/             # Comprehensive project documentation
├── deployment/       # Infrastructure and deployment configurations
└── README.md         # Basic project information
```

### Technology Stack (Planned)

**Frontend:**
- React.js 18+ with Next.js 14+ for SSR/SSG
- TailwindCSS for styling with HeadlessUI components
- Redux Toolkit with RTK Query for state management
- Framer Motion for animations

**Backend:**
- Node.js 18+ LTS with Express.js 4+
- RESTful APIs with OpenAPI/Swagger documentation
- JWT-based authentication
- Microservices architecture pattern

**Database:**
- PostgreSQL 15+ as primary database
- Redis 7+ for caching and session storage

**Infrastructure:**
- AWS cloud platform (EC2, RDS, ElastiCache, S3)
- Docker containerization
- GitHub Actions for CI/CD

## Development Commands

Currently, no package managers or build tools are configured as the project is in the planning phase. Once implementation begins, expect commands like:

- `npm install` - Install dependencies
- `npm run dev` - Start development server
- `npm run build` - Build for production
- `npm run test` - Run test suite
- `npm run lint` - Lint code
- `docker-compose up` - Start local development stack

## Key Business Requirements

Based on the SRS documentation, the platform must support:

1. **User Management:** Registration, authentication, profile management
2. **Product Catalog:** Advanced search, filtering, recommendations
3. **Shopping Cart & Checkout:** Multi-payment gateway support (Stripe, PayPal, UPI)
4. **Order Management:** Tracking, history, returns
5. **Admin Dashboard:** Product, inventory, order, and customer management
6. **Performance:** Sub-2-second page loads, 50k concurrent users
7. **Security:** PCI-DSS compliance, GDPR compliance

## Development Guidelines

- Target 99.9% uptime with comprehensive monitoring
- Maintain 80%+ unit test coverage
- Follow microservices patterns for scalability
- Implement responsive design for mobile-first approach
- Ensure WCAG 2.1 accessibility compliance
- Use conventional commits for version control

## Project Timeline

The project is planned for 6-month development cycle:
- **Phase 1 (Months 1-3):** Core platform (auth, catalog, cart, checkout)
- **Phase 2 (Months 4-5):** Enhanced features (recommendations, reviews, tracking)
- **Phase 3 (Month 6):** Optimization and production deployment