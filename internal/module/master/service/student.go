package service

import (
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	"context"
)

func (s *masterService) GetStudents(ctx context.Context, req *entity.GetStudentsReq) (*entity.GetStudentsResp, error) {
	ctx, span := tracing.StartSpan(ctx, "service.GetStudents")
	defer span.End()
	return s.repo.GetStudents(ctx, req)
}

func (s *masterService) CreateStudent(ctx context.Context, req *entity.CreateStudentReq) (*entity.CreateStudentResp, error) {
	ctx, span := tracing.StartSpan(ctx, "service.CreateStudent")
	defer span.End()
	return s.repo.CreateStudent(ctx, req)
}

func (s *masterService) GetStudent(ctx context.Context, req *entity.GetStudentReq) (*entity.GetStudentResp, error) {
	ctx, span := tracing.StartSpan(ctx, "service.GetStudent")
	defer span.End()
	return s.repo.GetStudent(ctx, req)
}

func (s *masterService) UpdateStudent(ctx context.Context, req *entity.UpdateStudentReq) error {
	ctx, span := tracing.StartSpan(ctx, "service.UpdateStudent")
	defer span.End()
	return s.repo.UpdateStudent(ctx, req)
}

func (s *masterService) DeleteStudent(ctx context.Context, req *entity.DeleteStudentReq) error {
	ctx, span := tracing.StartSpan(ctx, "service.DeleteStudent")
	defer span.End()
	return s.repo.DeleteStudent(ctx, req)
}
