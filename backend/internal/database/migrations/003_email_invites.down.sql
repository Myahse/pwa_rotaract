DELETE FROM club_role_permissions
WHERE permission_id IN (SELECT id FROM permissions WHERE key = 'club.invites.send');

DELETE FROM permissions WHERE key = 'club.invites.send';

DROP TABLE IF EXISTS email_invites;
