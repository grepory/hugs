create type notification_type as enum ('web_hook', 'slack_bot', 'email');
alter table notifications drop constraint if exists notifications_type_fkey;
alter table notifications alter column type set data type notification_type using type::notification_type;
drop table notification_types;
