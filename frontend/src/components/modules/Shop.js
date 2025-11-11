import React, { useEffect, useState } from 'react';
import { api } from '../../services/api';

const Shop = () => {
  const [products, setProducts] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchProducts();
  }, []);

  const fetchProducts = async () => {
    try {
      const response = await api.get('/shop/products');
      setProducts(response.data.products || []);
    } catch (error) {
      console.error('Failed to fetch products:', error);
    } finally {
      setLoading(false);
    }
  };

  if (loading) return <div className="loading">Loading...</div>;

  return (
    <div className="module-page">
      <h2>🛍️ Online Shop</h2>
      <div className="products-grid">
        {products.map((product) => (
          <div key={product.id} className="product-card">
            <div className="product-image">
              {product.image_url ? (
                <img src={product.image_url} alt={product.name} />
              ) : (
                <div className="placeholder-image">📦</div>
              )}
            </div>
            <h3>{product.name}</h3>
            <p className="product-description">{product.description}</p>
            <div className="product-footer">
              <span className="product-price">${product.price}</span>
              <span className="product-stock">Stock: {product.stock}</span>
            </div>
            <button className="btn btn-sm">Add to Cart</button>
          </div>
        ))}
      </div>
    </div>
  );
};

export default Shop;
