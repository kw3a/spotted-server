-- +goose Up
-- +goose StatementBegin
CREATE TRIGGER update_best_try_submission AFTER UPDATE ON submission
FOR EACH ROW
BEGIN
    IF NEW.accepted_test_cases > best_try.accepted_test_cases  THEN
        UPDATE best_try
        SET accepted_test_cases = NEW.accepted_test_cases 
        AND submission_id = NEW.id
        WHERE participation_id = NEW.participation_id 
        AND problem_id = NEW.problem_id;
    END IF; -- IF END
END; -- FUNCTION END
-- +goose StatementEnd

-- +goose Down
DROP TRIGGER IF EXISTS update_best_try_submission;