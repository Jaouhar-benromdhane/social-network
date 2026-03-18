DROP INDEX IF EXISTS idx_events_group_id;
DROP INDEX IF EXISTS idx_group_posts_group_id;
DROP INDEX IF EXISTS idx_group_members_user_id;
DROP INDEX IF EXISTS idx_groups_creator_id;

DROP TABLE IF EXISTS event_votes;
DROP TABLE IF EXISTS event_options;
DROP TABLE IF EXISTS events;
DROP TABLE IF EXISTS group_comments;
DROP TABLE IF EXISTS group_posts;
DROP TABLE IF EXISTS group_join_requests;
DROP TABLE IF EXISTS group_invites;
DROP TABLE IF EXISTS group_members;
DROP TABLE IF EXISTS groups;
