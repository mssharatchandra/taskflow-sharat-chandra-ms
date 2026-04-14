package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/sharatchandra/taskflow-sharat-chandra-ms/internal/handler"
	"github.com/sharatchandra/taskflow-sharat-chandra-ms/internal/middleware"
	"github.com/sharatchandra/taskflow-sharat-chandra-ms/internal/model"
	"github.com/sharatchandra/taskflow-sharat-chandra-ms/internal/repository"
	"github.com/sharatchandra/taskflow-sharat-chandra-ms/internal/service"
)

type inMemoryProjectRepo struct {
	projects map[uuid.UUID]*model.Project
	tasks    map[uuid.UUID]*model.Task
}

func newInMemoryProjectRepo() *inMemoryProjectRepo {
	return &inMemoryProjectRepo{
		projects: make(map[uuid.UUID]*model.Project),
		tasks:    make(map[uuid.UUID]*model.Task),
	}
}

func (r *inMemoryProjectRepo) Create(_ context.Context, project *model.Project) error {
	p := *project
	r.projects[p.ID] = &p
	return nil
}

func (r *inMemoryProjectRepo) FindByID(_ context.Context, id uuid.UUID) (*model.Project, error) {
	p, ok := r.projects[id]
	if !ok {
		return nil, nil
	}
	cp := *p
	return &cp, nil
}

func (r *inMemoryProjectRepo) FindWithTasksByID(ctx context.Context, id uuid.UUID) (*model.ProjectWithTasks, error) {
	project, err := r.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, nil
	}

	tasks := make([]model.Task, 0)
	for _, task := range r.tasks {
		if task.ProjectID == id {
			tasks = append(tasks, cloneTask(*task))
		}
	}

	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].CreatedAt.After(tasks[j].CreatedAt)
	})

	return &model.ProjectWithTasks{
		Project: *project,
		Tasks:   tasks,
	}, nil
}

func (r *inMemoryProjectRepo) CanAccess(_ context.Context, projectID uuid.UUID, userID uuid.UUID) (bool, error) {
	project, ok := r.projects[projectID]
	if !ok {
		return false, nil
	}
	if project.OwnerID == userID {
		return true, nil
	}

	for _, task := range r.tasks {
		if task.ProjectID == projectID && task.AssigneeID != nil && *task.AssigneeID == userID {
			return true, nil
		}
	}

	return false, nil
}

func (r *inMemoryProjectRepo) List(_ context.Context, userID uuid.UUID, pagination model.PaginationParams) ([]model.Project, int, error) {
	all := make([]model.Project, 0)
	for _, p := range r.projects {
		if p.OwnerID == userID {
			all = append(all, *p)
			continue
		}

		accessibleByTask := false
		for _, task := range r.tasks {
			if task.ProjectID == p.ID && task.AssigneeID != nil && *task.AssigneeID == userID {
				accessibleByTask = true
				break
			}
		}
		if accessibleByTask {
			all = append(all, *p)
		}
	}

	sort.Slice(all, func(i, j int) bool {
		return all[i].CreatedAt.After(all[j].CreatedAt)
	})

	total := len(all)
	start := pagination.Offset()
	if start > total {
		return []model.Project{}, total, nil
	}
	end := start + pagination.Limit
	if end > total {
		end = total
	}
	return all[start:end], total, nil
}

func (r *inMemoryProjectRepo) GetStats(_ context.Context, projectID uuid.UUID) (*model.ProjectStats, error) {
	stats := &model.ProjectStats{
		ByStatus: map[string]int{
			string(model.TaskStatusTodo):       0,
			string(model.TaskStatusInProgress): 0,
			string(model.TaskStatusDone):       0,
		},
		ByAssignee: make([]model.AssigneeTaskCount, 0),
	}

	assigneeCounts := make(map[uuid.UUID]int)
	unassignedCount := 0

	for _, task := range r.tasks {
		if task.ProjectID != projectID {
			continue
		}
		stats.ByStatus[string(task.Status)]++
		if task.AssigneeID == nil {
			unassignedCount++
		} else {
			assigneeCounts[*task.AssigneeID]++
		}
	}

	for assigneeID, count := range assigneeCounts {
		id := assigneeID
		stats.ByAssignee = append(stats.ByAssignee, model.AssigneeTaskCount{
			AssigneeID: &id,
			Count:      count,
		})
	}

	if unassignedCount > 0 {
		stats.ByAssignee = append(stats.ByAssignee, model.AssigneeTaskCount{
			AssigneeID: nil,
			Count:      unassignedCount,
		})
	}

	return stats, nil
}

