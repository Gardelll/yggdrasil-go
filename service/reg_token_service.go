/*
 * Copyright (C) 2025. Gardel <sunxinao@hotmail.com> and contributors
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package service

import (
	"bytes"
	"fmt"
	lru "github.com/hashicorp/golang-lru"
	"github.com/wneessen/go-mail"
	"regexp"
	"text/template"
	"yggdrasil-go/model"
	"yggdrasil-go/util"
)

type RegTokenService interface {
	SendTokenEmail(tokenType RegTokenType, email string) error
	VerifyToken(accessToken string) (string, error)
	IsSmtpEnabled() bool
}

type RegTokenType uint

const (
	RegisterToken RegTokenType = iota
	ResetPasswordToken
)

type SmtpConfig struct {
	Enabled               bool
	SmtpServer            string
	SmtpPort              int
	SmtpSsl               bool
	EmailFrom             string
	SmtpUser              string
	SmtpPassword          string
	TitlePrefix           string
	RegisterTemplate      string
	ResetPasswordTemplate string
}

type regTokenServiceImpl struct {
	tokenCache            *lru.Cache
	smtpEnabled           bool
	smtpServer            string
	smtpPort              int
	smtpSsl               bool
	smtpUser              string
	smtpPassword          string
	emailFrom             string
	titlePrefix           string
	registerTemplate      *template.Template
	resetPasswordTemplate *template.Template
}

// removeHtmlComments removes HTML comments from a string using regex
// This matches <!-- ... --> patterns including multiline comments
func removeHtmlComments(htmlContent string) string {
	// Regex pattern to match HTML comments: <!-- ... -->
	// The (?s) flag makes . match newlines as well
	commentRegex := regexp.MustCompile(`(?s)<!--.*?-->`)
	return commentRegex.ReplaceAllString(htmlContent, "")
}

func NewRegTokenService(smtpCfg *SmtpConfig) RegTokenService {
	cache, _ := lru.New(10000000)

	// Remove HTML comments from templates
	registerTemplate := removeHtmlComments(smtpCfg.RegisterTemplate)
	resetPasswordTemplate := removeHtmlComments(smtpCfg.ResetPasswordTemplate)

	impl := regTokenServiceImpl{
		tokenCache:            cache,
		smtpEnabled:           smtpCfg.Enabled,
		smtpServer:            smtpCfg.SmtpServer,
		smtpPort:              smtpCfg.SmtpPort,
		smtpSsl:               smtpCfg.SmtpSsl,
		smtpUser:              smtpCfg.SmtpUser,
		smtpPassword:          smtpCfg.SmtpPassword,
		emailFrom:             smtpCfg.EmailFrom,
		titlePrefix:           smtpCfg.TitlePrefix,
		registerTemplate:      template.Must(template.New("register").Parse(registerTemplate)),
		resetPasswordTemplate: template.Must(template.New("resetPassword").Parse(resetPasswordTemplate)),
	}
	return &impl
}

func (r *regTokenServiceImpl) SendTokenEmail(tokenType RegTokenType, email string) error {
	// If SMTP is not enabled, return immediately without sending email
	if !r.smtpEnabled {
		return nil
	}

	token := model.NewRegToken(email)
	r.tokenCache.Add(token.AccessToken, token)

	var subject, body string
	buf := bytes.Buffer{}
	switch tokenType {
	case RegisterToken:
		subject = fmt.Sprintf("%s 注册验证码", r.titlePrefix)
		if err := r.registerTemplate.Execute(&buf, token); err != nil {
			return fmt.Errorf("execute registerTemplate error: %v", err)
		}
		body = buf.String()
	case ResetPasswordToken:
		subject = fmt.Sprintf("%s 重置密码验证码", r.titlePrefix)
		if err := r.resetPasswordTemplate.Execute(&buf, token); err != nil {
			return fmt.Errorf("execute resetPasswordTemplate error: %v", err)
		}
		body = buf.String()
	default:
		return fmt.Errorf("unknown token type")
	}

	message := mail.NewMsg()
	if err := message.From(r.emailFrom); err != nil {
		return fmt.Errorf("failed to set From address: %s", err)
	}
	if err := message.To(email); err != nil {
		return fmt.Errorf("failed to set To address: %s", err)
	}
	message.Subject(subject)
	message.SetBodyString(mail.TypeTextHTML, body)
	client, err := mail.NewClient(r.smtpServer, mail.WithPort(r.smtpPort), mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(r.smtpUser), mail.WithPassword(r.smtpPassword))
	if err != nil {
		return fmt.Errorf("failed to create mail client: %s", err)
	}
	if r.smtpSsl {
		client.SetSSL(true)
	}
	if err := client.DialAndSend(message); err != nil {
		return fmt.Errorf("failed to send mail: %s", err)
	}
	return nil
}

func (r *regTokenServiceImpl) VerifyToken(accessToken string) (string, error) {
	token, ok := r.tokenCache.Get(accessToken)
	if !ok {
		return "", util.NewIllegalArgumentError(util.MessageInvalidToken)
	}

	if regToken, ok := token.(model.RegToken); ok {
		if regToken.IsValid() {
			r.tokenCache.Remove(accessToken)
			return regToken.Email, nil
		}
	}

	return "", util.NewIllegalArgumentError("wrong access token or email")
}

func (r *regTokenServiceImpl) IsSmtpEnabled() bool {
	return r.smtpEnabled
}
