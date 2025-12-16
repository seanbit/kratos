package biz

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/golang-jwt/jwt/v4"
	"github.com/pkg/errors"
	"github.com/seanbit/kratos/template/internal/conf"
	"github.com/seanbit/kratos/template/internal/data/model"
	"github.com/seanbit/kratos/template/internal/static"
	"github.com/seanbit/kratos/template/pkg/web3"
	"github.com/seanbit/kratos/webkit"
	"github.com/seanbit/kratos/webkit/cryptos"
	"github.com/segmentio/ksuid"
)

const (
	AuthSignatureExpiresDuration = time.Minute * 3
)

type IAuthRepo interface {
	GetUserAuthInfo(ctx context.Context, authType, authInfo string) (*model.UserAuthInfo, error)
	GetUserAuthInfoByAuthType(ctx context.Context, userId, authType string) (*model.UserAuthInfo, error)
	SetUserAuthInfo(ctx context.Context, userAuthInfo *model.UserAuthInfo) error
}

type UserLoginLog struct {
	UserId     string
	AuthType   string
	LoginIp    string
	LoginTime  time.Time
	IssueToken string
}

type IAuthLogRepo interface {
	PublishUserLoginEvent(ctx context.Context, userLoginLog *UserLoginLog) error
	SaveUserLoginLog(ctx context.Context, userLoginLog *model.UserLoginLog) error
}

type Auth struct {
	authRepo    IAuthRepo
	authLogRepo IAuthLogRepo
	geoIp       IGeoIp
	jwtKey      struct {
		private ed25519.PrivateKey
		public  ed25519.PublicKey
	}
	config *conf.Auth
}

func NewAuth(config *conf.Auth, authRepo IAuthRepo, authLogRepo IAuthLogRepo, geoIp IGeoIp) *Auth {
	privateKey, publicKey, err := web3.LoadEd25519Keys(config.JwtKey_25519, web3.Ed25519KeyPairEncodeHex)
	if err != nil {
		panic(fmt.Sprintf("Failed to load keys: %v\n", err))
	}
	return &Auth{
		authRepo:    authRepo,
		authLogRepo: authLogRepo,
		geoIp:       geoIp,
		jwtKey: struct {
			private ed25519.PrivateKey
			public  ed25519.PublicKey
		}{private: privateKey, public: publicKey},
		config: config,
	}
}

// LoginClaims defines the custom claims structure
type LoginClaims struct {
	*webkit.UserInfo
	jwt.RegisteredClaims
}

type LoginInfo struct {
	UserInfo *webkit.UserInfo
	Token    string
}

func (biz *Auth) LoginByWallet(ctx context.Context, blockchainType, originText, signature, address string) (*LoginInfo, error) {
	if err := biz.VerifyLoginSignature(ctx, blockchainType, originText, signature, address); err != nil {
		return nil, err
	}

	var authType AuthType
	switch blockchainType {
	case BlockChainTypeEvm:
		authType = AuthTypeWeb3WalletEvm
	default:
		return nil, ErrBlockChainTypeNotSupported
	}
	loginTime := time.Now()
	loginIp := webkit.GetRealIP(ctx)

	userAuthInfo, err := biz.authRepo.GetUserAuthInfo(ctx, authType, address)
	if err != nil {
		return nil, err
	}
	if userAuthInfo == nil {
		userAuthInfo = &model.UserAuthInfo{
			UserID:   ksuid.New().String(),
			AuthType: authType,
			AuthInfo: address,
		}
		if err := biz.authRepo.SetUserAuthInfo(ctx, userAuthInfo); err != nil {
			return nil, err
		}
	}
	loginInfo := &LoginInfo{
		UserInfo: &webkit.UserInfo{
			UserId:        userAuthInfo.UserID,
			Username:      "",
			UserType:      "",
			WalletAddress: address,
		},
	}
	loginInfo.Token, err = biz.GenerateToken(loginInfo.UserInfo, biz.config.LoginExpires.AsDuration())
	if err != nil {
		return nil, err
	}
	loginLog := &UserLoginLog{
		UserId:     userAuthInfo.UserID,
		AuthType:   authType,
		LoginIp:    loginIp,
		LoginTime:  loginTime,
		IssueToken: cryptos.MD5EncodeStringToHex(loginInfo.Token),
	}
	if err := biz.authLogRepo.PublishUserLoginEvent(ctx, loginLog); err != nil {
		log.Context(ctx).Errorf("Failed to publish user login event: %v", err)
	}
	return loginInfo, nil
}

