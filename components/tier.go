package components

import "strings"

// TierColorClass returns a Tailwind text color class for a ranked tier.
func TierColorClass(tier string) string {
	switch strings.ToUpper(tier) {
	case "IRON":
		return "text-stone-500"
	case "BRONZE":
		return "text-amber-800"
	case "SILVER":
		return "text-slate-400"
	case "GOLD":
		return "text-yellow-600"
	case "PLATINUM":
		return "text-cyan-600"
	case "EMERALD":
		return "text-emerald-600"
	case "DIAMOND":
		return "text-blue-500"
	case "MASTER":
		return "text-purple-600"
	case "GRANDMASTER":
		return "text-red-600"
	case "CHALLENGER":
		return "text-amber-400"
	default:
		return "text-slate-600"
	}
}

// TierTitle returns a properly cased tier + rank string (e.g. "Gold II").
func TierTitle(tier, rank string) string {
	t := strings.ToUpper(tier)
	cased := strings.ToUpper(tier[:1]) + strings.ToLower(tier[1:])
	// Master+ tiers have no subdivision
	if t == "MASTER" || t == "GRANDMASTER" || t == "CHALLENGER" {
		return cased
	}
	return cased + " " + rank
}
