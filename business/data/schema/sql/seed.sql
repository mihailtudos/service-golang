INSERT INTO users (user_id, name, email, roles, password_hash, date_created, date_updated)
VALUES
    ('5cf37266-3473-4006-984f-9325122678b7', 'Admin Gopher', 'admin@example.com', ARRAY['ADMIN', 'USER'], '$2a$10$ZNWmz0U/6FK.OUp4d/dWze4kS.U7CHrJaeIMHXx1nmBz3e7xhkT82', NOW(), NOW()),
    ('45b5fbd3-755f-4379-8f07-a58d4a30fa2f', 'User Gopher', 'user@example.com', ARRAY['USER'], '$2a$10$ZNWmz0U/6FK.OUp4d/dWze4kS.U7CHrJaeIMHXx1nmBz3e7xhkT82', NOW(), NOW()),
    ('28e57012-fbbb-4f67-854f-d15f8b9462c7', 'Alice Smith', 'alice@example.com', ARRAY['USER'], '$2a$10$Abc123ExampleHash1', NOW(), NOW()),
    ('df4bfa2c-5aa6-4016-9d8a-08e099441231', 'Bob Johnson', 'bob@example.com', ARRAY['USER'], '$2a$10$Def456ExampleHash2', NOW(), NOW()),
    ('763e2c62-5a9f-4d63-91d2-2c55b115ee41', 'Clara Gomez', 'clara@example.com', ARRAY['USER'], '$2a$10$Ghi789ExampleHash3', NOW(), NOW()),
    ('fa2a7f76-3a79-499f-bc04-9bbd71a4442a', 'Daniel Kim', 'daniel@example.com', ARRAY['USER'], '$2a$10$Jkl012ExampleHash4', NOW(), NOW()),
    ('e0b0fa31-63c1-49c7-afe3-dcfdc0be2682', 'Ella Zhang', 'ella@example.com', ARRAY['USER'], '$2a$10$Mno345ExampleHash5', NOW(), NOW()),
    ('07e6f012-2550-4f7a-9a0c-1d39e938d018', 'Frank Wu', 'frank@example.com', ARRAY['USER'], '$2a$10$Pqr678ExampleHash6', NOW(), NOW()),
    ('2f8c849e-06bb-4bc3-a805-fc6554b6e567', 'Grace Lee', 'grace@example.com', ARRAY['USER'], '$2a$10$Stu901ExampleHash7', NOW(), NOW()),
    ('82ac01b6-e38d-4a87-9004-94c6c30a3c56', 'Henry Davis', 'henry@example.com', ARRAY['USER'], '$2a$10$Vwx234ExampleHash8', NOW(), NOW())
    ON CONFLICT DO NOTHING;


INSERT INTO products (product_id, user_id, name, cost, quantity, date_created, date_updated)
VALUES
    ('a2b0639f-2cc6-44b8-b97b-15d69dbb511e', '45b5fbd3-755f-4379-8f07-a58d4a30fa2f', 'Comic Books', 50, 42, NOW(), NOW()),
    ('72f8b983-3eb4-48db-9ed0-e45cc6bd716b', '45b5fbd3-755f-4379-8f07-a58d4a30fa2f', 'McDonalds Toys', 75, 120, NOW(), NOW()),
    ('ccfb2ff4-3a1b-4373-9991-7dfb29471f99', '28e57012-fbbb-4f67-854f-d15f8b9462c7', 'Board Game', 100, 30, NOW(), NOW()),
    ('d901d770-b8ff-4f8e-9873-61fc7b030ad4', 'df4bfa2c-5aa6-4016-9d8a-08e099441231', 'Headphones', 200, 15, NOW(), NOW()),
    ('e63c8d58-6fc2-4968-b30c-58eab1e8bb17', '763e2c62-5a9f-4d63-91d2-2c55b115ee41', 'USB Drives', 20, 300, NOW(), NOW()),
    ('f8046b7e-9d37-407c-b4db-8634f52a8abf', 'fa2a7f76-3a79-499f-bc04-9bbd71a4442a', 'T-Shirts', 35, 80, NOW(), NOW()),
    ('49d1fc0c-d1e1-4dc6-9c94-d85a5b68704c', 'e0b0fa31-63c1-49c7-afe3-dcfdc0be2682', 'Smart Watch', 250, 10, NOW(), NOW()),
    ('60e8a97f-1720-456c-8c2e-f0b318c0277d', '07e6f012-2550-4f7a-9a0c-1d39e938d018', 'Sunglasses', 80, 60, NOW(), NOW()),
    ('76c2c03c-8b4d-4b30-865a-232c97358e11', '2f8c849e-06bb-4bc3-a805-fc6554b6e567', 'Notebooks', 15, 100, NOW(), NOW()),
    ('981c3f24-82e4-470f-bf1c-2e5de7e47c4c', '82ac01b6-e38d-4a87-9004-94c6c30a3c56', 'Bluetooth Speaker', 150, 25, NOW(), NOW())
    ON CONFLICT DO NOTHING;



