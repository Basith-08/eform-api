INSERT INTO roles (code, name)
VALUES
    ('admin', 'Administrator'),
    ('user', 'Employee')
ON CONFLICT (code) DO UPDATE SET
    name = EXCLUDED.name,
    updated_at = NOW();

INSERT INTO users (role_id, email, password_hash, full_name, phone, ktp_number, status)
SELECT
    roles.id,
    'admin@tia.co.id',
    '$2a$10$8/pcgycrZ4JVhy/Oaa7Ym.T.Jcp.I5h1JOyo9bsUQvODc5.Hh44Y2',
    'System Administrator',
    '081234567890',
    '3173000000000001',
    'active'
FROM roles
WHERE roles.code = 'admin'
ON CONFLICT (email) DO UPDATE SET
    role_id = EXCLUDED.role_id,
    password_hash = EXCLUDED.password_hash,
    full_name = EXCLUDED.full_name,
    phone = EXCLUDED.phone,
    ktp_number = EXCLUDED.ktp_number,
    status = EXCLUDED.status,
    updated_at = NOW();
