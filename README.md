# Kaffah Priority Finance Backend

Backend of Kaffah Priority Finance project.

## Developer Guide

### General Rules

* We use conventional commits to deal with git commits: <https://www.conventionalcommits.org>
  * Use `feat: commit message` to do git commit related to feature.
  * Use `refactor: commit message` to do git commit related to code refactorings.
  * Use `fix: commit message` to do git commit related to bugfix.
  * Use `test: commit message` to do git commit related to test files.
  * Use `docs: commit message` to do git commit related to documentations (including README.md files).
  * Use `style: commit message` to do git commit related to code style.

* Use git-chglog <https://github.com/git-chglog/git-chglog> to generate changelog (CHANGELOG.md) before merging to release branch.

### Branching Strategy

* Keep your branch strategy simple. Build your strategy from these three concepts:
  * Use feature branches for all new features and bug fixes.
  * Merge feature branches into the main branch using pull requests.
  * Keep a high quality, up-to-date main branch.

#### Use feature branches for your work

Develop your features and fix bugs in feature branches based off your main branch. These branches are also known as
topic branches. Feature branches isolate work in progress from the completed work in the main branch. Git branches are
inexpensive to create and maintain. Even small fixes and changes should have their own feature branch.

<!-- <p align="left"><img src="./featurebranching.png" width="360"></p> -->

#### Name your feature branches by convention

* Use a consistent naming convention for your feature branches to identify the work done in the branch. You can also
  include other information in the branch name, such as who created the branch.

* Some suggestions for naming your feature branches:
  * users/username/description
  * users/username/workitem
  * bugfix/description
  * feature/feature-name
  * feature/feature-area/feature-name
  * hotfix/description

#### Use release branches

* Create a release branch from the main branch when you get close to your release or other milestone, such as the end of
  a sprint. Give this branch a clear name associating it with the release, for example release/20.
* Create branches to fix bugs from the release branch and merge them back into the release branch in a pull request

<!-- <p align="left"><img src="./releasebranching_release.png" width="360"></p> -->

## API Endpoints

### General

* `GET /ping`: Health check
* `GET /storage/private/:filename`: Get private file

### User

* `POST /users/login`: User login
* `GET /users/me`: Get current user
* `POST /users/entities`: Create a new user
* `GET /users/entities`: Get all users
* `GET /users/entities/:id`: Get user by ID
* `PUT /users/entities/:id`: Update user by ID
* `DELETE /users/entities/:id`: Delete user by ID

### RBAC

* `GET /role-permissions/role-access-rights`: Get role and permissions
* `POST /role-permissions/roles`: Create a new role
* `PUT /role-permissions/roles/:id`: Update role by ID
* `GET /role-permissions/roles/:id`: Get role detail by ID
* `PUT /role-permissions/roles/:id/permissions`: Update role permissions
* `DELETE /role-permissions/roles/:id`: Delete role by ID
* `GET /role-permissions/permissions`: Get all permissions

### Master

* `GET /masters/student-managers`: Get all student managers
* `GET /masters/student-managers/:id`: Get student manager by ID
* `POST /masters/student-managers`: Create a new student manager
* `PUT /masters/student-managers/:id`: Update student manager by ID
* `DELETE /masters/student-managers/:id`: Delete student manager by ID
* `GET /masters/marketers`: Get all marketers
* `GET /masters/marketers/:id`: Get marketer by ID
* `POST /masters/marketers`: Create a new marketer
* `PUT /masters/marketers/:id`: Update marketer by ID
* `DELETE /masters/marketers/:id`: Delete marketer by ID
* `GET /masters/academic-managers`: Get all academic managers
* `GET /masters/academic-managers/:id`: Get academic manager by ID
* `POST /masters/academic-managers`: Create a new academic manager
* `PUT /masters/academic-managers/:id`: Update academic manager by ID
* `DELETE /masters/academic-managers/:id`: Delete academic manager by ID
* `GET /masters/lecturers`: Get all lecturers
* `GET /masters/lecturers/:id`: Get lecturer by ID
* `POST /masters/lecturers`: Create a new lecturer
* `PUT /masters/lecturers/:id`: Update lecturer by ID
* `DELETE /masters/lecturers/:id`: Delete lecturer by ID
* `GET /masters/students`: Get all students
* `POST /masters/students`: Create a new student
* `GET /masters/students/:id`: Get student by ID
* `PUT /masters/students/:id`: Update student by ID
* `DELETE /masters/students/:id`: Delete student by ID
* `POST /masters/programs`: Create a new program
* `GET /masters/programs`: Get all programs
* `GET /masters/programs/:id`: Get program by ID
* `PUT /masters/programs/:id`: Update program by ID
* `DELETE /masters/programs/:id`: Delete program by ID

### Report

* `POST /reports/templates`: Create a new template
* `GET /reports/templates`: Get all templates
* `PUT /reports/templates/:id`: Update template by ID
* `GET /reports/templates/:id`: Get template by ID
* `DELETE /reports/templates/:id`: Delete template by ID
* `POST /reports/registrations`: Create new registrations
* `POST /reports/copy-registrations`: Copy registrations
* `GET /reports/registration-summaries`: Get registration summaries
* `GET /reports/registration-summaries-for-cfo2`: Get registration summaries for CFO2
* `PATCH /reports/registration-paid-at-attributes`: Update registration paid at attributes
* `GET /reports/registrations`: Get all registrations
* `GET /reports/unused-registrations`: Get unused registrations
* `GET /reports/exported-registrations`: Get exported registrations
* `GET /reports/exported-registrations-for-cfo2-monthly`: Get exported registrations for CFO2 monthly
* `GET /reports/exported-registrations-for-cfo2-yearly`: Get exported registrations for CFO2 yearly
* `GET /reports/exported-registrations-for-wage-recap-monthly`: Get exported registrations for wage recap monthly
* `PUT /reports/registrations/:id`: Update registration by ID
* `GET /reports/registrations/:id`: Get registration by ID
* `DELETE /reports/registrations/:id`: Delete registration by ID
* `PUT /reports/registrations/:id/hr-fee-distributions`: HR distributions
* `PUT /reports/registrations/:id/lecturer-distributions`: Lecturer distributions
* `PUT /reports/registrations/:id/lecturers`: Update registration lecturer
* `PUT /reports/registrations/:id/is-paid`: Update registration is paid
* `GET /reports/registration-per-lecturers`: Get registration list per lecturer
* `PATCH /reports/lecturers-wages/:id`: Update lecturer wages
* `GET /reports/lecturers-wages`: Get lecturer wages
* `GET /reports/lecturers-wages-aggregate`: Get lecturer wages aggregate
* `GET /reports/acquisition-rights-aggregate`: Get acquisition rights aggregate
* `POST /reports/generate-registration-reports`: Generate registration reports
* `POST /reports/multi-allocation-registrations`: Registration multi-allocation
* `GET /reports/lecturer-programs`: Get lecturer programs
