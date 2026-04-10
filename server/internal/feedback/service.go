package feedback

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ruhuang/ink/server/internal/auth"
	"github.com/ruhuang/ink/server/internal/printer"
	"github.com/ruhuang/ink/server/internal/user"
	"github.com/ruhuang/ink/server/internal/workspace"
)

var (
	ErrInvalidInput     = errors.New("invalid feedback input")
	ErrNoAdminRecipient = errors.New("feedback admin recipient unavailable")
	ErrNoAdminDevice    = errors.New("feedback admin printer unavailable")
)

type Authenticator interface {
	GetCurrentUser(ctx context.Context, accessToken string) (auth.UserDTO, error)
}

type AdminRepository interface {
	FindPrimaryAdmin(ctx context.Context) (*user.User, error)
}

type WorkspaceRepository interface {
	FindByUserID(ctx context.Context, userID string) (*workspace.State, error)
}

type PrinterRepository interface {
	ListBindingsByUserID(ctx context.Context, userID string) ([]printer.Binding, error)
}

type PrinterJobCreator interface {
	CreatePrintJobForUser(ctx context.Context, userID string, input printer.CreateJobInput) (workspace.PrintJob, error)
}

type Clock interface {
	Now() time.Time
}

type SubmitInput struct {
	Content string `json:"content"`
}

type Service struct {
	auth       Authenticator
	admins     AdminRepository
	workspaces WorkspaceRepository
	printers   PrinterRepository
	jobs       PrinterJobCreator
	clock      Clock
}

func NewService(
	authenticator Authenticator,
	adminRepository AdminRepository,
	workspaceRepository WorkspaceRepository,
	printerRepository PrinterRepository,
	printerJobs PrinterJobCreator,
	clock Clock,
) *Service {
	return &Service{
		auth:       authenticator,
		admins:     adminRepository,
		workspaces: workspaceRepository,
		printers:   printerRepository,
		jobs:       printerJobs,
		clock:      clock,
	}
}

func (s *Service) Submit(ctx context.Context, accessToken string, input SubmitInput) error {
	currentUser, err := s.auth.GetCurrentUser(ctx, accessToken)
	if err != nil {
		return err
	}

	content := strings.TrimSpace(input.Content)
	if content == "" {
		return ErrInvalidInput
	}

	adminAccount, err := s.admins.FindPrimaryAdmin(ctx)
	if err != nil {
		return err
	}
	if adminAccount == nil {
		return ErrNoAdminRecipient
	}

	targetBindingID, err := s.resolveTargetBindingID(ctx, adminAccount.ID)
	if err != nil {
		return err
	}
	if targetBindingID == "" {
		return ErrNoAdminDevice
	}

	_, err = s.jobs.CreatePrintJobForUser(ctx, adminAccount.ID, printer.CreateJobInput{
		Title:             "用户反馈",
		Source:            "反馈（功能/建议/吐槽）",
		Content:           s.buildPrintContent(currentUser, content),
		PrinterBindingID:  targetBindingID,
		SubmitImmediately: true,
	})
	return err
}

func (s *Service) resolveTargetBindingID(ctx context.Context, adminUserID string) (string, error) {
	var preferredBindingID string

	state, err := s.workspaces.FindByUserID(ctx, adminUserID)
	if err != nil {
		return "", err
	}
	if state != nil {
		preferredBindingID = strings.TrimSpace(state.Preferences.DefaultDeviceID)
	}

	bindings, err := s.printers.ListBindingsByUserID(ctx, adminUserID)
	if err != nil {
		return "", err
	}

	if preferredBindingID != "" {
		for _, binding := range bindings {
			if binding.ID == preferredBindingID && binding.Status == workspace.DeviceStatusConnected {
				return binding.ID, nil
			}
		}
	}

	for _, binding := range bindings {
		if binding.Status == workspace.DeviceStatusConnected {
			return binding.ID, nil
		}
	}

	return "", nil
}

func (s *Service) buildPrintContent(sender auth.UserDTO, body string) string {
	submittedAt := s.clock.Now().Format("2006-01-02 15:04:05")
	senderName := strings.TrimSpace(sender.Name)
	if senderName == "" {
		senderName = sender.Email
	}

	return strings.Join([]string{
		"Ink 用户反馈",
		fmt.Sprintf("提交人：%s", senderName),
		fmt.Sprintf("账号：%s", sender.Email),
		fmt.Sprintf("角色：%s", sender.Role),
		fmt.Sprintf("时间：%s", submittedAt),
		"",
		"内容：",
		body,
	}, "\n")
}
