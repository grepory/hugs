create table customers (
    id UUID primary key
);

create table notification_types (
    id serial primary key,
    type varchar(255) unique not null
);

create table notifications (
    id serial primary key,
    check_id varchar(255) not null,
    customer_id UUID not null,
    user_id int not null,
    type varchar(255) references notification_types(type),
    value varchar(255) not null
);

create table slack_oath_response (
    id serial primary key,
    customer UUID references customers(id) on delete cascade,
    data jsonb not null
);

