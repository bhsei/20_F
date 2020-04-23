// Copyright 2014 The Gogs Authors. All rights reserved.
// Copyright 2017 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package setting

import (
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"code.gitea.io/gitea/modules/generate"
	"code.gitea.io/gitea/modules/git"
	"code.gitea.io/gitea/modules/log"
	"code.gitea.io/gitea/modules/user"

	shellquote "github.com/kballard/go-shellquote"
	version "github.com/mcuadros/go-version"
	"github.com/unknwon/cae/zip"
	"github.com/unknwon/com"
	ini "gopkg.in/ini.v1"
	"strk.kbt.io/projects/go/libravatar"
)

// Scheme describes protocol types
type Scheme string

// enumerates all the scheme types
const (
	HTTP       Scheme = "http"
	HTTPS      Scheme = "https"
	FCGI       Scheme = "fcgi"
	FCGIUnix   Scheme = "fcgi+unix"
	UnixSocket Scheme = "unix"
)

// LandingPage describes the default page
type LandingPage string

// enumerates all the landing page types
const (
	LandingPageHome          LandingPage = "/"
	LandingPageExplore       LandingPage = "/explore"
	LandingPageOrganizations LandingPage = "/explore/organizations"
	LandingPageLogin         LandingPage = "/user/login"
)

// enumerates all the types of captchas
const (
	ImageCaptcha = "image"
	ReCaptcha    = "recaptcha"
)

