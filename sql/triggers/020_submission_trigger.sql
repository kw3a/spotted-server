-- +goose Up
-- +goose StatementBegin
CREATE TRIGGER insert_test_case_result_submission 
AFTER INSERT ON submission
FOR EACH ROW
BEGIN
    DECLARE test_case_id CHAR(36);
    DECLARE cursor_finished BOOLEAN DEFAULT FALSE;

    DECLARE cur_test_cases CURSOR FOR 
        SELECT id
        FROM test_case
        WHERE problem_id = NEW.problem_id;

    DECLARE CONTINUE HANDLER FOR NOT FOUND 
        SET cursor_finished = TRUE;

    OPEN cur_test_cases;

    loop_test_cases: LOOP
        FETCH cur_test_cases INTO test_case_id;
        IF cursor_finished THEN
            LEAVE loop_test_cases;
        END IF; -- IF END

        INSERT INTO test_case_result (submission_id, test_case_id)
        VALUES (NEW.id, test_case_id);
    END LOOP loop_test_cases; -- LOOP END

    CLOSE cur_test_cases; -- CURSOR END
END; -- FUNCTION END

-- +goose StatementEnd
-- +goose Down
DROP TRIGGER IF EXISTS insert_test_case_result_submission;
