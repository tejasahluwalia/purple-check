package models

import (
	"math"
	"purple-check/internal/config"
	"testing"
)

func TestScoringAlgorithms(t *testing.T) {
	// 1. Test when total is 0 (no reviews)
	t.Run("ZeroReviews", func(t *testing.T) {
		if s := WeightedWilsonScoring(0, 0, 0); s != 0 {
			t.Errorf("WeightedWilsonScoring(0,0,0) = %v, want 0", s)
		}
		if s := DirichletScoring(0, 0, 0); s != 0 {
			t.Errorf("DirichletScoring(0,0,0) = %v, want 0", s)
		}
		if s := BayesianScoring(0, 0, 0); s != 0 {
			t.Errorf("BayesianScoring(0,0,0) = %v, want 0", s)
		}
	})

	// 2. Test specific calculations for each algorithm
	t.Run("WeightedWilsonScoring", func(t *testing.T) {
		// 1 positive, 0 mixed, 0 negative -> total = 1, effPos = 1
		s := WeightedWilsonScoring(1, 0, 0)
		expected := 0.2065 // standard 95% Wilson lower bound for 1/1
		if math.Abs(s-expected) > 0.01 {
			t.Errorf("WeightedWilsonScoring(1,0,0) = %v, want approx %v", s, expected)
		}

		// 1 positive, 1 mixed, 0 negative -> total = 2, effPos = 1.5
		s = WeightedWilsonScoring(1, 1, 0)
		expected = 0.1979
		if math.Abs(s-expected) > 0.01 {
			t.Errorf("WeightedWilsonScoring(1,1,0) = %v, want approx %v", s, expected)
		}
	})

	t.Run("DirichletScoring", func(t *testing.T) {
		// 1 positive, 0 mixed, 0 negative
		// p1 = 2/4 = 0.5, p2 = 1/4 = 0.25
		// mean = 0.5 + 0.125 = 0.625
		// meanSq = 0.5 + 0.0625 = 0.5625
		// var = (0.5625 - 0.625^2) / 5 = 0.034375
		// stddev = sqrt(var) = 0.1854
		// score = 0.625 - 1.96 * 0.1854 = 0.2616
		s := DirichletScoring(1, 0, 0)
		expected := 0.2616
		if math.Abs(s-expected) > 0.01 {
			t.Errorf("DirichletScoring(1,0,0) = %v, want approx %v", s, expected)
		}
	})

	t.Run("BayesianScoring", func(t *testing.T) {
		// 1 positive, 0 mixed, 0 negative -> total = 1, C = 5.0, m = 0.5
		// score = (1 + 0 + 5 * 0.5) / (1 + 5) = 3.5 / 6 = 0.5833
		s := BayesianScoring(1, 0, 0)
		expected := 0.5833
		if math.Abs(s-expected) > 0.01 {
			t.Errorf("BayesianScoring(1,0,0) = %v, want %v", s, expected)
		}
	})
}

func TestCalculateScoreSwitching(t *testing.T) {
	// Store original config
	origAlg := config.SCORING_ALGORITHM
	defer func() {
		config.SCORING_ALGORITHM = origAlg
	}()

	pos, mixed, neg := 10, 2, 1

	// Test wilson
	config.SCORING_ALGORITHM = "wilson"
	sWilson := CalculateScore(pos, mixed, neg)
	if sWilson != WeightedWilsonScoring(pos, mixed, neg) {
		t.Errorf("CalculateScore did not use Wilson scoring")
	}

	// Test dirichlet
	config.SCORING_ALGORITHM = "dirichlet"
	sDirichlet := CalculateScore(pos, mixed, neg)
	if sDirichlet != DirichletScoring(pos, mixed, neg) {
		t.Errorf("CalculateScore did not use Dirichlet scoring")
	}

	// Test bayesian
	config.SCORING_ALGORITHM = "bayesian"
	sBayesian := CalculateScore(pos, mixed, neg)
	if sBayesian != BayesianScoring(pos, mixed, neg) {
		t.Errorf("CalculateScore did not use Bayesian scoring")
	}

	// Test invalid fallback to wilson
	config.SCORING_ALGORITHM = "invalid_choice"
	sFallback := CalculateScore(pos, mixed, neg)
	if sFallback != WeightedWilsonScoring(pos, mixed, neg) {
		t.Errorf("CalculateScore fallback did not default to Wilson scoring")
	}
}

func BenchmarkWeightedWilsonScoring(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = WeightedWilsonScoring(10, 2, 1)
	}
}

func BenchmarkDirichletScoring(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = DirichletScoring(10, 2, 1)
	}
}

func BenchmarkBayesianScoring(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = BayesianScoring(10, 2, 1)
	}
}
