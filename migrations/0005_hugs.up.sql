create type notification_type_new as enum ('webhook', 'slack_bot', 'email', 'pagerduty');
alter table notifications add column type_new notification_type_new;
update notifications set type_new = 'webhook' where type = 'web_hook';
update notifications set type_new = type::text::notification_type_new;
alter table notifications drop column type;
drop type notification_type;
alter type notification_type_new rename to notification_type;
alter table notifications rename column type_new to type;
alter table notifications alter column type set data type notification_type using type::text::notification_type;
