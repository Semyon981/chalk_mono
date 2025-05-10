package repo

import (
	"chalk/internal/repo/models"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type CoursesRepo interface {
}

func NewCoursesRepo(db *pgx.Conn) CoursesRepo {
	return &coursesRepo{db: db}
}

type coursesRepo struct {
	db *pgx.Conn
}

/* ===================== Courses ===================== */

type CreateCourseParams struct {
	AccountID int64
	Name      string
}

func (r *coursesRepo) CreateCourse(ctx context.Context, params CreateCourseParams) (int64, error) {
	const query = `INSERT INTO courses (account_id, name) VALUES ($1, $2) RETURNING id`
	var id int64
	err := r.db.QueryRow(ctx, query, params.AccountID, params.Name).Scan(&id)
	if err != nil {
		var pqErr *pgconn.PgError
		if errors.As(err, &pqErr) {
			if pqErr.Code == "23503" {
				return 0, ErrAccountNotFound
			}
		}
		return 0, fmt.Errorf("insert course: %v", err)
	}
	return id, nil
}

type UpdateCourseParams struct {
	CourseID int64
	Name     string
}

func (r *coursesRepo) UpdateCourse(ctx context.Context, params UpdateCourseParams) error {
	const query = `UPDATE courses SET (name) = ($1) WHERE id = $2`
	cmdTag, err := r.db.Exec(ctx, query, params.Name, params.CourseID)
	if err != nil {
		return fmt.Errorf("update course: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return ErrCourseNotFound
	}
	return nil
}

type RemoveCourseParams struct {
	CourseID int64
}

func (r *coursesRepo) RemoveCourse(ctx context.Context, params RemoveCourseParams) error {
	const query = `DELETE FROM courses WHERE id = $1`
	cmdTag, err := r.db.Exec(ctx, query, params.CourseID)
	if err != nil {
		return fmt.Errorf("delete course: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return ErrCourseNotFound
	}
	return nil
}

func (r *coursesRepo) GetCoursesByAccountID(ctx context.Context, accountID int64) ([]*models.Course, error) {
	const checkAccountQuery = `SELECT EXISTS(SELECT 1 FROM accounts WHERE id = $1)`
	var accountExists bool
	err := r.db.QueryRow(ctx, checkAccountQuery, accountID).Scan(&accountExists)
	if err != nil {
		return nil, fmt.Errorf("check lesson existence: %w", err)
	}
	if !accountExists {
		return nil, ErrAccountNotFound
	}

	const query = `SELECT id, account_id, name FROM courses WHERE account_id = $1`
	rows, err := r.db.Query(ctx, query, accountID)
	if err != nil {
		return nil, fmt.Errorf("query courses: %w", err)
	}
	defer rows.Close()

	var courses []*models.Course
	for rows.Next() {
		var c models.Course
		if err := rows.Scan(&c.ID, &c.AccountID, &c.Name); err != nil {
			return nil, fmt.Errorf("scan course: %w", err)
		}
		courses = append(courses, &c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return courses, nil
}

func (r *coursesRepo) GetCourseByID(ctx context.Context, courseID int64) (*models.Course, error) {
	const query = `SELECT id, account_id, name FROM courses WHERE id = $1`
	var course models.Course
	err := r.db.QueryRow(ctx, query, courseID).Scan(&course.ID, &course.AccountID, &course.Name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCourseNotFound
		}
		return nil, fmt.Errorf("get course by id: %w", err)
	}
	return &course, nil
}

/* ===================== Course participants ===================== */

type EnrollUserParams struct {
	UserID, CourseID int64
}

// EnrollUser добавляет пользователя в участники курса
func (r *coursesRepo) EnrollUser(ctx context.Context, params EnrollUserParams) error {
	const query = `
        INSERT INTO course_participants (user_id, course_id)
        VALUES ($1, $2)
        ON CONFLICT (user_id, course_id) DO NOTHING
    `

	cmdTag, err := r.db.Exec(
		ctx,
		query,
		params.UserID,
		params.CourseID,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23503": // foreign key violation
				if strings.Contains(pgErr.Message, "fk_course_participants__user_id_account_id") {
					return ErrUserNotInAccount
				}
				if strings.Contains(pgErr.Message, "fk_course_participants__course_id_account_id") {
					return ErrCourseNotFound
				}
			}
		}
		return fmt.Errorf("enroll user: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return ErrAlreadyEnrolled
	}

	return nil
}

type UnenrollUserParams struct {
	UserID, CourseID int64
}

// UnenrollUser удаляет пользователя из участников курса
func (r *coursesRepo) UnenrollUser(ctx context.Context, params UnenrollUserParams) error {
	const query = `DELETE FROM course_participants WHERE user_id = $1 AND course_id = $2`

	cmdTag, err := r.db.Exec(ctx, query, params.UserID, params.CourseID)
	if err != nil {
		return fmt.Errorf("unenroll user: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return ErrNotEnrolled
	}

	return nil
}

// IsUserEnrolled проверяет участие пользователя в курсе
func (r *coursesRepo) IsUserEnrolled(ctx context.Context, userID, courseID int64) (bool, error) {
	const query = `SELECT EXISTS(SELECT 1 FROM course_participants WHERE user_id = $1 AND course_id = $2)`

	var exists bool
	err := r.db.QueryRow(ctx, query, userID, courseID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check enrollment: %w", err)
	}

	return exists, nil
}

// GetCourseParticipants возвращает список участников курса
func (r *coursesRepo) GetCourseParticipants(ctx context.Context, courseID int64) ([]int64, error) {
	const query = `SELECT user_id FROM course_participants WHERE course_id = $1`

	rows, err := r.db.Query(ctx, query, courseID)
	if err != nil {
		return nil, fmt.Errorf("get participants: %w", err)
	}
	defer rows.Close()

	var users []int64
	for rows.Next() {
		var userID int64
		if err := rows.Scan(&userID); err != nil {
			return nil, fmt.Errorf("scan user_id: %w", err)
		}
		users = append(users, userID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return users, nil
}

/* ===================== Modules ===================== */

type CreateModuleParams struct {
	CourseID int64
	Name     string
	OrderIdx *int64
}

func (r *coursesRepo) CreateModule(ctx context.Context, params CreateModuleParams) (int64, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var newOrderIdx int64

	// Шаг 1: Получаем максимальный order_idx для данного курса
	const maxQuery = `SELECT COALESCE(MAX(order_idx), 0) FROM modules WHERE course_id = $1`
	var maxOrderIdx int64
	if err := tx.QueryRow(ctx, maxQuery, params.CourseID).Scan(&maxOrderIdx); err != nil {
		return 0, fmt.Errorf("get max order_idx: %w", err)
	}

	// Шаг 2: Определяем порядок вставки
	if params.OrderIdx == nil || *params.OrderIdx > maxOrderIdx {
		newOrderIdx = maxOrderIdx + 1
	} else {
		newOrderIdx = max(*params.OrderIdx, 0)

		// Сдвигаем элементы с order_idx >= newOrderIdx на 1
		if newOrderIdx <= maxOrderIdx {
			const shiftQuery = `
				UPDATE modules
				SET order_idx = order_idx + 1
				WHERE course_id = $1 AND order_idx >= $2
			`
			if _, err := tx.Exec(ctx, shiftQuery, params.CourseID, newOrderIdx); err != nil {
				return 0, fmt.Errorf("shift modules: %w", err)
			}
		}
	}

	// Шаг 3: Вставляем новый модуль в выбранную позицию
	const insertQuery = `
		INSERT INTO modules (course_id, name, order_idx)
		VALUES ($1, $2, $3)
		RETURNING id
	`
	var id int64
	if err := tx.QueryRow(ctx, insertQuery, params.CourseID, params.Name, newOrderIdx).Scan(&id); err != nil {
		var pqErr *pgconn.PgError
		if errors.As(err, &pqErr) {
			if pqErr.Code == "23503" {
				return 0, ErrCourseNotFound
			}
		}
		return 0, fmt.Errorf("insert module: %w", err)
	}

	// Шаг 4: Подтверждаем транзакцию
	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit: %w", err)
	}
	return id, nil
}

type UpdateModuleParams struct {
	ModuleID int64
	Name     string
}

func (r *coursesRepo) UpdateModule(ctx context.Context, params UpdateModuleParams) error {
	const query = `UPDATE modules SET (name) = ($1) WHERE id = $2`
	cmdTag, err := r.db.Exec(ctx, query, params.Name, params.ModuleID)
	if err != nil {
		return fmt.Errorf("update module: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return ErrModuleNotFound
	}
	return nil
}

type UpdateModuleOrderIdxParams struct {
	ModuleID int64
	OrderIdx int64
}

func (r *coursesRepo) UpdateModuleOrderIdx(ctx context.Context, params UpdateModuleOrderIdxParams) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// получаем текущее course_id и order_idx для модуля
	const getQuery = `SELECT course_id, order_idx FROM modules WHERE id = $1`
	var courseID, oldIdx int64
	if err := tx.QueryRow(ctx, getQuery, params.ModuleID).Scan(&courseID, &oldIdx); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrModuleNotFound
		}
		return fmt.Errorf("select module: %w", err)
	}

	// находим максимальный индекс в этом курсе
	const maxQuery = `SELECT COALESCE(MAX(order_idx), 0) FROM modules WHERE course_id = $1`
	var maxIdx int64
	if err := tx.QueryRow(ctx, maxQuery, courseID).Scan(&maxIdx); err != nil {
		return fmt.Errorf("select max order_idx: %w", err)
	}

	// нормализуем новый индекс
	newIdx := params.OrderIdx
	switch {
	case newIdx < 0:
		newIdx = 0
	case newIdx > maxIdx:
		newIdx = maxIdx
	}

	// если ничего не меняется — выходим
	if newIdx == oldIdx {
		return tx.Commit(ctx)
	}

	// один SQL-запрос, безопасный для UNIQUE (course_id, order_idx)
	const updateOrderQuery = `
        UPDATE modules
           SET order_idx = CASE
               WHEN id = $1 THEN $2
               WHEN $2 < $3 AND course_id = $4 AND order_idx > $2 AND order_idx <= $3 THEN order_idx - 1
               WHEN $2 > $3 AND course_id = $4 AND order_idx >= $3 AND order_idx < $2 THEN order_idx + 1
               ELSE order_idx
           END
         WHERE course_id = $4
           AND (id = $1 OR (order_idx BETWEEN LEAST($2, $3) AND GREATEST($2, $3)))
    `
	if _, err := tx.Exec(ctx, updateOrderQuery, params.ModuleID, oldIdx, newIdx, courseID); err != nil {
		return fmt.Errorf("update order_idx: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

type RemoveModuleParams struct {
	ModuleID int64
}

func (r *coursesRepo) RemoveModule(ctx context.Context, params RemoveModuleParams) error {
	const query = `DELETE FROM modules WHERE id = $1`
	cmdTag, err := r.db.Exec(ctx, query, params.ModuleID)
	if err != nil {
		return fmt.Errorf("delete module: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return ErrModuleNotFound
	}
	return nil
}

func (r *coursesRepo) GetModulesByCourseID(ctx context.Context, courseID int64) ([]*models.Module, error) {
	_, err := r.GetCourseByID(ctx, courseID)
	if err != nil {
		if errors.Is(err, ErrCourseNotFound) {
			return nil, err
		}
		return nil, fmt.Errorf("check course existence: %v", err)
	}

	const query = `SELECT id, course_id, name, order_idx FROM modules WHERE course_id = $1 ORDER BY order_idx ASC`
	rows, err := r.db.Query(ctx, query, courseID)
	if err != nil {
		return nil, fmt.Errorf("query modules: %w", err)
	}
	defer rows.Close()

	var modules []*models.Module
	for rows.Next() {
		var m models.Module
		if err := rows.Scan(&m.ID, &m.CourseID, &m.Name, &m.OrderIdx); err != nil {
			return nil, fmt.Errorf("scan module: %w", err)
		}
		modules = append(modules, &m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return modules, nil
}

func (r *coursesRepo) GetModuleByID(ctx context.Context, moduleID int64) (*models.Module, error) {
	const query = `SELECT id, course_id, order_idx, name FROM modules WHERE id = $1`
	var module models.Module
	err := r.db.QueryRow(ctx, query, moduleID).Scan(&module.ID, &module.CourseID, &module.OrderIdx, &module.Name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrModuleNotFound
		}
		return nil, fmt.Errorf("get lesson by id: %w", err)
	}
	return &module, nil
}

/* ===================== Lessons ===================== */

type CreateLessonParams struct {
	ModuleID int64
	Name     string
	OrderIdx *int
}

func (r *coursesRepo) CreateLesson(ctx context.Context, params CreateLessonParams) (int64, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var newOrderIdx int

	// Получаем максимальный order_idx для модуля
	const maxQuery = `SELECT COALESCE(MAX(order_idx), 0) FROM lessons WHERE module_id = $1`
	var maxOrderIdx int
	if err := tx.QueryRow(ctx, maxQuery, params.ModuleID).Scan(&maxOrderIdx); err != nil {
		return 0, fmt.Errorf("get max order_idx: %w", err)
	}

	if params.OrderIdx == nil || *params.OrderIdx > maxOrderIdx {
		newOrderIdx = maxOrderIdx + 1
	} else {
		newOrderIdx = max(*params.OrderIdx, 0)
		if newOrderIdx <= maxOrderIdx {
			const shiftQuery = `
				UPDATE lessons
				SET order_idx = order_idx + 1
				WHERE module_id = $1 AND order_idx >= $2
			`
			if _, err := tx.Exec(ctx, shiftQuery, params.ModuleID, newOrderIdx); err != nil {
				return 0, fmt.Errorf("shift lessons: %w", err)
			}
		}
	}

	// Вставка
	const insertQuery = `
		INSERT INTO lessons (module_id, name, order_idx)
		VALUES ($1, $2, $3)
		RETURNING id
	`
	var id int64
	if err := tx.QueryRow(ctx, insertQuery, params.ModuleID, params.Name, newOrderIdx).Scan(&id); err != nil {
		var pqErr *pgconn.PgError
		if errors.As(err, &pqErr) && pqErr.Code == "23503" {
			return 0, ErrModuleNotFound
		}
		return 0, fmt.Errorf("insert lesson: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit: %w", err)
	}

	return id, nil
}

type UpdateLessonParams struct {
	LessonID int64
	Name     string
}

func (r *coursesRepo) UpdateLesson(ctx context.Context, params UpdateLessonParams) error {
	const query = `UPDATE lessons SET name = $1 WHERE id = $2`
	cmdTag, err := r.db.Exec(ctx, query, params.Name, params.LessonID)
	if err != nil {
		return fmt.Errorf("update lesson: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return ErrLessonNotFound
	}
	return nil
}

type UpdateLessonOrderIdxParams struct {
	LessonID int64
	OrderIdx int
}

func (r *coursesRepo) UpdateLessonOrderIdx(ctx context.Context, params UpdateLessonOrderIdxParams) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	const getQuery = `SELECT module_id, order_idx FROM lessons WHERE id = $1`
	var moduleID int64
	var oldIdx int
	if err := tx.QueryRow(ctx, getQuery, params.LessonID).Scan(&moduleID, &oldIdx); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrLessonNotFound
		}
		return fmt.Errorf("select lesson: %w", err)
	}

	const maxQuery = `SELECT COALESCE(MAX(order_idx), 0) FROM lessons WHERE module_id = $1`
	var maxIdx int
	if err := tx.QueryRow(ctx, maxQuery, moduleID).Scan(&maxIdx); err != nil {
		return fmt.Errorf("select max order_idx: %w", err)
	}

	newIdx := params.OrderIdx
	switch {
	case newIdx < 0:
		newIdx = 0
	case newIdx > maxIdx:
		newIdx = maxIdx
	}

	if newIdx == oldIdx {
		return tx.Commit(ctx)
	}

	const updateOrderQuery = `
		UPDATE lessons
		SET order_idx = CASE
			WHEN id = $1 THEN $2
			WHEN $2 < $3 AND module_id = $4 AND order_idx > $2 AND order_idx <= $3 THEN order_idx - 1
			WHEN $2 > $3 AND module_id = $4 AND order_idx >= $3 AND order_idx < $2 THEN order_idx + 1
			ELSE order_idx
		END
		WHERE module_id = $4
		  AND (id = $1 OR (order_idx BETWEEN LEAST($2, $3) AND GREATEST($2, $3)))
	`
	if _, err := tx.Exec(ctx, updateOrderQuery, params.LessonID, oldIdx, newIdx, moduleID); err != nil {
		return fmt.Errorf("update order_idx: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

type RemoveLessonParams struct {
	LessonID int64
}

func (r *coursesRepo) RemoveLesson(ctx context.Context, params RemoveLessonParams) error {
	const query = `DELETE FROM lessons WHERE id = $1`
	cmdTag, err := r.db.Exec(ctx, query, params.LessonID)
	if err != nil {
		return fmt.Errorf("delete lesson: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return ErrLessonNotFound
	}
	return nil
}

func (r *coursesRepo) GetLessonsByModuleID(ctx context.Context, moduleID int64) ([]*models.Lesson, error) {
	_, err := r.GetModuleByID(ctx, moduleID)
	if err != nil {
		if errors.Is(err, ErrModuleNotFound) {
			return nil, err
		}
		return nil, fmt.Errorf("check module existence: %v", err)
	}

	const query = `SELECT id, module_id, name, order_idx FROM lessons WHERE module_id = $1 ORDER BY order_idx ASC`
	rows, err := r.db.Query(ctx, query, moduleID)
	if err != nil {
		return nil, fmt.Errorf("query lessons: %w", err)
	}
	defer rows.Close()

	var lessons []*models.Lesson
	for rows.Next() {
		var l models.Lesson
		if err := rows.Scan(&l.ID, &l.ModuleID, &l.Name, &l.OrderIdx); err != nil {
			return nil, fmt.Errorf("scan lesson: %w", err)
		}
		lessons = append(lessons, &l)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return lessons, nil
}

func (r *coursesRepo) GetLessonByID(ctx context.Context, lessonID int64) (*models.Lesson, error) {
	const query = `SELECT id, module_id, order_idx, name FROM lessons WHERE id = $1`
	var lesson models.Lesson
	err := r.db.QueryRow(ctx, query, lessonID).Scan(&lesson.ID, &lesson.ModuleID, &lesson.OrderIdx, &lesson.Name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrLessonNotFound
		}
		return nil, fmt.Errorf("get lesson by id: %w", err)
	}
	return &lesson, nil
}

/* ===================== Blocks ===================== */

type createBaseBlockParams struct {
	LessonID int64
	OrderIdx *int
}

func createBaseBlock(ctx context.Context, tx pgx.Tx, params createBaseBlockParams) (int64, error) {
	const maxQuery = `SELECT COALESCE(MAX(order_idx), 0) FROM blocks WHERE lesson_id = $1`
	var maxOrderIdx int
	if err := tx.QueryRow(ctx, maxQuery, params.LessonID).Scan(&maxOrderIdx); err != nil {
		return 0, fmt.Errorf("get max order_idx: %w", err)
	}

	var newOrderIdx int
	if params.OrderIdx == nil || *params.OrderIdx > maxOrderIdx {
		newOrderIdx = maxOrderIdx + 1
	} else {
		newOrderIdx = max(*params.OrderIdx, 0)
		const shiftQuery = `UPDATE blocks SET order_idx = order_idx + 1 WHERE lesson_id = $1 AND order_idx >= $2`
		if _, err := tx.Exec(ctx, shiftQuery, params.LessonID, newOrderIdx); err != nil {
			return 0, fmt.Errorf("shift order_idx: %w", err)
		}
	}
	const insertBlockQuery = `INSERT INTO blocks (lesson_id, order_idx, type) VALUES ($1, $2, $3) RETURNING id`
	var blockID int64
	if err := tx.QueryRow(ctx, insertBlockQuery, params.LessonID, newOrderIdx, models.BlockTypeVideo).Scan(&blockID); err != nil {
		var pqErr *pgconn.PgError
		if errors.As(err, &pqErr) && pqErr.Code == "23503" {
			return 0, ErrLessonNotFound
		}
		return 0, fmt.Errorf("insert into blocks: %w", err)
	}

	return blockID, nil
}

type CreateVideoBlockParams struct {
	createBaseBlockParams
	FileID int64
}

func (r *coursesRepo) CreateVideoBlock(ctx context.Context, params CreateVideoBlockParams) (int64, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	blockID, err := createBaseBlock(ctx, tx, params.createBaseBlockParams)
	if err != nil {
		return 0, err
	}

	const insertVideoQuery = `INSERT INTO video_blocks (id, file_id) VALUES ($1, $2)`
	if _, err := tx.Exec(ctx, insertVideoQuery, blockID, params.FileID); err != nil {
		var pqErr *pgconn.PgError
		if errors.As(err, &pqErr) && pqErr.Code == "23503" {
			return 0, ErrFileNotFound
		}
		return 0, fmt.Errorf("insert into video_blocks: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit: %w", err)
	}
	return blockID, nil
}

type CreateTextBlockParams struct {
	createBaseBlockParams
	Content string
}

func (r *coursesRepo) CreateTextBlock(ctx context.Context, params CreateTextBlockParams) (int64, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	blockID, err := createBaseBlock(ctx, tx, params.createBaseBlockParams)
	if err != nil {
		return 0, err
	}

	const insertTextQuery = `INSERT INTO text_blocks (id, content) VALUES ($1, $2)`
	if _, err := tx.Exec(ctx, insertTextQuery, blockID, params.Content); err != nil {
		return 0, fmt.Errorf("insert into text_blocks: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit: %w", err)
	}
	return blockID, nil
}

type UpdateBlockOrderIdxParams struct {
	BlockID  int64
	OrderIdx int
}

func (r *coursesRepo) UpdateBlockOrderIdx(ctx context.Context, params UpdateBlockOrderIdxParams) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	const getQuery = `SELECT lesson_id, order_idx FROM blocks WHERE id = $1`
	var lessonID int64
	var oldIdx int
	if err := tx.QueryRow(ctx, getQuery, params.BlockID).Scan(&lessonID, &oldIdx); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrBlockNotFound
		}
		return fmt.Errorf("select block: %w", err)
	}

	const maxQuery = `SELECT COALESCE(MAX(order_idx), 0) FROM blocks WHERE lesson_id = $1`
	var maxIdx int
	if err := tx.QueryRow(ctx, maxQuery, lessonID).Scan(&maxIdx); err != nil {
		return fmt.Errorf("select max order_idx: %w", err)
	}

	newIdx := params.OrderIdx
	switch {
	case newIdx < 0:
		newIdx = 0
	case newIdx > maxIdx:
		newIdx = maxIdx
	}

	if newIdx == oldIdx {
		return tx.Commit(ctx)
	}

	const updateOrderQuery = `
		UPDATE blocks
		SET order_idx = CASE
			WHEN id = $1 THEN $2
			WHEN $2 < $3 AND lesson_id = $4 AND order_idx > $2 AND order_idx <= $3 THEN order_idx - 1
			WHEN $2 > $3 AND lesson_id = $4 AND order_idx >= $3 AND order_idx < $2 THEN order_idx + 1
			ELSE order_idx
		END
		WHERE lesson_id = $4
		  AND (id = $1 OR (order_idx BETWEEN LEAST($2, $3) AND GREATEST($2, $3)))
	`
	if _, err := tx.Exec(ctx, updateOrderQuery, params.BlockID, oldIdx, newIdx, lessonID); err != nil {
		return fmt.Errorf("update block order_idx: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

type RemoveBlockParams struct {
	BlockID int64
}

func (r *coursesRepo) RemoveBlock(ctx context.Context, params RemoveBlockParams) error {
	const query = `DELETE FROM blocks WHERE id = $1`
	cmdTag, err := r.db.Exec(ctx, query, params.BlockID)
	if err != nil {
		return fmt.Errorf("delete block: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return ErrBlockNotFound
	}
	return nil
}

func (r *coursesRepo) GetBlocksByLessonID(ctx context.Context, lessonID int64) ([]any, error) {
	_, err := r.GetLessonByID(ctx, lessonID)
	if err != nil {
		if errors.Is(err, ErrLessonNotFound) {
			return nil, ErrLessonNotFound
		}
		return nil, fmt.Errorf("check lesson existence: %w", err)
	}

	const query = `
        SELECT 
            b.id, b.lesson_id, b.order_idx, b.type,
            vb.file_id,
            tb.content
        FROM blocks b
        LEFT JOIN video_blocks vb ON b.id = vb.id
        LEFT JOIN text_blocks tb ON b.id = tb.id
        WHERE b.lesson_id = $1
        ORDER BY b.order_idx ASC
    `

	rows, err := r.db.Query(ctx, query, lessonID)
	if err != nil {
		return nil, fmt.Errorf("query blocks: %w", err)
	}
	defer rows.Close()

	var blocks []any
	for rows.Next() {
		var (
			base    models.BaseBlock
			fileID  *int64
			content *string
		)

		err := rows.Scan(
			&base.ID,
			&base.LessonID,
			&base.OrderIdx,
			&base.Type,
			&fileID,
			&content,
		)
		if err != nil {
			return nil, fmt.Errorf("scan block: %w", err)
		}

		switch base.Type {
		case models.BlockTypeVideo:
			if fileID == nil {
				return nil, fmt.Errorf("video block %d has no file_id", base.ID)
			}
			blocks = append(blocks, &models.VideoBlock{
				BaseBlock: base,
				FileID:    *fileID,
			})

		case models.BlockTypeText:
			if content == nil {
				return nil, fmt.Errorf("text block %d has no content", base.ID)
			}
			blocks = append(blocks, &models.TextBlock{
				BaseBlock: base,
				Content:   *content,
			})

		default:
			return nil, fmt.Errorf("unknown block type: %s", base.Type)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return blocks, nil
}
