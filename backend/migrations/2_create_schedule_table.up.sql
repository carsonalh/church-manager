CREATE TYPE repeat_interval AS ENUM ('Day', 'Week', 'Month', 'Year');

CREATE TYPE day_of_week AS ENUM ('Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday', 'Sunday');

CREATE TABLE schedule (
    id BIGSERIAL NOT NULL,
    begin_date TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    end_date TIMESTAMP WITHOUT TIME ZONE,
    -- if begin_date = end_date, then this is a one-off event
    repeat_interval_count INTEGER CHECK (repeat_interval_count >= 0),
    repeat_interval_unit repeat_interval,
    -- the previous two fields must either both be null or nonnull
    CHECK ((repeat_interval_count IS NULL) = (repeat_interval_unit IS NULL)),
    repeat_nth_day_of_month_day day_of_week,
    repeat_nth_day_of_month_n INTEGER,
    -- the previous two fields must either both be null or nonnull
    CHECK ((repeat_nth_day_of_month_day IS NULL) = (repeat_nth_day_of_month_n IS NULL)),
    -- out of the previous two pairs of fields, one must be null and the other nonnull
    CHECK ((repeat_interval_count IS NULL) <> (repeat_nth_day_of_month_day IS NULL))
);

COMMENT ON COLUMN schedule.begin_date IS 'The timestamp of the very first service in this schedule';
COMMENT ON COLUMN schedule.end_date IS 'The timestap after which no more services are to occur';
