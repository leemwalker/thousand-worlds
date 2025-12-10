package interview

// AllTopics defines the ordered list of topics for the interview
// AllTopics defines the ordered list of topics for the interview
var AllTopics = []Topic{
	// Core Concept
	{Category: CategoryTheme, Name: "Core Concept", Description: "Describe the world's theme, tone, and key inspirations in one sentence."},

	// Sentient Life
	{Category: CategoryCulture, Name: "Sentient Species", Description: "Who are the sentient people of this world? (Humans, Elves, Aliens, Robots, etc.)"},

	// Geography & Environment
	{Category: CategoryGeography, Name: "Environment", Description: "Describe the geography, climate, and any unique features of the world."},

	// Magic & Technology
	{Category: CategoryTechLevel, Name: "Magic & Tech", Description: "What is the level of technology and magic in this world?"},

	// Conflict
	{Category: CategoryTheme, Name: "Conflict", Description: "What is the central conflict or tension driving events in this world?"},

	// Geological Age
	{Category: CategoryGeography, Name: "Geological Age", Description: "How old is this world geologically? (Young with sharp peaks, Mature, or Old with smoothed mountains)"},

	// Branching Point (Q7)
	{Category: CategoryTheme, Name: "Branch", Description: "Decision point: Name world or continue."},

	// Factions (Q8)
	{Category: CategoryCulture, Name: "Factions", Description: "Describe the major factions or groups."},

	// World Name (Final Question)
	{Category: CategoryTheme, Name: "World Name", Description: "What should this world be called? Generate 2 creative, unique name suggestions based on the world's theme, tone, and characteristics."},
}

// GetTopicsByCategory returns topics filtered by category
func GetTopicsByCategory(category Category) []Topic {
	var topics []Topic
	for _, t := range AllTopics {
		if t.Category == category {
			topics = append(topics, t)
		}
	}
	return topics
}
