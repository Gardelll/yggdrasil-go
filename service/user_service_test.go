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
	"testing"
	"yggdrasil-go/model"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestRegisterWithoutSMTP tests user registration when SMTP is disabled
func TestRegisterWithoutSMTP(t *testing.T) {
	// Setup in-memory database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// Migrate schema
	err = db.AutoMigrate(&model.User{})
	if err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	// Create services with SMTP disabled
	smtpConfig := &SmtpConfig{
		Enabled: false,
	}
	regTokenService := NewRegTokenService(smtpConfig)
	tokenService := NewTokenService()
	userService := NewUserService(tokenService, regTokenService, db)

	// Test registration
	username := "test@example.com"
	password := "password123"
	profileName := "YggTestUser001"
	ip := "127.0.0.1"

	response, err := userService.Register(username, password, profileName, ip)
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	if response == nil {
		t.Fatal("Expected response, got nil")
	}

	// Verify user was created with EmailVerified = true
	var user model.User
	err = db.Where("email = ?", username).First(&user).Error
	if err != nil {
		t.Fatalf("Failed to find user: %v", err)
	}

	if !user.EmailVerified {
		t.Error("Expected user to be verified when SMTP is disabled, but EmailVerified is false")
	}

	if user.Email != username {
		t.Errorf("Expected email %s, got %s", username, user.Email)
	}

	if user.ProfileName != profileName {
		t.Errorf("Expected profile name %s, got %s", profileName, user.ProfileName)
	}
}

// TestRegisterWithSMTP tests user registration when SMTP is enabled
func TestRegisterWithSMTP(t *testing.T) {
	// Setup in-memory database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// Migrate schema
	err = db.AutoMigrate(&model.User{})
	if err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	// Create services with SMTP enabled
	smtpConfig := &SmtpConfig{
		Enabled:               true,
		SmtpServer:            "localhost",
		SmtpPort:              25,
		EmailFrom:             "test@example.com",
		TitlePrefix:           "[Test]",
		RegisterTemplate:      "Test template {{.AccessToken}}",
		ResetPasswordTemplate: "Reset template {{.AccessToken}}",
	}
	regTokenService := NewRegTokenService(smtpConfig)
	tokenService := NewTokenService()
	userService := NewUserService(tokenService, regTokenService, db)

	// Test registration
	username := "test2@example.com"
	password := "password123"
	profileName := "YggTestUser002"
	ip := "127.0.0.1"

	response, err := userService.Register(username, password, profileName, ip)
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	if response == nil {
		t.Fatal("Expected response, got nil")
	}

	// Verify user was created with EmailVerified = false
	var user model.User
	err = db.Where("email = ?", username).First(&user).Error
	if err != nil {
		t.Fatalf("Failed to find user: %v", err)
	}

	if user.EmailVerified {
		t.Error("Expected user to be unverified when SMTP is enabled, but EmailVerified is true")
	}
}

// TestResetPasswordWithoutSMTP tests password reset when SMTP is disabled
func TestResetPasswordWithoutSMTP(t *testing.T) {
	// Setup in-memory database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// Migrate schema
	err = db.AutoMigrate(&model.User{})
	if err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	// Create services with SMTP disabled
	smtpConfig := &SmtpConfig{
		Enabled: false,
	}
	regTokenService := NewRegTokenService(smtpConfig)
	tokenService := NewTokenService()
	userService := NewUserService(tokenService, regTokenService, db)

	// Create a test user
	user := model.User{
		ID:            uuid.New(),
		Email:         "test3@example.com",
		Password:      "hashedpassword",
		EmailVerified: true,
		ProfileName:   "YggTestUser003",
	}
	err = db.Create(&user).Error
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Test reset password - should fail with clear error
	err = userService.ResetPassword(user.Email, "newpassword123", "invalid-token")
	if err == nil {
		t.Error("Expected error when resetting password without SMTP, but got nil")
	}

	// Verify error message contains expected text
	expectedMsg := "密码重置功能当前不可用"
	if err != nil && !contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error message to contain '%s', got: %s", expectedMsg, err.Error())
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
