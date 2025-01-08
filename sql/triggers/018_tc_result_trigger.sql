-- +goose Up
-- +goose StatementBegin
CREATE TRIGGER update_submission_test_cases AFTER UPDATE ON test_case_result
FOR EACH ROW
BEGIN
    IF NEW.status = 'Accepted' THEN
        UPDATE submission
        SET accepted_test_cases = accepted_test_cases + 1
        WHERE id = NEW.submission_id;
    END IF; -- IF END
END; -- FUNCTION END
-- +goose StatementEnd

-- +goose Down
DROP TRIGGER IF EXISTS update_submission_test_cases;