// settings
var (
	// AppVer settings
	AppVer         string
	AppBuiltWith   string
	AppName        string
	AppURL         string
	AppSubURL      string
	AppSubURLDepth int // Number of slashes
	AppPath        string
	AppDataPath    string
	AppWorkPath    string

	// Server settings
	Protocol             Scheme
	Domain               string
	HTTPAddr             string
	HTTPPort             string
	LocalURL             string
	RedirectOtherPort    bool
	PortToRedirect       string
	OfflineMode          bool
	CertFile             string
	KeyFile              string
	StaticRootPath       string
	StaticCacheTime      time.Duration
	EnableGzip           bool
	LandingPageURL       LandingPage
	UnixSocketPermission uint32
	EnablePprof          bool
	PprofDataPath        string
	EnableLetsEncrypt    bool
	LetsEncryptTOS       bool
	LetsEncryptDirectory string
	LetsEncryptEmail     string
	GracefulRestartable  bool
	GracefulHammerTime   time.Duration
	StartupTimeout       time.Duration
	StaticURLPrefix      string

	SSH = struct {
		Disabled                 bool           `ini:"DISABLE_SSH"`
		StartBuiltinServer       bool           `ini:"START_SSH_SERVER"`
		BuiltinServerUser        string         `ini:"BUILTIN_SSH_SERVER_USER"`
		Domain                   string         `ini:"SSH_DOMAIN"`
		Port                     int            `ini:"SSH_PORT"`
		ListenHost               string         `ini:"SSH_LISTEN_HOST"`
		ListenPort               int            `ini:"SSH_LISTEN_PORT"`
		RootPath                 string         `ini:"SSH_ROOT_PATH"`
		ServerCiphers            []string       `ini:"SSH_SERVER_CIPHERS"`
		ServerKeyExchanges       []string       `ini:"SSH_SERVER_KEY_EXCHANGES"`
		ServerMACs               []string       `ini:"SSH_SERVER_MACS"`
		KeyTestPath              string         `ini:"SSH_KEY_TEST_PATH"`
		KeygenPath               string         `ini:"SSH_KEYGEN_PATH"`
		AuthorizedKeysBackup     bool           `ini:"SSH_AUTHORIZED_KEYS_BACKUP"`
		MinimumKeySizeCheck      bool           `ini:"-"`
		MinimumKeySizes          map[string]int `ini:"-"`
		CreateAuthorizedKeysFile bool           `ini:"SSH_CREATE_AUTHORIZED_KEYS_FILE"`
		ExposeAnonymous          bool           `ini:"SSH_EXPOSE_ANONYMOUS"`
	}{
		Disabled:           false,
		StartBuiltinServer: false,
		Domain:             "",
		Port:               22,
		ServerCiphers:      []string{"aes128-ctr", "aes192-ctr", "aes256-ctr", "aes128-gcm@openssh.com", "arcfour256", "arcfour128"},
		ServerKeyExchanges: []string{"diffie-hellman-group1-sha1", "diffie-hellman-group14-sha1", "ecdh-sha2-nistp256", "ecdh-sha2-nistp384", "ecdh-sha2-nistp521", "curve25519-sha256@libssh.org"},
		ServerMACs:         []string{"hmac-sha2-256-etm@openssh.com", "hmac-sha2-256", "hmac-sha1", "hmac-sha1-96"},
		KeygenPath:         "ssh-keygen",
		MinimumKeySizes:    map[string]int{"ed25519": 256, "ecdsa": 256, "rsa": 2048, "dsa": 1024},
	}

	LFS struct {
		StartServer     bool          `ini:"LFS_START_SERVER"`
		ContentPath     string        `ini:"LFS_CONTENT_PATH"`
		JWTSecretBase64 string        `ini:"LFS_JWT_SECRET"`
		JWTSecretBytes  []byte        `ini:"-"`
		HTTPAuthExpiry  time.Duration `ini:"LFS_HTTP_AUTH_EXPIRY"`
	}

	// Security settings
	InstallLock                        bool
	SecretKey                          string
	LogInRememberDays                  int
	CookieUserName                     string
	CookieRememberName                 string
	ReverseProxyAuthUser               string
	ReverseProxyAuthEmail              string
	MinPasswordLength                  int
	ImportLocalPaths                   bool
	DisableGitHooks                    bool
	OnlyAllowPushIfGiteaEnvironmentSet bool
	PasswordComplexity                 []string
	PasswordHashAlgo                   string

	// UI settings
	UI = struct {
		ExplorePagingNum      int
		IssuePagingNum        int
		RepoSearchPagingNum   int
		MembersPagingNum      int
		FeedMaxCommitNum      int
		GraphMaxCommitNum     int
		CodeCommentLines      int
		ReactionMaxUserNum    int
		ThemeColorMetaTag     string
		MaxDisplayFileSize    int64
		ShowUserEmail         bool
		DefaultShowFullName   bool
		DefaultTheme          string
		Themes                []string
		Reactions             []string
		ReactionsMap          map[string]bool
		SearchRepoDescription bool
		UseServiceWorker      bool

		Admin struct {
			UserPagingNum   int
			RepoPagingNum   int
			NoticePagingNum int
			OrgPagingNum    int
		} `ini:"ui.admin"`
		User struct {
			RepoPagingNum int
		} `ini:"ui.user"`
		Meta struct {
			Author      string
			Description string
			Keywords    string
		} `ini:"ui.meta"`
	}{
		ExplorePagingNum:    20,
		IssuePagingNum:      10,
		RepoSearchPagingNum: 10,
		MembersPagingNum:    20,
		FeedMaxCommitNum:    5,
		GraphMaxCommitNum:   100,
		CodeCommentLines:    4,
		ReactionMaxUserNum:  10,
		ThemeColorMetaTag:   `#6cc644`,
		MaxDisplayFileSize:  8388608,
		DefaultTheme:        `gitea`,
		Themes:              []string{`gitea`, `arc-green`},
		Reactions:           []string{`+1`, `-1`, `laugh`, `hooray`, `confused`, `heart`, `rocket`, `eyes`},
		Admin: struct {
			UserPagingNum   int
			RepoPagingNum   int
			NoticePagingNum int
			OrgPagingNum    int
		}{
			UserPagingNum:   50,
			RepoPagingNum:   50,
			NoticePagingNum: 25,
			OrgPagingNum:    50,
		},
		User: struct {
			RepoPagingNum int
		}{
			RepoPagingNum: 15,
		},
		Meta: struct {
			Author      string
			Description string
			Keywords    string
		}{
			Author:      "Gitea - Git with a cup of tea",
			Description: "Gitea (Git with a cup of tea) is a painless self-hosted Git service written in Go",
			Keywords:    "go,git,self-hosted,gitea",
		},
	}

	// Markdown settings
	Markdown = struct {
		EnableHardLineBreak bool
		CustomURLSchemes    []string `ini:"CUSTOM_URL_SCHEMES"`
		FileExtensions      []string
	}{
		EnableHardLineBreak: false,
		FileExtensions:      strings.Split(".md,.markdown,.mdown,.mkd", ","),
	}

	// Admin settings
	Admin struct {
		DisableRegularOrgCreation bool
		DefaultEmailNotification  string
	}

	// Picture settings
	AvatarUploadPath              string
	AvatarMaxWidth                int
	AvatarMaxHeight               int
	GravatarSource                string
	GravatarSourceURL             *url.URL
	DisableGravatar               bool
	EnableFederatedAvatar         bool
	LibravatarService             *libravatar.Libravatar
	AvatarMaxFileSize             int64
	RepositoryAvatarUploadPath    string
	RepositoryAvatarFallback      string
	RepositoryAvatarFallbackImage string

	// Log settings
	LogLevel           string
	StacktraceLogLevel string
	LogRootPath        string
	LogDescriptions    = make(map[string]*LogDescription)
	RedirectMacaronLog bool
	DisableRouterLog   bool
	RouterLogLevel     log.Level
	RouterLogMode      string
	EnableAccessLog    bool
	AccessLogTemplate  string
	EnableXORMLog      bool

	// Attachment settings
	AttachmentPath         string
	AttachmentAllowedTypes string
	AttachmentMaxSize      int64
	AttachmentMaxFiles     int
	AttachmentEnabled      bool

	// Time settings
	TimeFormat string
	// UILocation is the location on the UI, so that we can display the time on UI.
	DefaultUILocation = time.Local

	CSRFCookieName     = "_csrf"
	CSRFCookieHTTPOnly = true

	// Mirror settings
	Mirror struct {
		DefaultInterval time.Duration
		MinInterval     time.Duration
	}

	// API settings
	API = struct {
		EnableSwagger          bool
		SwaggerURL             string
		MaxResponseItems       int
		DefaultPagingNum       int
		DefaultGitTreesPerPage int
		DefaultMaxBlobSize     int64
	}{
		EnableSwagger:          true,
		SwaggerURL:             "",
		MaxResponseItems:       50,
		DefaultPagingNum:       30,
		DefaultGitTreesPerPage: 1000,
		DefaultMaxBlobSize:     10485760,
	}

	OAuth2 = struct {
		Enable                     bool
		AccessTokenExpirationTime  int64
		RefreshTokenExpirationTime int64
		InvalidateRefreshTokens    bool
		JWTSecretBytes             []byte `ini:"-"`
		JWTSecretBase64            string `ini:"JWT_SECRET"`
	}{
		Enable:                     true,
		AccessTokenExpirationTime:  3600,
		RefreshTokenExpirationTime: 730,
		InvalidateRefreshTokens:    false,
	}

	U2F = struct {
		AppID         string
		TrustedFacets []string
	}{}

	// Metrics settings
	Metrics = struct {
		Enabled bool
		Token   string
	}{
		Enabled: false,
		Token:   "",
	}

	// I18n settings
	Langs     []string
	Names     []string
	dateLangs map[string]string

	// Highlight settings are loaded in modules/template/highlight.go

	// Other settings
	ShowFooterBranding         bool
	ShowFooterVersion          bool
	ShowFooterTemplateLoadTime bool

	// Global setting objects
	Cfg           *ini.File
	CustomPath    string // Custom directory path
	CustomConf    string
	CustomPID     string
	ProdMode      bool
	RunUser       string
	IsWindows     bool
	HasRobotsTxt  bool
	InternalToken string // internal access token

	// UILocation is the location on the UI, so that we can display the time on UI.
	// Currently only show the default time.Local, it could be added to app.ini after UI is ready
	UILocation = time.Local
)

