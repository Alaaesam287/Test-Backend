-- ===============================
-- STORES & STORE OWNERS
-- ===============================

CREATE TABLE store_owner (
  store_owner_id  BIGSERIAL PRIMARY KEY,
  name            VARCHAR(255) NOT NULL,
  email           VARCHAR(255) UNIQUE NOT NULL,
  password_hash   TEXT NOT NULL,
  created_at      TIMESTAMP DEFAULT NOW()
);

CREATE TABLE store (
  store_id        BIGSERIAL PRIMARY KEY,
  store_owner_id  BIGINT UNIQUE NOT NULL REFERENCES store_owner(store_owner_id),
  name            VARCHAR(255) NOT NULL,
  domain          VARCHAR(255) UNIQUE,
  currency        VARCHAR(10) DEFAULT 'USD',
  timezone        VARCHAR(100) DEFAULT 'UTC',
  created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  updated_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ===============================
-- PRODUCT CATALOG
-- ===============================

CREATE TABLE product_category (
  category_id     BIGSERIAL PRIMARY KEY,
  store_id        BIGINT NOT NULL REFERENCES store(store_id),
  name            VARCHAR(255) NOT NULL,
  parent_id       BIGINT REFERENCES product_category(category_id),
  created_at      TIMESTAMP DEFAULT NOW(),
  UNIQUE (store_id, name)
);

CREATE TABLE product (
  product_id      BIGSERIAL PRIMARY KEY,
  store_id        BIGINT NOT NULL REFERENCES store(store_id),
  category_id     BIGINT NOT NULL REFERENCES product_category(category_id),
  name            VARCHAR(255) NOT NULL,
  slug            VARCHAR(255),
  description     TEXT,
  brand           VARCHAR(255),
  created_at      TIMESTAMP DEFAULT NOW(),
  updated_at      TIMESTAMP DEFAULT NOW(),
  in_stock        BOOLEAN DEFAULT TRUE NOT NULL,
  deleted_at      TIMESTAMP NULL,
  default_variant_id BIGINT REFERENCES product_variant(variant_id),
  CONSTRAINT unique_slug_per_store UNIQUE (store_id, slug)
);

CREATE TABLE product_image (
  image_id        BIGSERIAL PRIMARY KEY,
  product_id      BIGINT NOT NULL REFERENCES product(product_id),
  image_url       VARCHAR(500) NOT NULL,
  is_primary      BOOLEAN DEFAULT FALSE NOT NULL
);

CREATE TABLE attribute_definition (
  attribute_id    BIGSERIAL PRIMARY KEY,
  store_id        BIGINT NOT NULL REFERENCES store(store_id),
  name            VARCHAR(100) NOT NULL,
  data_type       VARCHAR(50) NOT NULL CHECK (data_type IN ('string', 'integer', 'decimal', 'boolean')),
  category_id     BIGINT REFERENCES product_category(category_id),
  UNIQUE(category_id, name)
);

CREATE TABLE product_attribute_value (
  product_id      BIGINT NOT NULL REFERENCES product(product_id),
  attribute_id    BIGINT NOT NULL REFERENCES attribute_definition(attribute_id),
  value_text      TEXT,
  value_number    DECIMAL(15,4),
  value_boolean   BOOLEAN,
  PRIMARY KEY (product_id, attribute_id)
);

CREATE TABLE option_type (
  option_type_id  BIGSERIAL PRIMARY KEY,
  store_id        BIGINT NOT NULL REFERENCES store(store_id),
  name            VARCHAR(100) NOT NULL
);

CREATE TABLE option_value (
  option_value_id BIGSERIAL PRIMARY KEY,
  option_type_id  BIGINT NOT NULL REFERENCES option_type(option_type_id),
  value           VARCHAR(100) NOT NULL
);

CREATE TABLE product_variant (
  variant_id      BIGSERIAL PRIMARY KEY,
  product_id      BIGINT NOT NULL REFERENCES product(product_id),
  sku             VARCHAR(100) NOT NULL,
  price           DECIMAL(10,2) NOT NULL,
  stock_quantity  INT DEFAULT 0 NOT NULL,
  image_url       VARCHAR(500) NOT NULL,
  created_at      TIMESTAMP DEFAULT NOW(),
  updated_at      TIMESTAMP DEFAULT NOW(),
  deleted_at      TIMESTAMP NULL,
  UNIQUE (store_id, sku)
);

CREATE TABLE variant_option (
  variant_id      BIGINT NOT NULL REFERENCES product_variant(variant_id),
  option_value_id BIGINT NOT NULL REFERENCES option_value(option_value_id),
  PRIMARY KEY (variant_id, option_value_id)
);

-- ===============================
-- CUSTOMERS
-- ===============================

CREATE TABLE customer (
  customer_id     BIGSERIAL PRIMARY KEY,
  store_id        BIGINT NOT NULL REFERENCES store(store_id),
  name            VARCHAR(255),
  email           VARCHAR(255),
  phone           VARCHAR(50),
  address         JSONB,
  created_at      TIMESTAMP DEFAULT NOW(),
  UNIQUE (store_id, email)
);

-- ===============================
-- SHOPPING CART & ORDERS
-- ===============================

CREATE TABLE cart (
  cart_id         BIGSERIAL PRIMARY KEY,
  customer_id     BIGINT REFERENCES customer(customer_id),
  store_id        BIGINT NOT NULL REFERENCES store(store_id),
  created_at      TIMESTAMP DEFAULT NOW(),
  updated_at      TIMESTAMP DEFAULT NOW()
);

CREATE TABLE cart_item (
  cart_item_id    BIGSERIAL PRIMARY KEY,
  cart_id         BIGINT NOT NULL REFERENCES cart(cart_id),
  variant_id      BIGINT NOT NULL REFERENCES product_variant(variant_id),
  quantity        INT NOT NULL CHECK (quantity > 0),
  unit_price      DECIMAL(10,2),
  created_at      TIMESTAMP DEFAULT NOW()
);

CREATE TABLE customer_order (
  order_id        BIGSERIAL PRIMARY KEY,
  store_id        BIGINT NOT NULL REFERENCES store(store_id),
  customer_id     BIGINT REFERENCES customer(customer_id),
  session_id      UUID NOT NULL REFERENCES visitor_session(session_id),
  total_amount    DECIMAL(10,2),
  status          VARCHAR(50) DEFAULT 'pending',
  created_at      TIMESTAMP DEFAULT NOW(),
  updated_at      TIMESTAMP DEFAULT NOW()
);

CREATE TABLE order_item (
  order_item_id   BIGSERIAL PRIMARY KEY,
  order_id        BIGINT NOT NULL REFERENCES customer_order(order_id),
  variant_id      BIGINT REFERENCES product_variant(variant_id),
  quantity        INT NOT NULL CHECK (quantity > 0),
  unit_price      DECIMAL(10,2),
  subtotal        DECIMAL(10,2)
);

CREATE TABLE payment (
  payment_id      BIGSERIAL PRIMARY KEY,
  order_id        BIGINT NOT NULL REFERENCES customer_order(order_id),
  method          VARCHAR(50),
  amount          DECIMAL(10,2),
  status          VARCHAR(50) DEFAULT 'pending',
  transaction_ref VARCHAR(255),
  created_at      TIMESTAMP DEFAULT NOW()
);

CREATE TABLE shipment (
  shipment_id     BIGSERIAL PRIMARY KEY,
  order_id        BIGINT NOT NULL REFERENCES customer_order(order_id),
  tracking_number VARCHAR(255),
  carrier         VARCHAR(255),
  shipped_at      TIMESTAMP,
  delivered_at    TIMESTAMP CHECK (delivered_at >= shipped_at),
  status          VARCHAR(50) DEFAULT 'pending'
);

CREATE TABLE visitor_session (
  session_id     UUID PRIMARY KEY,
  store_id       BIGINT NOT NULL REFERENCES store(store_id),
  customer_id    BIGINT REFERENCES customer(customer_id), -- nullable for guests
  ip_address     INET,
  user_agent     TEXT,
  first_seen_at  TIMESTAMP DEFAULT NOW(),
  last_seen_at   TIMESTAMP DEFAULT NOW(),
  is_returning   BOOLEAN DEFAULT FALSE
);

CREATE TABLE product_view (
  product_view_id BIGSERIAL PRIMARY KEY,
  product_id      BIGINT NOT NULL REFERENCES product(product_id),
  store_id        BIGINT NOT NULL REFERENCES store(store_id),
  session_id      UUID NOT NULL REFERENCES visitor_session(session_id),
  viewed_at       TIMESTAMP DEFAULT NOW()
);

CREATE TABLE cart_event (
  cart_event_id BIGSERIAL PRIMARY KEY,
  session_id    UUID NOT NULL REFERENCES visitor_session(session_id),
  product_id    BIGINT NOT NULL REFERENCES product(product_id),
  variant_id    BIGINT REFERENCES product_variant(variant_id),
  event_type    VARCHAR(20) CHECK (event_type IN ('add', 'remove')),
  created_at    TIMESTAMP DEFAULT NOW()
);

