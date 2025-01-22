package ports

import (
	"codebase-app/internal/module/master/entity"
	"context"
)

type MasterRepository interface {
	GetMarketers(ctx context.Context, req *entity.GetMarketersReq) (*entity.GetMarketersResp, error)
	GetLecturers(ctx context.Context, req *entity.GetLecturersReq) (*entity.GetLecturersResp, error)
	GetStudents(ctx context.Context, req *entity.GetStudentsReq) (*entity.GetStudentsResp, error)
	GetPrograms(ctx context.Context, req *entity.GetProgramsReq) (*entity.GetProgramsResp, error)
	GetProgram(ctx context.Context, req *entity.GetProgramReq) (*entity.GetProgramResp, error)
}

type MasterService interface {
	GetMarketers(ctx context.Context, req *entity.GetMarketersReq) (*entity.GetMarketersResp, error)
	GetLecturers(ctx context.Context, req *entity.GetLecturersReq) (*entity.GetLecturersResp, error)
	GetStudents(ctx context.Context, req *entity.GetStudentsReq) (*entity.GetStudentsResp, error)
	GetPrograms(ctx context.Context, req *entity.GetProgramsReq) (*entity.GetProgramsResp, error)
	GetProgram(ctx context.Context, req *entity.GetProgramReq) (*entity.GetProgramResp, error)
}
