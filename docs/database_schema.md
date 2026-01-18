# Database Schema

The project uses PostgreSQL. Schema changes are managed via SQL migrations using **Goose**.

## Migration Files
- Location: `db/migrations/`
- Format: `YYYYMMDDHHMMSS_name.sql`

## Key Tables

### Access Control (RBAC)
- **roles**: User roles (e.g., admin, manager).
- **permissions**: Granular permissions.
- **role_permissions**: Many-to-many link between roles and permissions.

### Users & Staff
- **users**: Authentication and base user data.
- **academic_managers**: Academic management staff.
- **lecturers**: Teaching staff.
- **student_managers**: Student management staff.
- **marketers**: Marketing staff.
- **students**: Students.

### Programs & Registration
- **programs**: Educational programs offered.
- **program_registrations**: Student registrations for programs.
- **program_registration_templates**: Templates for registrations.
- **prt_additional_students**: Additional students in templates.
- **pr_additional_students**: Additional students in registrations.

### System
- **activity_logs**: Audit logs for system activities.

### Payroll
- **payroll_runs**: Payroll periods/cycles.
- **payroll_items**: Individual payroll items for each mentor/program in a cycle.

## Workflow
1.  **Create Migration**: `task create-migration name=my_change`
2.  **Edit SQL**: Write the UP and DOWN SQL in the generated file.
3.  **Apply**: `task migrate cmd=up`
