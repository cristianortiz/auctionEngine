-- Script para poblar la base de datos con datos de prueba

-- Deshabilitar temporalmente los triggers para inserciones masivas si es necesario,
-- aunque para este caso simple no es estrictamente necesario.
-- SET session_replication_role = replica;

-- Insertar usuarios de prueba
-- Los IDs se generarán automáticamente por la base de datos (debido a DEFAULT gen_random_uuid() en la migración)
INSERT INTO users (username, email, password_hash) VALUES
('Jon', 'jon1@example.com', 'hashed_password_1'),
('Mary', 'mary2@example.com', 'hashed_password_2'),
('Paul', 'paul3@example.com', 'hashed_password_3');
-- Nota: Si ejecutas esto varias veces, creará nuevos usuarios cada vez.


-- Insertar lotes de subasta de prueba
-- Los IDs se generarán automáticamente por la base de datos (debido a DEFAULT gen_random_uuid() en la migración)
INSERT INTO auction_lots (title, description, initial_price, current_price, end_time, state, last_bid_time, time_extension) VALUES
('Test Car Auction', 'A nice car for testing.', 5000.00, 5000.00, NOW() + INTERVAL '10 minutes', 'active', NULL, INTERVAL '30 seconds'),
('Test Bike Auction', 'A fast bike for testing.', 1000.00, 1000.00, NOW() + INTERVAL '15 minutes', 'active', NULL, INTERVAL '45 seconds');
-- Nota: Si ejecutas esto varias veces, creará nuevos lotes cada vez.

-- Re-habilitar triggers si los deshabilitaste
-- SET session_replication_role = origin;

-- Mensaje de éxito (opcional, depende del cliente psql)
-- SELECT 'Database seeded successfully!' AS status;