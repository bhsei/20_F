// Copyright 2014 The Gogs Authors. All rights reserved.
// Copyright 2018 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package auth

import (
	"mime/multipart"
	"strings"

	"code.gitea.io/gitea/modules/setting"

	"gitea.com/macaron/binding"
	"gitea.com/macaron/macaron"
)

// InstallForm form for installation page
type InstallForm struct {
	DbType   string `binding:"Required"`
	DbHost   string
	DbUser   string
	DbPasswd string
	DbName   string
	SSLMode  string
	Charset  string `binding:"Required;In(utf8,utf8mb4)"`
	DbPath   string

	AppName      string `binding:"Required" locale:"install.app_name"`
	RepoRootPath string `binding:"Required"`
	LFSRootPath  string
	RunUser      string `binding:"Required"`
	Domain       string `binding:"Required"`
	SSHPort      int
	HTTPPort     string `binding:"Required"`
	AppURL       string `binding:"Required"`
	LogRootPath  string `binding:"Required"`

	SMTPHost        string
	SMTPFrom        string
	SMTPUser        string `binding:"OmitEmpty;MaxSize(254)" locale:"install.mailer_user"`
	SMTPPasswd      string
	RegisterConfirm bool
	MailNotify      bool
	WechatNotify    bool

	ModHost string
	ModPort int

	WechatAppId     string
	WechatAppSecret string
	WechatToken     string

	OfflineMode                    bool
	DisableGravatar                bool
	EnableFederatedAvatar          bool
	EnableOpenIDSignIn             bool
	EnableOpenIDSignUp             bool
	DisableRegistration            bool
	AllowOnlyExternalRegistration  bool
	EnableCaptcha                  bool
	RequireSignInView              bool
	DefaultKeepEmailPrivate        bool
	DefaultAllowCreateOrganization bool
	DefaultEnableTimetracking      bool
	NoReplyAddress                 string

	AdminName          string `binding:"OmitEmpty;AlphaDashDot;MaxSize(30)" locale:"install.admin_name"`
	AdminPasswd        string `binding:"OmitEmpty;MaxSize(255)" locale:"install.admin_password"`
	AdminConfirmPasswd string
	AdminEmail         string `binding:"OmitEmpty;MinSize(3);MaxSize(254);Include(@)" locale:"install.admin_email"`
}

// Validate validates the fields
func (f *InstallForm) Validate(ctx *macaron.Context, errs binding.Errors) binding.Errors {
	return validate(errs, ctx.Data, f, ctx.Locale)
}

//    _____   ____ _________________ ___
//   /  _  \ |    |   \__    ___/   |   \
//  /  /_\  \|    |   / |    | /    ~    \
// /    |    \    |  /  |    | \    Y    /
// \____|__  /______/   |____|  \___|_  /
//         \/                         \/

// RegisterForm form for registering
type RegisterForm struct {
	UserName           string `binding:"Required;AlphaDashDot;MaxSize(40)"`
	Email              string `binding:"Required;Email;MaxSize(254)"`
	Password           string `binding:"MaxSize(255)"`
	Retype             string
	GRecaptchaResponse string `form:"g-recaptcha-response"`
}

// Validate valideates the fields
func (f *RegisterForm) Validate(ctx *macaron.Context, errs binding.Errors) binding.Errors {
	return validate(errs, ctx.Data, f, ctx.Locale)
}

// IsEmailDomainWhitelisted validates that the email address
// provided by the user matches what has been configured .
// If the domain whitelist from the config is empty, it marks the
// email as whitelisted
func (f RegisterForm) IsEmailDomainWhitelisted() bool {
	if len(setting.Service.EmailDomainWhitelist) == 0 {
		return true
	}

	n := strings.LastIndex(f.Email, "@")
	if n <= 0 {
		return false
	}

	domain := strings.ToLower(f.Email[n+1:])

	for _, v := range setting.Service.EmailDomainWhitelist {
		if strings.ToLower(v) == domain {
			return true
		}
	}

	return false
}

// MustChangePasswordForm form for updating your password after account creation
// by an admin
type MustChangePasswordForm struct {
	Password string `binding:"Required;MaxSize(255)"`
	Retype   string
}

// Validate valideates the fields
func (f *MustChangePasswordForm) Validate(ctx *macaron.Context, errs binding.Errors) binding.Errors {
	return validate(errs, ctx.Data, f, ctx.Locale)
}

