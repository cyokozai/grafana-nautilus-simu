package main

import (
	"math"
	"testing"
)

// 座標更新のテスト
func TestUpdateBoid_Move(t *testing.T) {
	b := Boid{
		X: 0, Y: 0,
		Vx: 0.01, Vy: 0.005,
	}

	// Act
	UpdateBoid(&b)

	// Assert
	if math.Abs(b.X-0.01) > 1e-9 {
		t.Errorf("Expected X to be 0.01, got %f", b.X)
	}
	if math.Abs(b.Y-0.005) > 1e-9 {
		t.Errorf("Expected Y to be 0.005, got %f", b.Y)
	}
}

// 壁際での反転テスト
func TestUpdateBoid_WallBounce(t *testing.T) {
	boundary := 1.0 - Margin

	b := Boid{
		X:  boundary + 0.01,
		Y:  0,
		Vx: 0.005,
		Vy: 0,
	}
	initialVx := b.Vx

	// Act
	UpdateBoid(&b)

	// Assert
	expectedVx := initialVx - TurnFactor // 反転による速度変化

	if b.Vx >= initialVx {
		t.Errorf("Expected Vx to decrease (turn back), but got %f >= %f", b.Vx, initialVx)
	}
	
	if math.Abs(b.Vx-expectedVx) > 1e-9 {
		t.Errorf("Expected Vx %f, got %f", expectedVx, b.Vx)
	}
}

// 角度更新のテスト
func TestUpdateBoid_Rotation(t *testing.T) {
	b := Boid{
		X: 0, Y: 0,
		Vx: 0, Vy: 0.01,
	}

	// Act
	UpdateBoid(&b)

	// Assert
	expectedAngle := 90.0 // 上向きのベクトル
	
	if math.Abs(b.Angle-expectedAngle) > 1e-9 {
		t.Errorf("Expected Angle %f, got %f", expectedAngle, b.Angle)
	}
}