package models

import (
	"log"
	"math"
	"purple-check/internal/config"
)

// CalculateScore computes the score based on the configured SCORING_ALGORITHM.
func CalculateScore(positive, mixed, negative int) float64 {
	switch config.SCORING_ALGORITHM {
	case "dirichlet":
		return DirichletScoring(positive, mixed, negative)
	case "bayesian":
		return BayesianScoring(positive, mixed, negative)
	case "wilson":
		return WeightedWilsonScoring(positive, mixed, negative)
	default:
		// Fallback and log warning if the config is unknown
		log.Printf("Warning: unknown SCORING_ALGORITHM %q, falling back to 'wilson'", config.SCORING_ALGORITHM)
		return WeightedWilsonScoring(positive, mixed, negative)
	}
}

// DirichletScoring computes the lower bound of the Dirichlet credible interval
// using a Normal approximation with a uniform prior (alpha = 1 for Positive, Mixed, Negative).
func DirichletScoring(positive, mixed, negative int) float64 {
	n1 := float64(positive)
	n2 := float64(mixed)
	N := n1 + n2 + float64(negative)
	if N == 0 {
		return 0.0
	}

	// Dirichlet uniform prior (alpha = 1.0 for each of K = 3 categories)
	p1 := (n1 + 1.0) / (N + 3.0)
	p2 := (n2 + 1.0) / (N + 3.0)

	mean := p1 + 0.5*p2
	meanSq := p1 + 0.25*p2

	variance := (meanSq - (mean * mean)) / (N + 4.0)
	if variance < 0 {
		variance = 0
	}

	// 95% confidence interval (z = 1.96)
	const z = 1.96
	score := mean - z*math.Sqrt(variance)
	if score < 0 {
		return 0.0
	}
	return score
}

// BayesianScoring computes a weighted Bayesian average.
// It assigns weights (Positive = 1.0, Mixed = 0.5, Negative = 0.0) and applies Laplace smoothing
// with a virtual count prior of 5.0 and target prior mean of 0.85.
// with a virtual count prior of 5.0 and target prior mean of 0.5.
func BayesianScoring(positive, mixed, negative int) float64 {
	total := positive + mixed + negative
	if total == 0 {
		return 0.0
	}
	const priorWeight = 5.0 // C
	const priorMean = 0.5   // m

	numerator := float64(positive) + 0.5*float64(mixed) + (priorWeight * priorMean)
	denominator := float64(total) + priorWeight
	return numerator / denominator
}

// WeightedWilsonScoring computes the lower bound of the Wilson score interval
// on effective positive and total counts (Mixed count counts as 0.5 positive and 0.5 negative).
func WeightedWilsonScoring(positive, mixed, negative int) float64 {
	total := positive + mixed + negative
	if total == 0 {
		return 0.0
	}
	effPos := float64(positive) + 0.5*float64(mixed)
	n := float64(total)
	p := effPos / n
	const z = 1.96
	z2 := z * z

	numerator := p + z2/(2*n) - z*math.Sqrt((p*(1-p)+z2/(4*n))/n)
	denominator := 1 + z2/n
	score := numerator / denominator
	if score < 0 {
		return 0.0
	}
	return score
}