// SignInForm form for signing in with user/password
type SignInForm struct {
	UserName string `binding:"Required;MaxSize(254)"`
	// TODO remove required from password for SecondFactorAuthentication
	Password string `binding:"Required;MaxSize(255)"`
	Remember bool
}

// Validate valideates the fields
func (f *SignInForm) Validate(ctx *macaron.Context, errs binding.Errors) binding.Errors {
	return validate(errs, ctx.Data, f, ctx.Locale)
}

// AuthorizationForm form for authorizing oauth2 clients
type AuthorizationForm struct {
	ResponseType string `binding:"Required;In(code)"`
	ClientID     string `binding:"Required"`
	RedirectURI  string
	State        string

	// PKCE support
	CodeChallengeMethod string // S256, plain
	CodeChallenge       string
}

// Validate valideates the fields
func (f *AuthorizationForm) Validate(ctx *macaron.Context, errs binding.Errors) binding.Errors {
	return validate(errs, ctx.Data, f, ctx.Locale)
}

// GrantApplicationForm form for authorizing oauth2 clients
type GrantApplicationForm struct {
	ClientID    string `binding:"Required"`
	RedirectURI string
	State       string
}

// Validate valideates the fields
func (f *GrantApplicationForm) Validate(ctx *macaron.Context, errs binding.Errors) binding.Errors {
	return validate(errs, ctx.Data, f, ctx.Locale)
}

