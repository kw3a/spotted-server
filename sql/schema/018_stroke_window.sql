-- +goose Up
CREATE TABLE stroke_window (
	id CHAR(36) PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    stroke_amount INT NOT NULL,
    ud_mean INT NOT NULL,
    ud_std_dev INT NOT NULL,
    du1_mean INT NOT NULL,
    du1_std_dev INT NOT NULL,
    du2_mean INT NOT NULL,
    du2_std_dev INT NOT NULL,
    dd_mean INT NOT NULL,
    dd_std_dev INT NOT NULL,
    uu_mean INT NOT NULL,
    uu_std_dev INT NOT NULL,
	participation_id CHAR(36) NOT NULL,
    FOREIGN KEY (participation_id) REFERENCES participation(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE stroke_window;