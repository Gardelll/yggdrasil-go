/*
 * Copyright (C) 2025. Gardel <gardel741@outlook.com> and contributors
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
	"image"
	"image/color"
	"testing"
	"yggdrasil-go/model"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// createTestImage creates a simple test image
func createTestImage(width, height int, r, g, b uint8) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	c := color.RGBA{R: r, G: g, B: b, A: 255}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, c)
		}
	}
	return img
}

// TestSaveTexture_FirstTimeUpload tests reference counting when user uploads texture for the first time
func TestSaveTexture_FirstTimeUpload(t *testing.T) {
	// Setup in-memory database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// Migrate schema
	err = db.AutoMigrate(&model.User{}, &model.Texture{})
	if err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	// Create test user without texture
	userId := uuid.New()
	user := model.User{
		ID:            userId,
		Email:         "test@example.com",
		Password:      "hashedpassword",
		EmailVerified: true,
		ProfileName:   "TestPlayer",
	}
	profile := model.NewProfile(userId, "TestPlayer", model.STEVE, "")
	user.SetProfile(&profile)
	err = db.Create(&user).Error
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create texture service
	tokenService := NewTokenService()
	textureService := NewTextureService(tokenService, db)

	// Create test image
	testImage := createTestImage(64, 64, 255, 0, 0) // Red image

	// Save texture for first time
	err = textureService.(*textureServiceImpl).saveTexture(&user, testImage, "SKIN", nil)
	if err != nil {
		t.Fatalf("Failed to save texture: %v", err)
	}

	// Verify texture was created with used = 1
	var texture model.Texture
	hash := model.ComputeTextureId(testImage)
	err = db.First(&texture, "hash = ?", hash).Error
	if err != nil {
		t.Fatalf("Failed to find texture: %v", err)
	}

	if texture.Used != 1 {
		t.Errorf("Expected texture.Used = 1 for first upload, got %d", texture.Used)
	}
}

// TestSaveTexture_MultipleUsersSharedTexture tests reference counting when multiple users use same texture
func TestSaveTexture_MultipleUsersSharedTexture(t *testing.T) {
	// Setup in-memory database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// Migrate schema
	err = db.AutoMigrate(&model.User{}, &model.Texture{})
	if err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	// Create texture service
	tokenService := NewTokenService()
	textureService := NewTextureService(tokenService, db)

	// Create test image
	testImage := createTestImage(64, 64, 0, 255, 0) // Green image

	// User A uploads texture
	userA := model.User{
		ID:            uuid.New(),
		Email:         "userA@example.com",
		Password:      "hashedpassword",
		EmailVerified: true,
		ProfileName:   "PlayerA",
	}
	profileA := model.NewProfile(userA.ID, "PlayerA", model.STEVE, "")
	userA.SetProfile(&profileA)
	err = db.Create(&userA).Error
	if err != nil {
		t.Fatalf("Failed to create userA: %v", err)
	}

	err = textureService.(*textureServiceImpl).saveTexture(&userA, testImage, "SKIN", nil)
	if err != nil {
		t.Fatalf("Failed to save texture for userA: %v", err)
	}

	// Verify initial reference count
	var texture model.Texture
	hash := model.ComputeTextureId(testImage)
	err = db.First(&texture, "hash = ?", hash).Error
	if err != nil {
		t.Fatalf("Failed to find texture: %v", err)
	}

	if texture.Used != 1 {
		t.Errorf("Expected texture.Used = 1 after userA upload, got %d", texture.Used)
	}

	// User B uploads the same texture
	userB := model.User{
		ID:            uuid.New(),
		Email:         "userB@example.com",
		Password:      "hashedpassword",
		EmailVerified: true,
		ProfileName:   "PlayerB",
	}
	profileB := model.NewProfile(userB.ID, "PlayerB", model.STEVE, "")
	userB.SetProfile(&profileB)
	err = db.Create(&userB).Error
	if err != nil {
		t.Fatalf("Failed to create userB: %v", err)
	}

	err = textureService.(*textureServiceImpl).saveTexture(&userB, testImage, "SKIN", nil)
	if err != nil {
		t.Fatalf("Failed to save texture for userB: %v", err)
	}

	// Verify reference count incremented
	err = db.First(&texture, "hash = ?", hash).Error
	if err != nil {
		t.Fatalf("Failed to find texture after userB upload: %v", err)
	}

	if texture.Used != 2 {
		t.Errorf("Expected texture.Used = 2 after userB upload (shared texture), got %d", texture.Used)
	}
}

// TestSaveTexture_ReplaceTexture tests reference counting when user replaces their texture
func TestSaveTexture_ReplaceTexture(t *testing.T) {
	// Setup in-memory database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// Migrate schema
	err = db.AutoMigrate(&model.User{}, &model.Texture{})
	if err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	// Create texture service
	tokenService := NewTokenService()
	textureService := NewTextureService(tokenService, db)

	// Create test images
	testImage1 := createTestImage(64, 64, 255, 0, 0) // Red
	testImage2 := createTestImage(64, 64, 0, 0, 255) // Blue

	// Create user with initial texture
	user := model.User{
		ID:            uuid.New(),
		Email:         "test@example.com",
		Password:      "hashedpassword",
		EmailVerified: true,
		ProfileName:   "TestPlayer",
	}
	profile := model.NewProfile(user.ID, "TestPlayer", model.STEVE, "")
	user.SetProfile(&profile)
	err = db.Create(&user).Error
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Upload first texture
	err = textureService.(*textureServiceImpl).saveTexture(&user, testImage1, "SKIN", nil)
	if err != nil {
		t.Fatalf("Failed to save first texture: %v", err)
	}

	hash1 := model.ComputeTextureId(testImage1)
	var texture1 model.Texture
	err = db.First(&texture1, "hash = ?", hash1).Error
	if err != nil {
		t.Fatalf("Failed to find first texture: %v", err)
	}

	if texture1.Used != 1 {
		t.Errorf("Expected first texture.Used = 1, got %d", texture1.Used)
	}

	// Replace with second texture
	err = textureService.(*textureServiceImpl).saveTexture(&user, testImage2, "SKIN", nil)
	if err != nil {
		t.Fatalf("Failed to save second texture: %v", err)
	}

	// Verify first texture was deleted (used < 2)
	err = db.First(&texture1, "hash = ?", hash1).Error
	if err == nil {
		t.Error("Expected first texture to be deleted, but it still exists")
	}

	// Verify second texture has used = 1
	hash2 := model.ComputeTextureId(testImage2)
	var texture2 model.Texture
	err = db.First(&texture2, "hash = ?", hash2).Error
	if err != nil {
		t.Fatalf("Failed to find second texture: %v", err)
	}

	if texture2.Used != 1 {
		t.Errorf("Expected second texture.Used = 1, got %d", texture2.Used)
	}
}

// TestSaveTexture_ReuploadSameTexture tests that re-uploading same texture doesn't increment count
func TestSaveTexture_ReuploadSameTexture(t *testing.T) {
	// Setup in-memory database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// Migrate schema
	err = db.AutoMigrate(&model.User{}, &model.Texture{})
	if err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	// Create texture service
	tokenService := NewTokenService()
	textureService := NewTextureService(tokenService, db)

	// Create test image
	testImage := createTestImage(64, 64, 255, 255, 0) // Yellow

	// Create user
	user := model.User{
		ID:            uuid.New(),
		Email:         "test@example.com",
		Password:      "hashedpassword",
		EmailVerified: true,
		ProfileName:   "TestPlayer",
	}
	profile := model.NewProfile(user.ID, "TestPlayer", model.STEVE, "")
	user.SetProfile(&profile)
	err = db.Create(&user).Error
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Upload texture first time
	err = textureService.(*textureServiceImpl).saveTexture(&user, testImage, "SKIN", nil)
	if err != nil {
		t.Fatalf("Failed to save texture first time: %v", err)
	}

	hash := model.ComputeTextureId(testImage)
	var texture model.Texture
	err = db.First(&texture, "hash = ?", hash).Error
	if err != nil {
		t.Fatalf("Failed to find texture: %v", err)
	}

	initialUsed := texture.Used
	if initialUsed != 1 {
		t.Errorf("Expected texture.Used = 1 after first upload, got %d", initialUsed)
	}

	// Re-upload same texture
	err = textureService.(*textureServiceImpl).saveTexture(&user, testImage, "SKIN", nil)
	if err != nil {
		t.Fatalf("Failed to re-upload texture: %v", err)
	}

	// Verify reference count did NOT increment
	err = db.First(&texture, "hash = ?", hash).Error
	if err != nil {
		t.Fatalf("Failed to find texture after re-upload: %v", err)
	}

	if texture.Used != initialUsed {
		t.Errorf("Expected texture.Used to remain %d after re-upload, got %d", initialUsed, texture.Used)
	}
}