func (biz *Auth) GetUserInfoByAuthToken(ctx context.Context, authToken string) (*webkit.UserInfo, error) {
	return biz.ValidateToken(ctx, authToken)
}

func (biz *Auth) GetLoginSignatureText(ctx context.Context, blockchainType, address string, expiresDuration time.Duration) string {
	switch blockchainType {
	case BlockChainTypeEvm:
		expirationTime := time.Now().UTC().Add(expiresDuration).Format("2006-01-02T15:04:05Z")
		signatureTextPrefix := fmt.Sprintf(static.LoginSignaturePrefixTextFormat, static.LoginDomain)
		return fmt.Sprintf(static.LoginSignatureTextFormat, signatureTextPrefix, address, expirationTime)
	default:
		return ErrBlockChainTypeNotSupported.String()
	}
}

func (biz *Auth) VerifyLoginSignature(ctx context.Context, blockchainType, originText, signature, address string) error {
	if err := biz.CheckLoginSignatureText(blockchainType, originText, address); err != nil {
		return err
	}
	switch blockchainType {
	case BlockChainTypeEvm:
		return web3.VerifyEthereumSignature(ctx, originText, signature, address)
	default:
		return ErrBlockChainTypeNotSupported
	}
}

func (biz *Auth) CheckLoginSignatureText(blockchainType, originText, address string) error {
	switch blockchainType {
	case BlockChainTypeEvm:
		return biz.CheckEthereumLoginSignatureText(originText, address)
	default:
		return ErrBlockChainTypeNotSupported
	}
}

func (biz *Auth) CheckEthereumLoginSignatureText(originText, address string) error {
	re := regexp.MustCompile(static.LoginSignatureTextPattern)
	matches := re.FindStringSubmatch(originText)
	if matches == nil || len(matches) < 3 {
		return ErrSignatureTextInvalid
	}

	// 提取动作内容
	paramsAddress := matches[1]
	if strings.ToLower(strings.TrimSpace(paramsAddress)) != strings.ToLower(strings.TrimSpace(address)) {
		return ErrSignatureTextExpired
	}

	// 解析时间字符串
	expirationTime, err := time.Parse("2006-01-02T15:04:05Z", matches[2])
	if err != nil {
		return ErrSignatureTextInvalid
	}
	now := time.Now()
	if expirationTime.Before(now) {
		return ErrSignatureTextExpired
	}
	return nil
}

func (biz *Auth) GetJwtPublicKeyHex() (string, error) {
	return hex.EncodeToString([]byte(biz.jwtKey.public)), nil
}

func (biz *Auth) GetJwtPublicKeyB64() (string, error) {
	return base64.StdEncoding.EncodeToString([]byte(biz.jwtKey.public)), nil
}

// GenerateToken generates a JWT for the given user information
func (biz *Auth) GenerateToken(userInfo *webkit.UserInfo, expiration time.Duration) (string, error) {
	claims := LoginClaims{
		UserInfo: userInfo,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    static.LoginDomain,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	return token.SignedString(biz.jwtKey.private)
}

// ValidateToken validates the given JWT and returns the user information
func (biz *Auth) ValidateToken(ctx context.Context, tokenString string) (*webkit.UserInfo, error) {
	token, err := jwt.ParseWithClaims(tokenString, &LoginClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Ensure the signing method is EdDSA
		if _, ok := token.Method.(*jwt.SigningMethodEd25519); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return biz.jwtKey.public, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*LoginClaims)
	if !ok || !token.Valid {
		return nil, ErrLoginTokenInvalid
	}
	if claims.Issuer != static.LoginDomain {
		return nil, ErrLoginTokenInvalid
	}
	if claims.VerifyExpiresAt(time.Now(), true) == false {
		return nil, ErrLoginExpired
	}
	return claims.UserInfo, nil
}

func (biz *Auth) SaveUserLoginLog(ctx context.Context, userLoginLog *UserLoginLog) error {
	country, err := biz.geoIp.GetCountryFromIp(ctx, userLoginLog.LoginIp)
	if err != nil {
		log.Context(ctx).Errorf("get country by ip: %s error: %v", userLoginLog.LoginIp, err)
	}
	return biz.authLogRepo.SaveUserLoginLog(ctx, &model.UserLoginLog{
		UserID:      userLoginLog.UserId,
		AuthType:    userLoginLog.AuthType,
		IssueToken:  userLoginLog.IssueToken,
		IP:          userLoginLog.LoginIp,
		Country:     country.IsoCode,
		CreatedTime: userLoginLog.LoginTime,
	})
}
