package interview

// AllTopics defines the ordered list of topics for the interview
var AllTopics = []Topic{
	// Theme
	{Category: CategoryTheme, Name: "World Type", Description: "What kind of world is this? (fantasy, sci-fi, etc.)"},
	{Category: CategoryTheme, Name: "Tone", Description: "What is the overall tone? (grim, hopeful, etc.)"},
	{Category: CategoryTheme, Name: "Inspirations", Description: "Any specific inspirations? (books, movies, etc.)"},
	{Category: CategoryTheme, Name: "Uniqueness", Description: "What makes this world unique?"},
	{Category: CategoryTheme, Name: "Conflict", Description: "What conflicts or tensions exist?"},

	// Tech Level
	{Category: CategoryTechLevel, Name: "Tech Level", Description: "What is the technological advancement level?"},
	{Category: CategoryTechLevel, Name: "Magic", Description: "Is magic present? If so, how common?"},
	{Category: CategoryTechLevel, Name: "Advanced Tech", Description: "What is the most advanced technology available?"},
	{Category: CategoryTechLevel, Name: "Daily Life", Description: "How does magic/technology affect daily life?"},

	// Geography
	{Category: CategoryGeography, Name: "Planet Size", Description: "What is the planet size?"},
	{Category: CategoryGeography, Name: "Climate", Description: "What is the climate range?"},
	{Category: CategoryGeography, Name: "Features", Description: "Any unique geographical features?"},
	{Category: CategoryGeography, Name: "Land vs Water", Description: "How much land vs water?"},
	{Category: CategoryGeography, Name: "Extreme Environments", Description: "Any extreme environments?"},

	// Culture
	{Category: CategoryCulture, Name: "Species", Description: "What sentient species exist?"},
	{Category: CategoryCulture, Name: "Politics", Description: "What is the political structure?"},
	{Category: CategoryCulture, Name: "Values", Description: "What are the main cultural values?"},
	{Category: CategoryCulture, Name: "Economy", Description: "What is the economic system?"},
	{Category: CategoryCulture, Name: "Religion", Description: "What religions or belief systems exist?"},
	{Category: CategoryCulture, Name: "Taboos", Description: "What is considered taboo or forbidden?"},
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