// DateLang transforms standard language locale name to corresponding value in datetime plugin.
func DateLang(lang string) string {
	name, ok := dateLangs[lang]
	if ok {
		return name
	}
	return "en"
}

func getAppPath() (string, error) {
	var appPath string
	var err error
	if IsWindows && filepath.IsAbs(os.Args[0]) {
		appPath = filepath.Clean(os.Args[0])
	} else {
		appPath, err = exec.LookPath(os.Args[0])
	}

	if err != nil {
		return "", err
	}
	appPath, err = filepath.Abs(appPath)
	if err != nil {
		return "", err
	}
	// Note: we don't use path.Dir here because it does not handle case
	//	which path starts with two "/" in Windows: "//psf/Home/..."
	return strings.Replace(appPath, "\\", "/", -1), err
}

func getWorkPath(appPath string) string {
	workPath := AppWorkPath

	if giteaWorkPath, ok := os.LookupEnv("GITEA_WORK_DIR"); ok {
		workPath = giteaWorkPath
	}
	if len(workPath) == 0 {
		i := strings.LastIndex(appPath, "/")
		if i == -1 {
			workPath = appPath
		} else {
			workPath = appPath[:i]
		}
	}
	return strings.Replace(workPath, "\\", "/", -1)
}

func init() {
	IsWindows = runtime.GOOS == "windows"
	// We can rely on log.CanColorStdout being set properly because modules/log/console_windows.go comes before modules/setting/setting.go lexicographically
	log.NewLogger(0, "console", "console", fmt.Sprintf(`{"level": "trace", "colorize": %t, "stacktraceLevel": "none"}`, log.CanColorStdout))

	var err error
	if AppPath, err = getAppPath(); err != nil {
		log.Fatal("Failed to get app path: %v", err)
	}
	AppWorkPath = getWorkPath(AppPath)
}

func forcePathSeparator(path string) {
	if strings.Contains(path, "\\") {
		log.Fatal("Do not use '\\' or '\\\\' in paths, instead, please use '/' in all places")
	}
}

// IsRunUserMatchCurrentUser returns false if configured run user does not match
// actual user that runs the app. The first return value is the actual user name.
// This check is ignored under Windows since SSH remote login is not the main
// method to login on Windows.
func IsRunUserMatchCurrentUser(runUser string) (string, bool) {
	if IsWindows || SSH.StartBuiltinServer {
		return "", true
	}

	currentUser := user.CurrentUsername()
	return currentUser, runUser == currentUser
}

func createPIDFile(pidPath string) {
	currentPid := os.Getpid()
	if err := os.MkdirAll(filepath.Dir(pidPath), os.ModePerm); err != nil {
		log.Fatal("Failed to create PID folder: %v", err)
	}

	file, err := os.Create(pidPath)
	if err != nil {
		log.Fatal("Failed to create PID file: %v", err)
	}
	defer file.Close()
	if _, err := file.WriteString(strconv.FormatInt(int64(currentPid), 10)); err != nil {
		log.Fatal("Failed to write PID information: %v", err)
	}
}

// CheckLFSVersion will check lfs version, if not satisfied, then disable it.
func CheckLFSVersion() {
	if LFS.StartServer {
		//Disable LFS client hooks if installed for the current OS user
		//Needs at least git v2.1.2

		binVersion, err := git.BinVersion()
		if err != nil {
			log.Fatal("Error retrieving git version: %v", err)
		}

		if !version.Compare(binVersion, "2.1.2", ">=") {
			LFS.StartServer = false
			log.Error("LFS server support needs at least Git v2.1.2")
		} else {
			git.GlobalCommandArgs = append(git.GlobalCommandArgs, "-c", "filter.lfs.required=",
				"-c", "filter.lfs.smudge=", "-c", "filter.lfs.clean=")
		}
	}
}

// SetCustomPathAndConf will set CustomPath and CustomConf with reference to the
// GITEA_CUSTOM environment variable and with provided overrides before stepping
// back to the default
func SetCustomPathAndConf(providedCustom, providedConf, providedWorkPath string) {
	if len(providedWorkPath) != 0 {
		AppWorkPath = filepath.ToSlash(providedWorkPath)
	}
	if giteaCustom, ok := os.LookupEnv("GITEA_CUSTOM"); ok {
		CustomPath = giteaCustom
	}
	if len(providedCustom) != 0 {
		CustomPath = providedCustom
	}
	if len(CustomPath) == 0 {
		CustomPath = path.Join(AppWorkPath, "custom")
	} else if !filepath.IsAbs(CustomPath) {
		CustomPath = path.Join(AppWorkPath, CustomPath)
	}

	if len(providedConf) != 0 {
		CustomConf = providedConf
	}
	if len(CustomConf) == 0 {
		CustomConf = path.Join(CustomPath, "conf/app.ini")
	} else if !filepath.IsAbs(CustomConf) {
		CustomConf = path.Join(CustomPath, CustomConf)
	}
}

