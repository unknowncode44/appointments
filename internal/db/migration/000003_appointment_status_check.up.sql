ALTER TABLE appointments
    ADD CONSTRAINT appointments_status_check
    CHECK (status IN ('confirmed', 'cancelled', 'completed', 'no_show'));