INSERT INTO sales (sale_id, user_id, product_id, quantity, paid, date_created)
VALUES
    ('98b6d4b8-f04b-4c79-8c2e-a0aef46854b7', '45b5fbd3-755f-4379-8f07-a58d4a30fa2f', 'a2b0639f-2cc6-44b8-b97b-15d69dbb511e', 2, 100, NOW()),
    ('85f6fb09-eb05-4874-ae39-82d1a30fe0d7', '45b5fbd3-755f-4379-8f07-a58d4a30fa2f', 'a2b0639f-2cc6-44b8-b97b-15d69dbb511e', 5, 250, NOW()),
    ('a235be9e-ab5d-44e6-a987-facc749264c7', '45b5fbd3-755f-4379-8f07-a58d4a30fa2f', '72f8b983-3eb4-48db-9ed0-e45cc6bd716b', 3, 225, NOW()),
    ('b01cc2d1-25e7-4e83-9946-86e17d055bdf', '28e57012-fbbb-4f67-854f-d15f8b9462c7', 'ccfb2ff4-3a1b-4373-9991-7dfb29471f99', 1, 100, NOW()),
    ('c709bcfc-d0a4-49b7-aeb3-43de9c5ae749', 'df4bfa2c-5aa6-4016-9d8a-08e099441231', 'd901d770-b8ff-4f8e-9873-61fc7b030ad4', 1, 200, NOW()),
    ('d174c59c-88ed-4459-91e2-4ac7c71180b5', '763e2c62-5a9f-4d63-91d2-2c55b115ee41', 'e63c8d58-6fc2-4968-b30c-58eab1e8bb17', 10, 200, NOW()),
    ('e34bcf3c-556f-41f1-89b3-06f2e7c0c2ab', 'fa2a7f76-3a79-499f-bc04-9bbd71a4442a', 'f8046b7e-9d37-407c-b4db-8634f52a8abf', 3, 105, NOW()),
    ('f4426ed2-d1c2-4d47-9d46-d76bd42f4f86', 'e0b0fa31-63c1-49c7-afe3-dcfdc0be2682', '49d1fc0c-d1e1-4dc6-9c94-d85a5b68704c', 2, 500, NOW()),
    ('05e3f5c2-47a6-4e46-b7f4-9994f82347d9', '07e6f012-2550-4f7a-9a0c-1d39e938d018', '60e8a97f-1720-456c-8c2e-f0b318c0277d', 4, 320, NOW()),
    ('161b0aef-0a69-405c-83f1-391a318d4c37', '82ac01b6-e38d-4a87-9004-94c6c30a3c56', '981c3f24-82e4-470f-bf1c-2e5de7e47c4c', 2, 300, NOW())
    ON CONFLICT DO NOTHING;
