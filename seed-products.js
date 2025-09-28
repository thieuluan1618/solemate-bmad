const { Client } = require('pg');

const client = new Client({
  host: 'localhost',
  port: 5432,
  user: 'solemate',
  password: 'localdev',
  database: 'solemate_db',
});

const sampleProducts = [
  {
    name: 'Nike Air Max 270',
    description: 'The Nike Air Max 270 delivers visible cushioning under every step. Its bold design draws inspiration from Air Max icons, showcasing Nike\'s greatest innovation with its large window and fresh array of colors.',
    price: 150.00,
    category: 'Sneakers',
    brand: 'Nike',
    sku: 'NIKE-AM270-001',
    stock_quantity: 50,
    images: ['https://images.unsplash.com/photo-1542291026-7eec264c27ff?w=500']
  },
  {
    name: 'Adidas Ultraboost 22',
    description: 'The Adidas Ultraboost 22 features our most responsive Boost cushioning ever. Made with at least 50% recycled content, this running shoe delivers the energy return you need to go the distance.',
    price: 180.00,
    category: 'Running',
    brand: 'Adidas',
    sku: 'ADIDAS-UB22-001',
    stock_quantity: 30,
    images: ['https://images.unsplash.com/photo-1606107557195-0e29a4b5b4aa?w=500']
  },
  {
    name: 'Converse Chuck Taylor All Star',
    description: 'The Converse Chuck Taylor All Star is the most iconic sneaker in the world, recognized for its unmistakable silhouette, star-centered ankle patch and cultural authenticity.',
    price: 65.00,
    category: 'Casual',
    brand: 'Converse',
    sku: 'CONVERSE-CT-001',
    stock_quantity: 100,
    images: ['https://images.unsplash.com/photo-1525966222134-fcfa99b8ae77?w=500']
  },
  {
    name: 'Vans Old Skool',
    description: 'The Vans Old Skool features the classic low-top lace-up silhouette with the signature side stripe and padded collar for comfort and style.',
    price: 70.00,
    category: 'Skateboarding',
    brand: 'Vans',
    sku: 'VANS-OS-001',
    stock_quantity: 75,
    images: ['https://images.unsplash.com/photo-1551107696-a4b0c5a0d9a2?w=500']
  },
  {
    name: 'Jordan 1 Retro High',
    description: 'The Air Jordan 1 Retro High OG pays homage to the classic that started it all, with premium materials and the iconic colorway that made history.',
    price: 170.00,
    category: 'Basketball',
    brand: 'Jordan',
    sku: 'JORDAN-1RH-001',
    stock_quantity: 25,
    images: ['https://images.unsplash.com/photo-1552346154-21d32810aba3?w=500']
  },
  {
    name: 'New Balance 990v5',
    description: 'The New Balance 990v5 features premium pigskin and mesh uppers with ENCAP midsole technology for superior comfort and durability.',
    price: 185.00,
    category: 'Running',
    brand: 'New Balance',
    sku: 'NB-990V5-001',
    stock_quantity: 40,
    images: ['https://images.unsplash.com/photo-1549298916-b41d501d3772?w=500']
  }
];

async function seedProducts() {
  try {
    await client.connect();
    console.log('Connected to database');

    // Create categories if they don't exist
    const categories = ['Sneakers', 'Running', 'Casual', 'Skateboarding', 'Basketball'];
    for (const categoryName of categories) {
      await client.query(
        'INSERT INTO categories (name, slug, description, is_active, created_at) VALUES ($1, $2, $3, true, CURRENT_TIMESTAMP) ON CONFLICT (slug) DO NOTHING',
        [categoryName, categoryName.toLowerCase().replace(/\s+/g, '-'), `Great ${categoryName} shoes`]
      );
    }

    // Create brands if they don't exist
    const brands = ['Nike', 'Adidas', 'Converse', 'Vans', 'Jordan', 'New Balance'];
    for (const brandName of brands) {
      await client.query(
        'INSERT INTO brands (name, slug, description, is_active, created_at) VALUES ($1, $2, $3, true, CURRENT_TIMESTAMP) ON CONFLICT (slug) DO NOTHING',
        [brandName, brandName.toLowerCase().replace(/\s+/g, '-'), `Premium ${brandName} footwear`]
      );
    }

    // Insert products
    for (const product of sampleProducts) {
      // Get category ID
      const categoryResult = await client.query('SELECT id FROM categories WHERE name = $1', [product.category]);
      const categoryId = categoryResult.rows[0]?.id;

      // Get brand ID
      const brandResult = await client.query('SELECT id FROM brands WHERE name = $1', [product.brand]);
      const brandId = brandResult.rows[0]?.id;

      // Insert product
      const productResult = await client.query(
        `INSERT INTO products (
          name, slug, description, price, category_id, brand_id, sku, 
          is_active, created_at, updated_at
        ) VALUES (
          $1, $2, $3, $4, $5, $6, $7, true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
        ) RETURNING id`,
        [product.name, product.sku.toLowerCase(), product.description, product.price, categoryId, brandId, product.sku]
      );

      const productId = productResult.rows[0].id;

      // Insert product variant with stock
      await client.query(
        `INSERT INTO product_variants (
          product_id, sku, price, stock, is_active, created_at
        ) VALUES (
          $1, $2, $3, $4, true, CURRENT_TIMESTAMP
        )`,
        [productId, `${product.sku}-DEFAULT`, product.price, product.stock_quantity]
      );

      // Insert product images
      for (const imageUrl of product.images) {
        await client.query(
          'INSERT INTO product_images (product_id, url, alt_text, sort_order, is_primary, created_at) VALUES ($1, $2, $3, 0, true, CURRENT_TIMESTAMP)',
          [productId, imageUrl, `${product.name} image`]
        );
      }

      console.log(`Inserted product: ${product.name}`);
    }

    console.log('✅ Successfully seeded database with sample products!');
  } catch (error) {
    console.error('❌ Error seeding database:', error);
  } finally {
    await client.end();
  }
}

seedProducts();