// NewContext initializes configuration context.
// NOTE: do not print any log except error.
func NewContext() {
	Cfg = ini.Empty()

	if len(CustomPID) > 0 {
		createPIDFile(CustomPID)
	}

	if com.IsFile(CustomConf) {
		if err := Cfg.Append(CustomConf); err != nil {
			log.Fatal("Failed to load custom conf '%s': %v", CustomConf, err)
		}
	} else {
		log.Warn("Custom config '%s' not found, ignore this if you're running first time", CustomConf)
	}
	Cfg.NameMapper = ini.SnackCase

	homeDir, err := com.HomeDir()
	if err != nil {
		log.Fatal("Failed to get home directory: %v", err)
	}
	homeDir = strings.Replace(homeDir, "\\", "/", -1)

	LogLevel = getLogLevel(Cfg.Section("log"), "LEVEL", "Info")
	StacktraceLogLevel = getStacktraceLogLevel(Cfg.Section("log"), "STACKTRACE_LEVEL", "None")
	LogRootPath = Cfg.Section("log").Key("ROOT_PATH").MustString(path.Join(AppWorkPath, "log"))
	forcePathSeparator(LogRootPath)
	RedirectMacaronLog = Cfg.Section("log").Key("REDIRECT_MACARON_LOG").MustBool(false)
	RouterLogLevel = log.FromString(Cfg.Section("log").Key("ROUTER_LOG_LEVEL").MustString("Info"))

	sec := Cfg.Section("server")
	AppName = Cfg.Section("").Key("APP_NAME").MustString("Gitea: Git with a cup of tea")

	Protocol = HTTP
	switch sec.Key("PROTOCOL").String() {
	case "https":
		Protocol = HTTPS
		CertFile = sec.Key("CERT_FILE").String()
		KeyFile = sec.Key("KEY_FILE").String()
		if !filepath.IsAbs(CertFile) && len(CertFile) > 0 {
			CertFile = filepath.Join(CustomPath, CertFile)
		}
		if !filepath.IsAbs(KeyFile) && len(KeyFile) > 0 {
			KeyFile = filepath.Join(CustomPath, KeyFile)
		}
	case "fcgi":
		Protocol = FCGI
	case "fcgi+unix":
		Protocol = FCGIUnix
		UnixSocketPermissionRaw := sec.Key("UNIX_SOCKET_PERMISSION").MustString("666")
		UnixSocketPermissionParsed, err := strconv.ParseUint(UnixSocketPermissionRaw, 8, 32)
		if err != nil || UnixSocketPermissionParsed > 0777 {
			log.Fatal("Failed to parse unixSocketPermission: %s", UnixSocketPermissionRaw)
		}
		UnixSocketPermission = uint32(UnixSocketPermissionParsed)
	case "unix":
		Protocol = UnixSocket
		UnixSocketPermissionRaw := sec.Key("UNIX_SOCKET_PERMISSION").MustString("666")
		UnixSocketPermissionParsed, err := strconv.ParseUint(UnixSocketPermissionRaw, 8, 32)
		if err != nil || UnixSocketPermissionParsed > 0777 {
			log.Fatal("Failed to parse unixSocketPermission: %s", UnixSocketPermissionRaw)
		}
		UnixSocketPermission = uint32(UnixSocketPermissionParsed)
	}
	EnableLetsEncrypt = sec.Key("ENABLE_LETSENCRYPT").MustBool(false)
	LetsEncryptTOS = sec.Key("LETSENCRYPT_ACCEPTTOS").MustBool(false)
	if !LetsEncryptTOS && EnableLetsEncrypt {
		log.Warn("Failed to enable Let's Encrypt due to Let's Encrypt TOS not being accepted")
		EnableLetsEncrypt = false
	}
	LetsEncryptDirectory = sec.Key("LETSENCRYPT_DIRECTORY").MustString("https")
	LetsEncryptEmail = sec.Key("LETSENCRYPT_EMAIL").MustString("")
	Domain = sec.Key("DOMAIN").MustString("localhost")
	HTTPAddr = sec.Key("HTTP_ADDR").MustString("0.0.0.0")
	HTTPPort = sec.Key("HTTP_PORT").MustString("3000")
	GracefulRestartable = sec.Key("ALLOW_GRACEFUL_RESTARTS").MustBool(true)
	GracefulHammerTime = sec.Key("GRACEFUL_HAMMER_TIME").MustDuration(60 * time.Second)
	StartupTimeout = sec.Key("STARTUP_TIMEOUT").MustDuration(0 * time.Second)

	defaultAppURL := string(Protocol) + "://" + Domain
	if (Protocol == HTTP && HTTPPort != "80") || (Protocol == HTTPS && HTTPPort != "443") {
		defaultAppURL += ":" + HTTPPort
	}
	AppURL = sec.Key("ROOT_URL").MustString(defaultAppURL)
	AppURL = strings.TrimSuffix(AppURL, "/") + "/"

	// Check if has app suburl.
	appURL, err := url.Parse(AppURL)
	if err != nil {
		log.Fatal("Invalid ROOT_URL '%s': %s", AppURL, err)
	}
	// Suburl should start with '/' and end without '/', such as '/{subpath}'.
	// This value is empty if site does not have sub-url.
	AppSubURL = strings.TrimSuffix(appURL.Path, "/")
	StaticURLPrefix = strings.TrimSuffix(sec.Key("STATIC_URL_PREFIX").MustString(AppSubURL), "/")
	AppSubURLDepth = strings.Count(AppSubURL, "/")
	// Check if Domain differs from AppURL domain than update it to AppURL's domain
	// TODO: Can be replaced with url.Hostname() when minimal GoLang version is 1.8
	urlHostname := strings.SplitN(appURL.Host, ":", 2)[0]
	if urlHostname != Domain && net.ParseIP(urlHostname) == nil {
		Domain = urlHostname
	}

	var defaultLocalURL string
	switch Protocol {
	case UnixSocket:
		defaultLocalURL = "http://unix/"
	case FCGI:
		defaultLocalURL = AppURL
	case FCGIUnix:
		defaultLocalURL = AppURL
	default:
		defaultLocalURL = string(Protocol) + "://"
		if HTTPAddr == "0.0.0.0" {
			defaultLocalURL += "localhost"
		} else {
			defaultLocalURL += HTTPAddr
		}
		defaultLocalURL += ":" + HTTPPort + "/"
	}
	LocalURL = sec.Key("LOCAL_ROOT_URL").MustString(defaultLocalURL)
	RedirectOtherPort = sec.Key("REDIRECT_OTHER_PORT").MustBool(false)
	PortToRedirect = sec.Key("PORT_TO_REDIRECT").MustString("80")
	OfflineMode = sec.Key("OFFLINE_MODE").MustBool()
	DisableRouterLog = sec.Key("DISABLE_ROUTER_LOG").MustBool()
	StaticRootPath = sec.Key("STATIC_ROOT_PATH").MustString(AppWorkPath)
	StaticCacheTime = sec.Key("STATIC_CACHE_TIME").MustDuration(6 * time.Hour)
	AppDataPath = sec.Key("APP_DATA_PATH").MustString(path.Join(AppWorkPath, "data"))
	EnableGzip = sec.Key("ENABLE_GZIP").MustBool()
	EnablePprof = sec.Key("ENABLE_PPROF").MustBool(false)
	PprofDataPath = sec.Key("PPROF_DATA_PATH").MustString(path.Join(AppWorkPath, "data/tmp/pprof"))
	if !filepath.IsAbs(PprofDataPath) {
		PprofDataPath = filepath.Join(AppWorkPath, PprofDataPath)
	}

	switch sec.Key("LANDING_PAGE").MustString("home") {
	case "explore":
		LandingPageURL = LandingPageExplore
	case "organizations":
		LandingPageURL = LandingPageOrganizations
	case "login":
		LandingPageURL = LandingPageLogin
	default:
		LandingPageURL = LandingPageHome
	}

	if len(SSH.Domain) == 0 {
		SSH.Domain = Domain
	}
	SSH.RootPath = path.Join(homeDir, ".ssh")
	serverCiphers := sec.Key("SSH_SERVER_CIPHERS").Strings(",")
	if len(serverCiphers) > 0 {
		SSH.ServerCiphers = serverCiphers
	}
	serverKeyExchanges := sec.Key("SSH_SERVER_KEY_EXCHANGES").Strings(",")
	if len(serverKeyExchanges) > 0 {
		SSH.ServerKeyExchanges = serverKeyExchanges
	}
	serverMACs := sec.Key("SSH_SERVER_MACS").Strings(",")
	if len(serverMACs) > 0 {
		SSH.ServerMACs = serverMACs
	}
	SSH.KeyTestPath = os.TempDir()
	if err = Cfg.Section("server").MapTo(&SSH); err != nil {
		log.Fatal("Failed to map SSH settings: %v", err)
	}

	SSH.KeygenPath = sec.Key("SSH_KEYGEN_PATH").MustString("ssh-keygen")
	SSH.Port = sec.Key("SSH_PORT").MustInt(22)
	SSH.ListenPort = sec.Key("SSH_LISTEN_PORT").MustInt(SSH.Port)

	// When disable SSH, start builtin server value is ignored.
	if SSH.Disabled {
		SSH.StartBuiltinServer = false
	}

	if !SSH.Disabled && !SSH.StartBuiltinServer {
		if err := os.MkdirAll(SSH.RootPath, 0700); err != nil {
			log.Fatal("Failed to create '%s': %v", SSH.RootPath, err)
		} else if err = os.MkdirAll(SSH.KeyTestPath, 0644); err != nil {
			log.Fatal("Failed to create '%s': %v", SSH.KeyTestPath, err)
		}
	}

	SSH.MinimumKeySizeCheck = sec.Key("MINIMUM_KEY_SIZE_CHECK").MustBool()
	minimumKeySizes := Cfg.Section("ssh.minimum_key_sizes").Keys()
	for _, key := range minimumKeySizes {
		if key.MustInt() != -1 {
			SSH.MinimumKeySizes[strings.ToLower(key.Name())] = key.MustInt()
		}
	}
	SSH.AuthorizedKeysBackup = sec.Key("SSH_AUTHORIZED_KEYS_BACKUP").MustBool(true)
	SSH.CreateAuthorizedKeysFile = sec.Key("SSH_CREATE_AUTHORIZED_KEYS_FILE").MustBool(true)
	SSH.ExposeAnonymous = sec.Key("SSH_EXPOSE_ANONYMOUS").MustBool(false)

	sec = Cfg.Section("server")
	if err = sec.MapTo(&LFS); err != nil {
		log.Fatal("Failed to map LFS settings: %v", err)
	}
	LFS.ContentPath = sec.Key("LFS_CONTENT_PATH").MustString(filepath.Join(AppDataPath, "lfs"))
	if !filepath.IsAbs(LFS.ContentPath) {
		LFS.ContentPath = filepath.Join(AppWorkPath, LFS.ContentPath)
	}

	LFS.HTTPAuthExpiry = sec.Key("LFS_HTTP_AUTH_EXPIRY").MustDuration(20 * time.Minute)

	if LFS.StartServer {
		if err := os.MkdirAll(LFS.ContentPath, 0700); err != nil {
			log.Fatal("Failed to create '%s': %v", LFS.ContentPath, err)
		}

		LFS.JWTSecretBytes = make([]byte, 32)
		n, err := base64.RawURLEncoding.Decode(LFS.JWTSecretBytes, []byte(LFS.JWTSecretBase64))

		if err != nil || n != 32 {
			LFS.JWTSecretBase64, err = generate.NewJwtSecret()
			if err != nil {
				log.Fatal("Error generating JWT Secret for custom config: %v", err)
				return
			}

			// Save secret
			cfg := ini.Empty()
			if com.IsFile(CustomConf) {
				// Keeps custom settings if there is already something.
				if err := cfg.Append(CustomConf); err != nil {
					log.Error("Failed to load custom conf '%s': %v", CustomConf, err)
				}
			}

			cfg.Section("server").Key("LFS_JWT_SECRET").SetValue(LFS.JWTSecretBase64)

			if err := os.MkdirAll(filepath.Dir(CustomConf), os.ModePerm); err != nil {
				log.Fatal("Failed to create '%s': %v", CustomConf, err)
			}
			if err := cfg.SaveTo(CustomConf); err != nil {
				log.Fatal("Error saving generated JWT Secret to custom config: %v", err)
				return
			}
		}
	}

	if err = Cfg.Section("oauth2").MapTo(&OAuth2); err != nil {
		log.Fatal("Failed to OAuth2 settings: %v", err)
		return
	}

	if OAuth2.Enable {
		OAuth2.JWTSecretBytes = make([]byte, 32)
		n, err := base64.RawURLEncoding.Decode(OAuth2.JWTSecretBytes, []byte(OAuth2.JWTSecretBase64))

		if err != nil || n != 32 {
			OAuth2.JWTSecretBase64, err = generate.NewJwtSecret()
			if err != nil {
				log.Fatal("error generating JWT secret: %v", err)
				return
			}
			cfg := ini.Empty()
			if com.IsFile(CustomConf) {
				if err := cfg.Append(CustomConf); err != nil {
					log.Error("failed to load custom conf %s: %v", CustomConf, err)
					return
				}
			}
			cfg.Section("oauth2").Key("JWT_SECRET").SetValue(OAuth2.JWTSecretBase64)

			if err := os.MkdirAll(filepath.Dir(CustomConf), os.ModePerm); err != nil {
				log.Fatal("failed to create '%s': %v", CustomConf, err)
				return
			}
			if err := cfg.SaveTo(CustomConf); err != nil {
				log.Fatal("error saving generating JWT secret to custom config: %v", err)
				return
			}
		}
	}

	sec = Cfg.Section("admin")
	Admin.DefaultEmailNotification = sec.Key("DEFAULT_EMAIL_NOTIFICATIONS").MustString("enabled")

	sec = Cfg.Section("security")
	InstallLock = sec.Key("INSTALL_LOCK").MustBool(false)
	SecretKey = sec.Key("SECRET_KEY").MustString("!#@FDEWREWR&*(")
	LogInRememberDays = sec.Key("LOGIN_REMEMBER_DAYS").MustInt(7)
	CookieUserName = sec.Key("COOKIE_USERNAME").MustString("gitea_awesome")
	CookieRememberName = sec.Key("COOKIE_REMEMBER_NAME").MustString("gitea_incredible")
	ReverseProxyAuthUser = sec.Key("REVERSE_PROXY_AUTHENTICATION_USER").MustString("X-WEBAUTH-USER")
	ReverseProxyAuthEmail = sec.Key("REVERSE_PROXY_AUTHENTICATION_EMAIL").MustString("X-WEBAUTH-EMAIL")
	MinPasswordLength = sec.Key("MIN_PASSWORD_LENGTH").MustInt(6)
	ImportLocalPaths = sec.Key("IMPORT_LOCAL_PATHS").MustBool(false)
	DisableGitHooks = sec.Key("DISABLE_GIT_HOOKS").MustBool(false)
	OnlyAllowPushIfGiteaEnvironmentSet = sec.Key("ONLY_ALLOW_PUSH_IF_GITEA_ENVIRONMENT_SET").MustBool(true)
	PasswordHashAlgo = sec.Key("PASSWORD_HASH_ALGO").MustString("pbkdf2")
	CSRFCookieHTTPOnly = sec.Key("CSRF_COOKIE_HTTP_ONLY").MustBool(true)

	InternalToken = loadInternalToken(sec)

	cfgdata := sec.Key("PASSWORD_COMPLEXITY").Strings(",")
	PasswordComplexity = make([]string, 0, len(cfgdata))
	for _, name := range cfgdata {
		name := strings.ToLower(strings.Trim(name, `"`))
		if name != "" {
			PasswordComplexity = append(PasswordComplexity, name)
		}
	}

	sec = Cfg.Section("attachment")
	AttachmentPath = sec.Key("PATH").MustString(path.Join(AppDataPath, "attachments"))
	if !filepath.IsAbs(AttachmentPath) {
		AttachmentPath = path.Join(AppWorkPath, AttachmentPath)
	}
	AttachmentAllowedTypes = strings.Replace(sec.Key("ALLOWED_TYPES").MustString("image/jpeg,image/png,application/zip,application/gzip"), "|", ",", -1)
	AttachmentMaxSize = sec.Key("MAX_SIZE").MustInt64(4)
	AttachmentMaxFiles = sec.Key("MAX_FILES").MustInt(5)
	AttachmentEnabled = sec.Key("ENABLED").MustBool(true)

	timeFormatKey := Cfg.Section("time").Key("FORMAT").MustString("")
	if timeFormatKey != "" {
		TimeFormat = map[string]string{
			"ANSIC":       time.ANSIC,
			"UnixDate":    time.UnixDate,
			"RubyDate":    time.RubyDate,
			"RFC822":      time.RFC822,
			"RFC822Z":     time.RFC822Z,
			"RFC850":      time.RFC850,
			"RFC1123":     time.RFC1123,
			"RFC1123Z":    time.RFC1123Z,
			"RFC3339":     time.RFC3339,
			"RFC3339Nano": time.RFC3339Nano,
			"Kitchen":     time.Kitchen,
			"Stamp":       time.Stamp,
			"StampMilli":  time.StampMilli,
			"StampMicro":  time.StampMicro,
			"StampNano":   time.StampNano,
		}[timeFormatKey]
		// When the TimeFormatKey does not exist in the previous map e.g.'2006-01-02 15:04:05'
		if len(TimeFormat) == 0 {
			TimeFormat = timeFormatKey
			TestTimeFormat, _ := time.Parse(TimeFormat, TimeFormat)
			if TestTimeFormat.Format(time.RFC3339) != "2006-01-02T15:04:05Z" {
				log.Warn("Provided TimeFormat: %s does not create a fully specified date and time.", TimeFormat)
				log.Warn("In order to display dates and times correctly please check your time format has 2006, 01, 02, 15, 04 and 05")
			}
			log.Trace("Custom TimeFormat: %s", TimeFormat)
		}
	}

	zone := Cfg.Section("time").Key("DEFAULT_UI_LOCATION").String()
	if zone != "" {
		DefaultUILocation, err = time.LoadLocation(zone)
		if err != nil {
			log.Fatal("Load time zone failed: %v", err)
		} else {
			log.Info("Default UI Location is %v", zone)
		}
	}
	if DefaultUILocation == nil {
		DefaultUILocation = time.Local
	}

	RunUser = Cfg.Section("").Key("RUN_USER").MustString(user.CurrentUsername())
	// Does not check run user when the install lock is off.
	if InstallLock {
		currentUser, match := IsRunUserMatchCurrentUser(RunUser)
		if !match {
			log.Fatal("Expect user '%s' but current user is: %s", RunUser, currentUser)
		}
	}

	SSH.BuiltinServerUser = Cfg.Section("server").Key("BUILTIN_SSH_SERVER_USER").MustString(RunUser)

	newRepository()

	sec = Cfg.Section("picture")
	AvatarUploadPath = sec.Key("AVATAR_UPLOAD_PATH").MustString(path.Join(AppDataPath, "avatars"))
	forcePathSeparator(AvatarUploadPath)
	if !filepath.IsAbs(AvatarUploadPath) {
		AvatarUploadPath = path.Join(AppWorkPath, AvatarUploadPath)
	}
	RepositoryAvatarUploadPath = sec.Key("REPOSITORY_AVATAR_UPLOAD_PATH").MustString(path.Join(AppDataPath, "repo-avatars"))
	forcePathSeparator(RepositoryAvatarUploadPath)
	if !filepath.IsAbs(RepositoryAvatarUploadPath) {
		RepositoryAvatarUploadPath = path.Join(AppWorkPath, RepositoryAvatarUploadPath)
	}
	RepositoryAvatarFallback = sec.Key("REPOSITORY_AVATAR_FALLBACK").MustString("none")
	RepositoryAvatarFallbackImage = sec.Key("REPOSITORY_AVATAR_FALLBACK_IMAGE").MustString("/img/repo_default.png")
	AvatarMaxWidth = sec.Key("AVATAR_MAX_WIDTH").MustInt(4096)
	AvatarMaxHeight = sec.Key("AVATAR_MAX_HEIGHT").MustInt(3072)
	AvatarMaxFileSize = sec.Key("AVATAR_MAX_FILE_SIZE").MustInt64(1048576)
	switch source := sec.Key("GRAVATAR_SOURCE").MustString("gravatar"); source {
	case "duoshuo":
		GravatarSource = "http://gravatar.duoshuo.com/avatar/"
	case "gravatar":
		GravatarSource = "https://secure.gravatar.com/avatar/"
	case "libravatar":
		GravatarSource = "https://seccdn.libravatar.org/avatar/"
	default:
		GravatarSource = source
	}
	DisableGravatar = sec.Key("DISABLE_GRAVATAR").MustBool()
	EnableFederatedAvatar = sec.Key("ENABLE_FEDERATED_AVATAR").MustBool(!InstallLock)
	if OfflineMode {
		DisableGravatar = true
		EnableFederatedAvatar = false
	}
	if DisableGravatar {
		EnableFederatedAvatar = false
	}
	if EnableFederatedAvatar || !DisableGravatar {
		GravatarSourceURL, err = url.Parse(GravatarSource)
		if err != nil {
			log.Fatal("Failed to parse Gravatar URL(%s): %v",
				GravatarSource, err)
		}
	}

	if EnableFederatedAvatar {
		LibravatarService = libravatar.New()
		if GravatarSourceURL.Scheme == "https" {
			LibravatarService.SetUseHTTPS(true)
			LibravatarService.SetSecureFallbackHost(GravatarSourceURL.Host)
		} else {
			LibravatarService.SetUseHTTPS(false)
			LibravatarService.SetFallbackHost(GravatarSourceURL.Host)
		}
	}

	if err = Cfg.Section("ui").MapTo(&UI); err != nil {
		log.Fatal("Failed to map UI settings: %v", err)
	} else if err = Cfg.Section("markdown").MapTo(&Markdown); err != nil {
		log.Fatal("Failed to map Markdown settings: %v", err)
	} else if err = Cfg.Section("admin").MapTo(&Admin); err != nil {
		log.Fatal("Fail to map Admin settings: %v", err)
	} else if err = Cfg.Section("api").MapTo(&API); err != nil {
		log.Fatal("Failed to map API settings: %v", err)
	} else if err = Cfg.Section("metrics").MapTo(&Metrics); err != nil {
		log.Fatal("Failed to map Metrics settings: %v", err)
	}

	u := *appURL
	u.Path = path.Join(u.Path, "api", "swagger")
	API.SwaggerURL = u.String()

	newCron()
	newGit()

	sec = Cfg.Section("mirror")
	Mirror.MinInterval = sec.Key("MIN_INTERVAL").MustDuration(10 * time.Minute)
	Mirror.DefaultInterval = sec.Key("DEFAULT_INTERVAL").MustDuration(8 * time.Hour)
	if Mirror.MinInterval.Minutes() < 1 {
		log.Warn("Mirror.MinInterval is too low")
		Mirror.MinInterval = 1 * time.Minute
	}
	if Mirror.DefaultInterval < Mirror.MinInterval {
		log.Warn("Mirror.DefaultInterval is less than Mirror.MinInterval")
		Mirror.DefaultInterval = time.Hour * 8
	}

	Langs = Cfg.Section("i18n").Key("LANGS").Strings(",")
	if len(Langs) == 0 {
		Langs = []string{
			"en-US", "zh-CN", "zh-HK", "zh-TW", "de-DE", "fr-FR", "nl-NL", "lv-LV",
			"ru-RU", "uk-UA", "ja-JP", "es-ES", "pt-BR", "pl-PL", "bg-BG", "it-IT",
			"fi-FI", "tr-TR", "cs-CZ", "sr-SP", "sv-SE", "ko-KR"}
	}
	Names = Cfg.Section("i18n").Key("NAMES").Strings(",")
	if len(Names) == 0 {
		Names = []string{"English", "简体中文", "繁體中文（香港）", "繁體中文（台灣）", "Deutsch",
			"français", "Nederlands", "latviešu", "русский", "Українська", "日本語",
			"español", "português do Brasil", "polski", "български", "italiano",
			"suomi", "Türkçe", "čeština", "српски", "svenska", "한국어"}
	}
	dateLangs = Cfg.Section("i18n.datelang").KeysHash()

	ShowFooterBranding = Cfg.Section("other").Key("SHOW_FOOTER_BRANDING").MustBool(false)
	ShowFooterVersion = Cfg.Section("other").Key("SHOW_FOOTER_VERSION").MustBool(true)
	ShowFooterTemplateLoadTime = Cfg.Section("other").Key("SHOW_FOOTER_TEMPLATE_LOAD_TIME").MustBool(true)

	UI.ShowUserEmail = Cfg.Section("ui").Key("SHOW_USER_EMAIL").MustBool(true)
	UI.DefaultShowFullName = Cfg.Section("ui").Key("DEFAULT_SHOW_FULL_NAME").MustBool(false)
	UI.SearchRepoDescription = Cfg.Section("ui").Key("SEARCH_REPO_DESCRIPTION").MustBool(true)
	UI.UseServiceWorker = Cfg.Section("ui").Key("USE_SERVICE_WORKER").MustBool(true)

	HasRobotsTxt = com.IsFile(path.Join(CustomPath, "robots.txt"))

	newMarkup()

	sec = Cfg.Section("U2F")
	U2F.TrustedFacets, _ = shellquote.Split(sec.Key("TRUSTED_FACETS").MustString(strings.TrimRight(AppURL, "/")))
	U2F.AppID = sec.Key("APP_ID").MustString(strings.TrimRight(AppURL, "/"))

	zip.Verbose = false

	UI.ReactionsMap = make(map[string]bool)
	for _, reaction := range UI.Reactions {
		UI.ReactionsMap[reaction] = true
	}
}