func (r *inMemoryProjectRepo) Update(_ context.Context, project *model.Project) error {
	if _, ok := r.projects[project.ID]; !ok {
		return nil
	}
	cp := *project
	r.projects[project.ID] = &cp
	return nil
}

func (r *inMemoryProjectRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(r.projects, id)
	for taskID, task := range r.tasks {
		if task.ProjectID == id {
			delete(r.tasks, taskID)
		}
	}
	return nil
}

type inMemoryTaskRepo struct {
	tasks map[uuid.UUID]*model.Task
}

func newInMemoryTaskRepo(shared map[uuid.UUID]*model.Task) *inMemoryTaskRepo {
	return &inMemoryTaskRepo{tasks: shared}
}

func (r *inMemoryTaskRepo) Create(_ context.Context, task *model.Task) error {
	t := cloneTask(*task)
	r.tasks[t.ID] = &t
	return nil
}

func (r *inMemoryTaskRepo) FindByID(_ context.Context, id uuid.UUID) (*model.Task, error) {
	task, ok := r.tasks[id]
	if !ok {
		return nil, nil
	}
	cp := cloneTask(*task)
	return &cp, nil
}

func (r *inMemoryTaskRepo) ListByProject(_ context.Context, projectID uuid.UUID, filter model.TaskFilter, pagination model.PaginationParams) ([]model.Task, int, error) {
	all := make([]model.Task, 0)
	for _, task := range r.tasks {
		if task.ProjectID != projectID {
			continue
		}
		if filter.Status != nil && task.Status != *filter.Status {
			continue
		}
		if filter.Assignee != nil {
			if task.AssigneeID == nil || *task.AssigneeID != *filter.Assignee {
				continue
			}
		}
		all = append(all, cloneTask(*task))
	}

	sort.Slice(all, func(i, j int) bool {
		return all[i].CreatedAt.After(all[j].CreatedAt)
	})

	total := len(all)
	start := pagination.Offset()
	if start > total {
		return []model.Task{}, total, nil
	}
	end := start + pagination.Limit
	if end > total {
		end = total
	}
	return all[start:end], total, nil
}

func (r *inMemoryTaskRepo) Update(_ context.Context, task *model.Task) error {
	if _, ok := r.tasks[task.ID]; !ok {
		return nil
	}
	t := cloneTask(*task)
	r.tasks[t.ID] = &t
	return nil
}

func (r *inMemoryTaskRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(r.tasks, id)
	return nil
}

func cloneTask(task model.Task) model.Task {
	cp := task
	if task.AssigneeID != nil {
		assignee := *task.AssigneeID
		cp.AssigneeID = &assignee
	}
	if task.DueDate != nil {
		due := task.DueDate.UTC()
		cp.DueDate = &due
	}
	return cp
}

func setupProtectedRouter(secret string, projectRepo repository.ProjectRepository, taskRepo repository.TaskRepository) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	projectService := service.NewProjectService(projectRepo)
	taskService := service.NewTaskService(taskRepo, projectRepo)

	projectHandler := handler.NewProjectHandler(projectService)
	taskHandler := handler.NewTaskHandler(taskService)

	protected := router.Group("/")
	protected.Use(middleware.Auth(secret))
	{
		protected.GET("/projects", projectHandler.List)
		protected.GET("/projects/:id", projectHandler.Get)
		protected.DELETE("/projects/:id", projectHandler.Delete)
		protected.DELETE("/tasks/:id", taskHandler.Delete)
	}

	return router
}

func authHeaderForUser(t *testing.T, secret string, userID uuid.UUID, email string) string {
	t.Helper()

	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"email":   email,
		"exp":     time.Now().Add(1 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	return "Bearer " + signed
}

func TestProtectedRoutes_RequireValidJWT(t *testing.T) {
	projectRepo := newInMemoryProjectRepo()
	taskRepo := newInMemoryTaskRepo(projectRepo.tasks)
	router := setupProtectedRouter("test-secret", projectRepo, taskRepo)

	tests := []struct {
		name    string
		auth    string
		wantErr string
	}{
		{name: "missing token", auth: "", wantErr: "unauthorized"},
		{name: "invalid token", auth: "Bearer invalid.token.value", wantErr: "unauthorized"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, "/projects", nil)
			if tt.auth != "" {
				req.Header.Set("Authorization", tt.auth)
			}

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("expected status %d, got %d body=%s", http.StatusUnauthorized, rec.Code, rec.Body.String())
			}

			if !strings.Contains(rec.Body.String(), tt.wantErr) {
				t.Fatalf("expected response body to contain %q, got %s", tt.wantErr, rec.Body.String())
			}
		})
	}
}

