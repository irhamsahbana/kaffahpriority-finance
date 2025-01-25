package service

import (
	"codebase-app/internal/module/master/entity"
	"codebase-app/internal/module/master/ports"
	"context"
)

var _ ports.MasterService = &masterService{}

type masterService struct {
	repo ports.MasterRepository
}

func NewMasterService(repo ports.MasterRepository) *masterService {
	return &masterService{
		repo: repo,
	}
}

func (s *masterService) GetMarketers(ctx context.Context, req *entity.GetMarketersReq) (*entity.GetMarketersResp, error) {
	return s.repo.GetMarketers(ctx, req)
}

func (s *masterService) GetStudentManagers(ctx context.Context, req *entity.GetStudentManagersReq) (*entity.GetStudentManagersResp, error) {
	return s.repo.GetStudentManagers(ctx, req)
}

func (s *masterService) GetLecturers(ctx context.Context, req *entity.GetLecturersReq) (*entity.GetLecturersResp, error) {
	return s.repo.GetLecturers(ctx, req)
}

func (s *masterService) GetLecturer(ctx context.Context, req *entity.GetLecturerReq) (*entity.GetLecturerResp, error) {
	return s.repo.GetLecturer(ctx, req)
}

func (s *masterService) CreateLecturer(ctx context.Context, req *entity.CreateLecturerReq) (*entity.CreateLecturerResp, error) {
	return s.repo.CreateLecturer(ctx, req)
}

func (s *masterService) UpdateLecturer(ctx context.Context, req *entity.UpdateLecturerReq) error {
	return s.repo.UpdateLecturer(ctx, req)
}

func (s *masterService) DeleteLecturer(ctx context.Context, req *entity.DeleteLecturerReq) error {
	return s.repo.DeleteLecturer(ctx, req)
}

func (s *masterService) GetStudents(ctx context.Context, req *entity.GetStudentsReq) (*entity.GetStudentsResp, error) {
	return s.repo.GetStudents(ctx, req)
}

func (s *masterService) CreateStudent(ctx context.Context, req *entity.CreateStudentReq) (*entity.CreateStudentResp, error) {
	return s.repo.CreateStudent(ctx, req)
}

func (s *masterService) GetStudent(ctx context.Context, req *entity.GetStudentReq) (*entity.GetStudentResp, error) {
	return s.repo.GetStudent(ctx, req)
}

func (s *masterService) UpdateStudent(ctx context.Context, req *entity.UpdateStudentReq) error {
	return s.repo.UpdateStudent(ctx, req)
}

func (s *masterService) DeleteStudent(ctx context.Context, req *entity.DeleteStudentReq) error {
	return s.repo.DeleteStudent(ctx, req)
}

func (s *masterService) CreateProgram(ctx context.Context, req *entity.CreateProgramReq) (*entity.CreateProgramResp, error) {
	return s.repo.CreateProgram(ctx, req)
}

func (s *masterService) GetPrograms(ctx context.Context, req *entity.GetProgramsReq) (*entity.GetProgramsResp, error) {
	return s.repo.GetPrograms(ctx, req)
}

func (s *masterService) GetProgram(ctx context.Context, req *entity.GetProgramReq) (*entity.GetProgramResp, error) {
	return s.repo.GetProgram(ctx, req)
}

func (s *masterService) UpdateProgram(ctx context.Context, req *entity.UpdateProgramReq) (*entity.UpdateProgramResp, error) {
	return s.repo.UpdateProgram(ctx, req)
}

func (s *masterService) DeleteProgram(ctx context.Context, req *entity.DeleteProgramReq) error {
	return s.repo.DeleteProgram(ctx, req)
}
