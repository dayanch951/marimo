-- Insert default admin user
-- Password: admin123 (hashed with bcrypt)
-- This should be changed immediately in production!
INSERT INTO users (email, name, password, role, created_at, updated_at)
VALUES (
    'admin@marimo.com',
    'Administrator',
    '$2a$10$YourBcryptHashedPasswordHere', -- Change this to actual bcrypt hash
    'admin',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
)
ON CONFLICT (email) DO NOTHING;