func TestProjectGet_ForbiddenForInaccessibleUser(t *testing.T) {
	ownerID := uuid.New()
	outsiderID := uuid.New()
	projectID := uuid.New()

	projectRepo := newInMemoryProjectRepo()
	taskRepo := newInMemoryTaskRepo(projectRepo.tasks)
	projectRepo.projects[projectID] = &model.Project{
		ID:        projectID,
		Name:      "Roadmap",
		OwnerID:   ownerID,
		CreatedAt: time.Now().UTC(),
	}

	router := setupProtectedRouter("test-secret", projectRepo, taskRepo)

	req, _ := http.NewRequest(http.MethodGet, "/projects/"+projectID.String(), nil)
	req.Header.Set("Authorization", authHeaderForUser(t, "test-secret", outsiderID, "outsider@example.com"))

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusForbidden, rec.Code, rec.Body.String())
	}
}

func TestProjectDelete_ForbiddenForNonOwner(t *testing.T) {
	ownerID := uuid.New()
	otherUserID := uuid.New()
	projectID := uuid.New()

	projectRepo := newInMemoryProjectRepo()
	taskRepo := newInMemoryTaskRepo(projectRepo.tasks)
	projectRepo.projects[projectID] = &model.Project{
		ID:        projectID,
		Name:      "Core API",
		OwnerID:   ownerID,
		CreatedAt: time.Now().UTC(),
	}

	router := setupProtectedRouter("test-secret", projectRepo, taskRepo)

	req, _ := http.NewRequest(http.MethodDelete, "/projects/"+projectID.String(), nil)
	req.Header.Set("Authorization", authHeaderForUser(t, "test-secret", otherUserID, "dev@example.com"))

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusForbidden, rec.Code, rec.Body.String())
	}
	if _, ok := projectRepo.projects[projectID]; !ok {
		t.Fatal("project should remain when delete is forbidden")
	}
}

func TestTaskDelete_AuthzRules(t *testing.T) {
	projectOwnerID := uuid.New()
	taskCreatorID := uuid.New()
	outsiderID := uuid.New()
	projectID := uuid.New()
	taskID := uuid.New()
	now := time.Now().UTC()

	projectRepo := newInMemoryProjectRepo()
	taskRepo := newInMemoryTaskRepo(projectRepo.tasks)

	projectRepo.projects[projectID] = &model.Project{
		ID:        projectID,
		Name:      "Backlog",
		OwnerID:   projectOwnerID,
		CreatedAt: now,
	}
	projectRepo.tasks[taskID] = &model.Task{
		ID:        taskID,
		Title:     "Ship endpoint",
		Status:    model.TaskStatusTodo,
		Priority:  model.TaskPriorityHigh,
		ProjectID: projectID,
		CreatorID: taskCreatorID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	router := setupProtectedRouter("test-secret", projectRepo, taskRepo)

	// Outsider cannot delete
	forbiddenReq, _ := http.NewRequest(http.MethodDelete, "/tasks/"+taskID.String(), nil)
	forbiddenReq.Header.Set("Authorization", authHeaderForUser(t, "test-secret", outsiderID, "outsider@example.com"))

	forbiddenRec := httptest.NewRecorder()
	router.ServeHTTP(forbiddenRec, forbiddenReq)

	if forbiddenRec.Code != http.StatusForbidden {
		t.Fatalf("expected status %d for outsider, got %d body=%s", http.StatusForbidden, forbiddenRec.Code, forbiddenRec.Body.String())
	}

	// Project owner can delete
	ownerReq, _ := http.NewRequest(http.MethodDelete, "/tasks/"+taskID.String(), nil)
	ownerReq.Header.Set("Authorization", authHeaderForUser(t, "test-secret", projectOwnerID, "owner@example.com"))

	ownerRec := httptest.NewRecorder()
	router.ServeHTTP(ownerRec, ownerReq)

	if ownerRec.Code != http.StatusNoContent {
		t.Fatalf("expected status %d for owner delete, got %d body=%s", http.StatusNoContent, ownerRec.Code, ownerRec.Body.String())
	}

	if _, ok := taskRepo.tasks[taskID]; ok {
		t.Fatal("task should be deleted by project owner")
	}
}
