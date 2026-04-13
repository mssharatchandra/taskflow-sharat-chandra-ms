-- Seed user: test@example.com / password123
INSERT INTO users (id, name, email, password) VALUES
    ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'Test User', 'test@example.com',
     '$2a$12$i1UF0u4R7JSWIVX7rqKxH.m0ay2fennYhNgs0qGx4wadUI/YisL3.');

-- Seed project
INSERT INTO projects (id, name, description, owner_id) VALUES
    ('b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a22', 'Website Redesign', 'Q2 redesign of the marketing site',
     'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11');

-- Seed tasks with different statuses
INSERT INTO tasks (id, title, description, status, priority, project_id, assignee_id, creator_id, due_date) VALUES
    ('c2eebc99-9c0b-4ef8-bb6d-6bb9bd380a33', 'Design homepage mockups',
     'Create Figma mockups for the new homepage layout', 'done', 'high',
     'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a22', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
     'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '2026-04-20'),

    ('d3eebc99-9c0b-4ef8-bb6d-6bb9bd380a44', 'Implement responsive navigation',
     'Build the mobile-first nav component with hamburger menu', 'in_progress', 'medium',
     'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a22', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
     'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '2026-04-25'),

    ('e4eebc99-9c0b-4ef8-bb6d-6bb9bd380a55', 'Write API integration tests',
     'Add integration tests for all public API endpoints', 'todo', 'low',
     'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a22', NULL,
     'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '2026-05-01');
