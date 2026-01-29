-- Update OAuth Client Redirect URIs for Railzway Console
UPDATE oauth_clients
SET redirect_uris = ARRAY['https://cloud.railzway.com/auth/callback', 'http://localhost:3000/callback']
WHERE id = 1200;
