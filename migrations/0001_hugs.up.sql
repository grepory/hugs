create table notifications (
    id serial primary key,
    check_id varchar(255) not null,
    customer_id UUID not null,
    user_id int not null,
    type varchar(255) not null,
    value varchar(255) not null
);
