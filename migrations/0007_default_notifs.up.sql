create table default_notifications (
  id serial primary key,
  customer_id UUID not null,
  type notification_type not null,
  value varchar(255) not null
);

create index idx_default_notifications_customer on default_notifications(customer_id);