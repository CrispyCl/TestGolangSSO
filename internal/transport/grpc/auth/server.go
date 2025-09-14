package authgrpc

import (
	"context"
	"errors"
	"strings"

	"auth/internal/repository"
	"auth/internal/services/auth"

	ssov1 "github.com/CrispyCl/testprotos/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type GRPCServer struct {
	ssov1.UnimplementedAuthServer
	authServ AuthService
}

type AuthService interface {
	Login(ctx context.Context, email, password string, appID int, ip, userAgent string) (string, string, error)
	Register(ctx context.Context, email, password string) (userID int64, err error)
	Refresh(ctx context.Context, refreshToken string) (newAccess, newRefresh string, err error)
}

func Register(gRPCServer *grpc.Server, auth AuthService) {
	ssov1.RegisterAuthServer(gRPCServer, &GRPCServer{authServ: auth})
}

func extractMeta(ctx context.Context) (ip, ua string) {
	if p, ok := peer.FromContext(ctx); ok {
		ip = strings.Split(p.Addr.String(), ":")[0]
	}
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if uaVals := md.Get("user-agent"); len(uaVals) > 0 {
			ua = uaVals[0]
		}
	}
	return
}

func (s *GRPCServer) Login(
	ctx context.Context,
	req *ssov1.LoginRequest,
) (*ssov1.TokenPairResponse, error) {
	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	if req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

	if req.GetAppId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "app_id is required")
	}

	ip, ua := extractMeta(ctx)

	access, refresh, err := s.authServ.Login(ctx, req.Email, req.Password, int(req.AppId), ip, ua)

	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "invalid email or password")
		}

		return nil, status.Error(codes.Internal, "failed to login")
	}

	return &ssov1.TokenPairResponse{AccessToken: access, RefreshToken: refresh}, nil
}

func (s *GRPCServer) Register(
	ctx context.Context,
	req *ssov1.RegisterRequest,
) (*ssov1.RegisterResponse, error) {
	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	if req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

	uid, err := s.authServ.Register(ctx, req.Email, req.Password)

	if err != nil {
		if errors.Is(err, repository.ErrUserExists) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}

		return nil, status.Error(codes.Internal, "failed to register user")
	}

	return &ssov1.RegisterResponse{UserId: uid}, nil
}

func (s *GRPCServer) Refresh(ctx context.Context, req *ssov1.RefreshTokenRequest) (*ssov1.TokenPairResponse, error) {
	if req.RefreshToken == "" {
		return nil, status.Error(codes.InvalidArgument, "refresh_token is required")
	}

	access, refresh, err := s.authServ.Refresh(ctx, req.RefreshToken)

	if err != nil {
		return nil, status.Error(codes.Internal, "failed to refresh token")
	}

	return &ssov1.TokenPairResponse{AccessToken: access, RefreshToken: refresh}, nil
}
