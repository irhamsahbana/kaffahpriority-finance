package ports

import (
	"codebase-app/internal/module/master/entity"
	"context"
)

type MasterRepository interface {
	GetMarketers(ctx context.Context, req *entity.GetMarketersReq) (*entity.GetMarketersResp, error)
	GetMarketer(ctx context.Context, req *entity.GetMarketerReq) (*entity.GetMarketerResp, error)
	CreateMarketer(ctx context.Context, req *entity.CreateMarketerReq) (*entity.CreateMarketerResp, error)
	UpdateMarketer(ctx context.Context, req *entity.UpdateMarketerReq) error
	DeleteMarketer(ctx context.Context, req *entity.DeleteMarketerReq) error

	GetStudentManagers(ctx context.Context, req *entity.GetStudentManagersReq) (*entity.GetStudentManagersResp, error)
	CreateStudentManager(ctx context.Context, req *entity.CreateStudentManagerReq) (*entity.CreateStudentManagerResp, error)
	GetStudentManager(ctx context.Context, req *entity.GetStudentManagerReq) (*entity.GetStudentManagerResp, error)
	UpdateStudentManager(ctx context.Context, req *entity.UpdateStudentManagerReq) error
	DeleteStudentManager(ctx context.Context, req *entity.DeleteStudentManagerReq) error

	GetLecturers(ctx context.Context, req *entity.GetLecturersReq) (*entity.GetLecturersResp, error)
	GetLecturer(ctx context.Context, req *entity.GetLecturerReq) (*entity.GetLecturerResp, error)
	CreateLecturer(ctx context.Context, req *entity.CreateLecturerReq) (*entity.CreateLecturerResp, error)
	UpdateLecturer(ctx context.Context, req *entity.UpdateLecturerReq) error
	DeleteLecturer(ctx context.Context, req *entity.DeleteLecturerReq) error

	GetStudents(ctx context.Context, req *entity.GetStudentsReq) (*entity.GetStudentsResp, error)
	CreateStudent(ctx context.Context, req *entity.CreateStudentReq) (*entity.CreateStudentResp, error)
	GetStudent(ctx context.Context, req *entity.GetStudentReq) (*entity.GetStudentResp, error)
	UpdateStudent(ctx context.Context, req *entity.UpdateStudentReq) error
	DeleteStudent(ctx context.Context, req *entity.DeleteStudentReq) error

	CreateProgram(ctx context.Context, req *entity.CreateProgramReq) (*entity.CreateProgramResp, error)
	GetPrograms(ctx context.Context, req *entity.GetProgramsReq) (*entity.GetProgramsResp, error)
	GetProgram(ctx context.Context, req *entity.GetProgramReq) (*entity.GetProgramResp, error)
	UpdateProgram(ctx context.Context, req *entity.UpdateProgramReq) (*entity.UpdateProgramResp, error)
	DeleteProgram(ctx context.Context, req *entity.DeleteProgramReq) error
}

type MasterService interface {
	GetMarketers(ctx context.Context, req *entity.GetMarketersReq) (*entity.GetMarketersResp, error)
	GetMarketer(ctx context.Context, req *entity.GetMarketerReq) (*entity.GetMarketerResp, error)
	CreateMarketer(ctx context.Context, req *entity.CreateMarketerReq) (*entity.CreateMarketerResp, error)
	UpdateMarketer(ctx context.Context, req *entity.UpdateMarketerReq) error
	DeleteMarketer(ctx context.Context, req *entity.DeleteMarketerReq) error

	GetStudentManagers(ctx context.Context, req *entity.GetStudentManagersReq) (*entity.GetStudentManagersResp, error)
	CreateStudentManager(ctx context.Context, req *entity.CreateStudentManagerReq) (*entity.CreateStudentManagerResp, error)
	GetStudentManager(ctx context.Context, req *entity.GetStudentManagerReq) (*entity.GetStudentManagerResp, error)
	UpdateStudentManager(ctx context.Context, req *entity.UpdateStudentManagerReq) error
	DeleteStudentManager(ctx context.Context, req *entity.DeleteStudentManagerReq) error

	GetLecturers(ctx context.Context, req *entity.GetLecturersReq) (*entity.GetLecturersResp, error)
	GetLecturer(ctx context.Context, req *entity.GetLecturerReq) (*entity.GetLecturerResp, error)
	CreateLecturer(ctx context.Context, req *entity.CreateLecturerReq) (*entity.CreateLecturerResp, error)
	UpdateLecturer(ctx context.Context, req *entity.UpdateLecturerReq) error
	DeleteLecturer(ctx context.Context, req *entity.DeleteLecturerReq) error

	GetStudents(ctx context.Context, req *entity.GetStudentsReq) (*entity.GetStudentsResp, error)
	CreateStudent(ctx context.Context, req *entity.CreateStudentReq) (*entity.CreateStudentResp, error)
	GetStudent(ctx context.Context, req *entity.GetStudentReq) (*entity.GetStudentResp, error)
	UpdateStudent(ctx context.Context, req *entity.UpdateStudentReq) error
	DeleteStudent(ctx context.Context, req *entity.DeleteStudentReq) error

	CreateProgram(ctx context.Context, req *entity.CreateProgramReq) (*entity.CreateProgramResp, error)
	GetPrograms(ctx context.Context, req *entity.GetProgramsReq) (*entity.GetProgramsResp, error)
	GetProgram(ctx context.Context, req *entity.GetProgramReq) (*entity.GetProgramResp, error)
	UpdateProgram(ctx context.Context, req *entity.UpdateProgramReq) (*entity.UpdateProgramResp, error)
	DeleteProgram(ctx context.Context, req *entity.DeleteProgramReq) error
}
