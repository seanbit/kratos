package service

import (
	"context"

	pb "github.com/seanbit/kratos/template/api/web"
	"github.com/seanbit/kratos/template/internal/biz"
)

type AuthService struct {
	pb.UnimplementedAuthServer
	authBiz *biz.Auth
}

func NewAuthService(authBiz *biz.Auth) *AuthService {
	return &AuthService{authBiz: authBiz}
}

func (s *AuthService) LoginByWallet(ctx context.Context, req *pb.LoginByWalletRequest) (*pb.LoginByWalletResponse, error) {
	loginInfo, err := s.authBiz.LoginByWallet(ctx, blockchainTypes[req.BlockchainType], req.OriginText, req.Signature, req.Address)
	if err != nil {
		return nil, err
	}
	return &pb.LoginByWalletResponse{Token: loginInfo.Token}, err
}
func (s *AuthService) GetLoginSignatureText(ctx context.Context, req *pb.GetLoginSignTextRequest) (*pb.GetLoginSignTextResponse, error) {
	text := s.authBiz.GetLoginSignatureText(ctx, blockchainTypes[req.BlockchainType], req.Address, biz.AuthSignatureExpiresDuration)
	return &pb.GetLoginSignTextResponse{Text: text}, nil
}
