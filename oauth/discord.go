package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/i18n"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting/system_setting"
	"github.com/gin-gonic/gin"
)

func init() {
	Register("discord", &DiscordProvider{})
}

// DiscordProvider implements OAuth for Discord
type DiscordProvider struct{}

type discordOAuthResponse struct {
	AccessToken  string `json:"access_token"`
	IDToken      string `json:"id_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
}

type discordUser struct {
	UID  string `json:"id"`
	ID   string `json:"username"`
	Name string `json:"global_name"`
}

type discordGuildMember struct {
	Roles []string `json:"roles"`
}

func (p *DiscordProvider) GetName() string {
	return "Discord"
}

func (p *DiscordProvider) IsEnabled() bool {
	return system_setting.GetDiscordSettings().Enabled
}

func (p *DiscordProvider) ExchangeToken(ctx context.Context, code string, c *gin.Context) (*OAuthToken, error) {
	if code == "" {
		return nil, NewOAuthError(i18n.MsgOAuthInvalidCode, nil)
	}

	logger.LogDebug(ctx, "[OAuth-Discord] ExchangeToken: code=%s...", code[:min(len(code), 10)])

	settings := system_setting.GetDiscordSettings()
	redirectUri := fmt.Sprintf("%s/oauth/discord", system_setting.ServerAddress)
	values := url.Values{}
	values.Set("client_id", settings.ClientId)
	values.Set("client_secret", settings.ClientSecret)
	values.Set("code", code)
	values.Set("grant_type", "authorization_code")
	values.Set("redirect_uri", redirectUri)

	logger.LogDebug(ctx, "[OAuth-Discord] ExchangeToken: redirect_uri=%s", redirectUri)

	req, err := http.NewRequestWithContext(ctx, "POST", "https://discord.com/api/v10/oauth2/token", strings.NewReader(values.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	client := http.Client{
		Timeout: 5 * time.Second,
	}
	res, err := client.Do(req)
	if err != nil {
		logger.LogError(ctx, fmt.Sprintf("[OAuth-Discord] ExchangeToken error: %s", err.Error()))
		return nil, NewOAuthErrorWithRaw(i18n.MsgOAuthConnectFailed, map[string]any{"Provider": "Discord"}, err.Error())
	}
	defer res.Body.Close()

	logger.LogDebug(ctx, "[OAuth-Discord] ExchangeToken response status: %d", res.StatusCode)

	var discordResponse discordOAuthResponse

	err = json.NewDecoder(res.Body).Decode(&discordResponse)
	if err != nil {
		logger.LogError(ctx, fmt.Sprintf("[OAuth-Discord] ExchangeToken decode error: %s", err.Error()))
		return nil, err
	}

	if discordResponse.AccessToken == "" {
		logger.LogError(ctx, "[OAuth-Discord] ExchangeToken failed: empty access token")
		return nil, NewOAuthError(i18n.MsgOAuthTokenFailed, map[string]any{"Provider": "Discord"})
	}

	logger.LogDebug(ctx, "[OAuth-Discord] ExchangeToken success: scope=%s", discordResponse.Scope)

	return &OAuthToken{
		AccessToken:  discordResponse.AccessToken,
		TokenType:    discordResponse.TokenType,
		RefreshToken: discordResponse.RefreshToken,
		ExpiresIn:    discordResponse.ExpiresIn,
		Scope:        discordResponse.Scope,
		IDToken:      discordResponse.IDToken,
	}, nil
}

func (p *DiscordProvider) GetUserInfo(ctx context.Context, token *OAuthToken) (*OAuthUser, error) {
	logger.LogDebug(ctx, "[OAuth-Discord] GetUserInfo: fetching user info")

	req, err := http.NewRequestWithContext(ctx, "GET", "https://discord.com/api/v10/users/@me", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

	client := http.Client{
		Timeout: 5 * time.Second,
	}
	res, err := client.Do(req)
	if err != nil {
		logger.LogError(ctx, fmt.Sprintf("[OAuth-Discord] GetUserInfo error: %s", err.Error()))
		return nil, NewOAuthErrorWithRaw(i18n.MsgOAuthConnectFailed, map[string]any{"Provider": "Discord"}, err.Error())
	}
	defer res.Body.Close()

	logger.LogDebug(ctx, "[OAuth-Discord] GetUserInfo response status: %d", res.StatusCode)

	if res.StatusCode != http.StatusOK {
		logger.LogError(ctx, fmt.Sprintf("[OAuth-Discord] GetUserInfo failed: status=%d", res.StatusCode))
		return nil, NewOAuthError(i18n.MsgOAuthGetUserErr, nil)
	}

	var discordUser discordUser
	err = json.NewDecoder(res.Body).Decode(&discordUser)
	if err != nil {
		logger.LogError(ctx, fmt.Sprintf("[OAuth-Discord] GetUserInfo decode error: %s", err.Error()))
		return nil, err
	}

	if discordUser.UID == "" || discordUser.ID == "" {
		logger.LogError(ctx, "[OAuth-Discord] GetUserInfo failed: empty user fields")
		return nil, NewOAuthError(i18n.MsgOAuthUserInfoEmpty, map[string]any{"Provider": "Discord"})
	}

	// 黑名单检查（必须不在黑名单服务器）
	if err := validateDiscordGuildBan(ctx, token.AccessToken); err != nil {
		return nil, err
	}
	// 白名单检查（必须在白名单服务器）
	if err := validateDiscordGuildMembership(ctx, token.AccessToken); err != nil {
		return nil, err
	}

	logger.LogDebug(ctx, "[OAuth-Discord] GetUserInfo success: uid=%s, username=%s, name=%s", discordUser.UID, discordUser.ID, discordUser.Name)

	return &OAuthUser{
		ProviderUserID: discordUser.UID,
		Username:       discordUser.ID,
		DisplayName:    discordUser.Name,
	}, nil
}

func (p *DiscordProvider) IsUserIDTaken(providerUserID string) bool {
	return model.IsDiscordIdAlreadyTaken(providerUserID)
}

func (p *DiscordProvider) FillUserByProviderID(user *model.User, providerUserID string) error {
	user.DiscordId = providerUserID
	return user.FillUserByDiscordId()
}

func (p *DiscordProvider) SetProviderUserID(user *model.User, providerUserID string) {
	user.DiscordId = providerUserID
}

func (p *DiscordProvider) GetProviderPrefix() string {
	return "discord_"
}

// ============================================================================
// Discord Guild Validation (whitelist + blacklist)
// ============================================================================

// isUserInGuild 检查用户是否在指定服务器
func isUserInGuild(ctx context.Context, accessToken string, guildID string) (bool, error) {
	memberURL := fmt.Sprintf("https://discord.com/api/v10/users/@me/guilds/%s/member", guildID)
	req, err := http.NewRequestWithContext(ctx, "GET", memberURL, nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	client := http.Client{Timeout: 5 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()

	return res.StatusCode == http.StatusOK, nil
}

// validateDiscordGuildBan 黑名单检查：用户不能在黑名单服务器
func validateDiscordGuildBan(ctx context.Context, accessToken string) error {
	bannedGuildIDs := parseGuildIDs(os.Getenv("DISCORD_BANNED_GUILD_IDS"))
	if len(bannedGuildIDs) == 0 {
		return nil
	}

	for guildID := range bannedGuildIDs {
		inGuild, err := isUserInGuild(ctx, accessToken, guildID)
		if err != nil {
			logger.LogError(ctx, fmt.Sprintf("[OAuth-Discord] Banned guild check error: guild=%s err=%s", guildID, err.Error()))
			continue // 网络错误时不阻断，继续检查
		}
		if inGuild {
			logger.LogError(ctx, fmt.Sprintf("[OAuth-Discord] User is in banned guild: %s", guildID))
			return &AccessDeniedError{Message: "该账号无法登录，请联系管理员了解详情"}
		}
	}

	logger.LogDebug(ctx, "[OAuth-Discord] Banned guild validation passed")
	return nil
}

// validateDiscordGuildMembership 白名单检查：用户必须在白名单服务器
func validateDiscordGuildMembership(ctx context.Context, accessToken string) error {
	guildID := strings.TrimSpace(os.Getenv("DISCORD_ALLOWED_GUILD_ID"))
	if guildID == "" {
		return nil
	}

	allowedRoleIDs := parseDiscordRoleIDs(os.Getenv("DISCORD_ALLOWED_ROLE_IDS"))

	memberURL := fmt.Sprintf("https://discord.com/api/v10/users/@me/guilds/%s/member", guildID)
	req, err := http.NewRequestWithContext(ctx, "GET", memberURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	client := http.Client{Timeout: 5 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		logger.LogError(ctx, fmt.Sprintf("[OAuth-Discord] Guild membership validation error: %s", err.Error()))
		return NewOAuthErrorWithRaw(i18n.MsgOAuthConnectFailed, map[string]any{"Provider": "Discord"}, err.Error())
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		logger.LogError(ctx, fmt.Sprintf("[OAuth-Discord] Guild membership validation failed: guild=%s status=%d", guildID, res.StatusCode))
		return &AccessDeniedError{Message: "该账号未加入指定 Discord 社区，无法登录"}
	}

	var member discordGuildMember
	if err := json.NewDecoder(res.Body).Decode(&member); err != nil {
		logger.LogError(ctx, fmt.Sprintf("[OAuth-Discord] Guild membership decode error: %s", err.Error()))
		return err
	}

	if len(allowedRoleIDs) == 0 {
		logger.LogDebug(ctx, "[OAuth-Discord] Guild membership validation passed: guild=%s", guildID)
		return nil
	}

	if hasAnyDiscordRole(member.Roles, allowedRoleIDs) {
		logger.LogDebug(ctx, "[OAuth-Discord] Guild role validation passed: guild=%s", guildID)
		return nil
	}

	logger.LogError(ctx, fmt.Sprintf("[OAuth-Discord] Guild role validation failed: guild=%s user_role_count=%d", guildID, len(member.Roles)))
	return &AccessDeniedError{Message: "该账号未获得允许登录的 Discord 身份组，无法登录"}
}

func parseGuildIDs(raw string) map[string]struct{} {
	guildIDs := make(map[string]struct{})
	for _, id := range strings.Split(raw, ",") {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		guildIDs[id] = struct{}{}
	}
	return guildIDs
}

func parseDiscordRoleIDs(raw string) map[string]struct{} {
	roleIDs := make(map[string]struct{})
	for _, roleID := range strings.Split(raw, ",") {
		roleID = strings.TrimSpace(roleID)
		if roleID == "" {
			continue
		}
		roleIDs[roleID] = struct{}{}
	}
	return roleIDs
}

func hasAnyDiscordRole(userRoles []string, allowedRoles map[string]struct{}) bool {
	for _, roleID := range userRoles {
		if _, ok := allowedRoles[roleID]; ok {
			return true
		}
	}
	return false
}
