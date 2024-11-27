-- +goose Up
-- +goose StatementBegin
CREATE TRIGGER before_participation_insert
BEFORE INSERT ON participation
FOR EACH ROW
BEGIN
    DECLARE quiz_duration INT;
    SELECT duration INTO quiz_duration FROM quiz WHERE id = NEW.quiz_id;
    SET NEW.expires_at = NOW() + INTERVAL quiz_duration MINUTE;
END -- FUNCTION END
-- +goose StatementEnd
-- +goose Down
DROP TRIGGER IF EXISTS before_participation_insert;
