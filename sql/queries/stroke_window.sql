-- name: InsertStrokeWindow :exec
INSERT INTO stroke_window
(id, participation_id, stroke_amount, 
ud_mean, ud_std_dev, 
du1_mean, du1_std_dev, 
du2_mean, du2_std_dev, 
dd_mean, dd_std_dev, 
uu_mean, uu_std_dev,
feature_version, window_start_ms, window_end_ms,
ud_median, ud_mad,
du1_median, du1_mad,
du2_median, du2_mad,
dd_median, dd_mad,
uu_median, uu_mad,
long_pause_permille)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE id = VALUES(id);

-- name: SelectStrokeWindows :many
SELECT stroke_window.*
FROM stroke_window
WHERE participation_id = ?
ORDER BY created_at DESC;
