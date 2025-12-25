-- Insert sample organization
INSERT INTO organizations (id, name, slug) VALUES
('11111111-1111-1111-1111-111111111111', 'Acme Corp', 'acme-corp');

-- Insert sample users (password: "password123" for all)
-- Password hash generated using bcrypt for "password123"
INSERT INTO users (id, org_id, email, password_hash, first_name, last_name, role) VALUES
('22222222-2222-2222-2222-222222222222', '11111111-1111-1111-1111-111111111111', 'admin@acme.com', '$2a$10$YFvEjPKWqkCBvVY3VZ3yxOvqJSd5RnN/QVYoXGJmYnVQFpqZJxP2m', 'Admin', 'User', 'admin'),
('33333333-3333-3333-3333-333333333333', '11111111-1111-1111-1111-111111111111', 'manager@acme.com', '$2a$10$YFvEjPKWqkCBvVY3VZ3yxOvqJSd5RnN/QVYoXGJmYnVQFpqZJxP2m', 'Manager', 'User', 'manager'),
('44444444-4444-4444-4444-444444444444', '11111111-1111-1111-1111-111111111111', 'member@acme.com', '$2a$10$YFvEjPKWqkCBvVY3VZ3yxOvqJSd5RnN/QVYoXGJmYnVQFpqZJxP2m', 'Member', 'User', 'member');

-- Insert sample tasks
INSERT INTO tasks (org_id, title, description, status, priority, assigned_to, created_by) VALUES
('11111111-1111-1111-1111-111111111111', 'Setup CI/CD pipeline', 'Configure GitHub Actions for automated deployments', 'todo', 'high', '33333333-3333-3333-3333-333333333333', '22222222-2222-2222-2222-222222222222'),
('11111111-1111-1111-1111-111111111111', 'Implement user analytics', 'Add tracking for user behavior', 'in_progress', 'medium', '44444444-4444-4444-4444-444444444444', '22222222-2222-2222-2222-222222222222');

-- Insert sample issues
INSERT INTO issues (org_id, title, description, severity, status, reported_by) VALUES
('11111111-1111-1111-1111-111111111111', 'Login page not responsive', 'The login page does not display correctly on mobile devices', 'high', 'open', '44444444-4444-4444-4444-444444444444'),
('11111111-1111-1111-1111-111111111111', 'API timeout on large requests', 'Getting timeout errors when fetching large datasets', 'critical', 'in_progress', '33333333-3333-3333-3333-333333333333');
