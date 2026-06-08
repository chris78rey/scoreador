package sim

import "math"

const (
	airDensity            = 1.225
	ballArea              = 0.038
	ballMass              = 0.43
	ballDragCoefficient   = 0.24
	ballMagnusCoefficient = 0.08
	ballGravity           = 9.81
	ballGroundFriction    = 0.986
	ballAirDecay          = 0.999
	ballBounceDamping     = 0.54
	ballSpinDecay         = 0.996
	ballPhysicsSubstepDT  = 0.025
	ballMaxAcceleration   = 28.0
	ballMaxSpeed          = 32.0
)

func (e *Engine) ballAcceleration(velocity Vec2) Vec2 {
	speed := velocity.Len()
	if speed < 1e-9 {
		return Vec2{}
	}

	dragScale := 0.5 * airDensity * ballArea * ballDragCoefficient / ballMass
	magnusScale := 0.5 * airDensity * ballArea * ballMagnusCoefficient / ballMass

	drag := velocity.Scale(-dragScale * speed)
	perp := Vec2{X: -velocity.Y, Y: velocity.X}.Normalize()
	spin := clamp(e.ball.Spin/8.0, -1, 1)
	magnus := perp.Scale(magnusScale * speed * speed * spin)

	acceleration := drag.Add(magnus)
	if magnitude := acceleration.Len(); magnitude > ballMaxAcceleration {
		return acceleration.Normalize().Scale(ballMaxAcceleration)
	}
	return acceleration
}

func (e *Engine) integrateBall(dt float64) {
	substeps := int(math.Ceil(dt / ballPhysicsSubstepDT))
	if substeps < 1 {
		substeps = 1
	}
	subDt := dt / float64(substeps)

	for range substeps {
		position := e.ball.Position
		velocity := e.ball.Velocity

		acceleration := e.ballAcceleration(velocity)
		midVelocity := velocity.Add(acceleration.Scale(0.5 * subDt))
		midAcceleration := e.ballAcceleration(midVelocity)

		nextVelocity := velocity.Add(midAcceleration.Scale(subDt))
		nextPosition := position.Add(midVelocity.Scale(subDt))

		if e.ball.Height > 0 || e.ball.Vertical > 0 {
			e.ball.Vertical -= ballGravity * subDt
			e.ball.Height += e.ball.Vertical * subDt
			if e.ball.Height <= 0 {
				e.ball.Height = 0
				if math.Abs(e.ball.Vertical) > 1.4 {
					e.ball.Vertical = -e.ball.Vertical * 0.28
					e.ball.Height = 0.02
				} else {
					e.ball.Vertical = 0
				}
			}
		} else {
			e.ball.Height = 0
			e.ball.Vertical = 0
		}

		nextVelocity = nextVelocity.Scale(ballAirDecay)
		nextVelocity = nextVelocity.Scale(ballGroundFriction)
		e.ball.Spin *= ballSpinDecay

		if speed := nextVelocity.Len(); speed > ballMaxSpeed {
			nextVelocity = nextVelocity.Normalize().Scale(ballMaxSpeed)
		}

		if nextPosition.X < 0 {
			nextPosition.X = -nextPosition.X
			nextVelocity.X = -nextVelocity.X * ballBounceDamping
		} else if nextPosition.X > PitchLength {
			nextPosition.X = 2*PitchLength - nextPosition.X
			nextVelocity.X = -nextVelocity.X * ballBounceDamping
		}

		if nextPosition.Y < 0 {
			nextPosition.Y = -nextPosition.Y
			nextVelocity.Y = -nextVelocity.Y * ballBounceDamping
		} else if nextPosition.Y > PitchWidth {
			nextPosition.Y = 2*PitchWidth - nextPosition.Y
			nextVelocity.Y = -nextVelocity.Y * ballBounceDamping
		}

		if nextVelocity.Len() < 0.03 {
			nextVelocity = Vec2{}
		}

		e.ball.Position = nextPosition
		e.ball.Velocity = nextVelocity
	}
}
