-- # DROP TABLE auth.account_credentials;
-- # DROP TABLE auth.provider;
-- # DROP TABLE auth.account;

-- Accounts
INSERT INTO auth.account (name, email, id)
VALUES ('user_test_1', 'user_test_1@test.com', '06e1b677-a4fe-42cf-8afd-ceec867d1fa5'),
       ('user_test_2', 'user_test_2@test.com', '93a6437f-af81-47c1-9aa4-5caedc0a6869'),
       ('user_test_3', 'user_test_3@test.com', '9e0f95a2-6535-4679-9db6-c93a823702b1')
;

SELECT *
FROM auth.account;

-- Providers
INSERT INTO auth.provider (id, name)
VALUES ('jwt'),
       ('email')
;
SELECT *
FROM auth.provider;

INSERT INTO auth.account_credentials (account_id, provider_id, secret, verified, verified_at)
VALUES
    -- user_test_1
    ('06e1b677-a4fe-42cf-8afd-ceec867d1fa5',
     1,
     encode(digest('pass', 'sha256'), 'hex'),
     true, CURRENT_TIMESTAMP),
    ('06e1b677-a4fe-42cf-8afd-ceec867d1fa5',
     2,
     encode(digest('pass', 'sha256'), 'hex'),
     true, CURRENT_TIMESTAMP),

    -- user_test_2
    ('93a6437f-af81-47c1-9aa4-5caedc0a6869',
     1,
     encode(digest('pass', 'sha256'), 'hex'),
     true, CURRENT_TIMESTAMP),
    ('93a6437f-af81-47c1-9aa4-5caedc0a6869',
     2,
     encode(digest('pass', 'sha256'), 'hex'),
     false, NULL),

    -- user_test_3
    ('9e0f95a2-6535-4679-9db6-c93a823702b1',
     1,
     encode(digest('pass', 'sha256'), 'hex'),
     true, CURRENT_TIMESTAMP),
    ('9e0f95a2-6535-4679-9db6-c93a823702b1',
     2,
     encode(digest('pass', 'sha256'), 'hex'),
     false, NULL)
;


SELECT *
FROM auth.account_credentials;

SELECT * FROM auth.account WHERE id = '06e1b677-a4fe-42cf-8afd-ceec867d1fa5';
SELECT * FROM auth.account_credentials WHERE account_id = '06e1b677-a4fe-42cf-8afd-ceec867d1fa5';

SELECT account.*,  provider.name, credentials.*
FROM auth.account account
         JOIN auth.account_credentials credentials ON account.id = credentials.account_id
         JOIN auth.provider provider ON credentials.provider_id = provider.id
WHERE account.id = '06e1b677-a4fe-42cf-8afd-ceec867d1fa5'
  AND provider.disabled IS FALSE
  AND credentials.disabled IS FALSE
;