func loadInternalToken(sec *ini.Section) string {
	uri := sec.Key("INTERNAL_TOKEN_URI").String()
	if len(uri) == 0 {
		return loadOrGenerateInternalToken(sec)
	}
	tempURI, err := url.Parse(uri)
	if err != nil {
		log.Fatal("Failed to parse INTERNAL_TOKEN_URI (%s): %v", uri, err)
	}
	switch tempURI.Scheme {
	case "file":
		fp, err := os.OpenFile(tempURI.RequestURI(), os.O_RDWR, 0600)
		if err != nil {
			log.Fatal("Failed to open InternalTokenURI (%s): %v", uri, err)
		}
		defer fp.Close()

		buf, err := ioutil.ReadAll(fp)
		if err != nil {
			log.Fatal("Failed to read InternalTokenURI (%s): %v", uri, err)
		}
		// No token in the file, generate one and store it.
		if len(buf) == 0 {
			token, err := generate.NewInternalToken()
			if err != nil {
				log.Fatal("Error generate internal token: %v", err)
			}
			if _, err := io.WriteString(fp, token); err != nil {
				log.Fatal("Error writing to InternalTokenURI (%s): %v", uri, err)
			}
			return token
		}

		return string(buf)
	default:
		log.Fatal("Unsupported URI-Scheme %q (INTERNAL_TOKEN_URI = %q)", tempURI.Scheme, uri)
	}
	return ""
}

