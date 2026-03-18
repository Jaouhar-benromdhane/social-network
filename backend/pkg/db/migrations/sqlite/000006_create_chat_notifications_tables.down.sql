DROP INDEX IF EXISTS idx_notifications_user_id;
DROP INDEX IF EXISTS idx_group_messages_group_id;
DROP INDEX IF EXISTS idx_private_messages_recipient_id;

DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS group_messages;
DROP TABLE IF EXISTS private_messages;
