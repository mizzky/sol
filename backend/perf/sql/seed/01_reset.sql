\set ON_ERROR_STOP on

BEGIN;

\ir 00_guard.sql

TRUNCATE TABLE
    public.cart_items,
    public.order_items,
    public.payments,
    public.refresh_tokens,
    public.carts,
    public.orders,
    public.products,
    public.categories,
    public.users
RESTART IDENTITY;

COMMIT;