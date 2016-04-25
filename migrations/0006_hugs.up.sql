create table pagerduty_oauth_responses (
    id serial primary key,
    customer_id UUID not null,
    data jsonb not null
);

