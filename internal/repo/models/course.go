package models

type Course struct {
	ID        int64
	AccountID int64
	Name      string
}

type Module struct {
	ID       int64
	CourseID int64
	OrderIdx int
	Name     string
}

type Lesson struct {
	ID       int64
	ModuleID int64
	OrderIdx int
	Name     string
}

type BlockType string

const (
	BlockTypeVideo BlockType = "video"
	BlockTypeText  BlockType = "text"
)

type BaseBlock struct {
	ID       int64
	LessonID int64
	OrderIdx int
	Type     BlockType
}

type VideoBlock struct {
	BaseBlock
	FileID int64
}

type TextBlock struct {
	BaseBlock
	Content string
}

type CourseHierarchy struct {
	Course  *Course
	Modules []*ModuleHierarchy
}

type ModuleHierarchy struct {
	Module  *Module
	Lessons []*Lesson
}
