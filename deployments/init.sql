create table tests (
                       id uuid primary key,
                       url text not null,
                       rps integer default 100,
                       duration_seconds integer default 60,
                       status varchar not null check(status in ('running', 'finished', 'failed', 'cancelled')),
                       created_at timestamp default current_timestamp
);

create table metrics (
                         id uuid primary key default gen_random_uuid(),
                         test_id uuid not null references tests(id) on delete cascade,
                         bucket_time timestamptz not null,
                         requests_count integer not null check(requests_count >= 0),
                         success_count integer not null check(success_count >= 0),
                         error_count integer not null check(error_count >= 0),
                         avg_latency_ms numeric not null,
                         p95_latency_ms numeric not null
);

create index idx_metrics_test_id on metrics(test_id);