func loadOrGenerateInternalToken(sec *ini.Section) string {
	var err error
	token := sec.Key("INTERNAL_TOKEN").String()
	if len(token) == 0 {
		token, err = generate.NewInternalToken()
		if err != nil {
			log.Fatal("Error generate internal token: %v", err)
		}

		// Save secret
		cfgSave := ini.Empty()
		if com.IsFile(CustomConf) {
			// Keeps custom settings if there is already something.
			if err := cfgSave.Append(CustomConf); err != nil {
				log.Error("Failed to load custom conf '%s': %v", CustomConf, err)
			}
		}

		cfgSave.Section("security").Key("INTERNAL_TOKEN").SetValue(token)

		if err := os.MkdirAll(filepath.Dir(CustomConf), os.ModePerm); err != nil {
			log.Fatal("Failed to create '%s': %v", CustomConf, err)
		}
		if err := cfgSave.SaveTo(CustomConf); err != nil {
			log.Fatal("Error saving generated INTERNAL_TOKEN to custom config: %v", err)
		}
	}
	return token
}

// NewServices initializes the services
func NewServices() {
	InitDBConfig()
	newService()
	NewLogServices(false)
	newCacheService()
	newSessionService()
	newCORSService()
	newMailService()
	newWechatService()
	newRegisterMailService()
	newNotifyMailService()
	newNotifyWechatService()
	newWebhookService()
	newMigrationsService()
	newIndexerService()
	newTaskService()
	NewQueueService()
}
