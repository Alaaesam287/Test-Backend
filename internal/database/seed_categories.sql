-- ===============================
-- CATEGORY DEFINITIONS
-- ===============================

INSERT INTO category_definition (name, parent_id) VALUES
-- Electronics
('Electronics', NULL),
('Phones', (SELECT category_id FROM category_definition WHERE name = 'Electronics')),
('Laptops', (SELECT category_id FROM category_definition WHERE name = 'Electronics')),
('Tablets', (SELECT category_id FROM category_definition WHERE name = 'Electronics')),
('Accessories', (SELECT category_id FROM category_definition WHERE name = 'Electronics')),

-- Fashion
('Fashion', NULL),
('Men Clothing', (SELECT category_id FROM category_definition WHERE name = 'Fashion')),
('Women Clothing', (SELECT category_id FROM category_definition WHERE name = 'Fashion')),
('Shoes', (SELECT category_id FROM category_definition WHERE name = 'Fashion')),

-- Home & Living
('Home & Living', NULL),
('Furniture', (SELECT category_id FROM category_definition WHERE name = 'Home & Living')),
('Kitchen', (SELECT category_id FROM category_definition WHERE name = 'Home & Living')),
('Home Appliances', (SELECT category_id FROM category_definition WHERE name = 'Home & Living')),

-- Beauty & Health
('Beauty & Health', NULL),
('Skincare', (SELECT category_id FROM category_definition WHERE name = 'Beauty & Health')),
('Haircare', (SELECT category_id FROM category_definition WHERE name = 'Beauty & Health')),

-- Sports
('Sports & Outdoors', NULL),
('Fitness Equipment', (SELECT category_id FROM category_definition WHERE name = 'Sports & Outdoors')),

-- Books
('Books', NULL)
ON CONFLICT (name) DO NOTHING;

-- ===============================
-- ATTRIBUTE DEFINITIONS
-- ===============================

INSERT INTO attribute_definition (name) VALUES
-- Common (almost everything)
('Color'),
('Weight'),
('Material'),
('Dimensions'),

-- Electronics
('RAM'),
('Storage'),
('CPU'),
('GPU'),
('Screen Size'),
('Battery Capacity'),
('Operating System'),
('Camera Resolution'),

-- Clothing
('Size'),
('Fit'),
('Gender'),
('Fabric'),
('Season'),

-- Shoes
('Shoe Size'),
('Heel Height'),

-- Home
('Power Consumption'),
('Voltage'),
('Capacity'),

-- Books
('Author'),
('Publisher'),
('Language'),
('ISBN'),
('Pages'),

-- Sports
('Sport Type'),
('Usage Level')
ON CONFLICT (name) DO NOTHING;

-- Phones
INSERT INTO category_attribute (category_id, attribute_id, is_required)
SELECT c.category_id, a.attribute_id,
       a.name IN (
         'RAM', 'Storage', 'CPU',
         'Screen Size', 'Battery Capacity',
         'Operating System'
       )
FROM category_definition c, attribute_definition a
WHERE c.name = 'Phones'
  AND a.name IN (
    'RAM', 'Storage', 'CPU',
    'Screen Size', 'Battery Capacity',
    'Operating System', 'Camera Resolution',
    'Color', 'Weight'
)
ON CONFLICT DO NOTHING;

-- Laptops
INSERT INTO category_attribute (category_id, attribute_id, is_required)
SELECT c.category_id, a.attribute_id,
       a.name IN (
         'RAM', 'Storage', 'CPU',
         'Screen Size', 'Operating System'
       )
FROM category_definition c, attribute_definition a
WHERE c.name = 'Laptops'
  AND a.name IN (
    'RAM', 'Storage', 'CPU', 'GPU',
    'Screen Size', 'Operating System',
    'Weight', 'Dimensions'
)
ON CONFLICT DO NOTHING;

-- Clothing
INSERT INTO category_attribute (category_id, attribute_id, is_required)
SELECT c.category_id, a.attribute_id,
       a.name IN (
         'Size', 'Color',
         'Fabric', 'Gender'
       )
FROM category_definition c, attribute_definition a
WHERE c.name IN ('Men Clothing', 'Women Clothing')
  AND a.name IN (
    'Size', 'Color',
    'Fabric', 'Fit',
    'Gender', 'Season'
)
ON CONFLICT DO NOTHING;

-- Shoes
INSERT INTO category_attribute (category_id, attribute_id, is_required)
SELECT c.category_id, a.attribute_id,
       a.name IN (
         'Shoe Size', 'Color', 'Gender'
       )
FROM category_definition c, attribute_definition a
WHERE c.name = 'Shoes'
  AND a.name IN (
    'Shoe Size', 'Color',
    'Material', 'Gender',
    'Heel Height'
)
ON CONFLICT DO NOTHING;

-- Home Appliances
INSERT INTO category_attribute (category_id, attribute_id, is_required)
SELECT c.category_id, a.attribute_id,
       a.name IN (
         'Power Consumption', 'Voltage'
       )
FROM category_definition c, attribute_definition a
WHERE c.name = 'Home Appliances'
  AND a.name IN (
    'Power Consumption',
    'Voltage', 'Weight', 'Dimensions'
)
ON CONFLICT DO NOTHING;

-- Books
INSERT INTO category_attribute (category_id, attribute_id, is_required)
SELECT c.category_id, a.attribute_id,
       a.name IN (
         'Author', 'Language', 'ISBN'
       )
FROM category_definition c, attribute_definition a
WHERE c.name = 'Books'
  AND a.name IN (
    'Author', 'Publisher',
    'Language', 'ISBN', 'Pages'
)
ON CONFLICT DO NOTHING;

