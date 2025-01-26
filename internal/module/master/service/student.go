package service

import (
	"codebase-app/internal/module/master/entity"
	"context"
)

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