// AccessTokenForm for issuing access tokens from authorization codes or refresh tokens
type AccessTokenForm struct {
	GrantType    string `json:"grant_type"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RedirectURI  string `json:"redirect_uri"`
	Code         string `json:"code"`
	RefreshToken string `json:"refresh_token"`

	// PKCE support
	CodeVerifier string `json:"code_verifier"`
}

// Validate valideates the fields
func (f *AccessTokenForm) Validate(ctx *macaron.Context, errs binding.Errors) binding.Errors {
	return validate(errs, ctx.Data, f, ctx.Locale)
}

//   __________________________________________.___ _______    ________  _________
//  /   _____/\_   _____/\__    ___/\__    ___/|   |\      \  /  _____/ /   _____/
//  \_____  \  |    __)_   |    |     |    |   |   |/   |   \/   \  ___ \_____  \
//  /        \ |        \  |    |     |    |   |   /    |    \    \_\  \/        \
// /_______  //_______  /  |____|     |____|   |___\____|__  /\______  /_______  /
//         \/         \/                                   \/        \/        \/

// UpdateProfileForm form for updating profile
type UpdateProfileForm struct {
	Name             string `binding:"AlphaDashDot;MaxSize(40)"`
	FullName         string `binding:"MaxSize(100)"`
	Email            string `binding:"Required;Email;MaxSize(254)"`
	KeepEmailPrivate bool
	Website          string `binding:"ValidUrl;MaxSize(255)"`
	Location         string `binding:"MaxSize(50)"`
	Language         string `binding:"Size(5)"`
	Description      string `binding:"MaxSize(255)"`
}

// Validate validates the fields
func (f *UpdateProfileForm) Validate(ctx *macaron.Context, errs binding.Errors) binding.Errors {
	return validate(errs, ctx.Data, f, ctx.Locale)
}

// Avatar types
const (
	AvatarLocal  string = "local"
	AvatarByMail string = "bymail"
)

// AvatarForm form for changing avatar
type AvatarForm struct {
	Source      string
	Avatar      *multipart.FileHeader
	Gravatar    string `binding:"OmitEmpty;Email;MaxSize(254)"`
	Federavatar bool
}

// Validate validates the fields
func (f *AvatarForm) Validate(ctx *macaron.Context, errs binding.Errors) binding.Errors {
	return validate(errs, ctx.Data, f, ctx.Locale)
}

// AddEmailForm form for adding new email
type AddEmailForm struct {
	Email string `binding:"Required;Email;MaxSize(254)"`
}

// Validate validates the fields
func (f *AddEmailForm) Validate(ctx *macaron.Context, errs binding.Errors) binding.Errors {
	return validate(errs, ctx.Data, f, ctx.Locale)
}

// UpdateThemeForm form for updating a users' theme
type UpdateThemeForm struct {
	Theme string `binding:"Required;MaxSize(30)"`
}

// Validate validates the field
func (f *UpdateThemeForm) Validate(ctx *macaron.Context, errs binding.Errors) binding.Errors {
	return validate(errs, ctx.Data, f, ctx.Locale)
}

// IsThemeExists checks if the theme is a theme available in the config.
func (f UpdateThemeForm) IsThemeExists() bool {
	var exists bool

	for _, v := range setting.UI.Themes {
		if strings.EqualFold(v, f.Theme) {
			exists = true
			break
		}
	}

	return exists
}

// ChangePasswordForm form for changing password
type ChangePasswordForm struct {
	OldPassword string `form:"old_password" binding:"MaxSize(255)"`
	Password    string `form:"password" binding:"Required;MaxSize(255)"`
	Retype      string `form:"retype"`
}

// Validate validates the fields
func (f *ChangePasswordForm) Validate(ctx *macaron.Context, errs binding.Errors) binding.Errors {
	return validate(errs, ctx.Data, f, ctx.Locale)
}

// AddOpenIDForm is for changing openid uri
type AddOpenIDForm struct {
	Openid string `binding:"Required;MaxSize(256)"`
}

// Validate validates the fields
func (f *AddOpenIDForm) Validate(ctx *macaron.Context, errs binding.Errors) binding.Errors {
	return validate(errs, ctx.Data, f, ctx.Locale)
}

// AddKeyForm form for adding SSH/GPG key
type AddKeyForm struct {
	Type       string `binding:"OmitEmpty"`
	Title      string `binding:"Required;MaxSize(50)"`
	Content    string `binding:"Required"`
	IsWritable bool
}

// Validate validates the fields
func (f *AddKeyForm) Validate(ctx *macaron.Context, errs binding.Errors) binding.Errors {
	return validate(errs, ctx.Data, f, ctx.Locale)
}

// NewAccessTokenForm form for creating access token
type NewAccessTokenForm struct {
	Name string `binding:"Required;MaxSize(255)"`
}

// Validate valideates the fields
func (f *NewAccessTokenForm) Validate(ctx *macaron.Context, errs binding.Errors) binding.Errors {
	return validate(errs, ctx.Data, f, ctx.Locale)
}

// EditOAuth2ApplicationForm form for editing oauth2 applications
type EditOAuth2ApplicationForm struct {
	Name        string `binding:"Required;MaxSize(255)" form:"application_name"`
	RedirectURI string `binding:"Required" form:"redirect_uri"`
}

// Validate valideates the fields
func (f *EditOAuth2ApplicationForm) Validate(ctx *macaron.Context, errs binding.Errors) binding.Errors {
	return validate(errs, ctx.Data, f, ctx.Locale)
}

// TwoFactorAuthForm for logging in with 2FA token.
type TwoFactorAuthForm struct {
	Passcode string `binding:"Required"`
}

// Validate validates the fields
func (f *TwoFactorAuthForm) Validate(ctx *macaron.Context, errs binding.Errors) binding.Errors {
	return validate(errs, ctx.Data, f, ctx.Locale)
}

// TwoFactorScratchAuthForm for logging in with 2FA scratch token.
type TwoFactorScratchAuthForm struct {
	Token string `binding:"Required"`
}

// Validate valideates the fields
func (f *TwoFactorScratchAuthForm) Validate(ctx *macaron.Context, errs binding.Errors) binding.Errors {
	return validate(errs, ctx.Data, f, ctx.Locale)
}

// U2FRegistrationForm for reserving an U2F name
type U2FRegistrationForm struct {
	Name string `binding:"Required"`
}

// Validate valideates the fields
func (f *U2FRegistrationForm) Validate(ctx *macaron.Context, errs binding.Errors) binding.Errors {
	return validate(errs, ctx.Data, f, ctx.Locale)
}

// U2FDeleteForm for deleting U2F keys
type U2FDeleteForm struct {
	ID int64 `binding:"Required"`
}

// Validate valideates the fields
func (f *U2FDeleteForm) Validate(ctx *macaron.Context, errs binding.Errors) binding.Errors {
	return validate(errs, ctx.Data, f, ctx.Locale)
}

type ValidateForm struct {
	Signature string `form:"signature" binding:"Required"`
	Timestamp string `form:"timestamp" binding:"Required"`
	Nonce     string `form:"nonce" binding:"Required"`
	Echostr   string `form:"echostr" binding:"Required"`
}
