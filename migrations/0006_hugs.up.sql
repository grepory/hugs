create table pagerduty_oauth_responses (
    id serial primary key,
    customer_id UUId not null,
    data jsonb not null
);

